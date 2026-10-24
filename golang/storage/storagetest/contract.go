// Package storagetest is the behaviour contract every storage.Backend must
// meet. It does not depend on package testing, so the web build can run it
// against IndexedDB from an integration test.
package storagetest

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"cardflip/golang/storage"
)

// Opener opens a backend on the same underlying storage each time it is
// called, like the app opening its data directory after a restart.
type Opener func() (storage.Backend, error)

// Check is one contract check, run against a fresh, empty store.
type Check struct {
	Name string
	Run  func(open Opener) error
}

// Checks returns the contract checks.
func Checks() []Check {
	return []Check{
		{"get missing key", checkGetMissing},
		{"put and get", checkPutGet},
		{"put replaces", checkPutReplaces},
		{"append and records", checkAppendRecords},
		{"append several records at once", checkAppendAll},
		{"records of missing log", checkRecordsMissing},
		{"list by prefix", checkList},
		{"delete", checkDelete},
		{"keys are case- and byte-exact", checkExactKeys},
		{"invalid keys rejected", checkInvalidKeys},
		{"data survives reopen", checkReopen},
		{"put over a log", checkPutOverLog},
		{"backups: write, read, list, delete", checkBackups},
		{"backups: write replaces", checkBackupReplaces},
		{"backups: apart from keys", checkBackupsApart},
		{"backups: survive reopen", checkBackupsReopen},
		{"backups: invalid names rejected", checkBackupInvalidNames},
		{"trash: put, get, list, delete", checkTrash},
		{"trash: put replaces", checkTrashReplaces},
		{"trash: names are case- and byte-exact", checkTrashExactNames},
		{"trash: apart from keys and backups", checkTrashApart},
		{"trash: survives reopen", checkTrashReopen},
	}
}

// RunAll runs every check, each with a fresh store from newStore, and
// returns one error per failed check.
func RunAll(newStore func() (Opener, error)) []error {
	var errs []error
	for _, c := range Checks() {
		open, err := newStore()
		if err == nil {
			err = c.Run(open)
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", c.Name, err))
		}
	}
	return errs
}

func with(open Opener, fn func(b storage.Backend) error) error {
	b, err := open()
	if err != nil {
		return err
	}
	defer b.Close()
	return fn(b)
}

func expectValue(b storage.Backend, key string, want []byte) error {
	got, err := b.Get(key)
	if err != nil {
		return fmt.Errorf("Get(%q): %w", key, err)
	}
	if !bytes.Equal(got, want) {
		return fmt.Errorf("Get(%q) = %q, want %q", key, got, want)
	}
	return nil
}

func expectRecords(b storage.Backend, key string, want ...string) error {
	recs, err := b.Records(key)
	if err != nil {
		return fmt.Errorf("Records(%q): %w", key, err)
	}
	got := make([]string, len(recs))
	for i, r := range recs {
		got[i] = string(r)
	}
	if strings.Join(got, "|") != strings.Join(want, "|") || len(got) != len(want) {
		return fmt.Errorf("Records(%q) = %q, want %q", key, got, want)
	}
	return nil
}

func checkGetMissing(open Opener) error {
	return with(open, func(b storage.Backend) error {
		if _, err := b.Get("nothing/here"); !errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("err = %v, want ErrNotFound", err)
		}
		return nil
	})
}

func checkPutGet(open Opener) error {
	return with(open, func(b storage.Backend) error {
		big := bytes.Repeat([]byte("0123456789abcdef"), 1<<16) // 1 MiB
		for key, v := range map[string][]byte{"current": []byte("gen-1"), "gen-1/snapshot": big, "empty": {}} {
			if err := b.Put(key, v); err != nil {
				return err
			}
			if err := expectValue(b, key, v); err != nil {
				return err
			}
		}
		return nil
	})
}

