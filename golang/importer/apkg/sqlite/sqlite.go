// Package sqlite is a minimal read-only SQLite reader for Anki packages.
//
// It reads a whole database file from memory and supports exactly what the
// .apkg importer needs: full scans of rowid tables and WITHOUT ROWID tables,
// including overflow pages and INTEGER PRIMARY KEY aliases. There are no
// queries, indexes, WAL files or writes. Every malformed input is reported as
// an error, never a panic, so a corrupt package fails cleanly.
package sqlite

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strings"
)

// ErrCorrupt is wrapped by every error caused by a malformed database file.
var ErrCorrupt = errors.New("sqlite: database is malformed")

const (
	headerMagic = "SQLite format 3\x00"

	pageIndexInterior = 0x02
	pageTableInterior = 0x05
	pageIndexLeaf     = 0x0a
	pageTableLeaf     = 0x0d

	// maxDepth bounds B-tree recursion; real databases stay far below it.
	maxDepth = 64
)

// DB is an opened database file.
type DB struct {
	data     []byte
	pageSize int
	usable   int
	pages    int
	tables   map[string]*table
}

type table struct {
	name      string
	root      int
	columns   []string
	withoutID bool
	// rowidCol is the index of the INTEGER PRIMARY KEY column, or -1.
	rowidCol int
	// order maps record positions to column indexes (WITHOUT ROWID tables
	// store the primary key columns first).
	order []int
}

