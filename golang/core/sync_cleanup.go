package core

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"
)

// The clean-up after a merge (sync design decision 5) resolves the name
// clashes a merge can leave and repairs references to deleted objects. Both
// devices run it on the same objects and get the same result, so its
// changes converge even though each device stamps its own. It runs after
// every batch, on whatever the collection holds then, so it never creates
// cards: a note's cards may come in a later batch.

// syncCleanup cleans up after the batch op o and returns the local ops it
// committed. Callers hold the write lock.
func (c *Collection) syncCleanup(o *op) ([]*op, error) {
	var done []*op
	clashes := c.resolveClashes()
	if !clashes.empty() {
		if err := c.commit(clashes); err != nil {
			return done, err
		}
		done = append(done, clashes)
	}
	repair := c.repairReferences(o, clashes)
	if !repair.empty() {
		if err := c.commit(repair); err != nil {
			return done, err
		}
		done = append(done, repair)
	}
	return done, nil
}

// byName groups objects by name, ignoring case, keeping the groups with
// more than one, each sorted by ID.
func byName[T any](m map[ID]*T, name func(*T) string, id func(*T) ID) [][]*T {
	groups := map[string][]*T{}
	for _, v := range m {
		k := strings.ToLower(name(v))
		groups[k] = append(groups[k], v)
	}
	var out [][]*T
	for _, k := range slices.Sorted(maps.Keys(groups)) {
		g := groups[k]
		if len(g) > 1 {
			slices.SortFunc(g, func(a, b *T) int { return cmp.Compare(id(a), id(b)) })
			out = append(out, g)
		}
	}
	return out
}

// freeName returns name, or name with " (2)" or the next free number when
// taken reports it is taken.
func freeName(name string, taken func(string) bool) string {
	if !taken(name) {
		return name
	}
	for i := 2; ; i++ {
		if n := fmt.Sprintf("%s (%d)", name, i); !taken(n) {
			return n
		}
	}
}

// resolveClashes builds the op that resolves name clashes: two regular
// decks, or two filtered ones, are combined into the older (smaller ID),
// which the newer one's tombstone redirects to; of a regular and a
// filtered deck, the filtered one is renamed, as a filtered deck cannot
// hold the other's subdecks or cards. Note types with the same fields and
// card types are combined into the older one, other note types and presets
// are renamed. Callers hold the write lock.
func (c *Collection) resolveClashes() *op {
	o := &op{Kind: opSyncFix}
	taken := func(names map[string]bool) func(string) bool {
		return func(n string) bool { return names[strings.ToLower(n)] }
	}

	presetNames := map[string]bool{}
	for _, p := range c.presets {
		presetNames[strings.ToLower(p.Name)] = true
	}
	for _, g := range byName(c.presets, func(p *Preset) string { return p.Name }, func(p *Preset) ID { return p.ID }) {
		for _, p := range g[1:] {
			np := *p
			np.Name = freeName(p.Name, taken(presetNames))
			presetNames[strings.ToLower(np.Name)] = true
			o.Presets = append(o.Presets, &np)
		}
	}

	typeNames := map[string]bool{}
	for _, nt := range c.noteTypes {
		typeNames[strings.ToLower(nt.Name)] = true
	}
	retype := map[ID]ID{}
	for _, g := range byName(c.noteTypes, func(nt *NoteType) string { return nt.Name }, func(nt *NoteType) ID { return nt.ID }) {
		kept := g[0]
		for _, nt := range g[1:] {
			if sameNoteType(kept, nt) {
				retype[nt.ID] = kept.ID
				o.DelNoteTypes = append(o.DelNoteTypes, nt.ID)
				o.Redirects = append(o.Redirects, redirect{kindNoteType, nt.ID, kept.ID})
				continue
			}
			cp := cloneNoteType(nt)
			cp.Name = freeName(nt.Name, taken(typeNames))
			typeNames[strings.ToLower(cp.Name)] = true
			o.NoteTypes = append(o.NoteTypes, cp)
		}
	}
	for _, id := range sortedIDs(c.notes) {
		if to, ok := retype[c.notes[id].TypeID]; ok {
			nn := cloneNote(c.notes[id])
			nn.TypeID = to
			o.Notes = append(o.Notes, &nn)
		}
	}

	deckNames := map[string]bool{}
	for _, d := range c.decks {
		deckNames[strings.ToLower(d.Name)] = true
	}
	move := map[ID]ID{}
	for _, g := range byName(c.decks, func(d *Deck) string { return d.Name }, func(d *Deck) ID { return d.ID }) {
		regular := slices.IndexFunc(g, func(d *Deck) bool { return d.Filter == nil })
		filtered := slices.IndexFunc(g, func(d *Deck) bool { return d.Filter != nil })
		if regular >= 0 && filtered >= 0 {
			for _, d := range g {
				if d.Filter != nil {
					nd := *d
					nd.Name = freeName(d.Name, taken(deckNames))
					deckNames[strings.ToLower(nd.Name)] = true
					o.Decks = append(o.Decks, &nd)
				}
			}
			continue
		}
		for _, d := range g[1:] {
			move[d.ID] = g[0].ID
			o.DelDecks = append(o.DelDecks, d.ID)
			o.Redirects = append(o.Redirects, redirect{kindDeck, d.ID, g[0].ID})
		}
	}
	if len(move) > 0 {
		for _, id := range sortedIDs(c.cards) {
			card := c.cards[id]
			to, inMoved := move[card.DeckID]
			home, homeMoved := move[card.HomeDeckID]
			if !inMoved && !homeMoved {
				continue
			}
			cp := *card
			if inMoved {
				cp.DeckID = to
			}
			if homeMoved {
				cp.HomeDeckID = home
			}
			o.Cards = append(o.Cards, &cp)
		}
		for _, id := range sortedIDs(c.noteTypes) {
			nt := c.noteTypes[id]
			var cp *NoteType
			for i, t := range nt.Templates {
				if to, ok := move[t.DeckOverride]; ok {
					if cp == nil {
						cp = cloneNoteType(nt)
					}
					cp.Templates[i].DeckOverride = to
				}
			}
			if cp != nil {
				o.NoteTypes = replaceNoteType(o.NoteTypes, cp)
			}
		}
	}
	return o
}

