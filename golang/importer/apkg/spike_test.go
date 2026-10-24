package apkg

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/klauspost/compress/zstd"
	"google.golang.org/protobuf/encoding/protowire"

	"cardflip/golang/importer/apkg/sqlite"
)

// The reading spike (task 2.2): every fixture opens, its collection file is
// found and decompressed, the SQLite reader scans notes and cards, and a
// modern template config decodes as protobuf.

type fixtureManifest struct {
	Packages map[string]struct {
		Format string            `json:"format"`
		Notes  map[string]any    `json:"notes"`
		Cards  []map[string]any  `json:"cards"`
		Media  map[string]string `json:"media"`
		// Revlog is "<guid>/<ord>" -> [id, ease, ivl, lastIvl, factor,
		// time, type] per entry.
		Revlog map[string][][7]int64 `json:"revlog"`
		// DeckPresets is deck name -> the preset it uses.
		DeckPresets map[string]fixturePreset `json:"deckPresets"`
		// Decks and Notetypes are the Fixture decks and the used note
		// types of a deck package, and all of them in a collection package.
		Decks         []string `json:"decks"`
		FilteredDecks []string `json:"filteredDecks"`
		Notetypes     []struct {
			Name   string   `json:"name"`
			Fields []string `json:"fields"`
		} `json:"notetypes"`
	} `json:"packages"`
	// Updates are the update packages, with modification times in epoch
	// seconds.
	Updates map[string]struct {
		Notetypes []struct {
			Name   string   `json:"name"`
			Fields []string `json:"fields"`
			Mod    int64    `json:"mod"`
		} `json:"notetypes"`
		Notes map[string]struct {
			Fields []string `json:"fields"`
			Mod    int64    `json:"mod"`
		} `json:"notes"`
	} `json:"updates"`
	// Latex holds the packages with pre-generated LaTeX images.
	Latex struct {
		Packages map[string]struct {
			Media     map[string]string `json:"media"`
			Notetypes []struct {
				Name     string `json:"name"`
				LatexSVG bool   `json:"latexSvg"`
			} `json:"notetypes"`
		} `json:"packages"`
		// Names are the image names Anki gives each field's LaTeX blocks,
		// by note GUID.
		Names map[string][][]string `json:"names"`
		// Missing is the note whose images are not in the media.
		Missing string `json:"missing"`
	} `json:"latex"`
	// Occlusion holds the packages of Image Occlusion notes.
	Occlusion map[string]struct {
		Notes map[string]struct {
			Notetype string   `json:"notetype"`
			Fields   []string `json:"fields"`
		} `json:"notes"`
		Cards []struct {
			Note string `json:"note"`
			Ord  int    `json:"ord"`
		} `json:"cards"`
		Media     map[string]string `json:"media"`
		Notetypes []struct {
			Name   string   `json:"name"`
			Fields []string `json:"fields"`
		} `json:"notetypes"`
	} `json:"occlusion"`
}

// fixturePreset is a preset in fixtures.json, with Anki's values and enums.
type fixturePreset struct {
	Name                      string    `json:"name"`
	NewPerDay                 int       `json:"newPerDay"`
	ReviewsPerDay             int       `json:"reviewsPerDay"`
	LearnSteps                []float64 `json:"learnSteps"`
	RelearnSteps              []float64 `json:"relearnSteps"`
	DesiredRetention          float64   `json:"desiredRetention"`
	MaxIvl                    int       `json:"maxIvl"`
	MinInt                    int       `json:"minInt"`
	LeechThreshold            int       `json:"leechThreshold"`
	LeechAction               int       `json:"leechAction"`
	InsertOrder               int       `json:"insertOrder"`
	GatherPriority            int       `json:"gatherPriority"`
	SortOrder                 int       `json:"sortOrder"`
	NewMix                    int       `json:"newMix"`
	InterdayMix               int       `json:"interdayMix"`
	ReviewOrder               int       `json:"reviewOrder"`
	BuryNew                   bool      `json:"buryNew"`
	BuryReviews               bool      `json:"buryReviews"`
	BuryInterdayLearning      bool      `json:"buryInterdayLearning"`
	DisableAutoplay           bool      `json:"disableAutoplay"`
	SkipQuestionWhenReplaying bool      `json:"skipQuestionWhenReplaying"`
	SecondsToShowQuestion     float64   `json:"secondsToShowQuestion"`
	SecondsToShowAnswer       float64   `json:"secondsToShowAnswer"`
	QuestionAction            int       `json:"questionAction"`
	AnswerAction              int       `json:"answerAction"`
}

