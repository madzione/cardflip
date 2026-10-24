package storage

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// FormatVersion is the version of the on-disk layout, kept in the pointer.
const FormatVersion = 1

// Compaction thresholds for the journal (design D4).
const (
	CompactEntries = 2000
	CompactBytes   = 4 << 20
)

// Store is the collection's durable state on top of a Backend:
//
//	current           pointer {v, gen, rl, media, dl} naming the live sets
//	g<N>/snapshot     the collection as of generation N (opaque bytes)
//	g<N>/journal      ops applied after the snapshot
//	rl<R>/<YYYY-MM>   review log segment for one month
//	m<M>/<hash>       media blob
//	dl<D>/log         deletion log, append-only
//
// Media files compaction or the user removes go to the backend's media
// trash, apart from the keys.
//
// Every change goes to the journal first: a journal record holds the domain
// op and the review log and deletion log writes it implies, so the journal
// is the source of truth and segments are repaired from it at open. Replacing the snapshot
// (compaction, import into a new generation, restore) writes new sets and
// then flips the pointer, so a crash at any point leaves either the old or
// the new state.
type Store struct {
	b Backend

	mu       sync.Mutex
	ptr      pointer
	entries  int
	bytes    int
	pending  []RevlogWrite // segment writes that failed after their journal record
	segments map[string]map[int64]bool
	// pendingDel are deletion log writes that failed after their journal
	// record; deleted holds the IDs in the deletion log, once read.
	pendingDel []Deletion
	deleted    map[int64]bool
	// yield is called in long loops; see SetYield.
	yield func()
	// replacing is set while a Replacement is being built, which uses the
	// next generation: nothing else may flip to it meanwhile.
	replacing bool
}

// SetYield sets a function called periodically in long loops (media
// collection, cleanup), to keep the web UI responsive.
func (s *Store) SetYield(fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.yield = fn
}

func (s *Store) pause() {
	if s.yield != nil {
		s.yield()
	}
}

type pointer struct {
	V     int `json:"v"`
	Gen   int `json:"gen"`
	RL    int `json:"rl"`
	Media int `json:"media"`
	// DL is 0 in collections created before the deletion log.
	DL int `json:"dl,omitempty"`
}

// RevlogWrite adds a review log entry (Entry set) or deletes one (Entry nil)
// in the segment for Month ("YYYY-MM"). Seq is the collection's change
// sequence number of an added entry, which sync uses to find new entries
// (see ReadRevlogRecords); 0 for entries written before it existed.
type RevlogWrite struct {
	Month string          `json:"m"`
	ID    int64           `json:"id"`
	Entry json.RawMessage `json:"e,omitempty"`
	Seq   int64           `json:"s,omitempty"`
}

// Deletion is one deletion log entry. Its ID makes repairing the log from
// the journal add each entry once.
type Deletion struct {
	ID    int64           `json:"id"`
	Entry json.RawMessage `json:"e"`
}

type journalRecord struct {
	Op  json.RawMessage `json:"op"`
	Rev []RevlogWrite   `json:"rev,omitempty"`
	Del []Deletion      `json:"del,omitempty"`
}

type segmentRecord struct {
	ID    int64           `json:"id"`
	Entry json.RawMessage `json:"e,omitempty"`
	Del   bool            `json:"del,omitempty"`
	Seq   int64           `json:"s,omitempty"`
}

// Loaded is what Open read: the snapshot and the ops to replay on top of it.
type Loaded struct {
	// Fresh is true when there is no collection yet. Call Init.
	Fresh    bool
	Snapshot []byte
	Ops      []json.RawMessage
}

