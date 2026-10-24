package storage

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// MediaTrash holds media files removed from the collection until they are
// restored or purged. Like backups, trashed files are not keys of the
// Backend, so replacing or resetting the collection never touches them.
type MediaTrash interface {
	// List returns the trashed files, sorted by name.
	List() ([]TrashedMedia, error)
	// Put stores a file atomically with the time it was trashed, replacing
	// a trashed file with the same name. It returns after the data is
	// durable.
	Put(name string, data []byte, trashed time.Time) error
	// Get returns a trashed file's content, or ErrNotFound.
	Get(name string) ([]byte, error)
	// Delete removes a trashed file. Deleting a missing one is not an error.
	Delete(name string) error
}

// TrashedMedia is a file in the media trash. Trashed has millisecond
// precision.
type TrashedMedia struct {
	Name    string
	Size    int64
	Trashed time.Time
}

func validTrashName(name string) error {
	if name == "" {
		return errors.New("storage: empty media name")
	}
	return nil
}

// trashFileName maps a media name to a short name that is valid and
// distinct on every file system, like mediaKey.
func trashFileName(name string) string {
	h := sha256.Sum256([]byte(name))
	return hex.EncodeToString(h[:16])
}

// encodeTrashed frames a trashed file for backends that keep it in one
// byte string: the name as in encodeMedia, the trash time in Unix
// milliseconds (8 bytes, little-endian), the data.
func encodeTrashed(name string, data []byte, trashed time.Time) []byte {
	out := binary.AppendUvarint(nil, uint64(len(name)))
	out = append(out, name...)
	out = binary.LittleEndian.AppendUint64(out, uint64(trashed.UnixMilli()))
	return append(out, data...)
}

// decodeTrashedHeader reads the name and time of an encoded trashed file
// from its start and returns the length of the header.
func decodeTrashedHeader(v []byte) (name string, trashed time.Time, n int, err error) {
	l, k := binary.Uvarint(v)
	if k <= 0 || uint64(len(v)-k) < l+8 {
		return "", time.Time{}, 0, fmt.Errorf("%w: media trash header", ErrCorrupt)
	}
	name = string(v[k : k+int(l)])
	ms := int64(binary.LittleEndian.Uint64(v[k+int(l):]))
	return name, time.UnixMilli(ms), k + int(l) + 8, nil
}
