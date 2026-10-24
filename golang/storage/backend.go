// Package storage persists the collection: a small key/value Backend with
// files on native platforms and IndexedDB on the web, and the generation
// layer (snapshot + journal, review log segments, media) on top of it.
package storage

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"strings"
)

// ErrNotFound is returned by Get for a key that does not exist.
var ErrNotFound = errors.New("storage: not found")

// ErrCorrupt is wrapped by errors caused by unreadable stored data.
var ErrCorrupt = errors.New("storage: data is corrupt")

// ErrBusy is returned by operations that would replace the live state
// while a Replacement is being built; they can be retried after it.
var ErrBusy = errors.New("storage: a replacement is in progress")

// Backend is a durable key/value store. Keys are slash-separated paths such
// as "gen-1/snapshot" or "revlog/2026-10"; each segment is non-empty and must
// not start with a dot.
//
// A key holds either a value written with Put or a record log written with
// Append, never both. Every method returns only after its data is durable.
type Backend interface {
	// Get returns the value of key, or ErrNotFound.
	Get(key string) ([]byte, error)
	// Put replaces the value of key atomically: after a crash, Get returns
	// either the old or the new value, never a mix.
	Put(key string, value []byte) error
	// Append adds one record to the log at key, creating it when needed.
	Append(key string, record []byte) error
	// Records returns the records of the log at key in append order, or
	// none when it does not exist. A torn trailing record (a crash during
	// Append) is dropped; any other damage is an ErrCorrupt error.
	Records(key string) ([][]byte, error)
	// List returns all keys that start with prefix, sorted.
	List(prefix string) ([]string, error)
	// Delete removes key. Deleting a missing key is not an error.
	Delete(key string) error
	// Backups returns the store for backup files kept next to the keys.
	Backups() BackupStore
	// Trash returns the store for trashed media files kept next to the keys.
	Trash() MediaTrash
	// Close releases resources. The backend must not be used afterwards.
	Close() error
}

// AppendAller is implemented by backends that can append several records
// to a log in one durable write, much faster than one Append per record.
// After a crash, a prefix of the records may be present, as after a crash
// during a series of Appends.
type AppendAller interface {
	AppendAll(key string, records [][]byte) error
}

// appendAll appends records to the log at key, in one write when b
// supports it.
func appendAll(b Backend, key string, records [][]byte) error {
	if ba, ok := b.(AppendAller); ok && len(records) > 1 {
		return ba.AppendAll(key, records)
	}
	for _, r := range records {
		if err := b.Append(key, r); err != nil {
			return err
		}
	}
	return nil
}

// ValidKey reports whether key is acceptable to every backend.
func ValidKey(key string) error {
	if key == "" || len(key) > 200 {
		return fmt.Errorf("storage: invalid key length %d", len(key))
	}
	for _, seg := range strings.Split(key, "/") {
		if seg == "" || seg[0] == '.' {
			return fmt.Errorf("storage: invalid key %q", key)
		}
	}
	return nil
}

// Record framing shared by backends that keep a log in one byte stream:
// 4-byte little-endian length, 4-byte CRC-32 (IEEE) of the data, the data.
const frameHeader = 8

func frame(record []byte) []byte {
	out := make([]byte, frameHeader+len(record))
	binary.LittleEndian.PutUint32(out[0:4], uint32(len(record)))
	binary.LittleEndian.PutUint32(out[4:8], crc32.ChecksumIEEE(record))
	copy(out[frameHeader:], record)
	return out
}

// unframe splits a framed log into records. It returns the records and the
// length of the valid prefix: a torn final frame (cut short, or the last
// frame failing its checksum) ends the log there. A bad frame followed by
// more data means the log is damaged, which is an ErrCorrupt error.
func unframe(data []byte) ([][]byte, int, error) {
	var out [][]byte
	off := 0
	for off < len(data) {
		rest := data[off:]
		if len(rest) < frameHeader {
			return out, off, nil
		}
		n := int(binary.LittleEndian.Uint32(rest[0:4]))
		if n > len(rest)-frameHeader {
			return out, off, nil
		}
		rec := rest[frameHeader : frameHeader+n]
		if crc32.ChecksumIEEE(rec) != binary.LittleEndian.Uint32(rest[4:8]) {
			if frameHeader+n == len(rest) {
				return out, off, nil
			}
			return nil, 0, fmt.Errorf("%w: checksum mismatch at offset %d", ErrCorrupt, off)
		}
		out = append(out, rec)
		off += frameHeader + n
	}
	return out, off, nil
}
