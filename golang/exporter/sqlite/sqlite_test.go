package sqlite

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"testing"

	reader "cardflip/golang/importer/apkg/sqlite"
)

// testRow returns row i of the generated test table: integers of every
// encoding size, a float, text (long enough to overflow every 97th row),
// a blob on every 100th row and NULL otherwise.
func testRow(i int) []any {
	var blob any
	if i%100 == 0 {
		b := make([]byte, 5000+i%3000)
		for j := range b {
			b[j] = byte(i + j)
		}
		blob = b
	}
	text := fmt.Sprintf("row-%d-", i) + strings.Repeat("é", i%40)
	if i%97 == 0 {
		text += strings.Repeat("x", 9000)
	}
	ints := []int64{0, 1, -1, 127, -128, 32767, 1 << 20, -1 << 30, 1 << 40, math.MaxInt64, math.MinInt64}
	return []any{int64(i)*3 + 1, ints[i%len(ints)] - int64(i%2), float64(i)/8 - 50, text, blob, i%5 == 0}
}

var testColumns = []string{"id integer PRIMARY KEY", "n integer NOT NULL", "f real NOT NULL", "s text NOT NULL", "b blob", "flag integer NOT NULL"}

// writeTestDB writes a table "items" of n generated rows, inserted in
// reverse order, with indexes on (n, f) and on s, plus an empty table.
func writeTestDB(t testing.TB, n int) []byte {
	t.Helper()
	w := NewWriter()
	items, err := w.CreateTable("items", testColumns...)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.CreateTable("empty", "a integer NOT NULL", "b text"); err != nil {
		t.Fatal(err)
	}
	for i := n - 1; i >= 0; i-- {
		if err := items.Insert(testRow(i)...); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.CreateIndex("ix_items_nf", items, "n", "f"); err != nil {
		t.Fatal(err)
	}
	if err := w.CreateIndex("ix_items_s", items, "s"); err != nil {
		t.Fatal(err)
	}
	if err := w.CreateIndex("ix_items_id", items, "id"); err != nil {
		t.Fatal(err)
	}
	data, err := w.Bytes(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(data)%PageSize != 0 {
		t.Fatalf("file size %d is not a multiple of the page size", len(data))
	}
	return data
}

func TestRoundTripThroughReader(t *testing.T) {
	for _, n := range []int{0, 1, 100_000} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			db, err := reader.Open(writeTestDB(t, n))
			if err != nil {
				t.Fatal(err)
			}
			cols, _ := db.Columns("items")
			if strings.Join(cols, ",") != "id,n,f,s,b,flag" {
				t.Fatalf("columns = %v", cols)
			}
			i := 0
			err = db.Scan("items", func(r reader.Row) error {
				want := testRow(i)
				if r.Int("id") != want[0].(int64) || r.Int("n") != want[1].(int64) ||
					r.Float("f") != want[2].(float64) || r.Str("s") != want[3].(string) {
					return fmt.Errorf("row %d = %v %v %v %.20q", i, r.Int("id"), r.Int("n"), r.Float("f"), r.Str("s"))
				}
				if b, _ := want[4].([]byte); !bytes.Equal(r.Bytes("b"), b) || (b == nil) != (r.Value("b") == nil) {
					return fmt.Errorf("row %d: blob mismatch", i)
				}
				if (r.Int("flag") == 1) != want[5].(bool) {
					return fmt.Errorf("row %d: flag = %d", i, r.Int("flag"))
				}
				i++
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if i != n {
				t.Fatalf("scanned %d rows, want %d", i, n)
			}
			if c, err := db.Count("empty"); err != nil || c != 0 {
				t.Fatalf("empty table: %d rows, %v", c, err)
			}
		})
	}
}

func TestRowidOrderAndDefaults(t *testing.T) {
	w := NewWriter()
	tbl, _ := w.CreateTable("plain", "a text", "b integer")
	for _, v := range []string{"x", "y", "z"} {
		if err := tbl.Insert(v, len(v)); err != nil {
			t.Fatal(err)
		}
	}
	data, err := w.Bytes(nil)
	if err != nil {
		t.Fatal(err)
	}
	db, err := reader.Open(data)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	_ = db.Scan("plain", func(r reader.Row) error { got = append(got, r.Str("a")); return nil })
	if strings.Join(got, "") != "xyz" {
		t.Fatalf("rows = %v", got)
	}
}

func TestErrors(t *testing.T) {
	w := NewWriter()
	tbl, _ := w.CreateTable("t", "id integer PRIMARY KEY", "v text")
	if _, err := w.CreateTable("T", "a"); err == nil {
		t.Error("duplicate table name accepted")
	}
	if err := w.CreateIndex("ix", tbl, "missing"); err == nil {
		t.Error("index on a missing column accepted")
	}
	if err := tbl.Insert(int64(1)); err == nil {
		t.Error("wrong value count accepted")
	}
	if err := tbl.Insert("x", "y"); err == nil {
		t.Error("text rowid accepted")
	}
	if err := tbl.Insert(int64(1), struct{}{}); err == nil {
		t.Error("unsupported value accepted")
	}
	_ = tbl.Insert(int64(5), "a")
	_ = tbl.Insert(int64(5), "b")
	if _, err := w.Bytes(nil); err == nil {
		t.Error("duplicate rowid accepted")
	}
}

func TestVarint(t *testing.T) {
	for _, v := range []uint64{0, 1, 127, 128, 16383, 16384, 1<<56 - 1, 1 << 56, math.MaxUint64, 1<<63 + 12345} {
		b := appendVarint(nil, v)
		got, n := readVarint(b)
		if got != v || n != len(b) {
			t.Errorf("varint %d: got %d (%d of %d bytes)", v, got, n, len(b))
		}
	}
}