func checkPutReplaces(open Opener) error {
	return with(open, func(b storage.Backend) error {
		if err := b.Put("k", []byte("old value, longer")); err != nil {
			return err
		}
		if err := b.Put("k", []byte("new")); err != nil {
			return err
		}
		return expectValue(b, "k", []byte("new"))
	})
}

func checkAppendRecords(open Opener) error {
	return with(open, func(b storage.Backend) error {
		for _, r := range []string{"one", "", "three"} {
			if err := b.Append("gen-1/journal", []byte(r)); err != nil {
				return err
			}
		}
		return expectRecords(b, "gen-1/journal", "one", "", "three")
	})
}

func checkAppendAll(open Opener) error {
	err := with(open, func(b storage.Backend) error {
		ba, ok := b.(storage.AppendAller)
		if !ok {
			return nil
		}
		if err := b.Append("rl1/2026-01", []byte("one")); err != nil {
			return err
		}
		if err := ba.AppendAll("rl1/2026-01", [][]byte{[]byte("two"), {}, []byte("four")}); err != nil {
			return err
		}
		if err := b.Append("rl1/2026-01", []byte("five")); err != nil {
			return err
		}
		return expectRecords(b, "rl1/2026-01", "one", "two", "", "four", "five")
	})
	if err != nil {
		return err
	}
	return with(open, func(b storage.Backend) error {
		if _, ok := b.(storage.AppendAller); !ok {
			return nil
		}
		return expectRecords(b, "rl1/2026-01", "one", "two", "", "four", "five")
	})
}

func checkRecordsMissing(open Opener) error {
	return with(open, func(b storage.Backend) error {
		return expectRecords(b, "no/log")
	})
}

func checkList(open Opener) error {
	return with(open, func(b storage.Backend) error {
		for _, k := range []string{"gen-2/snapshot", "gen-1/snapshot", "media/x", "gen-10/snapshot"} {
			if err := b.Put(k, []byte(k)); err != nil {
				return err
			}
		}
		if err := b.Append("gen-1/journal", []byte("r")); err != nil {
			return err
		}
		got, err := b.List("gen-1/")
		if err != nil {
			return err
		}
		if strings.Join(got, ",") != "gen-1/journal,gen-1/snapshot" {
			return fmt.Errorf("List(gen-1/) = %v", got)
		}
		all, err := b.List("")
		if err != nil {
			return err
		}
		if len(all) != 5 {
			return fmt.Errorf("List('') = %v", all)
		}
		return nil
	})
}

func checkDelete(open Opener) error {
	return with(open, func(b storage.Backend) error {
		if err := b.Put("a", []byte("1")); err != nil {
			return err
		}
		if err := b.Append("log", []byte("1")); err != nil {
			return err
		}
		for _, k := range []string{"a", "log", "never-existed"} {
			if err := b.Delete(k); err != nil {
				return fmt.Errorf("Delete(%q): %w", k, err)
			}
		}
		if _, err := b.Get("a"); !errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("Get after Delete: %v", err)
		}
		if err := expectRecords(b, "log"); err != nil {
			return err
		}
		keys, err := b.List("")
		if err != nil || len(keys) != 0 {
			return fmt.Errorf("List after Delete = %v, %v", keys, err)
		}
		// A deleted log starts again from empty.
		if err := b.Append("log", []byte("2")); err != nil {
			return err
		}
		return expectRecords(b, "log", "2")
	})
}

func checkExactKeys(open Opener) error {
	return with(open, func(b storage.Backend) error {
		keys := []string{"media/Cat.jpg", "media/cat.jpg", "media/ñandú %20 é.png", "media/a:b*c?.png"}
		for _, k := range keys {
			if err := b.Put(k, []byte(k)); err != nil {
				return fmt.Errorf("Put(%q): %w", k, err)
			}
		}
		for _, k := range keys {
			if err := expectValue(b, k, []byte(k)); err != nil {
				return err
			}
		}
		got, err := b.List("media/")
		if err != nil {
			return err
		}
		if len(got) != len(keys) {
			return fmt.Errorf("List(media/) = %q", got)
		}
		return nil
	})
}

