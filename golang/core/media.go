package core

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"path"
	"sort"
	"strconv"
	"strings"
)

// maxMediaStem is the longest file name stem AddMedia keeps, in runes.
const maxMediaStem = 100

// AddMedia stores a media file added in the editor and returns the name it
// is stored under (card-browser spec: media file naming). The name keeps
// its extension; a name without a stem (".png") is named after the content.
// A file with the same content as a stored one is not stored again: the
// stored name is returned. A different file whose name is taken gets a
// numbered suffix (cat-1.jpg). Files added in this session count as
// referenced until the app closes, so a compaction before the note is saved
// keeps them; if no note uses them, a later compaction removes them.
func (c *Collection) AddMedia(name string, data []byte) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	sum := sha1.Sum(data)
	key := string(sum[:])
	stem, ext := splitMediaName(cleanMediaName(name))
	if stem == "" {
		stem = hex.EncodeToString(sum[:8])
	}
	sha := hex.EncodeToString(sum[:])
	if existing, err := c.mediaWithContent(key, data); err != nil {
		return "", err
	} else if existing != "" {
		c.added[existing] = true
		return existing, c.recordMedia(existing, sha)
	}
	stored := stem + ext
	for i := 1; ; i++ {
		old, err := c.store.GetMedia(stored)
		if err != nil {
			break // free
		}
		if bytes.Equal(old, data) {
			c.mediaIndex[key] = stored
			c.added[stored] = true
			return stored, c.recordMedia(stored, sha)
		}
		stored = stem + "-" + strconv.Itoa(i) + ext
	}
	if err := c.store.PutMedia(stored, data); err != nil {
		return "", userErr("mediaSaveFailed", "The file could not be saved: %v", err)
	}
	c.mediaIndex[key] = stored
	c.added[stored] = true
	o := &op{}
	c.mediaChanged(o, stored, sha, false)
	if len(o.Media) == 0 {
		c.events.publish(Change{Kind: "media", Media: true})
	}
	return stored, c.commitMedia(o)
}

// recordMedia records a stored file AddMedia returns in the media manifest,
// in case the manifest does not list it as stored. Callers hold the write
// lock.
func (c *Collection) recordMedia(name, sha string) error {
	o := &op{}
	c.mediaChanged(o, name, sha, false)
	return c.commitMedia(o)
}

// mediaWithContent returns the name of a stored file with this content, or
// "". The index of contents is built on first use; files removed since
// (compaction) are checked for and dropped.
func (c *Collection) mediaWithContent(key string, data []byte) (string, error) {
	if c.mediaIndex == nil {
		index := map[string]string{}
		err := c.store.EachMedia(func(name string, data []byte) error {
			sum := sha1.Sum(data)
			if _, ok := index[string(sum[:])]; !ok {
				index[string(sum[:])] = name
			}
			c.yield()
			return nil
		})
		if err != nil {
			return "", err
		}
		c.mediaIndex = index
	}
	name, ok := c.mediaIndex[key]
	if !ok {
		return "", nil
	}
	if old, err := c.store.GetMedia(name); err == nil && bytes.Equal(old, data) {
		return name, nil
	}
	delete(c.mediaIndex, key)
	return "", nil
}

// cleanMediaName reduces a file name to a base name that is safe in an
// <img src="…"> attribute, a [sound:…] tag and a file system.
func cleanMediaName(name string) string {
	name = path.Base(strings.ReplaceAll(nfc(name), `\`, "/"))
	if name == "." || name == "/" {
		return ""
	}
	name = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f || strings.ContainsRune(`[]<>:"/\?*|^&`, r) {
			return -1
		}
		return r
	}, name)
	// Windows does not allow names that end in a dot or space.
	return strings.TrimRight(strings.TrimSpace(name), ". ")
}

// hashedMediaName is name with a hash of data before the extension: the
// name a file gets when its name is taken by a different file.
func hashedMediaName(name string, data []byte) string {
	h := sha256.Sum256(data)
	ext := path.Ext(name)
	return strings.TrimSuffix(name, ext) + "_" + hex.EncodeToString(h[:4]) + ext
}

// splitMediaName splits a cleaned name into stem and extension, shortening
// a long stem.
func splitMediaName(name string) (stem, ext string) {
	ext = path.Ext(name)
	stem = strings.TrimSpace(strings.TrimSuffix(name, ext))
	if r := []rune(stem); len(r) > maxMediaStem {
		stem = string(r[:maxMediaStem])
	}
	return stem, ext
}

// MissingMediaTag is the tag TagMissingMedia adds.
const MissingMediaTag = "missing-media"

// MediaFile is a stored or trashed media file.
type MediaFile struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// MediaCheck is the media report (persistence spec: check media).
type MediaCheck struct {
	// Unused are the stored files that no note field, template or CSS
	// refers to, sorted by name. Static files ("_" names) are never unused.
	Unused      []MediaFile `json:"unused"`
	UnusedBytes int64       `json:"unusedBytes"`
	// Missing are the names notes refer to that are not stored, sorted.
	Missing []string `json:"missing"`
	// MissingNotes is the number of notes that refer to a missing file.
	MissingNotes int `json:"missingNotes"`
	// Trash lists the media trash, sorted by name.
	Trash      []MediaFile `json:"trash"`
	TrashBytes int64       `json:"trashBytes"`
}

// CheckMedia reports unused and missing media and the media trash.
func (c *Collection) CheckMedia() (*MediaCheck, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	r, _, err := c.checkMedia()
	return r, err
}

