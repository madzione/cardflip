// Package sqlite writes SQLite database files without an SQL engine, for
// Anki package export.
//
// A Writer collects the rows of rowid tables and builds the whole file at
// once: the header, the schema table on page 1, a B-tree per table and per
// index, and overflow pages for large values. Rows are sorted by rowid and
// index entries by key, so the trees are packed bottom-up and never
// rebalanced. The file has no free pages, no WAL and UTF-8 text, and is a
// valid database for SQLite itself (see the tests' integrity check).
package sqlite

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
)

// PageSize is the page size of written files, as Anki uses.
const PageSize = 4096

const (
	pageIndexInterior = 0x02
	pageTableInterior = 0x05
	pageIndexLeaf     = 0x0a
	pageTableLeaf     = 0x0d

	// sqliteVersion is the library version number stored in the header.
	sqliteVersion = 3045000
)

// Writer builds a database file.
type Writer struct {
	tables  []*Table
	indexes []*index
	names   map[string]bool
	pages   [][]byte
}

// Table is a rowid table being written.
type Table struct {
	name    string
	sql     string
	columns []string
	// rowidCol is the INTEGER PRIMARY KEY column, or -1.
	rowidCol int
	rows     []row
	root     int
}

type row struct {
	rowid int64
	vals  []any
}

type index struct {
	name  string
	table *Table
	sql   string
	cols  []int
	root  int
}

// NewWriter returns an empty database.
func NewWriter() *Writer { return &Writer{names: map[string]bool{}} }

// CreateTable adds a table. Each definition is a column definition
// ("id integer PRIMARY KEY", "guid text NOT NULL") whose first word is the
// column name. An "integer primary key" column is the rowid alias: its
// value in Insert is the row's rowid.
func (w *Writer) CreateTable(name string, defs ...string) (*Table, error) {
	if err := w.claim(name); err != nil {
		return nil, err
	}
	if len(defs) == 0 {
		return nil, fmt.Errorf("sqlite: table %s has no columns", name)
	}
	t := &Table{name: name, rowidCol: -1,
		sql: "CREATE TABLE " + name + " (\n  " + strings.Join(defs, ",\n  ") + "\n)"}
	for i, d := range defs {
		f := strings.Fields(d)
		if len(f) == 0 {
			return nil, fmt.Errorf("sqlite: table %s has an empty column definition", name)
		}
		t.columns = append(t.columns, f[0])
		if strings.Contains(strings.ToLower(strings.Join(f[1:], " ")), "integer primary key") {
			t.rowidCol = i
		}
	}
	w.tables = append(w.tables, t)
	return t, nil
}

// CreateIndex adds an index on columns of table t.
func (w *Writer) CreateIndex(name string, t *Table, columns ...string) error {
	if err := w.claim(name); err != nil {
		return err
	}
	ix := &index{name: name, table: t,
		sql: "CREATE INDEX " + name + " ON " + t.name + " (" + strings.Join(columns, ", ") + ")"}
	for _, c := range columns {
		i := t.column(c)
		if i < 0 {
			return fmt.Errorf("sqlite: index %s: table %s has no column %s", name, t.name, c)
		}
		ix.cols = append(ix.cols, i)
	}
	w.indexes = append(w.indexes, ix)
	return nil
}

func (w *Writer) claim(name string) error {
	key := strings.ToLower(name)
	if name == "" || w.names[key] {
		return fmt.Errorf("sqlite: duplicate or empty name %q", name)
	}
	w.names[key] = true
	return nil
}

func (t *Table) column(name string) int {
	for i, c := range t.columns {
		if strings.EqualFold(c, name) {
			return i
		}
	}
	return -1
}