// Open reads the live generation. Errors wrapping ErrCorrupt mean the data
// is unreadable; nothing has been modified in that case.
func Open(b Backend) (*Store, *Loaded, error) {
	s := &Store{b: b, segments: map[string]map[int64]bool{}}
	raw, err := b.Get("current")
	if errors.Is(err, ErrNotFound) {
		// Without a pointer, any keys are leftovers of an Init that did not
		// finish: nothing was ever committed.
		if err := s.deleteAll(); err != nil {
			return nil, nil, err
		}
		return s, &Loaded{Fresh: true}, nil
	}
	if err != nil {
		return nil, nil, err
	}
	if err := json.Unmarshal(raw, &s.ptr); err != nil || s.ptr.Gen < 1 {
		return nil, nil, fmt.Errorf("%w: unreadable pointer record", ErrCorrupt)
	}
	if s.ptr.V > FormatVersion {
		return nil, nil, fmt.Errorf("%w: stored data has format %d, this app reads up to %d; update Cardflip",
			ErrCorrupt, s.ptr.V, FormatVersion)
	}

	snap, err := b.Get(s.key("snapshot"))
	if errors.Is(err, ErrNotFound) {
		return nil, nil, fmt.Errorf("%w: snapshot of generation %d is missing", ErrCorrupt, s.ptr.Gen)
	}
	if err != nil {
		return nil, nil, err
	}
	recs, err := b.Records(s.key("journal"))
	if err != nil {
		return nil, nil, fmt.Errorf("journal: %w", err)
	}
	loaded := &Loaded{Snapshot: snap}
	var repairs []RevlogWrite
	var dels []Deletion
	for i, r := range recs {
		var jr journalRecord
		if err := json.Unmarshal(r, &jr); err != nil || len(jr.Op) == 0 {
			return nil, nil, fmt.Errorf("%w: journal record %d is unreadable", ErrCorrupt, i)
		}
		loaded.Ops = append(loaded.Ops, jr.Op)
		repairs = append(repairs, jr.Rev...)
		dels = append(dels, jr.Del...)
		s.entries++
		s.bytes += len(r)
	}
	// Read every segment the journal touches before writing anything, so
	// that a damaged segment is reported without modifying any data.
	for _, w := range repairs {
		if _, err := s.segment(w.Month); err != nil {
			return nil, nil, err
		}
	}
	if len(dels) > 0 {
		if _, err := s.deletedIDs(); err != nil {
			return nil, nil, err
		}
	}
	// Only the last write to an entry counts: an entry added and deleted
	// again (an answer undone) is not added back by the repair.
	type entryKey struct {
		month string
		id    int64
	}
	last := map[entryKey]int{}
	for i, w := range repairs {
		last[entryKey{w.Month, w.ID}] = i
	}
	final := repairs[:0:0]
	for i, w := range repairs {
		if last[entryKey{w.Month, w.ID}] == i {
			final = append(final, w)
		}
	}
	if err := s.applyRevlogs(final); err != nil {
		return nil, nil, err
	}
	for _, d := range dels {
		if err := s.applyDeletion(d); err != nil {
			return nil, nil, err
		}
	}
	if err := s.cleanup(); err != nil {
		return nil, nil, err
	}
	return s, loaded, nil
}

func (s *Store) key(name string) string { return fmt.Sprintf("g%d/%s", s.ptr.Gen, name) }

func segKey(rl int, month string) string { return fmt.Sprintf("rl%d/%s", rl, month) }

func delKey(dl int) string { return fmt.Sprintf("dl%d/log", dl) }

// MediaKey is the storage key of media file name in set m. Names are hashed
// so that any name is a valid, short, case-exact key on every backend.
func mediaKey(m int, name string) string {
	h := sha256.Sum256([]byte(name))
	return fmt.Sprintf("m%d/%s", m, hex.EncodeToString(h[:16]))
}

// Init creates the first generation from snapshot. Only valid when Open
// reported Fresh.
func (s *Store) Init(snapshot []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ptr.Gen != 0 {
		return errors.New("storage: collection already exists")
	}
	next := pointer{V: FormatVersion, Gen: 1, RL: 1, Media: 1, DL: 1}
	if err := s.b.Put(fmt.Sprintf("g%d/snapshot", next.Gen), snapshot); err != nil {
		return err
	}
	return s.flip(next)
}

