//go:build !js

package storage

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// Files is the native Backend: one file per key under a root directory,
// and backups and trashed media as files in separate directories.
type Files struct {
	root      string
	backupDir string
	trashDir  string

	mu   sync.Mutex
	logs map[string]*os.File // open append handles, torn tail already cut
}

// OpenFiles opens (creating if needed) a file backend rooted at dir that
// keeps backups in backupDir and trashed media in trashDir, both outside dir.
func OpenFiles(dir, backupDir, trashDir string) (*Files, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	b := &Files{root: dir, backupDir: backupDir, trashDir: trashDir, logs: map[string]*os.File{}}
	// Leftover temp files are from a write that crashed before its rename.
	for _, d := range []string{dir, backupDir, trashDir} {
		_ = filepath.WalkDir(d, func(p string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasPrefix(d.Name(), ".tmp-") {
				os.Remove(p)
			}
			return nil
		})
	}
	return b, nil
}

// OpenBackend opens the platform backend for dataDir: the collection in
// dataDir/collection, backups in dataDir/backups, trashed media in
// dataDir/media.trash.
func OpenBackend(dataDir string) (Backend, error) {
	if dataDir == "" {
		return nil, errors.New("storage: no data directory")
	}
	return OpenFiles(filepath.Join(dataDir, "collection"), filepath.Join(dataDir, "backups"),
		filepath.Join(dataDir, "media.trash"))
}

func (b *Files) path(key string) (string, error) {
	if err := ValidKey(key); err != nil {
		return "", err
	}
	segs := strings.Split(key, "/")
	for i, s := range segs {
		segs[i] = escapeSegment(s)
	}
	return filepath.Join(b.root, filepath.Join(segs...)), nil
}

func (b *Files) Get(key string) ([]byte, error) {
	p, err := b.path(key)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNotFound
	}
	return data, err
}

func (b *Files) Put(key string, value []byte) error {
	p, err := b.path(key)
	if err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closeLog(key)
	return writeAtomic(p, value)
}

// writeAtomic replaces the file p through a synced temp file and a rename.
func writeAtomic(p string, value []byte) error {
	dir := filepath.Dir(p)
	if err := mkdirs(dir); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		if !ok {
			tmp.Close()
			os.Remove(tmp.Name())
		}
	}()
	if _, err := tmp.Write(value); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp.Name(), p); err != nil {
		return err
	}
	ok = true
	return syncDir(dir)
}

func (b *Files) Append(key string, record []byte) error {
	return b.AppendAll(key, [][]byte{record})
}

// AppendAll appends records with one write and one sync.
func (b *Files) AppendAll(key string, records [][]byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	f, err := b.openLog(key)
	if err != nil {
		return err
	}
	var data []byte
	for _, r := range records {
		data = append(data, frame(r)...)
	}
	if _, err := f.Write(data); err != nil {
		b.closeLog(key)
		return err
	}
	return f.Sync()
}

// openLog returns the append handle for key, cutting a torn trailing record
// left by a crash so that new records follow the last complete one.
func (b *Files) openLog(key string) (*os.File, error) {
	if f := b.logs[key]; f != nil {
		return f, nil
	}
	p, err := b.path(key)
	if err != nil {
		return nil, err
	}
	dir := filepath.Dir(p)
	if err := mkdirs(dir); err != nil {
		return nil, err
	}
	_, statErr := os.Stat(p)
	created := errors.Is(statErr, fs.ErrNotExist)
	f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		f.Close()
		return nil, err
	}
	_, valid, err := unframe(data)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("%s: %w", key, err)
	}
	if valid < len(data) {
		if err := f.Truncate(int64(valid)); err != nil {
			f.Close()
			return nil, err
		}
	}
	if _, err := f.Seek(int64(valid), 0); err != nil {
		f.Close()
		return nil, err
	}
	if created {
		if err := syncDir(dir); err != nil {
			f.Close()
			return nil, err
		}
	}
	b.logs[key] = f
	return f, nil
}

func (b *Files) closeLog(key string) {
	if f := b.logs[key]; f != nil {
		f.Close()
		delete(b.logs, key)
	}
}

func (b *Files) Records(key string) ([][]byte, error) {
	data, err := b.Get(key)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	recs, _, err := unframe(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", key, err)
	}
	return recs, nil
}

func (b *Files) List(prefix string) ([]string, error) {
	var keys []string
	err := filepath.WalkDir(b.root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		rel, err := filepath.Rel(b.root, p)
		if err != nil {
			return err
		}
		segs := strings.Split(filepath.ToSlash(rel), "/")
		for i, s := range segs {
			u, ok := unescapeSegment(s)
			if !ok {
				return nil // not ours
			}
			segs[i] = u
		}
		if key := strings.Join(segs, "/"); strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
		return nil
	})
	sort.Strings(keys)
	return keys, err
}

func (b *Files) Delete(key string) error {
	p, err := b.path(key)
	if err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closeLog(key)
	if err := os.Remove(p); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	// Remove directories left empty, e.g. of a replaced generation.
	for dir := filepath.Dir(p); dir != b.root && strings.HasPrefix(dir, b.root); dir = filepath.Dir(dir) {
		if os.Remove(dir) != nil {
			break
		}
	}
	return nil
}

func (b *Files) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	for k := range b.logs {
		b.closeLog(k)
	}
	return nil
}

// mkdirs creates dir and syncs the parents of every directory it created,
// so that the new directory entries survive a power loss.
func mkdirs(dir string) error {
	if _, err := os.Stat(dir); err == nil {
		return nil
	}
	if err := mkdirs(filepath.Dir(dir)); err != nil {
		return err
	}
	if err := os.Mkdir(dir, 0o755); err != nil && !errors.Is(err, fs.ErrExist) {
		return err
	}
	return syncDir(filepath.Dir(dir))
}

