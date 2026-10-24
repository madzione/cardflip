//go:build !js

package apkg

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"cardflip/golang/core"
)

// Every collection file of every package is a valid SQLite database whose
// indexes answer the queries Anki makes (native test builds only).
func TestCollectionsPassIntegrityCheck(t *testing.T) {
	src := fixtureCollection(t)
	for _, o := range []Options{{}, {Scheduling: true, Presets: true, Media: true}, {Scheduling: true, Legacy: true}} {
		for name, data := range readZip(t, export(t, src, core.ExportScope{All: true}, o)) {
			if name != "collection.anki2" && name != "collection.anki21" {
				continue
			}
			path := filepath.Join(t.TempDir(), name)
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatal(err)
			}
			db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
			if err != nil {
				t.Fatal(err)
			}
			var res string
			if err := db.QueryRow("PRAGMA integrity_check").Scan(&res); err != nil || res != "ok" {
				t.Errorf("%+v %s: integrity_check = %q, %v", o, name, res, err)
			}
			var n int
			q := "SELECT count(*) FROM cards INDEXED BY ix_cards_sched WHERE did > 0 AND queue >= -3"
			if err := db.QueryRow(q).Scan(&n); err != nil || n == 0 {
				t.Errorf("%+v %s: cards by ix_cards_sched = %d, %v", o, name, n, err)
			}
			db.Close()
		}
	}
}