// Open parses the header and schema of the database file in data. data is
// retained and must not be modified while the DB is in use.
func Open(data []byte) (db *DB, err error) {
	defer recoverCorrupt(&err)
	if len(data) < 100 || string(data[:16]) != headerMagic {
		return nil, fmt.Errorf("%w: not an SQLite database", ErrCorrupt)
	}
	pageSize := int(binary.BigEndian.Uint16(data[16:18]))
	if pageSize == 1 {
		pageSize = 65536
	}
	if pageSize < 512 || pageSize&(pageSize-1) != 0 {
		return nil, fmt.Errorf("%w: invalid page size %d", ErrCorrupt, pageSize)
	}
	reserved := int(data[20])
	if enc := binary.BigEndian.Uint32(data[56:60]); enc != 0 && enc != 1 {
		return nil, fmt.Errorf("sqlite: unsupported text encoding %d (only UTF-8)", enc)
	}
	db = &DB{
		data:     data,
		pageSize: pageSize,
		usable:   pageSize - reserved,
		pages:    len(data) / pageSize,
		tables:   map[string]*table{},
	}
	if db.usable < 480 {
		return nil, fmt.Errorf("%w: usable page size %d too small", ErrCorrupt, db.usable)
	}
	if db.pages < 1 {
		return nil, fmt.Errorf("%w: truncated database", ErrCorrupt)
	}

	schema := &table{name: "sqlite_schema", root: 1, rowidCol: -1,
		columns: []string{"type", "name", "tbl_name", "rootpage", "sql"}}
	schema.order = identity(len(schema.columns))
	err = db.scan(schema, func(r Row) error {
		if r.Str("type") != "table" {
			return nil
		}
		t, err := parseCreateTable(r.Str("sql"))
		if err != nil {
			return fmt.Errorf("table %s: %w", r.Str("name"), err)
		}
		t.name = r.Str("name")
		t.root = int(r.Int("rootpage"))
		db.tables[strings.ToLower(t.name)] = t
		return nil
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}

// HasTable reports whether the database has a table with this name.
func (db *DB) HasTable(name string) bool {
	_, ok := db.tables[strings.ToLower(name)]
	return ok
}

// Columns returns the declared column names of table name.
func (db *DB) Columns(name string) ([]string, error) {
	t, ok := db.tables[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("sqlite: no table %q", name)
	}
	return append([]string(nil), t.columns...), nil
}

// Scan calls fn for every row of table name, in storage order. It stops at
// the first error from fn and returns it.
func (db *DB) Scan(name string, fn func(Row) error) (err error) {
	t, ok := db.tables[strings.ToLower(name)]
	if !ok {
		return fmt.Errorf("sqlite: no table %q", name)
	}
	defer recoverCorrupt(&err)
	return db.scan(t, fn)
}

// Count returns the number of rows in table name.
func (db *DB) Count(name string) (int, error) {
	n := 0
	err := db.Scan(name, func(Row) error { n++; return nil })
	return n, err
}

func (db *DB) scan(t *table, fn func(Row) error) error {
	cols := make(map[string]int, len(t.columns))
	for i, c := range t.columns {
		cols[strings.ToLower(c)] = i
	}
	visited := map[int]bool{}
	var walk func(page, depth int) error
	walk = func(page, depth int) error {
		if depth > maxDepth {
			return fmt.Errorf("%w: B-tree too deep", ErrCorrupt)
		}
		if visited[page] {
			return fmt.Errorf("%w: page %d referenced twice", ErrCorrupt, page)
		}
		visited[page] = true
		p, hdr, err := db.page(page)
		if err != nil {
			return err
		}
		kind := p[hdr]
		n := int(binary.BigEndian.Uint16(p[hdr+3 : hdr+5]))
		interior := kind == pageTableInterior || kind == pageIndexInterior
		ptrs := hdr + 8
		if interior {
			ptrs = hdr + 12
		}
		if ptrs+2*n > len(p) {
			return fmt.Errorf("%w: page %d cell array overflows", ErrCorrupt, page)
		}
		wantTable := !t.withoutID
		if (kind == pageTableLeaf || kind == pageTableInterior) != wantTable {
			return fmt.Errorf("%w: page %d has type %#x in table %s", ErrCorrupt, page, kind, t.name)
		}
		for i := 0; i < n; i++ {
			off := int(binary.BigEndian.Uint16(p[ptrs+2*i:]))
			if off < hdr || off >= len(p) {
				return fmt.Errorf("%w: page %d cell offset out of range", ErrCorrupt, page)
			}
			cell := p[off:]
			switch kind {
			case pageTableInterior:
				if err := walk(int(binary.BigEndian.Uint32(cell)), depth+1); err != nil {
					return err
				}
			case pageTableLeaf:
				size, k := varint(cell)
				rowid, k2 := varint(cell[k:])
				payload, err := db.payload(cell[k+k2:], int(size), true)
				if err != nil {
					return err
				}
				if err := emit(t, cols, payload, int64(rowid), fn); err != nil {
					return err
				}
			case pageIndexInterior:
				if err := walk(int(binary.BigEndian.Uint32(cell)), depth+1); err != nil {
					return err
				}
				size, k := varint(cell[4:])
				payload, err := db.payload(cell[4+k:], int(size), false)
				if err != nil {
					return err
				}
				if err := emit(t, cols, payload, 0, fn); err != nil {
					return err
				}
			case pageIndexLeaf:
				size, k := varint(cell)
				payload, err := db.payload(cell[k:], int(size), false)
				if err != nil {
					return err
				}
				if err := emit(t, cols, payload, 0, fn); err != nil {
					return err
				}
			}
		}
		if interior {
			return walk(int(binary.BigEndian.Uint32(p[hdr+8:hdr+12])), depth+1)
		}
		return nil
	}
	return walk(t.root, 0)
}

// page returns the bytes of page n (1-based) and the offset of its B-tree
// header, which is 100 on page 1 because of the file header.
func (db *DB) page(n int) ([]byte, int, error) {
	if n < 1 || n > db.pages {
		return nil, 0, fmt.Errorf("%w: page %d out of range (%d pages)", ErrCorrupt, n, db.pages)
	}
	p := db.data[(n-1)*db.pageSize : n*db.pageSize]
	hdr := 0
	if n == 1 {
		hdr = 100
	}
	switch p[hdr] {
	case pageIndexInterior, pageTableInterior, pageIndexLeaf, pageTableLeaf:
	default:
		return nil, 0, fmt.Errorf("%w: page %d is not a B-tree page", ErrCorrupt, n)
	}
	return p, hdr, nil
}

// payload assembles a cell payload of the given total size whose local part
// starts at local, following overflow pages as needed.
func (db *DB) payload(local []byte, size int, tableLeaf bool) ([]byte, error) {
	u := db.usable
	maxLocal := u - 35
	if !tableLeaf {
		maxLocal = (u-12)*64/255 - 23
	}
	if size < 0 {
		return nil, fmt.Errorf("%w: negative payload size", ErrCorrupt)
	}
	if size <= maxLocal {
		if size > len(local) {
			return nil, fmt.Errorf("%w: payload overflows page", ErrCorrupt)
		}
		return local[:size], nil
	}
	minLocal := (u-12)*32/255 - 23
	k := minLocal + (size-minLocal)%(u-4)
	n := k
	if k > maxLocal {
		n = minLocal
	}
	if n+4 > len(local) {
		return nil, fmt.Errorf("%w: payload overflows page", ErrCorrupt)
	}
	if size > len(db.data) {
		return nil, fmt.Errorf("%w: payload larger than the file", ErrCorrupt)
	}
	out := make([]byte, 0, size)
	out = append(out, local[:n]...)
	next := int(binary.BigEndian.Uint32(local[n:]))
	seen := map[int]bool{}
	for len(out) < size {
		if next < 1 || next > db.pages || seen[next] {
			return nil, fmt.Errorf("%w: bad overflow page %d", ErrCorrupt, next)
		}
		seen[next] = true
		p := db.data[(next-1)*db.pageSize : next*db.pageSize]
		chunk := p[4:u]
		if rest := size - len(out); rest < len(chunk) {
			chunk = chunk[:rest]
		}
		out = append(out, chunk...)
		next = int(binary.BigEndian.Uint32(p))
	}
	return out, nil
}

func emit(t *table, cols map[string]int, payload []byte, rowid int64, fn func(Row) error) error {
	rec, err := decodeRecord(payload)
	if err != nil {
		return err
	}
	vals := make([]any, len(t.columns))
	for i, v := range rec {
		if i < len(t.order) {
			vals[t.order[i]] = v
		}
	}
	if t.rowidCol >= 0 {
		vals[t.rowidCol] = rowid
	}
	return fn(Row{cols: cols, vals: vals})
}

// decodeRecord decodes an SQLite record into int64, float64, string, []byte
// and nil values.
func decodeRecord(b []byte) ([]any, error) {
	hsize, n := varint(b)
	if n == 0 || int(hsize) > len(b) || int(hsize) < n {
		return nil, fmt.Errorf("%w: bad record header", ErrCorrupt)
	}
	types := b[n:hsize]
	body := b[hsize:]
	var vals []any
	for len(types) > 0 {
		st, k := varint(types)
		if k == 0 {
			return nil, fmt.Errorf("%w: bad serial type", ErrCorrupt)
		}
		types = types[k:]
		size := serialSize(st)
		if size < 0 || size > len(body) {
			return nil, fmt.Errorf("%w: record value overflows", ErrCorrupt)
		}
		v := body[:size]
		body = body[size:]
		switch {
		case st == 0:
			vals = append(vals, nil)
		case st >= 1 && st <= 6:
			vals = append(vals, beInt(v))
		case st == 7:
			vals = append(vals, math.Float64frombits(binary.BigEndian.Uint64(v)))
		case st == 8:
			vals = append(vals, int64(0))
		case st == 9:
			vals = append(vals, int64(1))
		case st >= 12 && st%2 == 0:
			vals = append(vals, bytes.Clone(v))
		case st >= 13:
			vals = append(vals, string(v))
		default:
			return nil, fmt.Errorf("%w: reserved serial type %d", ErrCorrupt, st)
		}
	}
	return vals, nil
}

func serialSize(st uint64) int {
	switch st {
	case 0, 8, 9:
		return 0
	case 1:
		return 1
	case 2:
		return 2
	case 3:
		return 3
	case 4:
		return 4
	case 5:
		return 6
	case 6, 7:
		return 8
	case 10, 11:
		return -1
	}
	if st > 1<<40 {
		return -1
	}
	if st%2 == 0 {
		return int(st-12) / 2
	}
	return int(st-13) / 2
}

// beInt decodes a big-endian two's-complement integer of 1 to 8 bytes.
func beInt(b []byte) int64 {
	var v int64
	if len(b) > 0 && b[0]&0x80 != 0 {
		v = -1
	}
	for _, c := range b {
		v = v<<8 | int64(c)
	}
	return v
}

// varint decodes an SQLite varint and returns it with its length, or a
// length of 0 when b is too short.
func varint(b []byte) (uint64, int) {
	var v uint64
	for i := 0; i < 9; i++ {
		if i >= len(b) {
			return 0, 0
		}
		if i == 8 {
			return v<<8 | uint64(b[i]), 9
		}
		v = v<<7 | uint64(b[i]&0x7f)
		if b[i]&0x80 == 0 {
			return v, i + 1
		}
	}
	return v, 9
}

func identity(n int) []int {
	o := make([]int, n)
	for i := range o {
		o[i] = i
	}
	return o
}

// recoverCorrupt turns an index panic from malformed data into ErrCorrupt.
// The explicit checks above should make it unreachable; it is a backstop so
// that a hostile file can never crash the app.
func recoverCorrupt(err *error) {
	if r := recover(); r != nil {
		*err = fmt.Errorf("%w: %v", ErrCorrupt, r)
	}
}

// Row is one table row. Accessors return the zero value for NULL, missing
// columns (rows written before an ALTER TABLE ADD COLUMN) and type mismatches.
type Row struct {
	cols map[string]int
	vals []any
}

// Value returns the raw value of column name: int64, float64, string, []byte
// or nil.
func (r Row) Value(name string) any {
	i, ok := r.cols[strings.ToLower(name)]
	if !ok || i >= len(r.vals) {
		return nil
	}
	return r.vals[i]
}

// Int returns column name as an integer.
func (r Row) Int(name string) int64 {
	switch v := r.Value(name).(type) {
	case int64:
		return v
	case float64:
		return int64(v)
	}
	return 0
}

// Float returns column name as a float.
func (r Row) Float(name string) float64 {
	switch v := r.Value(name).(type) {
	case int64:
		return float64(v)
	case float64:
		return v
	}
	return 0
}

// Str returns column name as text; blobs are returned as their bytes.
func (r Row) Str(name string) string {
	switch v := r.Value(name).(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}
	return ""
}

// Bytes returns column name as bytes; text is returned as its UTF-8 bytes.
func (r Row) Bytes(name string) []byte {
	switch v := r.Value(name).(type) {
	case []byte:
		return v
	case string:
		return []byte(v)
	}
	return nil
}