// syncDir makes renames and new entries in dir durable. Windows cannot open
// directories for syncing; NTFS journals metadata itself.
func syncDir(dir string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	if err := d.Sync(); err != nil && !errors.Is(err, os.ErrInvalid) {
		return err
	}
	return nil
}

// escapeSegment maps a key segment to a file name that is valid and distinct
// on every file system, including case-insensitive ones: bytes other than
// [a-z0-9._-] become %XX.
func escapeSegment(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '.' || c == '_' || c == '-' {
			sb.WriteByte(c)
		} else {
			fmt.Fprintf(&sb, "%%%02X", c)
		}
	}
	return sb.String()
}

func unescapeSegment(s string) (string, bool) {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '%' {
			sb.WriteByte(s[i])
			continue
		}
		if i+2 >= len(s) {
			return "", false
		}
		var c byte
		if _, err := fmt.Sscanf(s[i+1:i+3], "%02X", &c); err != nil {
			return "", false
		}
		sb.WriteByte(c)
		i += 2
	}
	return sb.String(), true
}

// RemoveBackend deletes everything OpenBackend(dataDir) stored, backups
// and trashed media included.
func RemoveBackend(dataDir string) error {
	if dataDir == "" {
		return errors.New("storage: no data directory")
	}
	for _, d := range []string{"collection", "backups", "media.trash"} {
		if err := os.RemoveAll(filepath.Join(dataDir, d)); err != nil {
			return err
		}
	}
	return nil
}

func (b *Files) Backups() BackupStore { return filesBackups{b} }

// filesBackups stores each backup as a file named after it. Backup names
// need no escaping: ValidBackupName only allows portable file names.
type filesBackups struct{ b *Files }

func (fb filesBackups) path(name string) (string, error) {
	if err := ValidBackupName(name); err != nil {
		return "", err
	}
	return filepath.Join(fb.b.backupDir, name), nil
}

func (fb filesBackups) List() ([]BackupFile, error) {
	entries, err := os.ReadDir(fb.b.backupDir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []BackupFile
	for _, e := range entries {
		if e.IsDir() || ValidBackupName(e.Name()) != nil {
			continue
		}
		info, err := e.Info()
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue // deleted meanwhile
			}
			return nil, err
		}
		out = append(out, BackupFile{Name: e.Name(), Size: info.Size()})
	}
	return out, nil // ReadDir sorts by name
}

func (fb filesBackups) Write(name string, data []byte) error {
	p, err := fb.path(name)
	if err != nil {
		return err
	}
	fb.b.mu.Lock()
	defer fb.b.mu.Unlock()
	return writeAtomic(p, data)
}

func (fb filesBackups) Read(name string) ([]byte, error) {
	p, err := fb.path(name)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNotFound
	}
	return data, err
}

func (fb filesBackups) Delete(name string) error {
	p, err := fb.path(name)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

func (b *Files) Trash() MediaTrash { return filesTrash{b} }

// filesTrash stores each trashed file as a file named after the hash of its
// name, holding the encoded name and time before the data.
type filesTrash struct{ b *Files }

func (ft filesTrash) path(name string) (string, error) {
	if err := validTrashName(name); err != nil {
		return "", err
	}
	return filepath.Join(ft.b.trashDir, trashFileName(name)), nil
}

func (ft filesTrash) List() ([]TrashedMedia, error) {
	entries, err := os.ReadDir(ft.b.trashDir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []TrashedMedia
	for _, e := range entries {
		if e.IsDir() || len(e.Name()) != 32 || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		t, err := readTrashedHeader(filepath.Join(ft.b.trashDir, e.Name()))
		if errors.Is(err, fs.ErrNotExist) {
			continue // deleted meanwhile
		}
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// readTrashedHeader reads the name and time of the trashed file at p
// without reading its data.
func readTrashedHeader(p string) (TrashedMedia, error) {
	f, err := os.Open(p)
	if err != nil {
		return TrashedMedia{}, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return TrashedMedia{}, err
	}
	// The header is the name length (at most 10 bytes), the name and 8
	// bytes of time; read more only for long names.
	head := make([]byte, min(info.Size(), 512))
	if _, err := io.ReadFull(f, head); err != nil {
		return TrashedMedia{}, err
	}
	if l, k := binary.Uvarint(head); k > 0 && l < uint64(info.Size()) {
		if need := int64(k) + int64(l) + 8; need > int64(len(head)) && need <= info.Size() {
			rest := make([]byte, need-int64(len(head)))
			if _, err := io.ReadFull(f, rest); err != nil {
				return TrashedMedia{}, err
			}
			head = append(head, rest...)
		}
	}
	name, trashed, n, err := decodeTrashedHeader(head)
	if err != nil {
		return TrashedMedia{}, fmt.Errorf("%s: %w", filepath.Base(p), err)
	}
	return TrashedMedia{Name: name, Size: info.Size() - int64(n), Trashed: trashed}, nil
}

func (ft filesTrash) Put(name string, data []byte, trashed time.Time) error {
	p, err := ft.path(name)
	if err != nil {
		return err
	}
	ft.b.mu.Lock()
	defer ft.b.mu.Unlock()
	return writeAtomic(p, encodeTrashed(name, data, trashed))
}

func (ft filesTrash) Get(name string) ([]byte, error) {
	p, err := ft.path(name)
	if err != nil {
		return nil, err
	}
	v, err := os.ReadFile(p)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	stored, _, n, err := decodeTrashedHeader(v)
	if err != nil {
		return nil, err
	}
	if stored != name {
		return nil, ErrNotFound
	}
	return v[n:], nil
}

func (ft filesTrash) Delete(name string) error {
	p, err := ft.path(name)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}
