package core

import (
	"cmp"
	"encoding/json"
	"slices"
	"sort"

	"cardflip/golang/storage"
)

type storageRecord = storage.RevlogRecord

// Deltas (sync design decision 3): a device sends its peer everything that
// changed since the local sequence number the peer has received, ordered by
// kind and then by sequence number, in pages. The receiver applies each
// page as one op (ApplyBatch) and keeps the cursor of the last item, so an
// interrupted session continues where it stopped.

// DeltaKind is the kind of a delta item, in the order deltas send them:
// parents before what refers to them, deletions after everything.
type DeltaKind int

const (
	DeltaPresets DeltaKind = iota
	DeltaDecks
	DeltaNoteTypes
	DeltaNotes
	DeltaCards
	DeltaReviews
	DeltaSettings
	DeltaTombstones
	DeltaMedia
)

// DeltaKinds lists every kind, in delta order.
var DeltaKinds = []DeltaKind{DeltaPresets, DeltaDecks, DeltaNoteTypes, DeltaNotes, DeltaCards, DeltaReviews,
	DeltaSettings, DeltaTombstones, DeltaMedia}

// DeltaItem is one register, review log entry, tombstone or media manifest
// entry of a delta. Stamp is the register's stamp; a card has two, Stamp
// for its scheduling and Stamp2 for its placement and marks. Seq is the
// sender's sequence number of the item's last change. Data is the object,
// review log entry or settings as JSON; Of is the kind of object a
// tombstone deletes (and Aux the deck or note type it redirects to, or a
// review entry's month); Name, SHA1 and Trashed describe a media file.
type DeltaItem struct {
	Kind    DeltaKind       `json:"k"`
	ID      ID              `json:"id,omitempty"`
	Stamp   Stamp           `json:"t,omitempty"`
	Stamp2  Stamp           `json:"t2,omitempty"`
	Seq     int64           `json:"s"`
	Data    json.RawMessage `json:"d,omitempty"`
	Of      int             `json:"of,omitempty"`
	Aux     int64           `json:"a,omitempty"`
	Month   int64           `json:"m,omitempty"`
	Name    string          `json:"n,omitempty"`
	SHA1    string          `json:"h,omitempty"`
	Trashed bool            `json:"x,omitempty"`
}

// DeltaCursor is a position in a delta: the kind, and within it the
// sequence number and ID (review log entries are ordered by month first,
// media files by name instead of ID). The zero cursor is the start.
type DeltaCursor struct {
	Kind  DeltaKind `json:"k"`
	Month int64     `json:"m,omitempty"`
	Seq   int64     `json:"s"`
	ID    ID        `json:"id,omitempty"`
	Name  string    `json:"n,omitempty"`
}

func (a DeltaCursor) compare(b DeltaCursor) int {
	return cmp.Or(cmp.Compare(a.Kind, b.Kind), cmp.Compare(a.Month, b.Month), cmp.Compare(a.Seq, b.Seq),
		cmp.Compare(a.ID, b.ID), cmp.Compare(a.Name, b.Name))
}

// Cursor returns the item's position.
func (it *DeltaItem) Cursor() DeltaCursor {
	c := DeltaCursor{Kind: it.Kind, Seq: it.Seq, ID: it.ID, Name: it.Name}
	if it.Kind == DeltaReviews {
		c.Month = it.Month
	}
	return c
}

// SeqRange is an inclusive range of local sequence numbers.
type SeqRange struct {
	From int64 `json:"f"`
	To   int64 `json:"t"`
}

// SeqRanges is a sorted list of disjoint sequence number ranges.
type SeqRanges []SeqRange

// Contains reports whether seq is in one of the ranges.
func (r SeqRanges) Contains(seq int64) bool {
	i := sort.Search(len(r), func(i int) bool { return r[i].To >= seq })
	return i < len(r) && r[i].From <= seq
}

// Add returns the ranges with from..to added, merged with touching ones.
func (r SeqRanges) Add(from, to int64) SeqRanges {
	if from > to {
		return r
	}
	out := make(SeqRanges, 0, len(r)+1)
	for _, x := range r {
		switch {
		case x.To+1 < from || to+1 < x.From:
			out = append(out, x)
		default:
			from, to = min(from, x.From), max(to, x.To)
		}
	}
	out = append(out, SeqRange{from, to})
	slices.SortFunc(out, func(a, b SeqRange) int { return cmp.Compare(a.From, b.From) })
	return out
}