// mediaUse is which media the notes and note types refer to.
type mediaUse struct {
	// used holds every name a note, template or CSS refers to, as written
	// and decoded.
	used map[string]bool
	// byNote holds each note's references.
	byNote map[ID][]mediaName
}

// mediaName is a reference: the name it means (HTML entities decoded) and
// the forms a stored file may have (see mediaNameForms).
type mediaName struct {
	name  string
	forms []string
}

// in reports whether the reference names a stored file.
func (m mediaName) in(stored map[string]bool) bool {
	for _, f := range m.forms {
		if stored[f] {
			return true
		}
	}
	return false
}

func (c *Collection) mediaUse() mediaUse {
	u := mediaUse{used: map[string]bool{}, byNote: map[ID][]mediaName{}}
	for id, n := range c.notes {
		for _, raw := range mediaRefs(n.Fields...) {
			forms := mediaNameForms(raw)
			for _, f := range forms {
				u.used[f] = true
			}
			u.byNote[id] = append(u.byNote[id], mediaName{name: htmlUnescape(raw), forms: forms})
		}
	}
	for _, nt := range c.noteTypes {
		for _, name := range MediaNames(noteTypeHTML(nt)...) {
			u.used[name] = true
		}
	}
	// Files added in the editor in this session belong to notes that may
	// not be saved yet.
	for name := range c.added {
		u.used[name] = true
	}
	return u
}

// checkMedia builds the report and lists the notes that refer to missing
// files, sorted. Callers hold at least a read lock.
func (c *Collection) checkMedia() (*MediaCheck, []ID, error) {
	u := c.mediaUse()
	r := &MediaCheck{Unused: []MediaFile{}, Missing: []string{}, Trash: []MediaFile{}}
	stored := map[string]bool{}
	err := c.store.EachMedia(func(name string, data []byte) error {
		stored[name] = true
		if !u.used[name] && !staticMedia(name) {
			r.Unused = append(r.Unused, MediaFile{Name: name, Size: int64(len(data))})
			r.UnusedBytes += int64(len(data))
		}
		c.yield()
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	missing := map[string]bool{}
	var notes []ID
	for id, refs := range u.byNote {
		found := false
		for _, ref := range refs {
			if !ref.in(stored) {
				missing[ref.name] = true
				found = true
			}
		}
		if found {
			notes = append(notes, id)
		}
	}
	r.MissingNotes = len(notes)
	for name := range missing {
		r.Missing = append(r.Missing, name)
	}
	trash, err := c.store.TrashedMedia()
	if err != nil {
		return nil, nil, err
	}
	for _, f := range trash {
		r.Trash = append(r.Trash, MediaFile{Name: f.Name, Size: f.Size})
		r.TrashBytes += f.Size
	}
	sort.Slice(r.Unused, func(i, j int) bool { return r.Unused[i].Name < r.Unused[j].Name })
	sort.Strings(r.Missing)
	sort.Slice(notes, func(i, j int) bool { return notes[i] < notes[j] })
	return r, notes, nil
}

// DeleteUnusedMedia moves the files CheckMedia reports as unused to the
// media trash and returns how many it moved.
func (c *Collection) DeleteUnusedMedia() (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	r, _, err := c.checkMedia()
	if err != nil {
		return 0, err
	}
	if len(r.Unused) == 0 {
		return 0, nil
	}
	names := make([]string, len(r.Unused))
	for i, f := range r.Unused {
		names[i] = f.Name
	}
	if err := c.store.TrashMedia(names, c.now()); err != nil {
		return 0, userErr("mediaTrashFailed", "The unused files could not be moved to the trash: %v", err)
	}
	c.mediaIndex = nil
	o := &op{}
	for _, name := range names {
		if e, ok := c.media[name]; ok {
			c.mediaChanged(o, name, e.SHA1, true)
		} else if data, err := c.store.TrashedMediaData(name); err == nil {
			c.mediaChanged(o, name, sha1Hex(data), true)
		}
	}
	if len(o.Media) == 0 {
		c.events.publish(Change{Kind: "media", Media: true})
	}
	return len(names), c.commitMedia(o)
}

// TagMissingMedia adds the tag missing-media to every note that refers to
// a missing media file and returns how many notes it tagged. It is one
// undo step.
func (c *Collection) TagMissingMedia() (int, error) {
	c.mu.RLock()
	_, ids, err := c.checkMedia()
	c.mu.RUnlock()
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	if err := c.AddTags(ids, []string{MissingMediaTag}); err != nil {
		return 0, err
	}
	return len(ids), nil
}

// RestoreMediaTrash moves the trashed media files back and returns how many
// it restored. A file whose name a different file has taken since stays in
// the trash.
func (c *Collection) RestoreMediaTrash() (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	names, err := c.store.RestoreTrash()
	o := &op{}
	for _, name := range names {
		if data, err := c.store.GetMedia(name); err == nil {
			c.mediaChanged(o, name, sha1Hex(data), false)
		}
	}
	if len(names) > 0 {
		c.mediaIndex = nil
		if len(o.Media) == 0 {
			c.events.publish(Change{Kind: "media", Media: true})
		}
	}
	if cerr := c.commitMedia(o); err == nil {
		err = cerr
	}
	if err != nil {
		return len(names), userErr("mediaRestoreFailed", "The files could not be restored: %v", err)
	}
	return len(names), nil
}

// EmptyMediaTrash deletes every file in the media trash.
func (c *Collection) EmptyMediaTrash() error {
	if err := c.store.EmptyTrash(); err != nil {
		return userErr("mediaEmptyTrashFailed", "The trash could not be emptied: %v", err)
	}
	return nil
}
