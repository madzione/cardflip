package sqlite

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"testing"
)

// testdata/test.db is written by testdata/make_test_db.py.
func openTestDB(t *testing.T) (*DB, []byte) {
	t.Helper()
	data, err := os.ReadFile("testdata/test.db")
	if err != nil {
		t.Fatal(err)
	}
	db, err := Open(data)
	if err != nil {
		t.Fatal(err)
	}
	return db, data
}

func TestRowidTableWithOverflow(t *testing.T) {
	db, _ := openTestDB(t)
	cols, _ := db.Columns("items")
	if strings.Join(cols, ",") != "id,name,score,data" {
		t.Fatalf("columns = %v", cols)
	}
	seen := 0
	err := db.Scan("items", func(r Row) error {
		seen++
		id := r.Int("id")
		if id%7 != 0 {
			return fmt.Errorf("id %d is not the rowid alias", id)
		}
		i := id / 7
		if want := fmt.Sprintf("item-%d-", i) + strings.Repeat("x", int(i%50)); r.Str("name") != want {
			return fmt.Errorf("row %d name = %q", i, r.Str("name"))
		}
		if want := float64(i)/4 - 100; r.Float("score") != want {
			return fmt.Errorf("row %d score = %v, want %v", i, r.Float("score"), want)
		}
		blob := r.Bytes("data")
		if i%100 != 0 {
			if blob != nil {
				return fmt.Errorf("row %d: unexpected blob", i)
			}
			return nil
		}
		want := make([]byte, 5000)
		for j := range want {
			want[j] = byte((int(i) + j) % 256)
		}
		if !bytes.Equal(blob, want) {
			return fmt.Errorf("row %d: overflow blob mismatch (len %d)", i, len(blob))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if seen != 1500 {
		t.Fatalf("scanned %d rows, want 1500", seen)
	}
}

func TestWithoutRowidTable(t *testing.T) {
	db, _ := openTestDB(t)
	got := map[int64]bool{}
	err := db.Scan("PAIRS", func(r Row) error {
		a := r.Int("a")
		if r.Int("b") != -a*1000003 {
			return fmt.Errorf("a=%d b=%d", a, r.Int("b"))
		}
		if want := strings.Repeat("p", int(a%300)) + fmt.Sprint(a); r.Str("label") != want {
			return fmt.Errorf("a=%d label=%q", a, r.Str("label"))
		}
		got[a] = true
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 800 {
		t.Fatalf("got %d distinct rows, want 800", len(got))
	}
}

func TestAddedColumnAndIntegerLimits(t *testing.T) {
	db, _ := openTestDB(t)
	var rows [][3]any
	err := db.Scan("plain", func(r Row) error {
		rows = append(rows, [3]any{r.Value("x"), r.Value("y"), r.Value("z")})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	want := [][3]any{
		{int64(1), int64(2), nil}, // written before ALTER TABLE: z is missing
		{int64(3), int64(4), "z4"},
		{int64(math.MinInt64), int64(math.MaxInt64), nil},
	}
	if fmt.Sprint(rows) != fmt.Sprint(want) {
		t.Fatalf("rows = %v, want %v", rows, want)
	}
}

func TestUnknownTable(t *testing.T) {
	db, _ := openTestDB(t)
	if db.HasTable("nope") {
		t.Fatal("HasTable(nope)")
	}
	if err := db.Scan("nope", func(Row) error { return nil }); err == nil {
		t.Fatal("expected error")
	}
}

func TestNotADatabase(t *testing.T) {
	for _, data := range [][]byte{nil, []byte("hello"), bytes.Repeat([]byte{0}, 4096)} {
		if _, err := Open(data); !errors.Is(err, ErrCorrupt) {
			t.Errorf("Open(%d bytes) err = %v, want ErrCorrupt", len(data), err)
		}
	}
}

// Truncated and bit-flipped databases must fail with an error, never panic.
func TestCorruptDataNeverPanics(t *testing.T) {
	_, data := openTestDB(t)
	scanAll := func(b []byte) {
		db, err := Open(b)
		if err != nil {
			return
		}
		for _, name := range []string{"items", "pairs", "plain"} {
			_ = db.Scan(name, func(Row) error { return nil })
		}
	}
	for _, n := range []int{100, 1024, 1500, 4096, len(data) / 2, len(data) - 1} {
		scanAll(data[:n])
	}
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 300; i++ {
		b := bytes.Clone(data)
		for j := 0; j < 20; j++ {
			b[rng.Intn(len(b))] = byte(rng.Intn(256))
		}
		scanAll(b)
	}
}

func TestParseCreateTable(t *testing.T) {
	cases := []struct {
		sql      string
		cols     string
		rowidCol int
		order    string
	}{
		{"CREATE TABLE col (\n  id integer PRIMARY KEY, -- c\n  crt integer NOT NULL)", "id,crt", 0, "[0 1]"},
		{"CREATE TABLE t (a text, id INTEGER, PRIMARY KEY (id))", "a,id", 1, "[0 1]"},
		{"CREATE TABLE t (a text primary key, b int)", "a,b", -1, "[0 1]"},
		{"CREATE TABLE config (\n  KEY text NOT NULL PRIMARY KEY,\n  usn integer NOT NULL\n) without rowid", "KEY,usn", -1, "[0 1]"},
		{"CREATE TABLE \"p\" (label text, a integer, b integer, primary key (b, a)) WITHOUT ROWID", "label,a,b", -1, "[2 1 0]"},
		{"CREATE TABLE sqlite_stat1(tbl,idx,stat)", "tbl,idx,stat", -1, "[0 1 2]"},
	}
	for _, c := range cases {
		tb, err := parseCreateTable(c.sql)
		if err != nil {
			t.Errorf("%s: %v", c.sql, err)
			continue
		}
		if strings.Join(tb.columns, ",") != c.cols || tb.rowidCol != c.rowidCol || fmt.Sprint(tb.order) != c.order {
			t.Errorf("%s: cols=%v rowid=%d order=%v", c.sql, tb.columns, tb.rowidCol, tb.order)
		}
	}
}