// Prune returns the ranges without the part at or below seq.
func (r SeqRanges) Prune(seq int64) SeqRanges {
	var out SeqRanges
	for _, x := range r {
		if x.To > seq {
			out = append(out, SeqRange{max(x.From, seq+1), x.To})
		}
	}
	return out
}

// DeltaRequest asks for the changes after the local sequence number Since
// (-1 for everything), of Kinds (nil for all), after the cursor After,
// leaving out changes whose sequence number is in Skip (those received
// from the peer the delta is for). A page holds at most MaxItems items and
// about MaxBytes bytes of data.
type DeltaRequest struct {
	Since    int64
	Kinds    []DeltaKind
	After    DeltaCursor
	Skip     SeqRanges
	MaxItems int
	MaxBytes int
}

// DeltaPage is one page of a delta. Done is set when nothing follows it;
// Seq is the collection's sequence number when the page was made.
type DeltaPage struct {
	Items []DeltaItem `json:"items"`
	Done  bool        `json:"done"`
	Seq   int64       `json:"seq"`
}

// Default page limits (design decision 3).
const (
	DeltaMaxItems = 500
	DeltaMaxBytes = 4 << 20
)

// deltaKey is a candidate item: its position and what it is.
type deltaKey struct {
	pos  DeltaCursor
	key  metaKey
	name string
}

