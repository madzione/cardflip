//go:build !js

package storage_test

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"cardflip/golang/backup"
	"cardflip/golang/core"
	"cardflip/golang/storage"
)

func TestProfileDataDirs(t *testing.T) {
	dir := t.TempDir()
	if got := storage.ProfileDataDir(dir, storage.DefaultProfile); got != dir {
		t.Fatalf("default profile in %q, want the data directory itself", got)
	}
	ben := storage.ProfileDataDir(dir, "p2")
	if ben != filepath.Join(dir, "profiles", "p2") {
		t.Fatalf("profile p2 in %q", ben)
	}
}

func TestProfileRegistryFile(t *testing.T) {
	dir := t.TempDir()
	if _, err := storage.ReadProfiles(dir); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("read before write: %v", err)
	}
	if err := storage.WriteProfiles(dir, []byte(`{"a":1}`)); err != nil {
		t.Fatal(err)
	}
	if err := storage.WriteProfiles(dir, []byte(`{"a":2}`)); err != nil {
		t.Fatal(err)
	}
	got, err := storage.ReadProfiles(dir)
	if err != nil || string(got) != `{"a":2}` {
		t.Fatalf("read = %q, %v", got, err)
	}
}

// putIn stores a value in the backend of profile id under dir.
func putIn(t *testing.T, dir, id string) {
	t.Helper()
	b, err := storage.OpenBackend(storage.ProfileDataDir(dir, id))
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	if err := b.Put("gen-1/snapshot", []byte(id)); err != nil {
		t.Fatal(err)
	}
	if err := b.Backups().Write("backup-"+id+".cardflip", []byte(id)); err != nil {
		t.Fatal(err)
	}
}

func keysIn(t *testing.T, dir, id string) []string {
	t.Helper()
	b, err := storage.OpenBackend(storage.ProfileDataDir(dir, id))
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	keys, err := b.List("")
	if err != nil {
		t.Fatal(err)
	}
	return keys
}

func TestRemoveProfileData(t *testing.T) {
	dir := t.TempDir()
	putIn(t, dir, storage.DefaultProfile)
	putIn(t, dir, "p2")
	putIn(t, dir, "p3")
	if err := storage.WriteProfiles(dir, []byte("{}")); err != nil {
		t.Fatal(err)
	}

	// Removing another profile deletes its directory and nothing else.
	check(t, storage.RemoveProfileData(dir, "p2"))
	if _, err := os.Stat(storage.ProfileDataDir(dir, "p2")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("p2 still there: %v", err)
	}
	if len(keysIn(t, dir, storage.DefaultProfile)) != 1 || len(keysIn(t, dir, "p3")) != 1 {
		t.Fatal("other profiles lost data")
	}

	// Removing the default profile keeps the registry and other profiles.
	check(t, storage.RemoveProfileData(dir, storage.DefaultProfile))
	if keys := keysIn(t, dir, storage.DefaultProfile); len(keys) != 0 {
		t.Fatalf("default keys left: %v", keys)
	}
	if _, err := os.Stat(filepath.Join(dir, "backups", "backup-default.cardflip")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("default backup left: %v", err)
	}
	if _, err := storage.ReadProfiles(dir); err != nil {
		t.Fatalf("registry gone: %v", err)
	}
	if len(keysIn(t, dir, "p3")) != 1 {
		t.Fatal("p3 lost data")
	}
}

func check(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestMcpSettingsFile(t *testing.T) {
	dir := t.TempDir()
	if _, err := storage.ReadMcp(dir); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("read before write: %v", err)
	}
	if err := storage.WriteMcp(dir, []byte(`{"enabled":true}`)); err != nil {
		t.Fatal(err)
	}
	if err := storage.WriteMcp(dir, []byte(`{"enabled":false}`)); err != nil {
		t.Fatal(err)
	}
	got, err := storage.ReadMcp(dir)
	if err != nil || string(got) != `{"enabled":false}` {
		t.Fatalf("read = %q, %v", got, err)
	}
	// The file holds the token: only its owner may read it.
	fi, err := os.Stat(filepath.Join(dir, "mcp.json"))
	if err != nil {
		t.Fatal(err)
	}
	if perm := fi.Mode().Perm(); runtime.GOOS != "windows" && perm != 0o600 {
		t.Fatalf("mode = %v", perm)
	}
	// The setting is not the profile registry.
	if _, err := storage.ReadProfiles(dir); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("profiles: %v", err)
	}
}

