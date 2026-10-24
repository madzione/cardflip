package core

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"cardflip/golang/scheduler"
	"cardflip/golang/storage"
)

// Sync metadata (sync design decision 2) lives beside the objects: every
// syncable object is a last-writer-wins register with a stamp and the local
// sequence number of its last change, and deleted objects leave tombstones.
// The model structs do not change, and undo before-images carry no
// metadata: commitAs stamps whatever an op changes, in every mode.

// syncKind is the kind of object a register or tombstone belongs to.
type syncKind uint8

const (
	kindPreset syncKind = iota + 1
	kindDeck
	kindNoteType
	kindNote
	// kindCard is the kind of card tombstones. A card is two registers:
	// its scheduling (kindCardSched) and its placement and marks
	// (kindCardPlace), merged separately.
	kindCard
	kindCardSched
	kindCardPlace
	// kindReview registers exist only for review entries added again after
	// a deletion (redo of an answer), whose stamp must beat the tombstone.
	// Other entries are immutable and count as stamp 0.
	kindReview
	// kindSettings is the one register of the synced settings (ID 0).
	kindSettings
)

// metaKey identifies a register or tombstone.
type metaKey struct {
	kind syncKind
	id   ID
}

// tombKey returns the key of the tombstone that deletes the register k.
func tombKey(k metaKey) metaKey {
	if k.kind == kindCardSched || k.kind == kindCardPlace {
		k.kind = kindCard
	}
	return k
}

// objMeta is the metadata of a register or tombstone: the stamp and local
// sequence number of its last change. aux is the deck or note type a
// merged-away deck or note type redirects to (decision 5), or the month
// (yyyymm) of a deleted review entry.
type objMeta struct {
	stamp Stamp
	seq   int64
	aux   int64
}

// syncEntry is the journaled metadata of one register or tombstone, encoded
// as a compact array.
type syncEntry struct {
	Kind  syncKind
	ID    ID
	Stamp Stamp
	Seq   int64
	Aux   int64
}

func (e syncEntry) key() metaKey  { return metaKey{e.Kind, e.ID} }
func (e syncEntry) meta() objMeta { return objMeta{e.Stamp, e.Seq, e.Aux} }
func entryOf(k metaKey, m objMeta) syncEntry {
	return syncEntry{Kind: k.kind, ID: k.id, Stamp: m.stamp, Seq: m.seq, Aux: m.aux}
}

func (e syncEntry) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, 48)
	b = append(b, '[')
	b = strconv.AppendUint(b, uint64(e.Kind), 10)
	b = append(b, ',')
	b = strconv.AppendInt(b, int64(e.ID), 10)
	b = append(b, ',')
	b = strconv.AppendUint(b, uint64(e.Stamp), 10)
	b = append(b, ',')
	b = strconv.AppendInt(b, e.Seq, 10)
	if e.Aux != 0 {
		b = append(b, ',')
		b = strconv.AppendInt(b, e.Aux, 10)
	}
	return append(b, ']'), nil
}

func (e *syncEntry) UnmarshalJSON(data []byte) error {
	var a []json.Number
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	if len(a) < 4 || len(a) > 5 {
		return fmt.Errorf("sync metadata entry has %d values", len(a))
	}
	var v [5]int64
	var stamp uint64
	for i, n := range a {
		var err error
		if i == 2 {
			stamp, err = strconv.ParseUint(n.String(), 10, 64)
		} else {
			v[i], err = n.Int64()
		}
		if err != nil {
			return fmt.Errorf("sync metadata entry: %v", err)
		}
	}
	*e = syncEntry{Kind: syncKind(v[0]), ID: ID(v[1]), Stamp: Stamp(stamp), Seq: v[3], Aux: v[4]}
	return nil
}

// newEpoch returns a new collection lineage ID (decision 7).
func newEpoch() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// monthNumber turns "2026-10" into 202610, 0 if it is not a month.
func monthNumber(month string) int64 {
	y, m, ok := strings.Cut(month, "-")
	yn, err1 := strconv.Atoi(y)
	mn, err2 := strconv.Atoi(m)
	if !ok || err1 != nil || err2 != nil {
		return 0
	}
	return int64(yn*100 + mn)
}

// monthName turns 202610 into "2026-10".
func monthName(n int64) string { return fmt.Sprintf("%04d-%02d", n/100, n%100) }