// Delta returns a page of the changes req asks for.
func (c *Collection) Delta(req DeltaRequest) (DeltaPage, error) {
	if req.MaxItems <= 0 {
		req.MaxItems = DeltaMaxItems
	}
	if req.MaxBytes <= 0 {
		req.MaxBytes = DeltaMaxBytes
	}
	kinds := req.Kinds
	if kinds == nil {
		kinds = DeltaKinds
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	page := DeltaPage{Seq: c.seq, Items: []DeltaItem{}}
	fresh := func(seq int64) bool { return seq > req.Since && !req.Skip.Contains(seq) }
	bytes := 0
	full := func() bool { return len(page.Items) >= req.MaxItems || bytes >= req.MaxBytes }
	for _, kind := range DeltaKinds {
		if !slices.Contains(kinds, kind) || kind < req.After.Kind {
			continue
		}
		if full() {
			return page, nil
		}
		if kind == DeltaReviews {
			done, err := c.reviewDelta(req, &page, &bytes, fresh, full)
			if err != nil {
				return DeltaPage{}, err
			}
			if !done {
				return page, nil
			}
			continue
		}
		cands := c.deltaCandidates(kind, fresh)
		slices.SortFunc(cands, func(a, b deltaKey) int { return a.pos.compare(b.pos) })
		for _, k := range cands {
			if k.pos.compare(req.After) <= 0 {
				continue
			}
			if full() {
				return page, nil
			}
			it, err := c.deltaItem(kind, k)
			if err != nil {
				return DeltaPage{}, err
			}
			page.Items = append(page.Items, it)
			bytes += len(it.Data) + 64
		}
	}
	page.Done = true
	return page, nil
}

// deltaCandidates lists the positions of the items of kind that changed.
// Callers hold a lock.
func (c *Collection) deltaCandidates(kind DeltaKind, fresh func(int64) bool) []deltaKey {
	var out []deltaKey
	add := func(seq int64, k metaKey) {
		out = append(out, deltaKey{pos: DeltaCursor{Kind: kind, Seq: seq, ID: k.id}, key: k})
	}
	objKind := map[DeltaKind]syncKind{DeltaPresets: kindPreset, DeltaDecks: kindDeck, DeltaNoteTypes: kindNoteType,
		DeltaNotes: kindNote, DeltaSettings: kindSettings}
	switch kind {
	case DeltaCards:
		// A note goes with its cards: a peer that deleted the note gets
		// them back with it when the note changed after the deletion.
		for id, card := range c.cards {
			s, p := c.meta[metaKey{kindCardSched, id}], c.meta[metaKey{kindCardPlace, id}]
			n := c.meta[metaKey{kindNote, card.NoteID}]
			pos := max(s.seq, p.seq)
			if fresh(n.seq) {
				pos = max(pos, n.seq)
			}
			if fresh(s.seq) || fresh(p.seq) || fresh(n.seq) {
				add(pos, metaKey{kindCardSched, id})
			}
		}
	case DeltaTombstones:
		for k, m := range c.tombs {
			if fresh(m.seq) {
				out = append(out, deltaKey{pos: DeltaCursor{Kind: kind, Seq: m.seq, ID: k.id}, key: k})
			}
		}
	case DeltaMedia:
		for name, e := range c.media {
			if fresh(e.Seq) {
				out = append(out, deltaKey{pos: DeltaCursor{Kind: kind, Seq: e.Seq, Name: name}, name: name})
			}
		}
	default:
		sk := objKind[kind]
		for k, m := range c.meta {
			if k.kind == sk && fresh(m.seq) {
				add(m.seq, k)
			}
		}
	}
	return out
}

// deltaItem builds the item at k. Callers hold a lock.
func (c *Collection) deltaItem(kind DeltaKind, k deltaKey) (DeltaItem, error) {
	it := DeltaItem{Kind: kind, ID: k.key.id, Seq: k.pos.Seq, Stamp: c.meta[k.key].stamp}
	var obj any
	switch kind {
	case DeltaPresets:
		obj = c.presets[k.key.id]
	case DeltaDecks:
		obj = c.decks[k.key.id]
	case DeltaNoteTypes:
		obj = c.noteTypes[k.key.id]
	case DeltaNotes:
		obj = c.notes[k.key.id]
	case DeltaCards:
		obj = c.cards[k.key.id]
		it.Stamp2 = c.meta[metaKey{kindCardPlace, k.key.id}].stamp
	case DeltaSettings:
		obj = c.settings
	case DeltaTombstones:
		m := c.tombs[k.key]
		it.Stamp, it.Of, it.Aux = m.stamp, int(k.key.kind), m.aux
		return it, nil
	case DeltaMedia:
		e := c.media[k.name]
		it.ID, it.Stamp, it.Name, it.SHA1, it.Trashed = 0, e.Stamp, k.name, e.SHA1, e.Trashed
		return it, nil
	}
	raw, err := json.Marshal(obj)
	if err != nil {
		return DeltaItem{}, err
	}
	it.Data = raw
	return it, nil
}

// reviewDelta adds the review log entries of the delta to the page, month
// by month: within a month's segment, entries are stored in the order of
// their sequence numbers, so a page reads only the months it covers. It
// reports whether it got to the end. Callers hold a lock.
func (c *Collection) reviewDelta(req DeltaRequest, page *DeltaPage, bytes *int, fresh func(int64) bool, full func() bool) (bool, error) {
	var months []string
	if req.Since < 0 {
		all, err := c.store.RevlogMonths()
		if err != nil {
			return false, err
		}
		months = all
	} else {
		for m, seq := range c.revSeq {
			if seq > req.Since {
				months = append(months, m)
			}
		}
		slices.Sort(months)
	}
	for _, month := range months {
		mn := monthNumber(month)
		if req.After.Kind == DeltaReviews && mn < req.After.Month {
			continue
		}
		recs, err := c.store.ReadRevlogRecords(month)
		if err != nil {
			return false, err
		}
		slices.SortStableFunc(recs, func(a, b storageRecord) int {
			return cmp.Or(cmp.Compare(a.Seq, b.Seq), cmp.Compare(a.ID, b.ID))
		})
		for _, r := range recs {
			pos := DeltaCursor{Kind: DeltaReviews, Month: mn, Seq: r.Seq, ID: ID(r.ID)}
			if !fresh(r.Seq) || pos.compare(req.After) <= 0 {
				continue
			}
			if full() {
				return false, nil
			}
			it := DeltaItem{Kind: DeltaReviews, ID: ID(r.ID), Seq: r.Seq, Month: mn, Data: r.Entry,
				Stamp: c.meta[metaKey{kindReview, ID(r.ID)}].stamp}
			page.Items = append(page.Items, it)
			*bytes += len(r.Entry) + 64
		}
	}
	return true, nil
}
