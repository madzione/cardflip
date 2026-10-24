package core

import (
	"cardflip/golang/internal/usermsg"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"sort"
	"time"

	"cardflip/golang/scheduler"
)

// Problem kinds CheckDatabase reports, in the order of the report.
const (
	ProblemNotesWithoutType = "notesWithoutType"
	ProblemCardsWithoutNote = "cardsWithoutNote"
	ProblemInvalidCards     = "invalidCards"
	ProblemFieldCount       = "fieldCount"
	ProblemMissingCards     = "missingCards"
	ProblemMissingDeck      = "missingDeck"
	ProblemScheduling       = "scheduling"
	ProblemTags             = "tags"
)

var problemOrder = []string{ProblemNotesWithoutType, ProblemCardsWithoutNote, ProblemInvalidCards,
	ProblemFieldCount, ProblemMissingCards, ProblemMissingDeck, ProblemScheduling, ProblemTags}

// CheckProblem is one kind of problem CheckDatabase found and repaired.
type CheckProblem struct {
	Kind  string `json:"kind"`
	Count int    `json:"count"`
	// Message says what was found and how it was repaired, for the user.
	Message string `json:"message"`
}

// CheckReport is the result of CheckDatabase.
type CheckReport struct {
	// Problems lists the kinds of problems found; empty means none was.
	Problems []CheckProblem `json:"problems"`
	// OrphanedReviews counts the review log entries of cards that no longer
	// exist. They are not a problem: they are kept for statistics.
	OrphanedReviews int `json:"orphanedReviews"`
}

// CheckDatabase checks the collection for inconsistencies and repairs them
// in one change (persistence spec, "Check and repair the collection"):
//   - notes whose note type no longer exists are deleted with their cards,
//     and cards whose note no longer exists are deleted;
//   - cards for card types that no longer exist (or impossible cloze
//     numbers) are deleted; cards whose cloze number is just gone from the
//     note are empty cards, left to Empty Cards;
//   - notes get as many fields as their note type, keeping trimmed content
//     in the deletion log, and normalised tags;
//   - cards a note should have are generated;
//   - cards in a missing deck are moved to Default, created if needed;
//   - invalid scheduling values are repaired, memory states recomputed from
//     the review history where it has one.
//
// Review log entries of missing cards are kept and counted. The storage is
// compacted afterwards. If the repair cannot be saved nothing changes.
func (c *Collection) CheckDatabase(progress Progress) (CheckReport, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	k := &dbCheck{c: c, o: &op{Kind: "checkDatabase"}, now: c.now(), counts: map[string]int{},
		cards: map[ID]*Card{}, notes: map[ID]*Note{}, del: map[ID]bool{}, delNotes: map[ID]bool{}}
	k.checkNotes(progress)
	k.checkCards(progress)
	k.addMissingCards()
	orphans, err := k.checkHistory(progress)
	if err != nil {
		return CheckReport{}, err
	}
	progress("saving", 0, 1)
	sortIDs(k.o.DelNotes)
	sortIDs(k.o.DelCards)
	if err := c.commit(k.o); err != nil {
		return CheckReport{}, err
	}
	// The repair is already saved; a failed compaction is retried after the
	// next change, as in commit.
	_ = c.compact()
	progress("saving", 1, 1)
	return k.report(orphans), nil
}

// dbCheck is one run of CheckDatabase: the op of repairs it builds and what
// it found.
type dbCheck struct {
	c   *Collection
	o   *op
	now time.Time
	// cards and notes are the repaired copies in o, by ID; del and delNotes
	// the cards and notes o deletes.
	cards    map[ID]*Card
	notes    map[ID]*Note
	del      map[ID]bool
	delNotes map[ID]bool
	// needMemory are the repaired cards that need a memory state.
	needMemory []*Card
	// defaultDeck is where cards in missing decks go; 0 until needed.
	defaultDeck     ID
	defaultDeckName string

	counts map[string]int
	// typeCards counts the cards of the notes deleted for a missing note
	// type, trimmed whether trimmed fields had content, fromHistory the
	// memory states computed from review history.
	typeCards, fromHistory int
	trimmed                bool
}

func (k *dbCheck) deleteCard(id ID) {
	if !k.del[id] {
		k.del[id] = true
		k.o.DelCards = append(k.o.DelCards, id)
	}
}

// edit returns the repaired copy of card, adding it to the op.
func (k *dbCheck) edit(card *Card) *Card {
	if cp, ok := k.cards[card.ID]; ok {
		return cp
	}
	cp := *card
	k.cards[card.ID] = &cp
	k.o.Cards = append(k.o.Cards, &cp)
	return &cp
}