func TestSyncFiles(t *testing.T) {
	dir := t.TempDir()
	if _, err := storage.ReadSync(dir); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("read before write: %v", err)
	}
	if _, err := storage.ReadSyncLinks(dir, "p2"); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("read links before write: %v", err)
	}
	for _, data := range []string{`{"name":"a"}`, `{"name":"b"}`} {
		if err := storage.WriteSync(dir, []byte(data)); err != nil {
			t.Fatal(err)
		}
	}
	if got, err := storage.ReadSync(dir); err != nil || string(got) != `{"name":"b"}` {
		t.Fatalf("read = %q, %v", got, err)
	}
	for _, id := range []string{storage.DefaultProfile, "p2"} {
		if err := storage.WriteSyncLinks(dir, id, []byte(`["`+id+`"]`)); err != nil {
			t.Fatal(err)
		}
		if got, err := storage.ReadSyncLinks(dir, id); err != nil || string(got) != `["`+id+`"]` {
			t.Fatalf("links of %s = %q, %v", id, got, err)
		}
	}
	// Only the owner can read them: sync.json holds the device's key.
	for _, p := range []string{filepath.Join(dir, "sync.json"), filepath.Join(dir, "sync-links.json"),
		filepath.Join(dir, "profiles", "p2", "sync-links.json")} {
		fi, err := os.Stat(p)
		if err != nil {
			t.Fatal(err)
		}
		if runtime.GOOS != "windows" && fi.Mode().Perm() != 0o600 {
			t.Errorf("%s has mode %v, want 0600", p, fi.Mode().Perm())
		}
	}
	// They are outside the collection backends.
	putIn(t, dir, storage.DefaultProfile)
	putIn(t, dir, "p2")
	for _, id := range []string{storage.DefaultProfile, "p2"} {
		for _, k := range keysIn(t, dir, id) {
			if strings.Contains(k, "sync") {
				t.Errorf("backend of %s holds %s", id, k)
			}
		}
	}
	// Removing a profile removes its links, but not the device's state.
	for _, id := range []string{storage.DefaultProfile, "p2"} {
		if err := storage.RemoveProfileData(dir, id); err != nil {
			t.Fatal(err)
		}
		if _, err := storage.ReadSyncLinks(dir, id); !errors.Is(err, storage.ErrNotFound) {
			t.Errorf("links of removed profile %s: %v", id, err)
		}
	}
	if _, err := storage.ReadSync(dir); err != nil {
		t.Fatalf("device state after removing profiles: %v", err)
	}
}

// A backup of a linked profile holds neither the device's sync state nor
// the profile's links (device-pairing spec, Device identity).
func TestBackupLeavesOutSyncFiles(t *testing.T) {
	dir := t.TempDir()
	const secret = "PRIVATE-KEY-MARKER"
	if err := storage.WriteSync(dir, []byte(`{"key":"`+secret+`"}`)); err != nil {
		t.Fatal(err)
	}
	if err := storage.WriteSyncLinks(dir, storage.DefaultProfile, []byte(`[{"id":"`+secret+`"}]`)); err != nil {
		t.Fatal(err)
	}
	b, err := storage.OpenBackend(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	c, _, err := core.Open(b, core.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	for _, export := range []func(backup.Source, string, time.Time, func()) ([]byte, error){backup.Export, backup.ExportCollection} {
		data, err := export(c, "test", time.Now(), nil)
		if err != nil {
			t.Fatal(err)
		}
		zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range zr.File {
			if strings.Contains(f.Name, "sync") {
				t.Errorf("backup holds %s", f.Name)
			}
			rc, _ := f.Open()
			content, _ := io.ReadAll(rc)
			rc.Close()
			if bytes.Contains(content, []byte(secret)) {
				t.Errorf("backup entry %s holds sync state", f.Name)
			}
		}
	}
}