// replaceNoteType puts nt into list, replacing a version of it already
// there.
func replaceNoteType(list []*NoteType, nt *NoteType) []*NoteType {
	for i, x := range list {
		if x.ID == nt.ID {
			list[i].Templates = nt.Templates
			return list
		}
	}
	return append(list, nt)
}

// sameNoteType reports whether two note types have the same fields and
// card types, so that their notes can share one.
func sameNoteType(a, b *NoteType) bool {
	if !sameStructure(a, b) {
		return false
	}
	for i := range a.Templates {
		if !sameFormats(a.Templates[i], b.Templates[i]) {
			return false
		}
	}
	return true
}

// redirect records that a merged-away deck or note type redirects to the
// one it was combined into (decision 5): stampSync puts it in the
// tombstone.
type redirect struct {
	kind     syncKind
	from, to ID
}

// redirected follows the redirects of tombstoned decks or note types and
// returns the existing one id ends at. Callers hold a lock.
func (c *Collection) redirected(kind syncKind, id ID, exists func(ID) bool) (ID, bool) {
	for range 16 {
		if exists(id) {
			return id, true
		}
		t, ok := c.tombs[metaKey{kind, id}]
		if !ok || t.aux == 0 {
			return 0, false
		}
		id = ID(t.aux)
	}
	return 0, false
}

