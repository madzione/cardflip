//go:build !js

package sqlite

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// openReference opens a written file with a real SQLite (pure Go, so no
// cgo is needed; native test builds only).
func openReference(t *testing.T, data []byte) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func integrityCheck(t *testing.T, db *sql.DB) {
	t.Helper()
	rows, err := db.Query("PRAGMA integrity_check")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var msgs []string
	for rows.Next() {
		var m string
		if err := rows.Scan(&m); err != nil {
			t.Fatal(err)
		}
		msgs = append(msgs, m)
	}
	if strings.Join(msgs, "\n") != "ok" {
		t.Fatalf("integrity_check:\n%s", strings.Join(msgs, "\n"))
	}
}

func TestReferenceSQLite(t *testing.T) {
	for _, n := range []int{0, 1, 100_000} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			db := openReference(t, writeTestDB(t, n))
			integrityCheck(t, db)
			var count int
			if err := db.QueryRow("SELECT count(*) FROM items").Scan(&count); err != nil || count != n {
				t.Fatalf("count = %d, %v", count, err)
			}
			if n == 0 {
				return
			}
			// Look rows up through each index and compare every value.
			for _, i := range []int{0, n / 3, n - 1} {
				want := testRow(i)
				var (
					id, num, flag int64
					f             float64
					s             string
					b             []byte
				)
				q := "SELECT id, n, f, s, b, flag FROM items INDEXED BY ix_items_s WHERE s = ?"
				if err := db.QueryRow(q, want[3]).Scan(&id, &num, &f, &s, &b, &flag); err != nil {
					t.Fatalf("row %d by s: %v", i, err)
				}
				wb, _ := want[4].([]byte)
				if id != want[0].(int64) || num != want[1].(int64) || f != want[2].(float64) || !bytes.Equal(b, wb) || (flag == 1) != want[5].(bool) {
					t.Fatalf("row %d = %d %d %v len(b)=%d %d", i, id, num, f, len(b), flag)
				}
				var got int64
				q = "SELECT count(*) FROM items INDEXED BY ix_items_nf WHERE n = ? AND f = ?"
				if err := db.QueryRow(q, want[1], want[2]).Scan(&got); err != nil || got != 1 {
					t.Fatalf("row %d by (n, f): %d, %v", i, got, err)
				}
				if err := db.QueryRow("SELECT s FROM items WHERE id = ?", want[0]).Scan(&s); err != nil || s != want[3] {
					t.Fatalf("row %d by rowid: %v", i, err)
				}
			}
		})
	}
}

func TestReferenceLargeSchema(t *testing.T) {
	// Enough tables that the schema no longer fits on page 1.
	w := NewWriter()
	for i := range 80 {
		tbl, err := w.CreateTable(fmt.Sprintf("table_%02d", i),
			"id integer PRIMARY KEY", "value text NOT NULL -- "+strings.Repeat("c", 100))
		if err != nil {
			t.Fatal(err)
		}
		_ = tbl.Insert(int64(i), fmt.Sprint("v", i))
		if err := w.CreateIndex(fmt.Sprintf("ix_%02d", i), tbl, "value"); err != nil {
			t.Fatal(err)
		}
	}
	data, err := w.Bytes(nil)
	if err != nil {
		t.Fatal(err)
	}
	db := openReference(t, data)
	integrityCheck(t, db)
	var v string
	if err := db.QueryRow("SELECT value FROM table_79 WHERE id = 79").Scan(&v); err != nil || v != "v79" {
		t.Fatalf("value = %q, %v", v, err)
	}
}