func loadManifest(t *testing.T) fixtureManifest {
	t.Helper()
	data, err := os.ReadFile("testdata/fixtures.json")
	if err != nil {
		t.Fatal(err)
	}
	var m fixtureManifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func spikeOpen(path string) (*sqlite.DB, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, "", err
	}
	read := func(name string) ([]byte, error) {
		f, err := zr.Open(name)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return io.ReadAll(f)
	}
	for _, name := range []string{"collection.anki21b", "collection.anki21", "collection.anki2"} {
		raw, err := read(name)
		if err != nil {
			continue
		}
		if name == "collection.anki21b" {
			dec, err := zstd.NewReader(nil)
			if err != nil {
				return nil, "", err
			}
			raw, err = dec.DecodeAll(raw, nil)
			dec.Close()
			if err != nil {
				return nil, "", err
			}
		}
		db, err := sqlite.Open(raw)
		return db, name, err
	}
	return nil, "", os.ErrNotExist
}

func TestSpikeFixtureRowCounts(t *testing.T) {
	m := loadManifest(t)
	for name, want := range m.Packages {
		db, file, err := spikeOpen("testdata/" + name)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if file != "collection."+want.Format {
			t.Errorf("%s: picked %s, want collection.%s", name, file, want.Format)
		}
		notes, err := db.Count("notes")
		if err != nil {
			t.Fatalf("%s notes: %v", name, err)
		}
		cards, err := db.Count("cards")
		if err != nil {
			t.Fatalf("%s cards: %v", name, err)
		}
		t.Logf("%s (%s): %d notes, %d cards", name, file, notes, cards)
		if notes != len(want.Notes) || cards != len(want.Cards) {
			t.Errorf("%s: %d notes / %d cards, want %d / %d", name, notes, cards, len(want.Notes), len(want.Cards))
		}
	}
}

func TestSpikeTemplateConfigProtobuf(t *testing.T) {
	db, _, err := spikeOpen("testdata/modern.apkg")
	if err != nil {
		t.Fatal(err)
	}
	basic := int64(0)
	err = db.Scan("notetypes", func(r sqlite.Row) error {
		if r.Str("name") == "Basic" {
			basic = r.Int("id")
		}
		return nil
	})
	if err != nil || basic == 0 {
		t.Fatalf("Basic notetype not found (err %v)", err)
	}
	var qfmt, afmt string
	err = db.Scan("templates", func(r sqlite.Row) error {
		if r.Int("ntid") != basic || r.Int("ord") != 0 {
			return nil
		}
		// Notetype.Template.Config: q_format = 1, a_format = 2.
		b := r.Bytes("config")
		for len(b) > 0 {
			num, typ, n := protowire.ConsumeTag(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			b = b[n:]
			if typ == protowire.BytesType && (num == 1 || num == 2) {
				v, n := protowire.ConsumeBytes(b)
				if n < 0 {
					return protowire.ParseError(n)
				}
				if num == 1 {
					qfmt = string(v)
				} else {
					afmt = string(v)
				}
				b = b[n:]
				continue
			}
			n = protowire.ConsumeFieldValue(num, typ, b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			b = b[n:]
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if qfmt != "{{Front}}" || afmt != "{{FrontSide}}\n\n<hr id=answer>\n\n{{Back}}" {
		t.Fatalf("q_format = %q, a_format = %q", qfmt, afmt)
	}
}

func TestSpikeTruncatedPackageFails(t *testing.T) {
	if _, _, err := spikeOpen("testdata/truncated.apkg"); err == nil {
		t.Fatal("truncated package opened without error")
	}
}