func sortedIDs[T any](m map[ID]T) []ID {
	ids := make([]ID, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	sortIDs(ids)
	return ids
}

func (k *dbCheck) checkNotes(progress Progress) {
	c := k.c
	ids := sortedIDs(c.notes)
	for i, nid := range ids {
		if i%500 == 0 {
			progress("notes", i, len(ids))
			c.yield()
		}
		n := c.notes[nid]
		nt, ok := c.noteTypes[n.TypeID]
		if !ok {
			k.o.DelNotes = append(k.o.DelNotes, nid)
			k.delNotes[nid] = true
			for _, cid := range c.cardsByNote[nid] {
				k.deleteCard(cid)
			}
			k.counts[ProblemNotesWithoutType]++
			k.typeCards += len(c.cardsByNote[nid])
			continue
		}
		fields := n.Fields
		if len(fields) != len(nt.Fields) {
			k.counts[ProblemFieldCount]++
			if len(fields) > len(nt.Fields) && slices.ContainsFunc(fields[len(nt.Fields):], func(f string) bool { return f != "" }) {
				k.o.Trimmed = append(k.o.Trimmed, n)
				k.trimmed = true
			}
			fields = padFields(slices.Clone(fields[:min(len(fields), len(nt.Fields))]), len(nt.Fields))
		}
		tags := n.Tags
		if norm := NormalizeTags(n.Tags); !slices.Equal(norm, n.Tags) {
			k.counts[ProblemTags]++
			tags = norm
		}
		if len(fields) == len(n.Fields) && slices.Equal(tags, n.Tags) {
			continue
		}
		nn := cloneNote(n)
		nn.Fields, nn.Tags, nn.Modified = fields, tags, k.now
		k.notes[nid] = &nn
		k.o.Notes = append(k.o.Notes, &nn)
	}
	progress("notes", len(ids), len(ids))
}

// validOrd reports whether a card ord can exist in note type nt: a card
// type of a standard note type, or a cloze number of a cloze note type.
func validOrd(nt *NoteType, ord int) bool {
	if nt.Kind == Cloze {
		return ord >= 0 && len(nt.Templates) > 0
	}
	return ord >= 0 && ord < len(nt.Templates)
}

func (k *dbCheck) checkCards(progress Progress) {
	c := k.c
	ids := sortedIDs(c.cards)
	for i, cid := range ids {
		if i%500 == 0 {
			progress("cards", i, len(ids))
			c.yield()
		}
		if k.del[cid] {
			continue
		}
		card := c.cards[cid]
		n, ok := c.notes[card.NoteID]
		if !ok {
			k.deleteCard(cid)
			k.counts[ProblemCardsWithoutNote]++
			continue
		}
		// Notes without a note type were deleted with their cards above.
		if !validOrd(c.noteTypes[n.TypeID], card.Ord) {
			k.deleteCard(cid)
			k.counts[ProblemInvalidCards]++
			continue
		}
		if k.repairDeck(card) {
			k.counts[ProblemMissingDeck]++
		}
		if k.repairScheduling(card) {
			k.counts[ProblemScheduling]++
		}
	}
	progress("cards", len(ids), len(ids))
}

// defaultDeckID returns the default deck: the one with the built-in name,
// or the English one, adding it to the op when the collection has neither.
func (k *dbCheck) defaultDeckID() ID {
	if k.defaultDeck != 0 {
		return k.defaultDeck
	}
	name := k.c.builtin.deckName()
	d := k.c.deckByName(name)
	if d == nil {
		d = k.c.deckByName(DefaultDeckName)
	}
	if d == nil {
		d = &Deck{ID: k.c.ids.next(), Name: name, PresetID: DefaultPresetID}
		k.o.Decks = append(k.o.Decks, d)
	}
	k.defaultDeck, k.defaultDeckName = d.ID, d.Name
	return k.defaultDeck
}

// repairDeck puts a card whose deck is missing into the Default deck, and
// returns a card home whose filtered deck is missing or no longer filtered
// (to Default when its home is missing or filtered too). It reports
// whether it changed the card.
func (k *dbCheck) repairDeck(card *Card) bool {
	regular := func(id ID) bool {
		d, ok := k.c.decks[id]
		return ok && d.Filter == nil
	}
	deck, inDeck := k.c.decks[card.DeckID]
	switch {
	case card.HomeDeckID == 0 && inDeck && deck.Filter == nil:
		return false
	case card.HomeDeckID == 0:
		k.edit(card).DeckID = k.defaultDeckID()
	case !regular(card.HomeDeckID):
		e := k.edit(card)
		e.HomeDeckID = k.defaultDeckID()
		if !inDeck || deck.Filter == nil {
			e.returnHome()
		}
	case !inDeck || deck.Filter == nil:
		k.edit(card).returnHome()
	default:
		return false
	}
	return true
}

// validDue reports whether a (re)learning or review card's due time is in
// range: set, and no later than the longest interval from now.
func validDue(due, now time.Time) bool {
	return due.After(time.Unix(0, 0)) && due.Before(now.AddDate(0, 0, MaxIntervalDays+1))
}

func validMemory(m scheduler.Memory) bool {
	finite := func(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
	return finite(m.Stability) && m.Stability > 0 && finite(m.Difficulty) && m.Difficulty >= 1 && m.Difficulty <= 10
}

// repairScheduling repairs a card's invalid scheduling values and reports
// whether it had any. A card in an unknown state becomes New; review and
// relearning cards without a (valid) memory state are queued for
// checkHistory.
func (k *dbCheck) repairScheduling(card *Card) bool {
	cs := card.CardState
	if cs.State < scheduler.New || cs.State > scheduler.Relearning {
		e := k.edit(card)
		e.CardState = scheduler.CardState{State: scheduler.New}
		if e.NewPos <= 0 {
			e.NewPos = k.c.newPos()
		}
		return true
	}
	if cs.State == scheduler.New {
		return false
	}
	changed := false
	fix := func(bad bool, repair func()) {
		if bad {
			repair()
			changed = true
		}
	}
	fix(cs.Reps < 0, func() { cs.Reps = 0 })
	fix(cs.Lapses < 0, func() { cs.Lapses = 0 })
	fix(cs.Step < 0, func() { cs.Step = 0 })
	minIvl := 0
	if cs.State == scheduler.Review {
		minIvl = 1
	}
	fix(cs.ScheduledDays < minIvl, func() { cs.ScheduledDays = minIvl })
	fix(cs.ScheduledDays > MaxIntervalDays, func() { cs.ScheduledDays = MaxIntervalDays })
	fix(!validDue(cs.Due, k.now), func() { cs.Due = k.now })
	fix(cs.Memory != nil && !validMemory(*cs.Memory), func() { cs.Memory = nil })
	needMemory := cs.Memory == nil && (cs.State == scheduler.Review || cs.State == scheduler.Relearning)
	if !changed && !needMemory {
		return false
	}
	e := k.edit(card)
	e.CardState = cs
	if needMemory {
		k.needMemory = append(k.needMemory, e)
	}
	return true
}

// addMissingCards generates the cards notes should have but lack, New, in
// the deck of the note's first card.
func (k *dbCheck) addMissingCards() {
	c := k.c
	for _, nid := range sortedIDs(c.notes) {
		if k.delNotes[nid] {
			continue
		}
		n := c.notes[nid]
		if nn, ok := k.notes[nid]; ok {
			n = nn
		}
		var first *Card
		have := map[int]bool{}
		for _, cid := range c.cardsByNote[nid] {
			if k.del[cid] {
				continue
			}
			card := c.cards[cid]
			if cp, ok := k.cards[cid]; ok {
				card = cp
			}
			have[card.Ord] = true
			if first == nil || card.Ord < first.Ord {
				first = card
			}
		}
		deck := c.fallbackDeck()
		if first != nil {
			deck = first.DeckID
		}
		for _, ord := range cardOrds(c.noteTypes[n.TypeID], n.Fields) {
			if have[ord] {
				continue
			}
			have[ord] = true
			k.o.Cards = append(k.o.Cards, &Card{ID: c.ids.next(), NoteID: nid, DeckID: deck, Ord: ord,
				NewPos: c.newPos(), CardState: scheduler.CardState{State: scheduler.New}})
			k.counts[ProblemMissingCards]++
		}
	}
}

// checkHistory reads the whole review log: it gives the cards that need a
// memory state one, from their history where possible, and returns how
// many entries belong to cards that do not exist after the repair.
func (k *dbCheck) checkHistory(progress Progress) (int, error) {
	c := k.c
	months, err := c.store.RevlogMonths()
	if err != nil {
		return 0, err
	}
	history := make(map[ID][]*ReviewEntry, len(k.needMemory))
	for _, card := range k.needMemory {
		history[card.ID] = nil
	}
	orphans := 0
	for i, m := range months {
		progress("history", i, len(months))
		c.yield()
		entries, err := c.store.ReadRevlog(m)
		if err != nil {
			return 0, err
		}
		for j, raw := range entries {
			var e ReviewEntry
			if err := json.Unmarshal(raw, &e); err != nil {
				return 0, fmt.Errorf("%w: review log %s entry %d: %v", ErrCorrupt, m, j, err)
			}
			if _, ok := c.cards[e.CardID]; !ok || k.del[e.CardID] {
				orphans++
			}
			if h, ok := history[e.CardID]; ok && !e.Preview {
				history[e.CardID] = append(h, &e)
			}
		}
	}
	progress("history", len(months), len(months))
	for _, card := range k.needMemory {
		h := history[card.ID]
		sort.SliceStable(h, func(i, j int) bool { return h[i].Time.Before(h[j].Time) })
		rs := make([]scheduler.AnkiReview, len(h))
		for i, e := range h {
			rs[i] = ankiReview(e)
		}
		mem, fromHistory := scheduler.RecoverMemory(rs, card.ScheduledDays)
		card.Memory = &mem
		if fromHistory {
			k.fromHistory++
		}
	}
	return orphans, nil
}

// ankiReview describes a review log entry as an Anki review, for replaying
// FSRS over a card's history.
func ankiReview(e *ReviewEntry) scheduler.AnkiReview {
	r := scheduler.AnkiReview{Time: e.Time, Ease: int(e.Rating), Ivl: int64(e.IvlAfter), Factor: 1}
	switch {
	case e.Rating == 0:
		r.Type = 4
	case e.StateBefore == scheduler.Review:
		r.Type = 1
	case e.StateBefore == scheduler.Relearning:
		r.Type = 2
	}
	return r
}

// wasWere is "1 <singular> was" or "N <plural> were".
func wasWere(n int, singular, plural string) string {
	if n == 1 {
		return "1 " + singular + " was"
	}
	return fmt.Sprintf("%d %s were", n, plural)
}

func (k *dbCheck) report(orphans int) CheckReport {
	r := CheckReport{Problems: []CheckProblem{}, OrphanedReviews: orphans}
	for _, kind := range problemOrder {
		n := k.counts[kind]
		if n == 0 {
			continue
		}
		var msg string
		switch kind {
		case ProblemNotesWithoutType:
			msg = usermsg.Tag("checkNotesWithoutType", usermsg.Args(n, k.typeCards),
				wasWere(n, "note whose note type no longer exists", "notes whose note type no longer exists")+
					" removed, with "+plural(k.typeCards, "card")+". The notes are in the deletion log.")
		case ProblemCardsWithoutNote:
			msg = usermsg.Tag("checkCardsWithoutNote", usermsg.Args(n),
				wasWere(n, "card without a note", "cards without a note")+" removed.")
		case ProblemInvalidCards:
			msg = usermsg.Tag("checkInvalidCards", usermsg.Args(n),
				wasWere(n, "card for a card type that no longer exists", "cards for card types that no longer exist")+" removed.")
		case ProblemFieldCount:
			msg = usermsg.Tag("checkFieldCount", usermsg.Args(n),
				wasWere(n, "note with the wrong number of fields", "notes with the wrong number of fields")+" fixed.")
			if k.trimmed {
				msg += " " + usermsg.Text("checkFieldCountTrimmed", "The removed field content is in the deletion log.")
			}
		case ProblemMissingCards:
			msg = usermsg.Tag("checkMissingCards", usermsg.Args(n), wasWere(n, "missing card", "missing cards")+" generated.")
		case ProblemMissingDeck:
			msg = usermsg.Tag("checkMissingDeck", usermsg.Args(n, k.defaultDeckName),
				wasWere(n, "card in a missing deck", "cards in missing decks")+" moved to "+k.defaultDeckName+".")
		case ProblemScheduling:
			msg = usermsg.Tag("checkScheduling", usermsg.Args(n),
				wasWere(n, "card with invalid scheduling", "cards with invalid scheduling")+" repaired.")
			if k.fromHistory > 0 {
				msg += " " + usermsg.Tag("checkSchedulingFromHistory", usermsg.Args(k.fromHistory),
					fmt.Sprintf("Memory states of %s were recomputed from the review history.", plural(k.fromHistory, "card")))
			}
		case ProblemTags:
			msg = usermsg.Tag("checkTags", usermsg.Args(n),
				wasWere(n, "note with duplicate or invalid tags", "notes with duplicate or invalid tags")+" fixed.")
		}
		r.Problems = append(r.Problems, CheckProblem{Kind: kind, Count: n, Message: msg})
	}
	return r
}
