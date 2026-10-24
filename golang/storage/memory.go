package storage

import (
	"bytes"
	"sort"
	"strings"
	"sync"
	"time"
)

// Memory is an in-memory Backend for tests. Handles from the same MemoryStore
// share data, which models closing and reopening the app.
type Memory struct {
	store *MemoryStore
}

// MemoryStore is the data behind Memory backends.
type MemoryStore struct {
	mu     sync.Mutex
	values map[string][]byte
	logs   map[string][][]byte
	// backups holds the backup store's files, apart from the keys.
	backups map[string][]byte
	// trash holds the trashed media files, apart from the keys.
	trash map[string]memoryTrashed
	// FailPut, when set, makes Put fail for keys it returns true for,
	// backup writes for "backups/<name>" and trash writes for
	// "trash/<name>". Tests use it to simulate a full
	// disk.
	FailPut func(key string) bool
}

// NewMemoryStore returns an empty store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{values: map[string][]byte{}, logs: map[string][][]byte{}, backups: map[string][]byte{},
		trash: map[string]memoryTrashed{}}
}

// Open returns a backend on the store.
func (s *MemoryStore) Open() *Memory { return &Memory{store: s} }

// Clone returns a copy of the store, as the data a crash at this moment
// would leave behind.
func (s *MemoryStore) Clone() *MemoryStore {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := NewMemoryStore()
	for k, v := range s.values {
		out.values[k] = bytes.Clone(v)
	}
	for k, recs := range s.logs {
		for _, r := range recs {
			out.logs[k] = append(out.logs[k], bytes.Clone(r))
		}
	}
	for k, v := range s.backups {
		out.backups[k] = bytes.Clone(v)
	}
	for k, v := range s.trash {
		out.trash[k] = memoryTrashed{data: bytes.Clone(v.data), trashed: v.trashed}
	}
	return out
}

// NewMemory returns a backend on a fresh store.
func NewMemory() *Memory { return NewMemoryStore().Open() }

func (m *Memory) Get(key string) ([]byte, error) {
	if err := ValidKey(key); err != nil {
		return nil, err
	}
	s := m.store
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.values[key]
	if !ok {
		return nil, ErrNotFound
	}
	return bytes.Clone(v), nil
}

func (m *Memory) Put(key string, value []byte) error {
	if err := ValidKey(key); err != nil {
		return err
	}
	s := m.store
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.FailPut != nil && s.FailPut(key) {
		return errSimulated
	}
	delete(s.logs, key)
	s.values[key] = bytes.Clone(value)
	return nil
}

func (m *Memory) Append(key string, record []byte) error {
	return m.AppendAll(key, [][]byte{record})
}

func (m *Memory) AppendAll(key string, records [][]byte) error {
	if err := ValidKey(key); err != nil {
		return err
	}
	s := m.store
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range records {
		s.logs[key] = append(s.logs[key], bytes.Clone(r))
	}
	return nil
}

func (m *Memory) Records(key string) ([][]byte, error) {
	if err := ValidKey(key); err != nil {
		return nil, err
	}
	s := m.store
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([][]byte, len(s.logs[key]))
	for i, r := range s.logs[key] {
		out[i] = bytes.Clone(r)
	}
	return out, nil
}

func (m *Memory) List(prefix string) ([]string, error) {
	s := m.store
	s.mu.Lock()
	defer s.mu.Unlock()
	var keys []string
	for k := range s.values {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	for k := range s.logs {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys, nil
}

func (m *Memory) Delete(key string) error {
	if err := ValidKey(key); err != nil {
		return err
	}
	s := m.store
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.values, key)
	delete(s.logs, key)
	return nil
}

func (m *Memory) Close() error { return nil }

func (m *Memory) Backups() BackupStore { return memoryBackups{m.store} }

type memoryBackups struct{ s *MemoryStore }

func (b memoryBackups) List() ([]BackupFile, error) {
	b.s.mu.Lock()
	defer b.s.mu.Unlock()
	out := make([]BackupFile, 0, len(b.s.backups))
	for name, data := range b.s.backups {
		out = append(out, BackupFile{Name: name, Size: int64(len(data))})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (b memoryBackups) Write(name string, data []byte) error {
	if err := ValidBackupName(name); err != nil {
		return err
	}
	b.s.mu.Lock()
	defer b.s.mu.Unlock()
	if b.s.FailPut != nil && b.s.FailPut("backups/"+name) {
		return errSimulated
	}
	b.s.backups[name] = bytes.Clone(data)
	return nil
}

func (b memoryBackups) Read(name string) ([]byte, error) {
	if err := ValidBackupName(name); err != nil {
		return nil, err
	}
	b.s.mu.Lock()
	defer b.s.mu.Unlock()
	data, ok := b.s.backups[name]
	if !ok {
		return nil, ErrNotFound
	}
	return bytes.Clone(data), nil
}

func (b memoryBackups) Delete(name string) error {
	if err := ValidBackupName(name); err != nil {
		return err
	}
	b.s.mu.Lock()
	defer b.s.mu.Unlock()
	delete(b.s.backups, name)
	return nil
}

func (m *Memory) Trash() MediaTrash { return memoryTrash{m.store} }

type memoryTrashed struct {
	data    []byte
	trashed time.Time
}

type memoryTrash struct{ s *MemoryStore }

func (t memoryTrash) List() ([]TrashedMedia, error) {
	t.s.mu.Lock()
	defer t.s.mu.Unlock()
	out := make([]TrashedMedia, 0, len(t.s.trash))
	for name, v := range t.s.trash {
		out = append(out, TrashedMedia{Name: name, Size: int64(len(v.data)), Trashed: v.trashed})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (t memoryTrash) Put(name string, data []byte, trashed time.Time) error {
	if err := validTrashName(name); err != nil {
		return err
	}
	t.s.mu.Lock()
	defer t.s.mu.Unlock()
	if t.s.FailPut != nil && t.s.FailPut("trash/"+name) {
		return errSimulated
	}
	t.s.trash[name] = memoryTrashed{data: bytes.Clone(data), trashed: time.UnixMilli(trashed.UnixMilli())}
	return nil
}

func (t memoryTrash) Get(name string) ([]byte, error) {
	if err := validTrashName(name); err != nil {
		return nil, err
	}
	t.s.mu.Lock()
	defer t.s.mu.Unlock()
	v, ok := t.s.trash[name]
	if !ok {
		return nil, ErrNotFound
	}
	return bytes.Clone(v.data), nil
}

func (t memoryTrash) Delete(name string) error {
	if err := validTrashName(name); err != nil {
		return err
	}
	t.s.mu.Lock()
	defer t.s.mu.Unlock()
	delete(t.s.trash, name)
	return nil
}

type simulatedError struct{}

func (simulatedError) Error() string { return "storage: simulated write failure" }

var errSimulated error = simulatedError{}