// Insert adds a row with one value per column: nil, a bool, an integer, a
// float64, a string or a []byte. The rowid alias column takes an integer,
// the rowid; tables without one number their rows 1, 2, … in insert order.
// Rows can be inserted in any order.
func (t *Table) Insert(values ...any) error {
	if len(values) != len(t.columns) {
		return fmt.Errorf("sqlite: %s: %d values for %d columns", t.name, len(values), len(t.columns))
	}
	vals := make([]any, len(values))
	for i, v := range values {
		n, err := normalize(v)
		if err != nil {
			return fmt.Errorf("sqlite: %s.%s: %w", t.name, t.columns[i], err)
		}
		vals[i] = n
	}
	rowid := int64(len(t.rows) + 1)
	if t.rowidCol >= 0 {
		id, ok := vals[t.rowidCol].(int64)
		if !ok {
			return fmt.Errorf("sqlite: %s: rowid must be an integer", t.name)
		}
		rowid = id
		vals[t.rowidCol] = nil // stored as NULL; readers use the rowid
	}
	t.rows = append(t.rows, row{rowid, vals})
	return nil
}

// Len returns the number of rows inserted.
func (t *Table) Len() int { return len(t.rows) }

func normalize(v any) (any, error) {
	switch x := v.(type) {
	case nil, int64, float64, string, []byte:
		return x, nil
	case bool:
		if x {
			return int64(1), nil
		}
		return int64(0), nil
	case int:
		return int64(x), nil
	case int32:
		return int64(x), nil
	case int16:
		return int64(x), nil
	case int8:
		return int64(x), nil
	case uint8:
		return int64(x), nil
	case uint16:
		return int64(x), nil
	case uint32:
		return int64(x), nil
	case float32:
		return float64(x), nil
	}
	return nil, fmt.Errorf("unsupported value type %T", v)
}

// Bytes builds the database file. yield is called periodically while large
// tables are encoded; it may be nil. The Writer must not be used afterwards.
func (w *Writer) Bytes(yield func()) ([]byte, error) {
	if yield == nil {
		yield = func() {}
	}
	w.pages = [][]byte{nil} // page 1, written last
	for _, t := range w.tables {
		sort.SliceStable(t.rows, func(i, j int) bool { return t.rows[i].rowid < t.rows[j].rowid })
		cells := make([]tableCell, len(t.rows))
		for i, r := range t.rows {
			if i > 0 && r.rowid == t.rows[i-1].rowid {
				return nil, fmt.Errorf("sqlite: %s: duplicate rowid %d", t.name, r.rowid)
			}
			cells[i] = tableCell{rowid: r.rowid, cell: w.tableLeafCell(r.rowid, encodeRecord(r.vals))}
			if i%2000 == 1999 {
				yield()
			}
		}
		t.root = w.place(w.buildTable(cells))
	}
	for _, ix := range w.indexes {
		keys := make([][]any, len(ix.table.rows))
		for i, r := range ix.table.rows {
			k := make([]any, 0, len(ix.cols)+1)
			for _, c := range ix.cols {
				v := r.vals[c]
				if c == ix.table.rowidCol {
					v = r.rowid
				}
				k = append(k, v)
			}
			keys[i] = append(k, r.rowid)
		}
		sort.SliceStable(keys, func(i, j int) bool { return compareKeys(keys[i], keys[j]) < 0 })
		cells := make([][]byte, len(keys))
		for i, k := range keys {
			cells[i] = encodeRecord(k)
			if i%2000 == 1999 {
				yield()
			}
		}
		ix.root = w.place(w.buildIndex(cells))
	}

	var schema []tableCell
	add := func(typ, name, tbl string, root int, sql string) {
		rowid := int64(len(schema) + 1)
		rec := encodeRecord([]any{typ, name, tbl, int64(root), sql})
		schema = append(schema, tableCell{rowid: rowid, cell: w.tableLeafCell(rowid, rec)})
	}
	for _, t := range w.tables {
		add("table", t.name, t.name, t.root, t.sql)
	}
	for _, ix := range w.indexes {
		add("index", ix.name, ix.table.name, ix.root, ix.sql)
	}
	root := w.buildTable(schema)
	if !root.fits(100) && len(root.cells) < 2 {
		return nil, errors.New("sqlite: schema entry too large")
	}
	if !root.fits(100) {
		// The root does not fit next to the file header: split it under a
		// new root.
		a, b := root.split()
		root = &node{kind: pageTableInterior,
			cells: [][]byte{interiorTableCell(w.place(a), a.maxKey)}, right: w.place(b), maxKey: b.maxKey}
	}
	w.pages[0] = root.encode(100)
	w.header(w.pages[0])

	out := make([]byte, 0, len(w.pages)*PageSize)
	for _, p := range w.pages {
		out = append(out, p...)
	}
	return out, nil
}