// Commit durably records one change: op is appended to the journal together
// with its review log writes and deletion log entries, then those are
// applied to the segments and the deletion log. Commit returns after the
// journal record is durable; a failure to update a segment or the deletion
// log is retried later and repaired from the journal at the next open.
func (s *Store) Commit(op json.RawMessage, rev []RevlogWrite, del []Deletion) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ptr.Gen == 0 {
		return errors.New("storage: collection not initialised")
	}
	rec, err := json.Marshal(journalRecord{Op: op, Rev: rev, Del: del})
	if err != nil {
		return err
	}
	if err := s.b.Append(s.key("journal"), rec); err != nil {
		return err
	}
	s.entries++
	s.bytes += len(rec)
	s.pending = append(s.pending, rev...)
	s.pendingDel = append(s.pendingDel, del...)
	s.flushPending()
	return nil
}

func (s *Store) flushPending() {
	if len(s.pending) > 0 && s.applyRevlogs(s.pending) == nil {
		s.pending = nil
	}
	for len(s.pendingDel) > 0 {
		if err := s.applyDeletion(s.pendingDel[0]); err != nil {
			break
		}
		s.pendingDel = s.pendingDel[1:]
	}
}

// deletedIDs returns the IDs of the entries in the deletion log.
func (s *Store) deletedIDs() (map[int64]bool, error) {
	if s.deleted != nil {
		return s.deleted, nil
	}
	recs, err := s.b.Records(delKey(s.ptr.DL))
	if err != nil {
		return nil, fmt.Errorf("deletion log: %w", err)
	}
	ids := map[int64]bool{}
	for i, r := range recs {
		var d Deletion
		if err := json.Unmarshal(r, &d); err != nil {
			return nil, fmt.Errorf("%w: deletion log record %d is unreadable", ErrCorrupt, i)
		}
		ids[d.ID] = true
	}
	s.deleted = ids
	return ids, nil
}

// applyDeletion adds d to the deletion log unless it is there already.
func (s *Store) applyDeletion(d Deletion) error {
	ids, err := s.deletedIDs()
	if err != nil {
		return err
	}
	if ids[d.ID] {
		return nil
	}
	data, err := json.Marshal(d)
	if err != nil {
		return err
	}
	if err := s.b.Append(delKey(s.ptr.DL), data); err != nil {
		return err
	}
	ids[d.ID] = true
	return nil
}

// ReadDeletions returns the deletion log entries in the order they were
// written.
func (s *Store) ReadDeletions() ([]json.RawMessage, error) {
	s.mu.Lock()
	dl := s.ptr.DL
	s.mu.Unlock()
	recs, err := s.b.Records(delKey(dl))
	if err != nil {
		return nil, fmt.Errorf("deletion log: %w", err)
	}
	seen := map[int64]bool{}
	out := make([]json.RawMessage, 0, len(recs))
	for i, r := range recs {
		var d Deletion
		if err := json.Unmarshal(r, &d); err != nil {
			return nil, fmt.Errorf("%w: deletion log record %d is unreadable", ErrCorrupt, i)
		}
		if !seen[d.ID] {
			seen[d.ID] = true
			out = append(out, d.Entry)
		}
	}
	return out, nil
}

// ClearDeletions empties the deletion log. snapshot, which must include
// every committed op, starts a new generation, so that the journal can no
// longer repair cleared entries back into the log.
func (s *Store) ClearDeletions(snapshot []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ptr.Gen == 0 {
		return errors.New("storage: collection not initialised")
	}
	if s.replacing {
		return ErrBusy
	}
	if err := s.flushed(); err != nil {
		return err
	}
	next := s.ptr
	next.Gen++
	next.DL++
	for _, k := range []string{fmt.Sprintf("g%d/journal", next.Gen), delKey(next.DL)} {
		if err := s.b.Delete(k); err != nil {
			return err
		}
	}
	if err := s.b.Put(fmt.Sprintf("g%d/snapshot", next.Gen), snapshot); err != nil {
		return err
	}
	if err := s.flip(next); err != nil {
		return err
	}
	s.entries, s.bytes = 0, 0
	return nil
}

