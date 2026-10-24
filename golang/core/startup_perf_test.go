package core_test

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/importer/csv"
	"cardflip/golang/scheduler"
	"cardflip/golang/storage"
)

// startupFixture builds the collection of the 50,000-card perf fixture
// (integration_test/perf/stats_flow.dart) in dir: 50,000 Basic notes
// imported from a tab-separated file, 60 of them answered.
func startupFixture(t *testing.T, dir string) {
	t.Helper()
	b, err := storage.OpenBackend(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	c, _, err := core.Open(b, core.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	var tsv strings.Builder
	for i := range 50000 {
		fmt.Fprintf(&tsv, "Question %d about topic %d\tAnswer %d\n", i, i%97, i)
	}
	f, err := csv.Parse([]byte(tsv.String()))
	if err != nil {
		t.Fatal(err)
	}
	types := c.NoteTypes()
	basic := csv.NoteTypeNamed(types, "Basic")
	batch, err := csv.Build(f, csv.Options{TypeID: basic.ID, NoteTypes: types, DeckID: c.Decks()[0].ID,
		Mapping: csv.DefaultMapping(len(f.Rows[0]), len(basic.Fields)), Duplicates: core.DupDuplicate})
	if err != nil {
		t.Fatal(err)
	}
	if sum, err := c.Import(batch, nil); err != nil || sum.NotesAdded != 50000 {
		t.Fatalf("import: %+v, %v", sum, err)
	}
	ids, err := c.Search("", core.ColCreated, false)
	if err != nil {
		t.Fatal(err)
	}
	for i, id := range ids[:60] {
		if _, err := c.Answer(id, scheduler.Rating(1+i%4), 2*time.Second); err != nil {
			t.Fatal(err)
		}
	}
}

// openTime opens the collection in dir and returns how long it took.
func openTime(t *testing.T, dir string) time.Duration {
	t.Helper()
	start := time.Now()
	b, err := storage.OpenBackend(dir)
	if err != nil {
		t.Fatal(err)
	}
	c, _, err := core.Open(b, core.Options{})
	if err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)
	if n := len(c.Decks()); n == 0 {
		t.Fatal("no decks")
	}
	c.Close()
	b.Close()
	return elapsed
}

func copyDir(t *testing.T, from, to string) {
	t.Helper()
	err := filepath.WalkDir(from, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(from, path)
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(to, rel), 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(to, rel), data, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}

// Opening the perf fixture's 50,000 cards stays under 3 seconds, the
// collection's startup target. CARDFLIP_STARTUP_DIR names a data directory
// to measure instead (such as one written by an earlier version, to measure
// its migration), and CARDFLIP_STARTUP_KEEP one to copy the fixture to.
func TestStartup50k(t *testing.T) {
	if testing.Short() {
		t.Skip("large collection test")
	}
	dir := filepath.Join(t.TempDir(), "profile")
	if from := os.Getenv("CARDFLIP_STARTUP_DIR"); from != "" {
		copyDir(t, from, dir)
	} else {
		startupFixture(t, dir)
		// The first open compacts the journal of the import, as the app's
		// first start after it would.
		openTime(t, dir)
	}
	if keep := os.Getenv("CARDFLIP_STARTUP_KEEP"); keep != "" {
		copyDir(t, dir, keep)
	}
	first := openTime(t, dir)
	var times []time.Duration
	for range 5 {
		times = append(times, openTime(t, dir))
	}
	slices.Sort(times)
	t.Logf("first open %v, then median %v (min %v, max %v)", first, times[2], times[0], times[4])
	if first > 3*time.Second || times[2] > 3*time.Second {
		t.Fatalf("opening 50,000 cards took %v (first) and %v (median), want < 3s", first, times[2])
	}
}