// sameSchedule reports whether two cards have the same scheduling register:
// everything that answering, resetting or rescheduling a card changes.
func sameSchedule(a, b *Card) bool {
	return sameCardState(a.CardState, b.CardState) && a.NewPos == b.NewPos && equalPtr(a.OriginalPos, b.OriginalPos) &&
		a.BuriedUntil.Equal(b.BuriedUntil) && a.PreviewDue.Equal(b.PreviewDue)
}

func sameCardState(a, b scheduler.CardState) bool {
	return a.State == b.State && a.Due.Equal(b.Due) && equalPtr(a.Memory, b.Memory) && a.Step == b.Step &&
		a.ScheduledDays == b.ScheduledDays && a.Interday == b.Interday && a.Reps == b.Reps && a.Lapses == b.Lapses &&
		a.LastReview.Equal(b.LastReview)
}

// samePlacement reports whether two cards have the same placement and marks
// register: their note and card type, deck, flag and suspension.
func samePlacement(a, b *Card) bool {
	return a.NoteID == b.NoteID && a.Ord == b.Ord && a.DeckID == b.DeckID && a.HomeDeckID == b.HomeDeckID &&
		a.FilteredPos == b.FilteredPos && a.Flag == b.Flag && a.Suspended == b.Suspended
}

func equalPtr[T comparable](a, b *T) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// syncStamper collects the metadata of the registers and tombstones an op
// changes, with one fresh stamp and consecutive sequence numbers.
type syncStamper struct {
	c     *Collection
	o     *op
	stamp Stamp
	seq   int64
}

func (s *syncStamper) next() (Stamp, int64) {
	if s.stamp == 0 {
		s.stamp = s.c.clock.now()
	}
	s.seq++
	return s.stamp, s.seq
}

func (s *syncStamper) put(kind syncKind, id ID) {
	st, seq := s.next()
	s.o.Meta = append(s.o.Meta, syncEntry{Kind: kind, ID: id, Stamp: st, Seq: seq})
}

func (s *syncStamper) tomb(kind syncKind, id ID, aux int64) {
	for _, r := range s.o.Redirects {
		if r.kind == kind && r.from == id {
			aux = int64(r.to)
		}
	}
	st, seq := s.next()
	s.o.Tombs = append(s.o.Tombs, syncEntry{Kind: kind, ID: id, Stamp: st, Seq: seq, Aux: aux})
}

// stampObjects stamps the objects of one kind that an op puts and deletes:
// those it adds or changes get a register update, those that exist and it
// deletes a tombstone. Objects put unchanged are left alone, so they cannot
// win over a real change made elsewhere.
func stampObjects[T any](s *syncStamper, kind syncKind, cur map[ID]*T, put []*T, del []ID, idOf func(*T) ID) {
	deleted := make(map[ID]bool, len(del))
	for _, id := range del {
		deleted[id] = true
	}
	final := make(map[ID]*T, len(put))
	var order []ID
	for _, v := range put {
		id := idOf(v)
		if _, ok := final[id]; !ok {
			order = append(order, id)
		}
		final[id] = v
	}
	for _, id := range order {
		if deleted[id] {
			continue
		}
		if old, ok := cur[id]; !ok || !reflect.DeepEqual(old, final[id]) {
			s.put(kind, id)
		}
	}
	seen := map[ID]bool{}
	for _, id := range del {
		if seen[id] {
			continue
		}
		seen[id] = true
		if _, ok := cur[id]; ok {
			s.tomb(kind, id, 0)
		}
	}
}