// repairReferences builds the op that repairs what the batch op o and the
// clash op left pointing at deleted objects: decks and note types that were
// combined are followed to the kept one; otherwise the rules of Check
// database apply (notes without a note type are deleted, cards without a
// note or card type are deleted, cards without a deck go to Default, notes
// get as many fields as their note type). Only what o and clashes touched
// is checked. Callers hold the write lock.
func (c *Collection) repairReferences(o, clashes *op) *op {
	k := &dbCheck{c: c, o: &op{Kind: opSyncFix}, now: c.now(), counts: map[string]int{},
		cards: map[ID]*Card{}, notes: map[ID]*Note{}, del: map[ID]bool{}, delNotes: map[ID]bool{}}
	notes, cards := map[ID]bool{}, map[ID]bool{}
	decks := map[ID]bool{}
	delTypes, delDecks, delPresets, types := map[ID]bool{}, map[ID]bool{}, map[ID]bool{}, map[ID]bool{}
	for _, x := range []*op{o, clashes} {
		for _, n := range x.Notes {
			notes[n.ID] = true
		}
		for _, card := range x.Cards {
			cards[card.ID] = true
		}
		for _, d := range x.Decks {
			decks[d.ID] = true
		}
		for _, nt := range x.NoteTypes {
			types[nt.ID] = true
		}
		for _, id := range x.DelNoteTypes {
			delTypes[id] = true
		}
		for _, id := range x.DelDecks {
			delDecks[id] = true
		}
		for _, id := range x.DelPresets {
			delPresets[id] = true
		}
	}
	if len(delTypes)+len(types) > 0 {
		for id, n := range c.notes {
			if delTypes[n.TypeID] || types[n.TypeID] {
				notes[id] = true
			}
		}
	}
	if len(delDecks) > 0 {
		for id, card := range c.cards {
			if delDecks[card.DeckID] || delDecks[card.HomeDeckID] {
				cards[id] = true
			}
		}
	}
	if len(delPresets) > 0 {
		for id, d := range c.decks {
			if delPresets[d.PresetID] {
				decks[id] = true
			}
		}
	}
	for id := range notes {
		for _, cid := range c.cardsByNote[id] {
			cards[cid] = true
		}
	}

	for _, id := range slices.Sorted(maps.Keys(decks)) {
		d, ok := c.decks[id]
		if !ok || d.Filter != nil {
			continue
		}
		if _, ok := c.presets[d.PresetID]; !ok {
			nd := *d
			nd.PresetID = DefaultPresetID
			k.o.Decks = append(k.o.Decks, &nd)
		}
	}
	typeExists := func(id ID) bool { _, ok := c.noteTypes[id]; return ok }
	for _, id := range slices.Sorted(maps.Keys(notes)) {
		n, ok := c.notes[id]
		if !ok {
			continue
		}
		nt, ok := c.noteTypes[n.TypeID]
		if !ok {
			if to, ok := c.redirected(kindNoteType, n.TypeID, typeExists); ok {
				nn := cloneNote(n)
				nn.TypeID = to
				k.notes[id] = &nn
				k.o.Notes = append(k.o.Notes, &nn)
				n, nt = &nn, c.noteTypes[to]
			} else {
				k.o.DelNotes = append(k.o.DelNotes, id)
				k.delNotes[id] = true
				for _, cid := range c.cardsByNote[id] {
					k.deleteCard(cid)
				}
				continue
			}
		}
		if len(n.Fields) != len(nt.Fields) {
			if len(n.Fields) > len(nt.Fields) && slices.ContainsFunc(n.Fields[len(nt.Fields):], func(f string) bool { return f != "" }) {
				k.o.Trimmed = append(k.o.Trimmed, n)
			}
			nn, ok := k.notes[id]
			if !ok {
				cp := cloneNote(n)
				nn = &cp
				k.notes[id] = nn
				k.o.Notes = append(k.o.Notes, nn)
			}
			nn.Fields = padFields(slices.Clone(n.Fields[:min(len(n.Fields), len(nt.Fields))]), len(nt.Fields))
		}
	}
	deckExists := func(id ID) bool { _, ok := c.decks[id]; return ok }
	for _, id := range slices.Sorted(maps.Keys(cards)) {
		card, ok := c.cards[id]
		if !ok || k.del[id] {
			continue
		}
		n, ok := k.notes[card.NoteID]
		if !ok {
			n, ok = c.notes[card.NoteID]
		}
		if !ok {
			k.deleteCard(id)
			continue
		}
		if !validOrd(c.noteTypes[n.TypeID], card.Ord) {
			k.deleteCard(id)
			continue
		}
		cur := card
		if !deckExists(card.DeckID) {
			if to, ok := c.redirected(kindDeck, card.DeckID, deckExists); ok {
				cur = k.edit(card)
				cur.DeckID = to
			}
		}
		if card.HomeDeckID != 0 && !deckExists(card.HomeDeckID) {
			if to, ok := c.redirected(kindDeck, card.HomeDeckID, deckExists); ok {
				cur = k.edit(card)
				cur.HomeDeckID = to
			}
		}
		k.repairDeck(cur)
	}
	sortIDs(k.o.DelNotes)
	sortIDs(k.o.DelCards)
	return k.o
}