// flushed writes pending segment and deletion log writes, which must be
// done before the journal that records them goes away.
func (s *Store) flushed() error {
	s.flushPending()
	if len(s.pending) > 0 {
		return fmt.Errorf("storage: review log segments could not be updated")
	}
	if len(s.pendingDel) > 0 {
		return fmt.Errorf("storage: the deletion log could not be updated")
	}
	return nil
}

// segment returns which entry IDs month's segment holds: true for present,
// false for deleted.
func (s *Store) segment(month string) (map[int64]bool, error) {
	if seg, ok := s.segments[month]; ok {
		return seg, nil
	}
	recs, err := s.b.Records(segKey(s.ptr.RL, month))
	if err != nil {
		return nil, fmt.Errorf("review log %s: %w", month, err)
	}
	seg := map[int64]bool{}
	for i, r := range recs {
		var sr segmentRecord
		if err := json.Unmarshal(r, &sr); err != nil {
			return nil, fmt.Errorf("%w: review log %s record %d is unreadable", ErrCorrupt, month, i)
		}
		seg[sr.ID] = !sr.Del
	}
	s.segments[month] = seg
	return seg, nil
}

// applyRevlogs makes the segments agree with ws, with one write per
// segment. It is idempotent, which makes repairing from the journal and
// retrying after an error safe: a record written twice is read once.
func (s *Store) applyRevlogs(ws []RevlogWrite) error {
	var months []string
	byMonth := map[string][]RevlogWrite{}
	for _, w := range ws {
		if _, ok := byMonth[w.Month]; !ok {
			months = append(months, w.Month)
		}
		byMonth[w.Month] = append(byMonth[w.Month], w)
	}
	for _, m := range months {
		if err := s.applySegment(m, byMonth[m]); err != nil {
			return err
		}
		s.pause()
	}
	return nil
}

func (s *Store) applySegment(month string, ws []RevlogWrite) error {
	seg, err := s.segment(month)
	if err != nil {
		return err
	}
	changed := map[int64]bool{}
	var recs [][]byte
	for _, w := range ws {
		present, ok := changed[w.ID]
		if !ok {
			present = seg[w.ID]
		}
		var rec segmentRecord
		switch {
		case w.Entry != nil && !present:
			rec = segmentRecord{ID: w.ID, Entry: w.Entry, Seq: w.Seq}
		case w.Entry == nil && present:
			rec = segmentRecord{ID: w.ID, Del: true}
		default:
			continue
		}
		data, err := json.Marshal(rec)
		if err != nil {
			return err
		}
		recs = append(recs, data)
		changed[w.ID] = w.Entry != nil
	}
	if len(recs) == 0 {
		return nil
	}
	if err := appendAll(s.b, segKey(s.ptr.RL, month), recs); err != nil {
		return err
	}
	for id, present := range changed {
		seg[id] = present
	}
	return nil
}

// ReadRevlog returns the review log entries of month that were not deleted,
// in the order they were written.
func (s *Store) ReadRevlog(month string) ([]json.RawMessage, error) {
	s.mu.Lock()
	rl := s.ptr.RL
	s.mu.Unlock()
	return readSegment(s.b, segKey(rl, month), month)
}

func readSegment(b Backend, key, month string) ([]json.RawMessage, error) {
	recs, err := readSegmentRecords(b, key, month)
	if err != nil {
		return nil, err
	}
	out := make([]json.RawMessage, len(recs))
	for i, r := range recs {
		out[i] = r.Entry
	}
	return out, nil
}