func checkInvalidKeys(open Opener) error {
	return with(open, func(b storage.Backend) error {
		for _, k := range []string{"", "a//b", "/a", "a/", ".hidden", "a/.tmp-1", strings.Repeat("x", 201)} {
			if err := b.Put(k, []byte("x")); err == nil {
				return fmt.Errorf("Put(%q) succeeded", k)
			}
		}
		return nil
	})
}

func checkReopen(open Opener) error {
	err := with(open, func(b storage.Backend) error {
		if err := b.Put("current", []byte("gen-3")); err != nil {
			return err
		}
		if err := b.Append("gen-3/journal", []byte("a")); err != nil {
			return err
		}
		return b.Append("gen-3/journal", []byte("b"))
	})
	if err != nil {
		return err
	}
	return with(open, func(b storage.Backend) error {
		if err := expectValue(b, "current", []byte("gen-3")); err != nil {
			return err
		}
		if err := b.Append("gen-3/journal", []byte("c")); err != nil {
			return err
		}
		return expectRecords(b, "gen-3/journal", "a", "b", "c")
	})
}

func checkPutOverLog(open Opener) error {
	return with(open, func(b storage.Backend) error {
		if err := b.Append("k", []byte("rec")); err != nil {
			return err
		}
		if err := b.Put("k", []byte("value")); err != nil {
			return err
		}
		return expectValue(b, "k", []byte("value"))
	})
}

func expectBackups(bs storage.BackupStore, want ...storage.BackupFile) error {
	got, err := bs.List()
	if err != nil {
		return fmt.Errorf("List: %w", err)
	}
	if fmt.Sprint(got) != fmt.Sprint(want) || len(got) != len(want) {
		return fmt.Errorf("backups = %v, want %v", got, want)
	}
	return nil
}

func expectBackup(bs storage.BackupStore, name string, want []byte) error {
	got, err := bs.Read(name)
	if err != nil {
		return fmt.Errorf("Read(%q): %w", name, err)
	}
	if !bytes.Equal(got, want) {
		return fmt.Errorf("Read(%q) = %d bytes, want %d", name, len(got), len(want))
	}
	return nil
}

func checkBackups(open Opener) error {
	return with(open, func(b storage.Backend) error {
		bs := b.Backups()
		if err := expectBackups(bs); err != nil {
			return err
		}
		big := bytes.Repeat([]byte("backup data "), 1<<17) // 1.5 MiB
		for name, data := range map[string][]byte{"b-2.cardflip": []byte("two"), "b-1.cardflip": big, "b-3": {}} {
			if err := bs.Write(name, data); err != nil {
				return fmt.Errorf("Write(%q): %w", name, err)
			}
		}
		if err := expectBackup(bs, "b-1.cardflip", big); err != nil {
			return err
		}
		if err := expectBackups(bs, storage.BackupFile{Name: "b-1.cardflip", Size: int64(len(big))},
			storage.BackupFile{Name: "b-2.cardflip", Size: 3}, storage.BackupFile{Name: "b-3", Size: 0}); err != nil {
			return err
		}
		for _, name := range []string{"b-1.cardflip", "never-existed"} {
			if err := bs.Delete(name); err != nil {
				return fmt.Errorf("Delete(%q): %w", name, err)
			}
		}
		if _, err := bs.Read("b-1.cardflip"); !errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("Read after Delete: %v, want ErrNotFound", err)
		}
		return expectBackups(bs, storage.BackupFile{Name: "b-2.cardflip", Size: 3}, storage.BackupFile{Name: "b-3", Size: 0})
	})
}

