package core

import (
	"encoding/json"
	"fmt"
	"sort"

	"cardflip/golang/storage"
)

// Snapshot returns the collection as JSON, for backups.
func (c *Collection) Snapshot() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.snapshot()
}

// RevlogMonths lists the months of the stored review log.
func (c *Collection) RevlogMonths() ([]string, error) { return c.store.RevlogMonths() }

// ReadRevlog returns one month of the stored review log.
func (c *Collection) ReadRevlog(month string) ([]json.RawMessage, error) {
	return c.store.ReadRevlog(month)
}

// EachMedia calls fn for every stored media file.
func (c *Collection) EachMedia(fn func(name string, data []byte) error) error {
	return c.store.EachMedia(fn)
}

// RestoreError explains why a backup cannot be restored.
type RestoreError struct{ Msg string }

func (e *RestoreError) Error() string { return e.Msg }

func damaged(format string, args ...any) error {
	return &RestoreError{Msg: "This backup is damaged: " + fmt.Sprintf(format, args...)}
}

// validateState checks a restored collection's internal references.
func validateState(s *state) error {
	if s.Version > snapshotVersion {
		return &RestoreError{Msg: "This backup was made by a newer version of Cardflip. Update Cardflip to restore it."}
	}
	if len(s.Decks) == 0 {
		return damaged("it has no decks")
	}
	presets, decks, types, notes := map[ID]bool{}, map[ID]bool{}, map[ID]*NoteType{}, map[ID]*Note{}
	for _, p := range s.Presets {
		if p == nil || p.ID == 0 || p.Name == "" {
			return damaged("a preset is invalid")
		}
		presets[p.ID] = true
	}
	if s.Version >= 2 && !presets[DefaultPresetID] {
		return damaged("it has no Default preset")
	}
	for _, d := range s.Decks {
		if d == nil || d.ID == 0 || d.Name == "" {
			return damaged("a deck is invalid")
		}
		if d.legacy == nil && !presets[d.PresetID] {
			return damaged("a deck uses a missing preset")
		}
		decks[d.ID] = true
	}
	for _, nt := range s.NoteTypes {
		if nt == nil || len(nt.Fields) == 0 || len(nt.Templates) == 0 {
			return damaged("a note type is invalid")
		}
		types[nt.ID] = nt
	}
	for _, n := range s.Notes {
		if n == nil || types[n.TypeID] == nil {
			return damaged("a note has no note type")
		}
		notes[n.ID] = n
	}
	for _, card := range s.Cards {
		if card == nil || notes[card.NoteID] == nil || !decks[card.DeckID] || card.HomeDeckID != 0 && !decks[card.HomeDeckID] {
			return damaged("a card refers to a missing note or deck")
		}
	}
	return nil
}

// Restore replaces the whole collection, review log and deletion log with a
// backup's content. Nothing changes unless the backup is valid and
// completely saved.
func (c *Collection) Restore(collection []byte, revlog, deletions []json.RawMessage, media map[string][]byte) error {
	return c.restore(collection, revlog, deletions, media, false)
}

// RestoreKeepingMedia replaces the collection, review log and deletion log
// with a backup's content and leaves the media files as they are, for
// backups made without media.
func (c *Collection) RestoreKeepingMedia(collection []byte, revlog, deletions []json.RawMessage) error {
	return c.restore(collection, revlog, deletions, nil, true)
}

// parseState reads and checks a collection that replaces this one, and
// brings it up to date.
func (c *Collection) parseState(collection []byte) (*state, error) {
	var s state
	if err := json.Unmarshal(collection, &s); err != nil {
		return nil, damaged("collection.json cannot be read (%v)", err)
	}
	if err := validateState(&s); err != nil {
		return nil, err
	}
	compactPositions(s.Cards)
	c.ids.observe(s.NextID)
	migrateState(&s, c.ids.next)
	return &s, nil
}

// restore replaces the collection. The replacement starts a new lineage
// (sync design decision 7), so that the next sync of each link asks how to
// combine it.
func (c *Collection) restore(collection []byte, revlog, deletions []json.RawMessage, media map[string][]byte, keepMedia bool) error {
	s, err := c.parseState(collection)
	if err != nil {
		return err
	}
	if !s.Normalized {
		// A backup of an earlier version (see migrateNFC). Without media,
		// the stored files are kept, and those are in NFC already.
		var clashes map[string]string
		media, clashes = nfcMedia(media)
		nfcState(s, clashes)
	}
	entries := make([]*ReviewEntry, 0, len(revlog))
	for i, raw := range revlog {
		var e ReviewEntry
		if err := json.Unmarshal(raw, &e); err != nil || e.ID == 0 || e.Time.IsZero() {
			return damaged("review log entry %d cannot be read", i+1)
		}
		entries = append(entries, &e)
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].ID < entries[j].ID })
	deleted := make([]ID, len(deletions))
	for i, raw := range deletions {
		var e DeletedNote
		if err := json.Unmarshal(raw, &e); err != nil || e.ID == 0 {
			return damaged("deletion log entry %d cannot be read", i+1)
		}
		deleted[i] = e.ID
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.continueState(s, newEpoch())
	// The media files are replaced too; the manifest is built again when
	// sync needs it.
	s.Media, s.MediaBuilt = nil, false
	replace := c.store.Replace
	if keepMedia {
		replace = c.store.ReplaceKeepingMedia
	}
	r, err := replace()
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		if !ok {
			r.Abort()
		}
	}()
	names := make([]string, 0, len(media))
	for name := range media {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := r.PutMedia(name, media[name]); err != nil {
			return fmt.Errorf("could not save media file %s: %w", name, err)
		}
		c.yield()
	}
	for i, e := range entries {
		raw, err := json.Marshal(e)
		if err != nil {
			return err
		}
		if err := r.AppendRevlog(monthKey(e.Time), int64(e.ID), raw); err != nil {
			return err
		}
		if i%500 == 0 {
			c.yield()
		}
	}
	for i, raw := range deletions {
		if err := r.AppendDeletion(int64(deleted[i]), raw); err != nil {
			return err
		}
		if i%500 == 0 {
			c.yield()
		}
	}
	if err := c.commitState(r, s); err != nil {
		return err
	}
	ok = true
	return nil
}

// continueState makes a replacing collection carry on this one's sequence
// numbers and IDs, so that changes after it are newer than anything a link
// has seen, and gives it its lineage. Callers hold the write lock.
func (c *Collection) continueState(s *state, epoch string) {
	s.Epoch, s.Seq, s.NextID = epoch, max(s.Seq, c.seq), max(s.NextID, c.ids.high())
}

// commitState saves s as the snapshot of the replacement r and makes it
// the collection. Callers hold the write lock.
func (c *Collection) commitState(r *storage.Replacement, s *state) error {
	snap, err := json.Marshal(s)
	if err != nil {
		return err
	}
	if err := r.Commit(snap); err != nil {
		return err
	}
	c.applyState(s)
	c.ensureMeta()
	c.mediaIndex = nil
	c.hist.clear()
	c.revision.Add(1)
	if err := c.loadRevlog(); err != nil {
		return err
	}
	c.events.publish(Change{Kind: "restore", All: true, Settings: true, Decks: true, NoteTypes: true, Media: true,
		Notes: true, Cards: true, Reviews: true})
	return nil
}