func (w *Writer) header(p []byte) {
	copy(p, "SQLite format 3\x00")
	binary.BigEndian.PutUint16(p[16:], PageSize)
	p[18], p[19] = 1, 1 // legacy (rollback journal) file format
	p[20] = 0           // reserved bytes per page
	p[21], p[22], p[23] = 64, 32, 32
	binary.BigEndian.PutUint32(p[24:], 1) // file change counter
	binary.BigEndian.PutUint32(p[28:], uint32(len(w.pages)))
	binary.BigEndian.PutUint32(p[40:], 1) // schema cookie
	binary.BigEndian.PutUint32(p[44:], 4) // schema format
	binary.BigEndian.PutUint32(p[56:], 1) // UTF-8
	binary.BigEndian.PutUint32(p[92:], 1) // version-valid-for = change counter
	binary.BigEndian.PutUint32(p[96:], sqliteVersion)
}

// alloc adds a page and returns its number.
func (w *Writer) alloc(p []byte) int {
	w.pages = append(w.pages, p)
	return len(w.pages)
}

// place writes a node to a new page and returns its number.
func (w *Writer) place(n *node) int { return w.alloc(n.encode(0)) }

// Payload limits for a usable page size of PageSize (file format spec,
// "B-tree Page Format").
const (
	usable        = PageSize
	tableMaxLocal = usable - 35
	indexMaxLocal = (usable-12)*64/255 - 23
	minLocal      = (usable-12)*32/255 - 23
)

// localSize returns how many bytes of a payload stay on the B-tree page.
func localSize(size, maxLocal int) int {
	if size <= maxLocal {
		return size
	}
	k := minLocal + (size-minLocal)%(usable-4)
	if k <= maxLocal {
		return k
	}
	return minLocal
}

// spill keeps the local part of a payload and writes the rest to overflow
// pages, returning the local part followed by the first overflow page
// number when there is one.
func (w *Writer) spill(payload []byte, maxLocal int) []byte {
	n := localSize(len(payload), maxLocal)
	if n == len(payload) {
		return payload
	}
	rest := payload[n:]
	var chunks [][]byte
	for len(rest) > 0 {
		k := min(len(rest), usable-4)
		chunks = append(chunks, rest[:k])
		rest = rest[k:]
	}
	// Overflow pages are numbered consecutively; each points to the next.
	first := len(w.pages) + 1
	for i, c := range chunks {
		p := make([]byte, PageSize)
		if i+1 < len(chunks) {
			binary.BigEndian.PutUint32(p, uint32(first+i+1))
		}
		copy(p[4:], c)
		w.alloc(p)
	}
	out := make([]byte, n+4)
	copy(out, payload[:n])
	binary.BigEndian.PutUint32(out[n:], uint32(first))
	return out
}

func (w *Writer) tableLeafCell(rowid int64, rec []byte) []byte {
	var b []byte
	b = appendVarint(b, uint64(len(rec)))
	b = appendVarint(b, uint64(rowid))
	return append(b, w.spill(rec, tableMaxLocal)...)
}

func interiorTableCell(child int, key int64) []byte {
	b := binary.BigEndian.AppendUint32(nil, uint32(child))
	return appendVarint(b, uint64(key))
}

// indexCell is the cell of an index entry: on a leaf page (child 0) or on
// an interior page with its left child.
func (w *Writer) indexCell(child int, rec []byte) []byte {
	var b []byte
	if child != 0 {
		b = binary.BigEndian.AppendUint32(b, uint32(child))
	}
	b = appendVarint(b, uint64(len(rec)))
	return append(b, w.spill(rec, indexMaxLocal)...)
}

