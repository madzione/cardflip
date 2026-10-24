package core

import (
	"encoding/json"
	"fmt"

	"cardflip/golang/storage"
)

// SnapshotVersion is the version of the snapshots this build writes; a
// peer can replace this collection only with snapshots up to it.
const SnapshotVersion = snapshotVersion

// Epoch returns the collection's lineage (sync design decision 7), which
// changes when the collection is replaced.
func (c *Collection) Epoch() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.epoch
}

// Seq returns the sequence number of the collection's latest change.
func (c *Collection) Seq() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.seq
}

// Replacing receives a peer's collection that replaces this one (the
// first-sync choice "replace this"): the peer's review log arrives in
// pages, then its snapshot. The media files are kept: the files the peer's
// media manifest lists that are not stored with the same content are
// fetched afterwards (MissingMedia), and the others go at the next
// compaction. The deletion log is kept. Nothing changes until Commit, and
// the collection can be used meanwhile, but changes made before Commit
// are replaced too.
type Replacing struct {
	c *Collection
	r *storage.Replacement
}

// BeginReplace starts receiving a replacement.
func (c *Collection) BeginReplace() (*Replacing, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	r, err := c.store.ReplaceKeepingMedia()
	if err != nil {
		return nil, err
	}
	dels, err := c.store.ReadDeletions()
	if err == nil {
		for i, raw := range dels {
			var e DeletedNote
			if err = json.Unmarshal(raw, &e); err != nil {
				err = fmt.Errorf("%w: deletion log entry %d: %v", ErrCorrupt, i, err)
				break
			}
			if err = r.AppendDeletion(int64(e.ID), raw); err != nil {
				break
			}
		}
	}
	if err != nil {
		r.Abort()
		return nil, err
	}
	return &Replacing{c: c, r: r}, nil
}

// AddReviews adds review log entries of the replacement.
func (rp *Replacing) AddReviews(entries []json.RawMessage) error {
	for i, raw := range entries {
		var e ReviewEntry
		if err := json.Unmarshal(raw, &e); err != nil || e.ID == 0 || e.Time.IsZero() {
			return fmt.Errorf("review log entry %d of the replacement cannot be read", i+1)
		}
		if err := rp.r.AppendRevlog(monthKey(e.Time), int64(e.ID), raw); err != nil {
			return err
		}
		if i%500 == 0 {
			rp.c.yield()
		}
	}
	return nil
}

// Abort drops what was received.
func (rp *Replacing) Abort() { rp.r.Abort() }

// Commit replaces the collection with snapshot, which takes the lineage
// epoch of the peer it came from.
func (rp *Replacing) Commit(snapshot []byte, epoch string) error {
	c := rp.c
	s, err := c.parseState(snapshot)
	if err != nil {
		rp.r.Abort()
		return err
	}
	// Which manifest entries have their file stored already.
	stored := map[string]string{}
	if s.MediaBuilt {
		if err := c.store.EachMedia(func(name string, data []byte) error {
			if _, ok := s.Media[name]; ok {
				stored[name] = sha1Hex(data)
			}
			c.yield()
			return nil
		}); err != nil {
			rp.r.Abort()
			return err
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.continueState(s, epoch)
	for name, e := range s.Media {
		e.Missing = stored[name] != e.SHA1
		s.Media[name] = e
	}
	if err := c.commitState(rp.r, s); err != nil {
		rp.r.Abort()
		return err
	}
	return nil
}
