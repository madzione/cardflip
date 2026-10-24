package core

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"maps"
	"slices"

	"cardflip/golang/storage"
)

// The media manifest (sync design decision 13) lists the collection's media
// files by name, each a last-writer-wins register with the file's SHA-1 and
// whether it is in the media trash. It is built when a profile gets its
// first link (BuildMediaManifest) and kept up to date after that through
// the Media changes of ops; collections that never sync have none and pay
// nothing. Only the explicit trash actions of Check media are synced: files
// that compaction moves to the trash because nothing refers to them keep
// their entry, as each device compacts on its own.

// mediaEntry is the manifest entry of one file name. Missing is set, on
// this device only, while a received entry's file is still to be fetched.
type mediaEntry struct {
	SHA1    string `json:"h"`
	Stamp   Stamp  `json:"t"`
	Seq     int64  `json:"s"`
	Trashed bool   `json:"x,omitempty"`
	Missing bool   `json:"m,omitempty"`
}

// mediaChange is a journaled change of the manifest: the new content of a
// name, or its move to or from the trash.
type mediaChange struct {
	Name    string `json:"n"`
	SHA1    string `json:"h"`
	Trashed bool   `json:"x,omitempty"`
	Stamp   Stamp  `json:"t,omitempty"`
	Seq     int64  `json:"s,omitempty"`
	Missing bool   `json:"m,omitempty"`
}

// opMedia is the kind of the op that records media changes outside other
// ops (AddMedia, Check media, the media trash, building the manifest). It
// is not an undo step.
const opMedia = "media"

func sha1Hex(data []byte) string {
	sum := sha1.Sum(data)
	return hex.EncodeToString(sum[:])
}

// mediaChanged records a change of a media file for the manifest, if there
// is one: as a change of o, which the caller commits. Callers hold the
// write lock.
func (c *Collection) mediaChanged(o *op, name, sha string, trashed bool) {
	if c.mediaBuild != nil {
		c.mediaBuild[name] = true
	}
	if c.media == nil {
		return
	}
	if e, ok := c.media[name]; ok && e.SHA1 == sha && e.Trashed == trashed && !e.Missing {
		return
	}
	o.Media = append(o.Media, mediaChange{Name: name, SHA1: sha, Trashed: trashed})
}

// commitMedia commits the media changes of o, if any, as an op of its own.
// Callers hold the write lock.
func (c *Collection) commitMedia(o *op) error {
	if len(o.Media) == 0 {
		return nil
	}
	o.Kind = opMedia
	return c.commit(o)
}

// stampMedia gives the media changes of a local op a stamp and sequence
// numbers.
func (s *syncStamper) stampMedia() {
	for i := range s.o.Media {
		s.o.Media[i].Stamp, s.o.Media[i].Seq = s.next()
	}
}

// applyMedia applies the media changes of o to the manifest.
func (c *Collection) applyMedia(o *op) {
	if o.MediaInit && c.media == nil {
		c.media = map[string]mediaEntry{}
	}
	for _, ch := range o.Media {
		c.clock.observe(ch.Stamp)
		c.seq = max(c.seq, ch.Seq)
		if c.media != nil {
			c.media[ch.Name] = mediaEntry{SHA1: ch.SHA1, Stamp: ch.Stamp, Seq: ch.Seq, Trashed: ch.Trashed, Missing: ch.Missing}
		}
	}
}

// BuildMediaManifest builds the media manifest from the stored files, if
// the collection has none yet. The files are hashed without holding the
// collection's lock, with yields; files that change meanwhile are hashed
// again before the manifest is saved.
func (c *Collection) BuildMediaManifest() error {
	c.mu.Lock()
	if c.media != nil || c.mediaBuild != nil {
		c.mu.Unlock()
		return nil
	}
	c.mediaBuild = map[string]bool{}
	c.mu.Unlock()
	hashes := map[string]string{}
	err := c.store.EachMedia(func(name string, data []byte) error {
		hashes[name] = sha1Hex(data)
		c.yield()
		return nil
	})
	c.mu.Lock()
	defer c.mu.Unlock()
	dirty := c.mediaBuild
	c.mediaBuild = nil
	if err != nil {
		return err
	}
	for name := range dirty {
		data, err := c.store.GetMedia(name)
		switch {
		case err == nil:
			hashes[name] = sha1Hex(data)
		case errors.Is(err, storage.ErrNotFound):
			delete(hashes, name)
		default:
			return err
		}
	}
	o := &op{Kind: opMedia, MediaInit: true}
	for _, name := range slices.Sorted(maps.Keys(hashes)) {
		o.Media = append(o.Media, mediaChange{Name: name, SHA1: hashes[name]})
	}
	return c.commit(o)
}

// HasMediaManifest reports whether the media manifest was built.
func (c *Collection) HasMediaManifest() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.media != nil
}