// node is a B-tree page before it is written.
type node struct {
	kind  byte
	cells [][]byte
	// right is the rightmost child of an interior page.
	right int
	// maxKey is the largest rowid in a table subtree.
	maxKey int64
	size   int
	// keys and lefts are an index node's entries (records and left
	// children), kept so it can be split.
	keys  [][]byte
	lefts []int
}

func headerSize(kind byte) int {
	if kind == pageTableInterior || kind == pageIndexInterior {
		return 12
	}
	return 8
}

// fits reports whether the node fits a page after offset header bytes.
func (n *node) fits(offset int) bool {
	return offset+headerSize(n.kind)+2*len(n.cells)+n.size <= PageSize
}

func (n *node) add(cell []byte) {
	n.cells = append(n.cells, cell)
	n.size += len(cell)
}

// room reports whether a cell of size bytes still fits on a normal page.
func (n *node) room(size int) bool {
	return headerSize(n.kind)+2*(len(n.cells)+1)+n.size+size <= PageSize
}

// encode writes the node as a page whose B-tree header starts at offset.
func (n *node) encode(offset int) []byte {
	p := make([]byte, PageSize)
	h := p[offset:]
	h[0] = n.kind
	binary.BigEndian.PutUint16(h[3:], uint16(len(n.cells)))
	ptrs := offset + headerSize(n.kind)
	if n.kind == pageTableInterior || n.kind == pageIndexInterior {
		binary.BigEndian.PutUint32(h[8:], uint32(n.right))
	}
	end := PageSize
	for i, c := range n.cells {
		end -= len(c)
		copy(p[end:], c)
		binary.BigEndian.PutUint16(p[ptrs+2*i:], uint16(end))
	}
	// A content start of 65536 is stored as 0; with 4096-byte pages it
	// never happens.
	binary.BigEndian.PutUint16(h[5:], uint16(end))
	return p
}

// split divides a table node's cells in two halves. Only used for the
// schema root, which is a table leaf or interior node.
func (n *node) split() (*node, *node) {
	if n.kind == pageTableLeaf {
		a, b := &node{kind: n.kind}, &node{kind: n.kind, maxKey: n.maxKey}
		half := len(n.cells) / 2
		for _, c := range n.cells[:half] {
			a.add(c)
		}
		for _, c := range n.cells[half:] {
			b.add(c)
		}
		a.maxKey = leafCellRowid(n.cells[half-1])
		return a, b
	}
	// Interior: the left half's last child becomes its right pointer.
	half := len(n.cells) / 2
	a, b := &node{kind: n.kind}, &node{kind: n.kind, right: n.right, maxKey: n.maxKey}
	for _, c := range n.cells[:half] {
		a.add(c)
	}
	last := n.cells[half]
	a.right = int(binary.BigEndian.Uint32(last))
	key, _ := readVarint(last[4:])
	a.maxKey = int64(key)
	for _, c := range n.cells[half+1:] {
		b.add(c)
	}
	return a, b
}

func leafCellRowid(cell []byte) int64 {
	_, k := readVarint(cell)
	rowid, _ := readVarint(cell[k:])
	return int64(rowid)
}

type tableCell struct {
	rowid int64
	cell  []byte
}

// buildTable packs sorted leaf cells into a table B-tree. Every node but
// the root is placed on a page; the root is returned unplaced.
func (w *Writer) buildTable(cells []tableCell) *node {
	var level []*node
	cur := &node{kind: pageTableLeaf}
	for _, c := range cells {
		if len(cur.cells) > 0 && !cur.room(len(c.cell)) {
			level = append(level, cur)
			cur = &node{kind: pageTableLeaf}
		}
		cur.add(c.cell)
		cur.maxKey = c.rowid
	}
	level = append(level, cur)
	for len(level) > 1 {
		// Each child but a node's last gets a cell keyed by its largest
		// rowid; the last is the node's right pointer.
		var up []*node
		cur := &node{kind: pageTableInterior}
		for _, child := range level {
			page := w.place(child)
			if cur.right != 0 {
				cell := interiorTableCell(cur.right, cur.maxKey)
				if cur.room(len(cell)) {
					cur.add(cell)
				} else {
					up = append(up, cur)
					cur = &node{kind: pageTableInterior}
				}
			}
			cur.right, cur.maxKey = page, child.maxKey
		}
		if len(cur.cells) == 0 && len(up) > 0 {
			// A node with only a right pointer: move the previous node's
			// last child over.
			prev := up[len(up)-1]
			last := prev.cells[len(prev.cells)-1]
			prev.cells, prev.size = prev.cells[:len(prev.cells)-1], prev.size-len(last)
			cur.add(interiorTableCell(prev.right, prev.maxKey))
			prev.right = int(binary.BigEndian.Uint32(last))
			key, _ := readVarint(last[4:])
			prev.maxKey = int64(key)
		}
		up = append(up, cur)
		level = up
	}
	return level[0]
}

