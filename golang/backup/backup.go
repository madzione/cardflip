// Package backup writes and reads .cardflip archives: a zip with
// manifest.json, collection.json (the snapshot), revlog.jsonl (one review
// log entry per line), deleted.jsonl (one deletion log entry per line; older
// backups have none) and media/<name>.
package backup

import (
	"archive/zip"
	"bufio"
	"bytes"
	"cardflip/golang/internal/usermsg"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// Format is the backup format this app writes and the newest it reads.
// Format 2 has deck option presets; format 1 backups are migrated when
// they are restored.
const Format = 2

// Manifest describes a backup.
type Manifest struct {
	Format     int       `json:"format"`
	AppVersion string    `json:"appVersion"`
	CreatedAt  time.Time `json:"createdAt"`
}

// Source is what a backup is made from.
type Source interface {
	Snapshot() ([]byte, error)
	RevlogMonths() ([]string, error)
	ReadRevlog(month string) ([]json.RawMessage, error)
	ReadDeletionLog() ([]json.RawMessage, error)
	EachMedia(fn func(name string, data []byte) error) error
}

// Errors returned by Read. Each says why a file cannot be restored.
var (
	ErrNotBackup   = errors.New(usermsg.Text("notBackup", "this file is not a Cardflip backup"))
	ErrNewerFormat = errors.New(usermsg.Text("backupNewerFormat", "this backup was made by a newer version of Cardflip; update Cardflip to restore it"))
	ErrDamaged     = errors.New(usermsg.Text("backupDamaged", "this backup is damaged"))
)

// FileName is the suggested name of a backup made on day t.
func FileName(t time.Time) string {
	return "cardflip-" + t.Format("2006-01-02") + ".cardflip"
}

// Export writes a backup of src. yield is called between media files.
func Export(src Source, appVersion string, now time.Time, yield func()) ([]byte, error) {
	return export(src, appVersion, now, yield, true)
}

// ExportCollection writes a backup of src without media files.
func ExportCollection(src Source, appVersion string, now time.Time, yield func()) ([]byte, error) {
	return export(src, appVersion, now, yield, false)
}

func export(src Source, appVersion string, now time.Time, yield func(), media bool) ([]byte, error) {
	if yield == nil {
		yield = func() {}
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	put := func(name string, method uint16, data []byte) error {
		w, err := zw.CreateHeader(&zip.FileHeader{Name: name, Method: method, Modified: now})
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	}
	manifest, _ := json.MarshalIndent(Manifest{Format: Format, AppVersion: appVersion, CreatedAt: now.UTC()}, "", "  ")
	if err := put("manifest.json", zip.Deflate, manifest); err != nil {
		return nil, err
	}
	snap, err := src.Snapshot()
	if err != nil {
		return nil, err
	}
	if err := put("collection.json", zip.Deflate, snap); err != nil {
		return nil, err
	}
	months, err := src.RevlogMonths()
	if err != nil {
		return nil, err
	}
	var revlog bytes.Buffer
	for _, m := range months {
		entries, err := src.ReadRevlog(m)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			revlog.Write(e)
			revlog.WriteByte('\n')
		}
		yield()
	}
	if err := put("revlog.jsonl", zip.Deflate, revlog.Bytes()); err != nil {
		return nil, err
	}
	deleted, err := src.ReadDeletionLog()
	if err != nil {
		return nil, err
	}
	if err := put("deleted.jsonl", zip.Deflate, jsonLines(deleted)); err != nil {
		return nil, err
	}
	if media {
		err = src.EachMedia(func(name string, data []byte) error {
			yield()
			// Images are compressed already.
			return put("media/"+name, zip.Store, data)
		})
		if err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func jsonLines(entries []json.RawMessage) []byte {
	var buf bytes.Buffer
	for _, e := range entries {
		buf.Write(e)
		buf.WriteByte('\n')
	}
	return buf.Bytes()
}

// Archive is a backup's validated content.
type Archive struct {
	Manifest   Manifest
	Collection []byte
	Revlog     []json.RawMessage
	Deletions  []json.RawMessage
	Media      map[string][]byte
}

// Read opens and validates a backup. Its errors wrap ErrNotBackup,
// ErrNewerFormat or ErrDamaged.
func Read(data []byte) (*Archive, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, ErrNotBackup
	}
	files := map[string]*zip.File{}
	for _, f := range zr.File {
		files[f.Name] = f
	}
	read := func(name string) ([]byte, error) {
		f, ok := files[name]
		if !ok {
			return nil, fmt.Errorf("%w: %s is missing", ErrDamaged, name)
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("%w: %s: %v", ErrDamaged, name, err)
		}
		defer rc.Close()
		b, err := io.ReadAll(rc)
		if err != nil {
			return nil, fmt.Errorf("%w: %s: %v", ErrDamaged, name, err)
		}
		return b, nil
	}
	if files["manifest.json"] == nil {
		return nil, ErrNotBackup
	}
	raw, err := read("manifest.json")
	if err != nil {
		return nil, err
	}
	a := &Archive{Media: map[string][]byte{}}
	if err := json.Unmarshal(raw, &a.Manifest); err != nil || a.Manifest.Format < 1 {
		return nil, ErrNotBackup
	}
	if a.Manifest.Format > Format {
		return nil, fmt.Errorf("%w (backup format %d, this app reads up to %d)", ErrNewerFormat, a.Manifest.Format, Format)
	}
	if a.Collection, err = read("collection.json"); err != nil {
		return nil, err
	}
	if !json.Valid(a.Collection) {
		return nil, fmt.Errorf("%w: collection.json is not valid JSON", ErrDamaged)
	}
	rl, err := read("revlog.jsonl")
	if err != nil {
		return nil, err
	}
	if a.Revlog, err = readLines("revlog.jsonl", rl); err != nil {
		return nil, err
	}
	if files["deleted.jsonl"] != nil {
		dl, err := read("deleted.jsonl")
		if err != nil {
			return nil, err
		}
		if a.Deletions, err = readLines("deleted.jsonl", dl); err != nil {
			return nil, err
		}
	}
	for name := range files {
		if media, ok := strings.CutPrefix(name, "media/"); ok && media != "" {
			if a.Media[media], err = read(name); err != nil {
				return nil, err
			}
		}
	}
	return a, nil
}

// readLines reads a file of one JSON value per line.
func readLines(name string, data []byte) ([]json.RawMessage, error) {
	var out []json.RawMessage
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 64<<10), 16<<20)
	for line := 1; sc.Scan(); line++ {
		l := bytes.TrimSpace(sc.Bytes())
		if len(l) == 0 {
			continue
		}
		if !json.Valid(l) {
			return nil, fmt.Errorf("%w: %s line %d is not valid JSON", ErrDamaged, name, line)
		}
		out = append(out, json.RawMessage(bytes.Clone(l)))
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrDamaged, name, err)
	}
	return out, nil
}