// RevlogHas reports which of ids month's segment holds.
func (s *Store) RevlogHas(month string, ids []int64) (map[int64]bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	seg, err := s.segment(month)
	if err != nil {
		return nil, err
	}
	out := map[int64]bool{}
	for _, id := range ids {
		if seg[id] {
			out[id] = true
		}
	}
	return out, nil
}

// RevlogRecord is a stored review log entry with its ID and change sequence
// number (see RevlogWrite).
type RevlogRecord struct {
	ID    int64
	Seq   int64
	Entry json.RawMessage
}

// ReadRevlogRecords returns the review log entries of month that were not
// deleted, in the order they were written, with their sequence numbers.
func (s *Store) ReadRevlogRecords(month string) ([]RevlogRecord, error) {
	s.mu.Lock()
	rl := s.ptr.RL
	s.mu.Unlock()
	return readSegmentRecords(s.b, segKey(rl, month), month)
}

func readSegmentRecords(b Backend, key, month string) ([]RevlogRecord, error) {
	recs, err := b.Records(key)
	if err != nil {
		return nil, fmt.Errorf("review log %s: %w", month, err)
	}
	var order []int64
	entries := map[int64]segmentRecord{}
	for i, r := range recs {
		var sr segmentRecord
		if err := json.Unmarshal(r, &sr); err != nil {
			return nil, fmt.Errorf("%w: review log %s record %d is unreadable", ErrCorrupt, month, i)
		}
		if sr.Del {
			delete(entries, sr.ID)
			continue
		}
		if _, dup := entries[sr.ID]; !dup {
			order = append(order, sr.ID)
		}
		entries[sr.ID] = sr
	}
	out := make([]RevlogRecord, 0, len(entries))
	for _, id := range order {
		if e, ok := entries[id]; ok {
			out = append(out, RevlogRecord{ID: id, Seq: e.Seq, Entry: e.Entry})
			delete(entries, id)
		}
	}
	return out, nil
}

// RevlogMonths lists the months that have a review log segment, oldest first.
func (s *Store) RevlogMonths() ([]string, error) {
	s.mu.Lock()
	prefix := fmt.Sprintf("rl%d/", s.ptr.RL)
	s.mu.Unlock()
	keys, err := s.b.List(prefix)
	if err != nil {
		return nil, err
	}
	months := make([]string, len(keys))
	for i, k := range keys {
		months[i] = strings.TrimPrefix(k, prefix)
	}
	return months, nil
}

// PutMedia stores a media file. Call it before committing the op that
// references the file; files nothing references are moved to the media
// trash at compaction.
func (s *Store) PutMedia(name string, data []byte) error {
	s.mu.Lock()
	m := s.ptr.Media
	s.mu.Unlock()
	return s.b.Put(mediaKey(m, name), encodeMedia(name, data))
}

// GetMedia returns a media file, or ErrNotFound.
func (s *Store) GetMedia(name string) ([]byte, error) {
	s.mu.Lock()
	m := s.ptr.Media
	s.mu.Unlock()
	return s.getMediaIn(m, name)
}

// getMedia is GetMedia for callers holding s.mu.
func (s *Store) getMedia(name string) ([]byte, error) { return s.getMediaIn(s.ptr.Media, name) }

func (s *Store) getMediaIn(m int, name string) ([]byte, error) {
	v, err := s.b.Get(mediaKey(m, name))
	if err != nil {
		return nil, err
	}
	stored, data, err := decodeMedia(v)
	if err != nil {
		return nil, err
	}
	if stored != name {
		return nil, ErrNotFound
	}
	return data, nil
}

// HasMedia reports whether media file name is stored.
func (s *Store) HasMedia(name string) bool {
	_, err := s.GetMedia(name)
	return err == nil
}