func checkBackupReplaces(open Opener) error {
	return with(open, func(b storage.Backend) error {
		bs := b.Backups()
		if err := bs.Write("b", []byte("old, longer value")); err != nil {
			return err
		}
		if err := bs.Write("b", []byte("new")); err != nil {
			return err
		}
		if err := expectBackup(bs, "b", []byte("new")); err != nil {
			return err
		}
		return expectBackups(bs, storage.BackupFile{Name: "b", Size: 3})
	})
}

// Replacing or resetting the collection deletes every key; backups must
// not be among them.
func checkBackupsApart(open Opener) error {
	return with(open, func(b storage.Backend) error {
		if err := b.Put("backups", []byte("a key")); err != nil {
			return err
		}
		if err := b.Backups().Write("backups", []byte("a backup")); err != nil {
			return err
		}
		keys, err := b.List("")
		if err != nil {
			return err
		}
		if strings.Join(keys, ",") != "backups" {
			return fmt.Errorf("List('') = %v, want only the key", keys)
		}
		if err := b.Delete("backups"); err != nil {
			return err
		}
		if err := expectBackup(b.Backups(), "backups", []byte("a backup")); err != nil {
			return err
		}
		if keys, err := b.List(""); err != nil || len(keys) != 0 {
			return fmt.Errorf("List after Delete = %v, %v", keys, err)
		}
		return nil
	})
}

func checkBackupsReopen(open Opener) error {
	err := with(open, func(b storage.Backend) error {
		return b.Backups().Write("b-1", []byte("kept"))
	})
	if err != nil {
		return err
	}
	return with(open, func(b storage.Backend) error {
		if err := expectBackup(b.Backups(), "b-1", []byte("kept")); err != nil {
			return err
		}
		return expectBackups(b.Backups(), storage.BackupFile{Name: "b-1", Size: 4})
	})
}

func checkBackupInvalidNames(open Opener) error {
	return with(open, func(b storage.Backend) error {
		bs := b.Backups()
		for _, name := range []string{"", ".hidden", "a/b", "../escape", "Upper", "sp ace", strings.Repeat("x", 101)} {
			if err := bs.Write(name, []byte("x")); err == nil {
				return fmt.Errorf("Write(%q) succeeded", name)
			}
		}
		return expectBackups(bs)
	})
}

// Trash times are kept to the millisecond.
var (
	trashT1 = time.Date(2026, 10, 1, 9, 30, 0, 123_000_000, time.UTC)
	trashT2 = time.Date(2026, 10, 13, 18, 5, 7, 456_000_000, time.UTC)
)

func expectTrash(t storage.MediaTrash, want ...storage.TrashedMedia) error {
	got, err := t.List()
	if err != nil {
		return fmt.Errorf("List: %w", err)
	}
	ok := len(got) == len(want)
	for i := 0; ok && i < len(got); i++ {
		ok = got[i].Name == want[i].Name && got[i].Size == want[i].Size && got[i].Trashed.Equal(want[i].Trashed)
	}
	if !ok {
		return fmt.Errorf("trash = %v, want %v", got, want)
	}
	return nil
}

func expectTrashed(t storage.MediaTrash, name string, want []byte) error {
	got, err := t.Get(name)
	if err != nil {
		return fmt.Errorf("Get(%q): %w", name, err)
	}
	if !bytes.Equal(got, want) {
		return fmt.Errorf("Get(%q) = %d bytes, want %d", name, len(got), len(want))
	}
	return nil
}

