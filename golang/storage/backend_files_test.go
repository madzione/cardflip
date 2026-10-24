//go:build !js

package storage_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cardflip/golang/storage"
	"cardflip/golang/storage/storagetest"
)

func runContract(t *testing.T, newStore func(t *testing.T) storagetest.Opener) {
	for _, c := range storagetest.Checks() {
		t.Run(c.Name, func(t *testing.T) {
			if err := c.Run(newStore(t)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestFilesContract(t *testing.T) {
	runContract(t, func(t *testing.T) storagetest.Opener {
		dir, backups, trash := t.TempDir(), t.TempDir(), t.TempDir()
		return func() (storage.Backend, error) { return storage.OpenFiles(dir, backups, trash) }
	})
}

func TestMemoryContract(t *testing.T) {
	runContract(t, func(t *testing.T) storagetest.Opener {
		s := storage.NewMemoryStore()
		return func() (storage.Backend, error) { return s.Open(), nil }
	})
}

// A crash during Put leaves a temp file next to the old value: the old value
// is still readable, and the next open removes the temp file.
func TestFilesAtomicReplace(t *testing.T) {
	dir := t.TempDir()
	b, _ := storage.OpenFiles(dir, filepath.Join(t.TempDir(), "backups"), filepath.Join(t.TempDir(), "trash"))
	if err := b.Put("gen-1/snapshot", []byte("old snapshot")); err != nil {
		t.Fatal(err)
	}
	b.Close()
	// Simulate the half-written temp file of a Put killed before rename.
	tmp := filepath.Join(dir, "gen-1", ".tmp-123")
	if err := os.WriteFile(tmp, []byte("new snap"), 0o644); err != nil {
		t.Fatal(err)
	}

	b, _ = storage.OpenFiles(dir, filepath.Join(t.TempDir(), "backups"), filepath.Join(t.TempDir(), "trash"))
	defer b.Close()
	got, err := b.Get("gen-1/snapshot")
	if err != nil || string(got) != "old snapshot" {
		t.Fatalf("Get = %q, %v", got, err)
	}
	if _, err := os.Stat(tmp); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temp file not cleaned up: %v", err)
	}
	keys, _ := b.List("")
	if strings.Join(keys, ",") != "gen-1/snapshot" {
		t.Fatalf("List = %v", keys)
	}
	if err := b.Put("gen-1/snapshot", []byte("new snapshot")); err != nil {
		t.Fatal(err)
	}
	if got, _ := b.Get("gen-1/snapshot"); string(got) != "new snapshot" {
		t.Fatalf("Get after replace = %q", got)
	}
}

func journalFile(dir string) string { return filepath.Join(dir, "gen-1", "journal") }

func writeJournal(t *testing.T, dir string, recs ...string) {
	t.Helper()
	b, _ := storage.OpenFiles(dir, filepath.Join(t.TempDir(), "backups"), filepath.Join(t.TempDir(), "trash"))
	defer b.Close()
	for _, r := range recs {
		if err := b.Append("gen-1/journal", []byte(r)); err != nil {
			t.Fatal(err)
		}
	}
}

func records(t *testing.T, b storage.Backend) []string {
	t.Helper()
	recs, err := b.Records("gen-1/journal")
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, r := range recs {
		out = append(out, string(r))
	}
	return out
}

// A crash during Append leaves a partial last record. It is dropped when
// reading, and cut before the next append so new records stay readable.
func TestFilesTornTrailingRecordDropped(t *testing.T) {
	for _, cut := range []int{1, 4, 8, 12} {
		dir := t.TempDir()
		writeJournal(t, dir, "first", "second", "third record")
		data, _ := os.ReadFile(journalFile(dir))
		os.WriteFile(journalFile(dir), data[:len(data)-cut], 0o644)

		b, _ := storage.OpenFiles(dir, filepath.Join(t.TempDir(), "backups"), filepath.Join(t.TempDir(), "trash"))
		if got := records(t, b); strings.Join(got, ",") != "first,second" {
			t.Fatalf("cut %d: records = %q", cut, got)
		}
		if err := b.Append("gen-1/journal", []byte("fourth")); err != nil {
			t.Fatal(err)
		}
		b.Close()

		b, _ = storage.OpenFiles(dir, filepath.Join(t.TempDir(), "backups"), filepath.Join(t.TempDir(), "trash"))
		if got := records(t, b); strings.Join(got, ",") != "first,second,fourth" {
			t.Fatalf("cut %d: after append records = %q", cut, got)
		}
		b.Close()
	}
}

// A last record whose bytes were not all flushed (garbage in place of data)
// fails its checksum and counts as torn.
func TestFilesTornRecordBadChecksum(t *testing.T) {
	dir := t.TempDir()
	writeJournal(t, dir, "first", "second")
	data, _ := os.ReadFile(journalFile(dir))
	data[len(data)-1] ^= 0xff
	os.WriteFile(journalFile(dir), data, 0o644)

	b, _ := storage.OpenFiles(dir, filepath.Join(t.TempDir(), "backups"), filepath.Join(t.TempDir(), "trash"))
	defer b.Close()
	if got := records(t, b); strings.Join(got, ",") != "first" {
		t.Fatalf("records = %q", got)
	}
}

// Damage before the last record is not a torn write: it is reported, and the
// file is left untouched.
func TestFilesCorruptMiddleRecordReported(t *testing.T) {
	dir := t.TempDir()
	writeJournal(t, dir, "first", "second", "third")
	data, _ := os.ReadFile(journalFile(dir))
	data[9] ^= 0xff // inside "first"
	os.WriteFile(journalFile(dir), data, 0o644)

	b, _ := storage.OpenFiles(dir, filepath.Join(t.TempDir(), "backups"), filepath.Join(t.TempDir(), "trash"))
	defer b.Close()
	if _, err := b.Records("gen-1/journal"); !errors.Is(err, storage.ErrCorrupt) {
		t.Fatalf("Records err = %v, want ErrCorrupt", err)
	}
	if err := b.Append("gen-1/journal", []byte("x")); !errors.Is(err, storage.ErrCorrupt) {
		t.Fatalf("Append err = %v, want ErrCorrupt", err)
	}
	after, _ := os.ReadFile(journalFile(dir))
	if !bytes.Equal(after, data) {
		t.Fatal("corrupt journal was modified")
	}
}

func TestFilesEscapedNamesOnDisk(t *testing.T) {
	dir := t.TempDir()
	b, _ := storage.OpenFiles(dir, filepath.Join(t.TempDir(), "backups"), filepath.Join(t.TempDir(), "trash"))
	defer b.Close()
	if err := b.Put("media/Cat.JPG", []byte("x")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "media", "%43at.%4A%50%47")); err != nil {
		t.Fatal(err)
	}
}

func TestFilesDeleteRemovesEmptyDirectories(t *testing.T) {
	dir := t.TempDir()
	b, _ := storage.OpenFiles(dir, filepath.Join(t.TempDir(), "backups"), filepath.Join(t.TempDir(), "trash"))
	defer b.Close()
	b.Put("g1/snapshot", []byte("x"))
	b.Append("g1/journal", []byte("x"))
	b.Put("g2/snapshot", []byte("y"))
	b.Delete("g1/snapshot")
	if _, err := os.Stat(filepath.Join(dir, "g1")); err != nil {
		t.Fatal("directory removed while not empty")
	}
	b.Delete("g1/journal")
	if _, err := os.Stat(filepath.Join(dir, "g1")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("empty directory kept: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatal("root removed")
	}
}