// EachMedia calls fn for every stored media file.
func (s *Store) EachMedia(fn func(name string, data []byte) error) error {
	s.mu.Lock()
	prefix := fmt.Sprintf("m%d/", s.ptr.Media)
	s.mu.Unlock()
	keys, err := s.b.List(prefix)
	if err != nil {
		return err
	}
	for _, k := range keys {
		v, err := s.b.Get(k)
		if err != nil {
			return err
		}
		name, data, err := decodeMedia(v)
		if err != nil {
			return err
		}
		if err := fn(name, data); err != nil {
			return err
		}
	}
	return nil
}

func encodeMedia(name string, data []byte) []byte {
	out := binary.AppendUvarint(nil, uint64(len(name)))
	out = append(out, name...)
	return append(out, data...)
}

func decodeMedia(v []byte) (string, []byte, error) {
	n, k := binary.Uvarint(v)
	if k <= 0 || uint64(len(v)-k) < n {
		return "", nil, fmt.Errorf("%w: media blob header", ErrCorrupt)
	}
	return string(v[k : k+int(n)]), v[k+int(n):], nil
}

// NeedsCompaction reports whether the journal passed a threshold.
func (s *Store) NeedsCompaction() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.entries >= CompactEntries || s.bytes >= CompactBytes
}

// JournalEntries returns the number of ops in the live journal.
func (s *Store) JournalEntries() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.entries
}

// MediaTrashRetention is how long compaction keeps files in the media trash.
const MediaTrashRetention = 30 * 24 * time.Hour

// Compact writes snapshot, which must include every committed op, as a new
// generation with an empty journal, then moves media to the trash, trashed
// at now, unless they are named in keep or keepName (if not nil) returns
// true for them. It deletes files trashed more than MediaTrashRetention
// before now.
func (s *Store) Compact(snapshot []byte, keep []string, keepName func(string) bool, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ptr.Gen == 0 {
		return errors.New("storage: collection not initialised")
	}
	if s.replacing {
		return ErrBusy
	}
	if err := s.flushed(); err != nil {
		return err
	}
	next := s.ptr
	next.Gen++
	if err := s.b.Delete(fmt.Sprintf("g%d/journal", next.Gen)); err != nil {
		return err
	}
	if err := s.b.Put(fmt.Sprintf("g%d/snapshot", next.Gen), snapshot); err != nil {
		return err
	}
	if err := s.flip(next); err != nil {
		return err
	}
	s.entries, s.bytes = 0, 0
	if err := s.collectMedia(keep, keepName, now); err != nil {
		return err
	}
	return s.purgeTrash(func(t TrashedMedia) bool { return now.Sub(t.Trashed) > MediaTrashRetention })
}

func (s *Store) collectMedia(keep []string, keepName func(string) bool, now time.Time) error {
	want := map[string]bool{}
	for _, name := range keep {
		want[mediaKey(s.ptr.Media, name)] = true
	}
	keys, err := s.b.List(fmt.Sprintf("m%d/", s.ptr.Media))
	if err != nil {
		return err
	}
	for _, k := range keys {
		if !want[k] {
			s.pause()
			if err := s.trashKey(k, keepName, now); err != nil {
				return err
			}
		}
	}
	return nil
}

// trashKey moves the media blob at key to the trash, unless keepName
// returns true for its name. The trash is written first, so a crash in
// between leaves the file in both places, never in neither.
func (s *Store) trashKey(key string, keepName func(string) bool, now time.Time) error {
	v, err := s.b.Get(key)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	// A blob without a readable name cannot be restored by name.
	if name, data, err := decodeMedia(v); err == nil {
		if keepName != nil && keepName(name) {
			return nil
		}
		if err := s.b.Trash().Put(name, data, now); err != nil {
			return err
		}
	}
	return s.b.Delete(key)
}

// TrashMedia moves the named media files to the trash, trashed at now.
// Names that are not stored are skipped.
func (s *Store) TrashMedia(names []string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, name := range names {
		s.pause()
		if _, err := s.getMedia(name); errors.Is(err, ErrNotFound) {
			continue
		} else if err != nil {
			return err
		}
		if err := s.trashKey(mediaKey(s.ptr.Media, name), nil, now); err != nil {
			return err
		}
	}
	return nil
}