// stampSync adds to o the metadata of everything it changes (decision 2),
// and returns rev with the sequence numbers of the review log entries it
// adds. Callers hold the write lock.
func (c *Collection) stampSync(o *op, rev []storage.RevlogWrite) []storage.RevlogWrite {
	if o.Kind == opSync {
		// ApplyBatch sets the metadata of incoming changes itself.
		return rev
	}
	s := &syncStamper{c: c, o: o, seq: c.seq}
	o.Meta, o.Tombs = nil, nil
	if o.Settings != nil && !syncedSettingsEqual(c.settings, *o.Settings) {
		s.put(kindSettings, 0)
	}
	stampObjects(s, kindPreset, c.presets, o.Presets, o.DelPresets, func(p *Preset) ID { return p.ID })
	stampObjects(s, kindDeck, c.decks, o.Decks, o.DelDecks, func(d *Deck) ID { return d.ID })
	stampObjects(s, kindNoteType, c.noteTypes, o.NoteTypes, o.DelNoteTypes, func(nt *NoteType) ID { return nt.ID })
	stampObjects(s, kindNote, c.notes, o.Notes, o.DelNotes, func(n *Note) ID { return n.ID })
	c.stampCards(s, o)
	s.stampMedia()
	if len(rev) > 0 {
		rev = append([]storage.RevlogWrite(nil), rev...)
		for i, w := range rev {
			if w.Entry == nil {
				s.tomb(kindReview, ID(w.ID), monthNumber(w.Month))
				continue
			}
			if _, ok := c.tombs[metaKey{kindReview, ID(w.ID)}]; ok {
				s.put(kindReview, ID(w.ID))
			}
			_, rev[i].Seq = s.next()
			if o.RevSeq == nil {
				o.RevSeq = map[string]int64{}
			}
			o.RevSeq[w.Month] = rev[i].Seq
		}
	}
	if o.Kind == opRenumber {
		// The old IDs are the peer's objects now.
		o.Tombs = nil
	}
	return rev
}

// stampCards stamps the scheduling and placement registers of the cards o
// changes, each only when it changed, and tombstones deleted cards. Cards
// deleted with their note get no tombstone: the note's deletes them (see
// sync_apply.go).
func (c *Collection) stampCards(s *syncStamper, o *op) {
	delNotes := make(map[ID]bool, len(o.DelNotes))
	for _, id := range o.DelNotes {
		if _, ok := c.notes[id]; ok {
			delNotes[id] = true
		}
	}
	deleted := make(map[ID]bool, len(o.DelCards))
	for _, id := range o.DelCards {
		deleted[id] = true
	}
	final := make(map[ID]*Card, len(o.Cards))
	var order []ID
	for _, card := range o.Cards {
		if _, ok := final[card.ID]; !ok {
			order = append(order, card.ID)
		}
		final[card.ID] = card
	}
	for _, id := range order {
		if deleted[id] {
			continue
		}
		card, old := final[id], c.cards[id]
		if old == nil || !sameSchedule(old, card) {
			s.put(kindCardSched, id)
		}
		if old == nil || !samePlacement(old, card) {
			s.put(kindCardPlace, id)
		}
	}
	seen := map[ID]bool{}
	for _, id := range o.DelCards {
		if seen[id] {
			continue
		}
		seen[id] = true
		if card, ok := c.cards[id]; ok && !delNotes[card.NoteID] {
			s.tomb(kindCard, id, 0)
		}
	}
}

// applyMeta applies the metadata of o. Like apply, it must not fail.
func (c *Collection) applyMeta(o *op) {
	for _, e := range o.Meta {
		c.clock.observe(e.Stamp)
		c.seq = max(c.seq, e.Seq)
		c.meta[e.key()] = e.meta()
		delete(c.tombs, tombKey(e.key()))
	}
	for _, e := range o.Tombs {
		c.clock.observe(e.Stamp)
		c.seq = max(c.seq, e.Seq)
		k := e.key()
		c.tombs[k] = e.meta()
		if k.kind == kindCard {
			delete(c.meta, metaKey{kindCardSched, k.id})
			delete(c.meta, metaKey{kindCardPlace, k.id})
		} else {
			delete(c.meta, k)
		}
	}
	for _, e := range o.Pruned {
		delete(c.tombs, e.key())
	}
	// Deleted objects have no registers, tombstone or not (cards deleted
	// with their note have none).
	for kind, ids := range map[syncKind][]ID{kindPreset: o.DelPresets, kindDeck: o.DelDecks,
		kindNoteType: o.DelNoteTypes, kindNote: o.DelNotes, kindCardSched: o.DelCards, kindCardPlace: o.DelCards} {
		for _, id := range ids {
			delete(c.meta, metaKey{kind, id})
		}
	}
	for m, seq := range o.RevSeq {
		c.revSeq[m] = max(c.revSeq[m], seq)
		c.seq = max(c.seq, seq)
	}
	if o.Epoch != "" {
		c.epoch = o.Epoch
	}
	if o.Slots != nil {
		c.slots = o.Slots
		c.useSlot()
	}
}

// ownSlot returns this device's slot in the slot table, if it has one.
// Callers hold a lock.
func (c *Collection) ownSlot() (int, bool) {
	slot, ok := c.slots[c.device]
	return slot, ok && c.device != ""
}

