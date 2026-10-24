package core

import (
	"encoding/json"
	"fmt"
	"time"

	"cardflip/golang/storage"
)

// DeletedNote is a deletion log entry: a note as it was when it was deleted.
type DeletedNote struct {
	// ID identifies the entry, not the note.
	ID       ID        `json:"id"`
	Time     time.Time `json:"time"`
	NoteID   ID        `json:"noteId"`
	NoteType string    `json:"noteType"`
	Fields   []string  `json:"fields"`
	Tags     []string  `json:"tags,omitempty"`
}

// deletions builds the deletion log entries of the notes o deletes, for
// whatever reason: deleting notes or decks, undoing their creation, or
// redoing a deletion; and of the notes whose fields with content Check
// database trims, as they were. Callers hold the write lock.
func (c *Collection) deletions(o *op) ([]storage.Deletion, error) {
	var out []storage.Deletion
	now := c.now()
	notes := o.Trimmed[:len(o.Trimmed):len(o.Trimmed)]
	for _, id := range o.DelNotes {
		if n, ok := c.notes[id]; ok {
			notes = append(notes, n)
		}
	}
	for _, n := range notes {
		e := DeletedNote{ID: c.ids.next(), Time: now, NoteID: n.ID, Fields: n.Fields, Tags: n.Tags}
		if nt, ok := c.noteTypes[n.TypeID]; ok {
			e.NoteType = nt.Name
		}
		raw, err := json.Marshal(e)
		if err != nil {
			return nil, err
		}
		out = append(out, storage.Deletion{ID: int64(e.ID), Entry: raw})
	}
	return out, nil
}

// DeletionLog returns the deletion log, oldest entry first.
func (c *Collection) DeletionLog() ([]DeletedNote, error) {
	raw, err := c.ReadDeletionLog()
	if err != nil {
		return nil, err
	}
	out := make([]DeletedNote, 0, len(raw))
	for i, r := range raw {
		var e DeletedNote
		if err := json.Unmarshal(r, &e); err != nil {
			return nil, fmt.Errorf("%w: deletion log entry %d: %v", ErrCorrupt, i, err)
		}
		out = append(out, e)
	}
	return out, nil
}

// ReadDeletionLog returns the stored deletion log entries, for backups.
func (c *Collection) ReadDeletionLog() ([]json.RawMessage, error) { return c.store.ReadDeletions() }

// ClearDeletionLog empties the deletion log.
func (c *Collection) ClearDeletionLog() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	snap, err := c.snapshot()
	if err != nil {
		return err
	}
	if err := c.store.ClearDeletions(snap); err != nil {
		return fmt.Errorf("could not clear the deletion log: %w", err)
	}
	c.revision.Add(1)
	return nil
}