// TrashedMedia lists the media trash, sorted by name.
func (s *Store) TrashedMedia() ([]TrashedMedia, error) { return s.b.Trash().List() }

// RestoreTrash moves the trashed files back into the media and returns the
// names it restored. A file whose name a different stored file has taken
// since stays in the trash; one identical to the stored file is dropped.
func (s *Store) RestoreTrash() ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := s.b.Trash()
	list, err := t.List()
	if err != nil {
		return nil, err
	}
	var restored []string
	for _, f := range list {
		s.pause()
		data, err := t.Get(f.Name)
		if errors.Is(err, ErrNotFound) {
			continue
		}
		if err != nil {
			return restored, err
		}
		cur, err := s.getMedia(f.Name)
		switch {
		case err == nil && !bytes.Equal(cur, data):
			continue
		case err == nil:
		case errors.Is(err, ErrNotFound):
			if err := s.b.Put(mediaKey(s.ptr.Media, f.Name), encodeMedia(f.Name, data)); err != nil {
				return restored, err
			}
			restored = append(restored, f.Name)
		default:
			return restored, err
		}
		if err := t.Delete(f.Name); err != nil {
			return restored, err
		}
	}
	return restored, nil
}

// TrashedMediaData returns the content of a file in the media trash.
func (s *Store) TrashedMediaData(name string) ([]byte, error) { return s.b.Trash().Get(name) }

// EmptyTrash deletes every file in the media trash.
func (s *Store) EmptyTrash() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.purgeTrash(func(TrashedMedia) bool { return true })
}

// PurgeTrash deletes the files trashed before t.
func (s *Store) PurgeTrash(before time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.purgeTrash(func(t TrashedMedia) bool { return t.Trashed.Before(before) })
}

