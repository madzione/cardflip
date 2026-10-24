package core

import (
	"encoding/json"

	"cardflip/golang/storage"
)

// ReplaceWithImport replaces the whole collection with a new one that
// holds the batch and nothing else (an Anki collection package): its decks,
// note types, notes, cards with their history, presets and media. The
// settings and the deletion log are kept; the settings' references to
// decks and note types of the old collection are dropped. As with Restore,
// nothing changes unless the new collection is completely saved.
func (c *Collection) ReplaceWithImport(b *ImportBatch, progress Progress) (ImportSummary, error) {
	if progress == nil {
		progress = func(string, int, int) {}
	}
	// Build the new collection in memory, starting without decks and note
	// types: the batch brings its own.
	tmp, _, err := Open(storage.NewMemory(), Options{Now: c.now, Yield: c.yield, Names: c.builtinNames()})
	if err != nil {
		return ImportSummary{}, err
	}
	defer tmp.Close()
	tmp.decks, tmp.noteTypes = map[ID]*Deck{}, map[ID]*NoteType{}
	// An empty collection has no media to clash with, so the media files
	// go straight to Restore instead of through the temporary collection.
	b = nfcBatch(b)
	batch := *b
	batch.Media = nil
	sum, err := tmp.Import(&batch, progress)
	if err != nil {
		return ImportSummary{}, err
	}
	sum.Media = len(b.Media)
	if len(tmp.decks) == 0 {
		d := &Deck{ID: tmp.ids.next(), Name: tmp.builtin.deckName(), PresetID: DefaultPresetID}
		tmp.decks[d.ID] = d
	}
	if len(tmp.noteTypes) == 0 {
		for _, nt := range builtinNoteTypes(tmp.builtin) {
			nt.ID = tmp.ids.next()
			tmp.noteTypes[nt.ID] = nt
		}
	}
	settings := c.Settings()
	settings.LastNoteTypeID, settings.LastDeckID, settings.LastDeckByNoteType = 0, 0, nil
	tmp.settings = settings

	progress("saving", 0, 1)
	snap, err := tmp.Snapshot()
	if err != nil {
		return ImportSummary{}, err
	}
	var revlog []json.RawMessage
	months, err := tmp.RevlogMonths()
	if err != nil {
		return ImportSummary{}, err
	}
	for _, m := range months {
		entries, err := tmp.ReadRevlog(m)
		if err != nil {
			return ImportSummary{}, err
		}
		revlog = append(revlog, entries...)
	}
	deletions, err := c.ReadDeletionLog()
	if err != nil {
		return ImportSummary{}, err
	}
	if err := c.Restore(snap, revlog, deletions, b.Media); err != nil {
		return ImportSummary{}, err
	}
	progress("saving", 1, 1)
	return sum, nil
}