// buildIndex packs sorted index records into an index B-tree. Unlike table
// trees, interior pages hold entries themselves: between two neighbouring
// pages, one entry moves up to their parent. The root is returned unplaced.
func (w *Writer) buildIndex(recs [][]byte) *node {
	// An entry of a level is a record with the page left of it (0 on the
	// leaf level); right is the level's rightmost page.
	type level struct {
		keys  [][]byte
		lefts []int
		right int
	}
	cur := level{keys: recs, lefts: make([]int, len(recs))}
	leaf := true
	for {
		kind := byte(pageIndexInterior)
		if leaf {
			kind = pageIndexLeaf
		}
		// Greedily fill nodes; the entry that does not fit moves up.
		var nodes []*node
		var up level
		n := &node{kind: kind}
		for i, k := range cur.keys {
			cell := w.indexCellSize(cur.lefts[i], k)
			if len(n.keys) > 0 && !n.room(cell) {
				nodes = append(nodes, n)
				up.keys = append(up.keys, k)
				up.lefts = append(up.lefts, cur.lefts[i]) // becomes n's right child
				n = &node{kind: kind}
				continue
			}
			n.keys = append(n.keys, k)
			n.lefts = append(n.lefts, cur.lefts[i])
			n.size += cell
			n.cells = append(n.cells, nil)
		}
		if len(n.keys) == 0 && len(nodes) > 0 {
			// The last entry moved up and left an empty node: move the
			// previous node's last entry up instead, and the entry back down.
			prev := nodes[len(nodes)-1]
			j := len(prev.keys) - 1
			k, l := up.keys[len(up.keys)-1], up.lefts[len(up.lefts)-1]
			up.keys[len(up.keys)-1], up.lefts[len(up.lefts)-1] = prev.keys[j], prev.lefts[j]
			prev.keys, prev.lefts = prev.keys[:j], prev.lefts[:j]
			prev.cells = prev.cells[:j]
			n.keys, n.lefts, n.cells = [][]byte{k}, []int{l}, [][]byte{nil}
		}
		nodes = append(nodes, n)
		// The right child of each node is the left page of the entry after
		// it; the last node's is the level's right page.
		for i, nd := range nodes {
			if i+1 < len(nodes) {
				nd.right = up.lefts[i]
			} else {
				nd.right = cur.right
			}
			nd.cells, nd.size = nil, 0
			for j, k := range nd.keys {
				nd.add(w.indexCell(nd.lefts[j], k))
			}
		}
		if len(nodes) == 1 {
			return nodes[0]
		}
		for i, nd := range nodes {
			page := w.place(nd)
			if i < len(up.lefts) {
				up.lefts[i] = page
			} else {
				up.right = page
			}
		}
		cur = up
		leaf = false
	}
}

// indexCellSize returns the size of an index cell without writing its
// overflow pages.
func (w *Writer) indexCellSize(child int, rec []byte) int {
	n := len(appendVarint(nil, uint64(len(rec)))) + localSize(len(rec), indexMaxLocal)
	if localSize(len(rec), indexMaxLocal) < len(rec) {
		n += 4 // first overflow page
	}
	if child != 0 {
		n += 4
	}
	return n
}