func (s *Store) purgeTrash(drop func(TrashedMedia) bool) error {
	t := s.b.Trash()
	list, err := t.List()
	if err != nil {
		return err
	}
	for _, f := range list {
		if drop(f) {
			s.pause()
			if err := t.Delete(f.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

// flip makes next the live state and removes what the old one used.
func (s *Store) flip(next pointer) error {
	raw, err := json.Marshal(next)
	if err != nil {
		return err
	}
	if err := s.b.Put("current", raw); err != nil {
		return err
	}
	if next.RL != s.ptr.RL {
		s.segments = map[string]map[int64]bool{}
		s.pending = nil
	}
	if next.DL != s.ptr.DL {
		s.deleted = nil
		s.pendingDel = nil
	}
	s.ptr = next
	// Best effort: leftovers are removed again at the next open.
	_ = s.cleanup()
	return nil
}

// cleanup deletes generations and sets the pointer does not name. They are
// older states already replaced, or new ones whose flip never happened.
func (s *Store) cleanup() error {
	keys, err := s.b.List("")
	if err != nil {
		return err
	}
	for _, k := range keys {
		if k == "current" {
			continue
		}
		head, _, _ := strings.Cut(k, "/")
		if live, ok := s.liveSet(head); ok && !live {
			s.pause()
			if err := s.b.Delete(k); err != nil {
				return err
			}
		}
	}
	return nil
}

// liveSet reports whether a key's first segment names a set this store
// owns (ok) and whether that set is the live one.
func (s *Store) liveSet(head string) (live, ok bool) {
	for _, p := range []struct {
		prefix string
		n      int
	}{{"rl", s.ptr.RL}, {"g", s.ptr.Gen}, {"m", s.ptr.Media}, {"dl", s.ptr.DL}} {
		if rest, found := strings.CutPrefix(head, p.prefix); found {
			n, err := strconv.Atoi(rest)
			if err != nil {
				return false, false
			}
			return n == p.n, true
		}
	}
	return false, false
}

func (s *Store) deleteAll() error {
	keys, err := s.b.List("")
	if err != nil {
		return err
	}
	for _, k := range keys {
		if err := s.b.Delete(k); err != nil {
			return err
		}
	}
	return nil
}

// Reset deletes all stored data, after the user explicitly chose to start a
// new collection over unreadable data. The store is then Fresh.
func (s *Store) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.deleteAll(); err != nil {
		return err
	}
	s.ptr = pointer{}
	s.entries, s.bytes = 0, 0
	s.pending = nil
	s.segments = map[string]map[int64]bool{}
	s.pendingDel = nil
	s.deleted = nil
	return nil
}

// Replacement builds a complete new state (snapshot, review log, deletion
// log and media)
// next to the live one. Nothing changes for readers until Commit flips the
// pointer; Abort, or a crash before the flip, leaves the live state as it was.
type Replacement struct {
	s    *Store
	next pointer
	done bool
}

// Replace starts building a replacement state, used by restore.
func (s *Store) Replace() (*Replacement, error) { return s.replace(false) }

// ReplaceKeepingMedia starts building a replacement of the snapshot and
// review log that keeps the live media files. PutMedia must not be called
// on it.
func (s *Store) ReplaceKeepingMedia() (*Replacement, error) { return s.replace(true) }

func (s *Store) replace(keepMedia bool) (*Replacement, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.replacing {
		return nil, ErrBusy
	}
	next := pointer{V: FormatVersion, Gen: s.ptr.Gen + 1, RL: s.ptr.RL + 1, Media: s.ptr.Media + 1, DL: s.ptr.DL + 1}
	prefixes := []string{fmt.Sprintf("g%d/", next.Gen), fmt.Sprintf("rl%d/", next.RL), fmt.Sprintf("dl%d/", next.DL)}
	if keepMedia {
		next.Media = s.ptr.Media
	} else {
		prefixes = append(prefixes, fmt.Sprintf("m%d/", next.Media))
	}
	// Start from empty sets, in case an earlier attempt left keys behind.
	for _, prefix := range prefixes {
		keys, err := s.b.List(prefix)
		if err != nil {
			return nil, err
		}
		for _, k := range keys {
			if err := s.b.Delete(k); err != nil {
				return nil, err
			}
		}
	}
	s.replacing = true
	return &Replacement{s: s, next: next}, nil
}

// PutMedia adds a media file to the replacement.
func (r *Replacement) PutMedia(name string, data []byte) error {
	return r.s.b.Put(mediaKey(r.next.Media, name), encodeMedia(name, data))
}

// AppendRevlog adds a review log entry to the replacement.
func (r *Replacement) AppendRevlog(month string, id int64, entry json.RawMessage) error {
	data, err := json.Marshal(segmentRecord{ID: id, Entry: entry})
	if err != nil {
		return err
	}
	return r.s.b.Append(segKey(r.next.RL, month), data)
}

// AppendDeletion adds a deletion log entry to the replacement.
func (r *Replacement) AppendDeletion(id int64, entry json.RawMessage) error {
	data, err := json.Marshal(Deletion{ID: id, Entry: entry})
	if err != nil {
		return err
	}
	return r.s.b.Append(delKey(r.next.DL), data)
}

// Commit writes the snapshot and makes the replacement live.
func (r *Replacement) Commit(snapshot []byte) error {
	if r.done {
		return errors.New("storage: replacement already finished")
	}
	s := r.s
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.b.Put(fmt.Sprintf("g%d/snapshot", r.next.Gen), snapshot); err != nil {
		return err
	}
	if err := s.flip(r.next); err != nil {
		return err
	}
	r.done = true
	s.replacing = false
	s.entries, s.bytes = 0, 0
	return nil
}

// Abort removes what the replacement wrote.
func (r *Replacement) Abort() {
	if r.done {
		return
	}
	r.done = true
	s := r.s
	s.mu.Lock()
	defer s.mu.Unlock()
	s.replacing = false
	_ = s.cleanup()
}