// useSlot makes the clock and the IDs use this device's slot (decision 4).
// Without one, IDs are allocated as before and stamps have slot 0.
func (c *Collection) useSlot() {
	slot, ok := c.ownSlot()
	c.clock.setSlot(slot)
	c.ids.setSlot(slot, ok)
}

// SetDevice tells the collection the ID of the device it is on, which
// names its slot in the slot table. The app calls it once the device has a
// sync identity.
func (c *Collection) SetDevice(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.device = id
	c.useSlot()
}

// Slot returns this device's slot, if the collection has one.
func (c *Collection) Slot() (int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ownSlot()
}

// SetSlots saves a new slot table (sync design decision 4). It is not an
// undo step.
func (c *Collection) SetSlots(slots map[string]int) error {
	for dev, slot := range slots {
		if dev == "" || slot < 0 || slot > MaxSlot {
			return fmt.Errorf("invalid slot %d for device %q", slot, dev)
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if maps.Equal(slots, c.slots) {
		return nil
	}
	return c.commit(&op{Kind: opSlots, Slots: maps.Clone(slots)})
}

// Slots returns a copy of the slot table: device ID to slot.
func (c *Collection) Slots() map[string]int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return maps.Clone(c.slots)
}

// ensureMeta gives every object without metadata a stamp from its
// modification time or, failing that, its ID's time, with sequence number 0
// (migration from snapshot version 3, and new collections), and drops the
// metadata of objects that no longer exist. It is deterministic, so it
// needs no saving. Callers hold the write lock (or are in Open).
func (c *Collection) ensureMeta() {
	idTime := func(id ID) Stamp { return makeStamp(int64(id), 0, 0) }
	need := func(k metaKey, st Stamp) {
		if _, ok := c.meta[k]; !ok {
			c.meta[k] = objMeta{stamp: st}
			c.clock.observe(st)
		}
	}
	for id := range c.presets {
		need(metaKey{kindPreset, id}, idTime(id))
	}
	for id := range c.decks {
		need(metaKey{kindDeck, id}, idTime(id))
	}
	for id, nt := range c.noteTypes {
		st := stampAt(nt.Modified)
		if st == 0 {
			st = idTime(id)
		}
		need(metaKey{kindNoteType, id}, st)
	}
	for id, n := range c.notes {
		st := stampAt(n.Modified)
		if st == 0 {
			st = idTime(id)
		}
		need(metaKey{kindNote, id}, st)
	}
	for id, card := range c.cards {
		st := stampAt(card.Modified)
		if st == 0 {
			st = idTime(id)
		}
		need(metaKey{kindCardSched, id}, st)
		need(metaKey{kindCardPlace, id}, st)
	}
	need(metaKey{kindSettings, 0}, 0)
	for k := range c.meta {
		var ok bool
		switch k.kind {
		case kindPreset:
			_, ok = c.presets[k.id]
		case kindDeck:
			_, ok = c.decks[k.id]
		case kindNoteType:
			_, ok = c.noteTypes[k.id]
		case kindNote:
			_, ok = c.notes[k.id]
		case kindCardSched, kindCardPlace:
			_, ok = c.cards[k.id]
		default:
			ok = true
		}
		if !ok {
			delete(c.meta, k)
		}
	}
	if c.epoch == "" {
		c.epoch = newEpoch()
	}
}

// metaTable is the snapshot encoding of registers or tombstones: columns
// per kind, sorted by ID, with IDs as differences to the previous one and
// stamps as milliseconds after the ID's time. An object that has not
// changed since the migration (stamp from its ID's time, sequence number 0)
// costs a few bytes.
type metaTable map[metaKey]objMeta

type metaColumns struct {
	Kind syncKind `json:"k"`
	// IDs are differences to the previous ID.
	IDs []int64 `json:"i"`
	// Times are the stamps' milliseconds minus the ID, Low their counter
	// and slot bits; omitted when all are 0.
	Times []int64  `json:"t,omitempty"`
	Low   []uint64 `json:"l,omitempty"`
	Seqs  []int64  `json:"s,omitempty"`
	Aux   []int64  `json:"a,omitempty"`
}

const stampLowBits = stampCounterBits + stampSlotBits

func (t metaTable) MarshalJSON() ([]byte, error) {
	byKind := map[syncKind][]ID{}
	for k := range t {
		byKind[k.kind] = append(byKind[k.kind], k.id)
	}
	var out []metaColumns
	for kind := kindPreset; kind <= kindSettings; kind++ {
		ids := byKind[kind]
		if len(ids) == 0 {
			continue
		}
		slices.Sort(ids)
		col := metaColumns{Kind: kind, IDs: make([]int64, len(ids))}
		times, low, seqs, aux := make([]int64, len(ids)), make([]uint64, len(ids)), make([]int64, len(ids)), make([]int64, len(ids))
		var anyTime, anyLow, anySeq, anyAux bool
		prev := int64(0)
		for i, id := range ids {
			m := t[metaKey{kind, id}]
			col.IDs[i] = int64(id) - prev
			prev = int64(id)
			times[i] = m.stamp.Millis() - int64(id)
			low[i] = uint64(m.stamp) & (1<<stampLowBits - 1)
			seqs[i], aux[i] = m.seq, m.aux
			anyTime = anyTime || times[i] != 0
			anyLow = anyLow || low[i] != 0
			anySeq = anySeq || seqs[i] != 0
			anyAux = anyAux || aux[i] != 0
		}
		if anyTime {
			col.Times = times
		}
		if anyLow {
			col.Low = low
		}
		if anySeq {
			col.Seqs = seqs
		}
		if anyAux {
			col.Aux = aux
		}
		out = append(out, col)
	}
	return json.Marshal(out)
}

func (t *metaTable) UnmarshalJSON(data []byte) error {
	var cols []metaColumns
	if err := json.Unmarshal(data, &cols); err != nil {
		return err
	}
	n := 0
	for _, col := range cols {
		n += len(col.IDs)
	}
	m := make(metaTable, n)
	for _, col := range cols {
		n := len(col.IDs)
		if col.Times != nil && len(col.Times) != n || col.Low != nil && len(col.Low) != n ||
			col.Seqs != nil && len(col.Seqs) != n || col.Aux != nil && len(col.Aux) != n {
			return fmt.Errorf("sync metadata of kind %d has columns of different lengths", col.Kind)
		}
		id := int64(0)
		for i, d := range col.IDs {
			id += d
			ms := id
			var om objMeta
			if col.Times != nil {
				ms += col.Times[i]
			}
			var lo uint64
			if col.Low != nil {
				lo = col.Low[i]
			}
			if ms < 0 {
				return fmt.Errorf("sync metadata of kind %d has a negative time", col.Kind)
			}
			om.stamp = Stamp(uint64(ms)<<stampLowBits | lo&(1<<stampLowBits-1))
			if col.Seqs != nil {
				om.seq = col.Seqs[i]
			}
			if col.Aux != nil {
				om.aux = col.Aux[i]
			}
			m[metaKey{col.Kind, ID(id)}] = om
		}
	}
	*t = m
	return nil
}

// saveMeta puts the sync metadata into a snapshot. Callers hold a lock.
func (c *Collection) saveMeta(s *state) {
	s.Epoch, s.Seq, s.Meta, s.Tombs, s.RevSeq, s.Slots = c.epoch, c.seq, c.meta, c.tombs, c.revSeq, c.slots
	s.Media, s.MediaBuilt = c.media, c.media != nil
}

// loadMeta takes the sync metadata of a snapshot, which apply has loaded.
// Snapshots before version 4 have none; ensureMeta fills it in.
func (c *Collection) loadMeta(s *state) {
	c.epoch, c.seq, c.slots = s.Epoch, s.Seq, s.Slots
	c.meta, c.tombs = map[metaKey]objMeta(s.Meta), map[metaKey]objMeta(s.Tombs)
	if c.meta == nil {
		c.meta = map[metaKey]objMeta{}
	}
	if c.tombs == nil {
		c.tombs = map[metaKey]objMeta{}
	}
	c.revSeq = s.RevSeq
	if c.revSeq == nil {
		c.revSeq = map[string]int64{}
	}
	c.media = nil
	if s.MediaBuilt {
		c.media = s.Media
		if c.media == nil {
			c.media = map[string]mediaEntry{}
		}
		for _, e := range c.media {
			c.clock.observe(e.Stamp)
			c.seq = max(c.seq, e.Seq)
		}
	}
	for _, t := range []map[metaKey]objMeta{c.meta, c.tombs} {
		for _, m := range t {
			c.clock.observe(m.stamp)
			c.seq = max(c.seq, m.seq)
		}
	}
	for _, seq := range c.revSeq {
		c.seq = max(c.seq, seq)
	}
	c.useSlot()
}