func checkTrash(open Opener) error {
	return with(open, func(b storage.Backend) error {
		t := b.Trash()
		if err := expectTrash(t); err != nil {
			return err
		}
		if _, err := t.Get("cat.jpg"); !errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("Get of missing file: %v, want ErrNotFound", err)
		}
		big := bytes.Repeat([]byte("media data "), 1<<17) // 1.4 MiB
		if err := t.Put("old.jpg", big, trashT1); err != nil {
			return fmt.Errorf("Put: %w", err)
		}
		if err := t.Put("empty.mp3", nil, trashT2); err != nil {
			return fmt.Errorf("Put: %w", err)
		}
		if err := expectTrashed(t, "old.jpg", big); err != nil {
			return err
		}
		if err := expectTrashed(t, "empty.mp3", nil); err != nil {
			return err
		}
		if err := expectTrash(t, storage.TrashedMedia{Name: "empty.mp3", Size: 0, Trashed: trashT2},
			storage.TrashedMedia{Name: "old.jpg", Size: int64(len(big)), Trashed: trashT1}); err != nil {
			return err
		}
		for _, name := range []string{"old.jpg", "never-existed"} {
			if err := t.Delete(name); err != nil {
				return fmt.Errorf("Delete(%q): %w", name, err)
			}
		}
		if _, err := t.Get("old.jpg"); !errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("Get after Delete: %v, want ErrNotFound", err)
		}
		if err := t.Put("", []byte("x"), trashT1); err == nil {
			return errors.New("Put with an empty name succeeded")
		}
		return expectTrash(t, storage.TrashedMedia{Name: "empty.mp3", Size: 0, Trashed: trashT2})
	})
}

func checkTrashReplaces(open Opener) error {
	return with(open, func(b storage.Backend) error {
		t := b.Trash()
		if err := t.Put("a.png", []byte("old, longer content"), trashT1); err != nil {
			return err
		}
		if err := t.Put("a.png", []byte("new"), trashT2); err != nil {
			return err
		}
		if err := expectTrashed(t, "a.png", []byte("new")); err != nil {
			return err
		}
		return expectTrash(t, storage.TrashedMedia{Name: "a.png", Size: 3, Trashed: trashT2})
	})
}

func checkTrashExactNames(open Opener) error {
	return with(open, func(b storage.Backend) error {
		t := b.Trash()
		names := []string{"Cat.jpg", "cat.jpg", "ñandú %20 é.png", "a:b*c?/d.png", strings.Repeat("long ", 120) + ".mp3"}
		for _, name := range names {
			if err := t.Put(name, []byte(name), trashT1); err != nil {
				return fmt.Errorf("Put(%q): %w", name, err)
			}
		}
		for _, name := range names {
			if err := expectTrashed(t, name, []byte(name)); err != nil {
				return err
			}
		}
		list, err := t.List()
		if err != nil {
			return err
		}
		if len(list) != len(names) {
			return fmt.Errorf("List = %v", list)
		}
		for _, f := range list {
			if f.Size != int64(len(f.Name)) {
				return fmt.Errorf("%q: size %d, want %d", f.Name, f.Size, len(f.Name))
			}
		}
		return nil
	})
}

// Resetting the collection deletes every key; trashed files must not be
// among them, nor among the backups.
func checkTrashApart(open Opener) error {
	return with(open, func(b storage.Backend) error {
		if err := b.Put("media", []byte("a key")); err != nil {
			return err
		}
		if err := b.Trash().Put("media", []byte("trashed"), trashT1); err != nil {
			return err
		}
		keys, err := b.List("")
		if err != nil {
			return err
		}
		if strings.Join(keys, ",") != "media" {
			return fmt.Errorf("List('') = %v, want only the key", keys)
		}
		if err := expectBackups(b.Backups()); err != nil {
			return err
		}
		if err := b.Delete("media"); err != nil {
			return err
		}
		return expectTrashed(b.Trash(), "media", []byte("trashed"))
	})
}

func checkTrashReopen(open Opener) error {
	err := with(open, func(b storage.Backend) error {
		return b.Trash().Put("kept.ogg", []byte("kept"), trashT2)
	})
	if err != nil {
		return err
	}
	return with(open, func(b storage.Backend) error {
		if err := expectTrashed(b.Trash(), "kept.ogg", []byte("kept")); err != nil {
			return err
		}
		return expectTrash(b.Trash(), storage.TrashedMedia{Name: "kept.ogg", Size: 4, Trashed: trashT2})
	})
}