// encodeRecord encodes values in SQLite's record format.
func encodeRecord(vals []any) []byte {
	var types, body []byte
	for _, v := range vals {
		switch x := v.(type) {
		case nil:
			types = appendVarint(types, 0)
		case int64:
			st, b := intSerial(x)
			types = appendVarint(types, st)
			body = append(body, b...)
		case float64:
			types = appendVarint(types, 7)
			body = binary.BigEndian.AppendUint64(body, math.Float64bits(x))
		case string:
			types = appendVarint(types, uint64(len(x))*2+13)
			body = append(body, x...)
		case []byte:
			types = appendVarint(types, uint64(len(x))*2+12)
			body = append(body, x...)
		}
	}
	// The header size includes its own varint, whose length depends on it.
	hs := len(types) + 1
	for len(appendVarint(nil, uint64(hs)))+len(types) != hs {
		hs++
	}
	out := appendVarint(make([]byte, 0, hs+len(body)), uint64(hs))
	out = append(out, types...)
	return append(out, body...)
}

// intSerial returns the serial type and big-endian bytes of an integer in
// the smallest encoding.
func intSerial(v int64) (uint64, []byte) {
	switch {
	case v == 0:
		return 8, nil
	case v == 1:
		return 9, nil
	case v >= -1<<7 && v < 1<<7:
		return 1, []byte{byte(v)}
	case v >= -1<<15 && v < 1<<15:
		return 2, binary.BigEndian.AppendUint16(nil, uint16(v))
	case v >= -1<<23 && v < 1<<23:
		return 3, []byte{byte(v >> 16), byte(v >> 8), byte(v)}
	case v >= -1<<31 && v < 1<<31:
		return 4, binary.BigEndian.AppendUint32(nil, uint32(v))
	case v >= -1<<47 && v < 1<<47:
		b := binary.BigEndian.AppendUint64(nil, uint64(v))
		return 5, b[2:]
	}
	return 6, binary.BigEndian.AppendUint64(nil, uint64(v))
}

// appendVarint appends an SQLite varint: big-endian groups of 7 bits with
// the high bit set on all but the last byte, and a full ninth byte.
func appendVarint(b []byte, v uint64) []byte {
	if v > 1<<56-1 {
		var buf [9]byte
		buf[8] = byte(v)
		v >>= 8
		for i := 7; i >= 0; i-- {
			buf[i] = byte(v&0x7f) | 0x80
			v >>= 7
		}
		return append(b, buf[:]...)
	}
	var buf [9]byte
	i := 8
	buf[i] = byte(v & 0x7f)
	v >>= 7
	for v > 0 {
		i--
		buf[i] = byte(v&0x7f) | 0x80
		v >>= 7
	}
	return append(b, buf[i:]...)
}

func readVarint(b []byte) (uint64, int) {
	var v uint64
	for i := 0; i < 9 && i < len(b); i++ {
		if i == 8 {
			return v<<8 | uint64(b[i]), 9
		}
		v = v<<7 | uint64(b[i]&0x7f)
		if b[i]&0x80 == 0 {
			return v, i + 1
		}
	}
	return v, len(b)
}

// compareKeys orders index keys as SQLite does with the BINARY collation:
// NULL, then numbers by value, then text, then blobs, column by column.
func compareKeys(a, b []any) int {
	for i := range min(len(a), len(b)) {
		if c := compareValues(a[i], b[i]); c != 0 {
			return c
		}
	}
	return len(a) - len(b)
}

func compareValues(a, b any) int {
	ra, rb := rank(a), rank(b)
	if ra != rb {
		return ra - rb
	}
	switch x := a.(type) {
	case int64:
		switch y := b.(type) {
		case int64:
			return cmp(x, y)
		case float64:
			return cmpFloat(float64(x), y)
		}
	case float64:
		switch y := b.(type) {
		case int64:
			return cmpFloat(x, float64(y))
		case float64:
			return cmpFloat(x, y)
		}
	case string:
		return strings.Compare(x, b.(string))
	case []byte:
		return bytes.Compare(x, b.([]byte))
	}
	return 0
}

func rank(v any) int {
	switch v.(type) {
	case nil:
		return 0
	case int64, float64:
		return 1
	case string:
		return 2
	}
	return 3
}

func cmp(a, b int64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

func cmpFloat(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}
