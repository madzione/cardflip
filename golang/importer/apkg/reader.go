package apkg

import (
	"archive/zip"
	"bytes"
	"cardflip/golang/internal/usermsg"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/klauspost/compress/zstd"
	"google.golang.org/protobuf/encoding/protowire"

	"cardflip/golang/importer/apkg/sqlite"
)

// Package is the content of an Anki package, independent of its format.
type Package struct {
	// Format is the collection file used: anki21b, anki21 or anki2.
	Format string
	// Crt is the collection's creation time in epoch seconds; review due
	// days count from it.
	Crt   int64
	Decks map[int64]Deck
	// Presets are the deck options presets (Anki's deck configs) by ID.
	Presets   map[int64]Preset
	NoteTypes map[int64]NoteType
	Notes     []Note
	Cards     []Card
	// Revlog is the review history of the cards, in ID (time) order.
	Revlog []Review
	Media  map[string][]byte
}

type Deck struct {
	Name     string
	Filtered bool
	// PresetID is the deck's preset (Anki's deck config ID).
	PresetID int64
	// Description is the deck's description, HTML or (in decks that set
	// Anki's markdown option) markdown, kept as it is.
	Description string
}

// AnkiDefaultPreset is the ID of Anki's default preset.
const AnkiDefaultPreset = 1

// Preset is a deck options preset with the options Cardflip maps, in
// Anki's units and enums (deck_config.proto).
type Preset struct {
	Name                     string
	NewPerDay, ReviewsPerDay int
	// LearnSteps and RelearnSteps are in minutes.
	LearnSteps, RelearnSteps []float64
	DesiredRetention         float64
	// MaxInterval and MinLapseInterval are in days.
	MaxInterval, MinLapseInterval int
	LeechThreshold                int
	// LeechAction: 0 suspend, 1 tag only.
	LeechAction int
	// InsertOrder: 0 due (sequential), 1 random.
	InsertOrder int
	// GatherPriority, SortOrder, ReviewOrder are NewCardGatherPriority,
	// NewCardSortOrder and ReviewCardOrder; NewMix and InterdayMix are
	// ReviewMix (0 mix, 1 after reviews, 2 before reviews).
	GatherPriority, SortOrder, ReviewOrder, NewMix, InterdayMix int
	BuryNew, BuryReviews, BuryInterdayLearning                  bool
	DisableAutoplay, SkipQuestionWhenReplaying                  bool
	// SecondsToShowQuestion and SecondsToShowAnswer are auto advance's
	// times; QuestionAction (0 show answer, 1 show reminder) and
	// AnswerAction (0 bury card, 1 again, 2 good, 3 hard, 4 show reminder)
	// are Anki's enums.
	SecondsToShowQuestion, SecondsToShowAnswer float64
	QuestionAction, AnswerAction               int
}

type NoteType struct {
	Name   string
	Cloze  bool
	Fields []string
	// FieldConfigs are the fields' options, parallel to Fields.
	FieldConfigs []FieldConfig
	// SortField is the index of the field the browser sorts by.
	SortField int
	Templates []Template
	CSS       string
	// Mod is the modification time in epoch seconds, 0 when unknown.
	Mod int64
	// LatexSVG is set when the note type's LaTeX images are SVG.
	LatexSVG bool
}

// FieldConfig holds the options Anki stores for a field.
type FieldConfig struct {
	Sticky, RTL bool
	// Font and Size are the editing font and size.
	Font string
	Size int
	// PlainText opens the field in the HTML editor.
	PlainText, Collapsed, ExcludeFromSearch bool
}

type Template struct {
	Name, QFmt, AFmt string
	// BrowserQFmt and BrowserAFmt are the browser appearance templates,
	// empty for none.
	BrowserQFmt, BrowserAFmt string
	// DeckID is the deck new cards of the card type go to (Anki's deck
	// override), 0 for none.
	DeckID int64
}

type Note struct {
	ID     int64
	GUID   string
	TypeID int64
	Fields []string
	Tags   []string
	// Mod is the modification time in epoch seconds.
	Mod int64
}

type Card struct {
	ID, NoteID, DeckID                   int64
	Ord                                  int
	Type, Queue                          int
	Due, Ivl, Factor, Reps, Lapses, Left int64
	ODue, ODeckID                        int64
	// Flag is the card's flag, the low 3 bits of cards.flags.
	Flag int
	Data string
	// Mod is the modification time in epoch seconds.
	Mod int64
}

// Review is one entry of Anki's review log.
type Review struct {
	// ID is the time of the entry in epoch milliseconds.
	ID, CardID int64
	// Ease is the rating, 1 to 4, or 0 for a manual entry.
	Ease int
	// Ivl and LastIvl are the intervals after and before the entry: days
	// when positive, seconds when negative.
	Ivl, LastIvl int64
	// Factor is the ease factor in permille; 0 for a cram review in a
	// filtered deck that does not reschedule.
	Factor int64
	// Time is the answer time in milliseconds.
	Time int64
	// Type: 0 learning, 1 review, 2 relearning, 3 filtered, 4 manual,
	// 5 rescheduled.
	Type int
}

// ErrUnreadable wraps errors for files that are not readable Anki packages.
var ErrUnreadable = errors.New(usermsg.Text("apkgUnreadable", "not a readable Anki package"))

func unreadable(code, format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrUnreadable, usermsg.Text(code, format, args...))
}

// Package format versions in the "meta" file (PackageMetadata.version).
const (
	metaLegacy1 = 1
	metaLegacy2 = 2
	metaLatest  = 3
)

// Read parses an .apkg or .colpkg file. yield is called periodically in long loops.
func Read(data []byte, yield func()) (*Package, error) {
	if yield == nil {
		yield = func() {}
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, unreadable("apkgNotZip", "the file is damaged or not a zip archive (%v)", err)
	}
	files := map[string]*zip.File{}
	for _, f := range zr.File {
		files[f.Name] = f
	}
	read := func(name string) ([]byte, error) {
		f, ok := files[name]
		if !ok {
			return nil, unreadable("apkgFileMissing", "%s is missing", name)
		}
		rc, err := f.Open()
		if err != nil {
			return nil, unreadable("apkgFileError", "%s: %v", name, err)
		}
		defer rc.Close()
		b, err := io.ReadAll(rc)
		if err != nil {
			return nil, unreadable("apkgFileDamaged", "%s is damaged (%v)", name, err)
		}
		return b, nil
	}

	if _, ok := files["meta"]; ok {
		meta, err := read("meta")
		if err != nil {
			return nil, err
		}
		if v := protoVarint(meta, 1); v > metaLatest {
			return nil, unreadable("apkgNewerVersion", "it was made by a newer Anki version (package format %d) that Cardflip cannot read yet", v)
		}
	}

	dec, err := zstd.NewReader(nil, zstd.WithDecoderConcurrency(1))
	if err != nil {
		return nil, err
	}
	defer dec.Close()

	p := &Package{Decks: map[int64]Deck{}, Presets: map[int64]Preset{}, NoteTypes: map[int64]NoteType{}, Media: map[string][]byte{}}
	var raw []byte
	switch {
	case files["collection.anki21b"] != nil:
		p.Format = "anki21b"
		if raw, err = read("collection.anki21b"); err != nil {
			return nil, err
		}
		if raw, err = dec.DecodeAll(raw, nil); err != nil {
			return nil, unreadable("apkgCollectionDamaged", "the collection is damaged (%v)", err)
		}
	case files["collection.anki21"] != nil:
		p.Format = "anki21"
		raw, err = read("collection.anki21")
	case files["collection.anki2"] != nil:
		p.Format = "anki2"
		raw, err = read("collection.anki2")
	default:
		return nil, unreadable("apkgNoCollection", "it contains no Anki collection")
	}
	if err != nil {
		return nil, err
	}
	db, err := sqlite.Open(raw)
	if err != nil {
		return nil, unreadable("apkgDatabaseDamaged", "the collection database is damaged (%v)", err)
	}
	if err := readCollection(db, p, yield); err != nil {
		if errors.Is(err, sqlite.ErrCorrupt) {
			return nil, unreadable("apkgDatabaseDamaged", "the collection database is damaged (%v)", err)
		}
		return nil, err
	}
	if err := readMedia(p, files, read, dec, yield); err != nil {
		return nil, err
	}
	return p, nil
}

func readCollection(db *sqlite.DB, p *Package, yield func()) error {
	var colModels, colDecks, colDconf string
	err := db.Scan("col", func(r sqlite.Row) error {
		p.Crt = r.Int("crt")
		colModels = r.Str("models")
		colDecks = r.Str("decks")
		colDconf = r.Str("dconf")
		return nil
	})
	if err != nil {
		return err
	}
	if db.HasTable("notetypes") {
		if err := readModernTypes(db, p); err != nil {
			return err
		}
	} else if err := readLegacyTypes(colModels, colDecks, p); err != nil {
		return err
	}
	if db.HasTable("deck_config") {
		if err := readModernPresets(db, p); err != nil {
			return err
		}
	} else {
		readLegacyPresets(colDconf, p)
	}

	n := 0
	err = db.Scan("notes", func(r sqlite.Row) error {
		if n++; n%500 == 0 {
			yield()
		}
		p.Notes = append(p.Notes, Note{
			ID: r.Int("id"), GUID: r.Str("guid"), TypeID: r.Int("mid"),
			Fields: strings.Split(r.Str("flds"), "\x1f"), Tags: strings.Fields(r.Str("tags")), Mod: r.Int("mod"),
		})
		return nil
	})
	if err != nil {
		return err
	}
	sort.Slice(p.Notes, func(i, j int) bool { return p.Notes[i].ID < p.Notes[j].ID })
	err = db.Scan("cards", func(r sqlite.Row) error {
		if n++; n%500 == 0 {
			yield()
		}
		p.Cards = append(p.Cards, Card{
			ID: r.Int("id"), NoteID: r.Int("nid"), DeckID: r.Int("did"), Ord: int(r.Int("ord")),
			Type: int(r.Int("type")), Queue: int(r.Int("queue")), Due: r.Int("due"), Ivl: r.Int("ivl"),
			Factor: r.Int("factor"), Reps: r.Int("reps"), Lapses: r.Int("lapses"), Left: r.Int("left"),
			ODue: r.Int("odue"), ODeckID: r.Int("odid"), Flag: int(r.Int("flags") & 7), Data: r.Str("data"),
			Mod: r.Int("mod"),
		})
		return nil
	})
	if err != nil {
		return err
	}
	sort.Slice(p.Cards, func(i, j int) bool { return p.Cards[i].ID < p.Cards[j].ID })
	if !db.HasTable("revlog") {
		return nil
	}
	err = db.Scan("revlog", func(r sqlite.Row) error {
		if n++; n%500 == 0 {
			yield()
		}
		p.Revlog = append(p.Revlog, Review{
			ID: r.Int("id"), CardID: r.Int("cid"), Ease: int(r.Int("ease")), Ivl: r.Int("ivl"),
			LastIvl: r.Int("lastIvl"), Factor: r.Int("factor"), Time: r.Int("time"), Type: int(r.Int("type")),
		})
		return nil
	})
	if err != nil {
		return err
	}
	sort.Slice(p.Revlog, func(i, j int) bool { return p.Revlog[i].ID < p.Revlog[j].ID })
	return nil
}

// readModernTypes reads note types and decks from the anki21b tables, whose
// configs are protobuf messages (field numbers from Anki's notetypes.proto
// and decks.proto). Unknown fields are skipped.
func readModernTypes(db *sqlite.DB, p *Package) error {
	err := db.Scan("notetypes", func(r sqlite.Row) error {
		cfg := r.Bytes("config")
		// Notetype.Config: kind = 1 (0 normal, 1 cloze), sort_field_idx
		// = 2, css = 3, latex_svg = 7.
		p.NoteTypes[r.Int("id")] = NoteType{
			Name:      r.Str("name"),
			Cloze:     protoVarint(cfg, 1) == 1,
			SortField: int(protoVarint(cfg, 2)),
			CSS:       protoString(cfg, 3),
			Mod:       r.Int("mtime_secs"),
			LatexSVG:  protoVarint(cfg, 7) != 0,
		}
		return nil
	})
	if err != nil {
		return err
	}
	type ordered[T any] struct {
		ord int64
		v   T
	}
	type field struct {
		name string
		cfg  FieldConfig
	}
	fields := map[int64][]ordered[field]{}
	err = db.Scan("fields", func(r sqlite.Row) error {
		f := field{r.Str("name"), modernFieldConfig(r.Bytes("config"))}
		fields[r.Int("ntid")] = append(fields[r.Int("ntid")], ordered[field]{r.Int("ord"), f})
		return nil
	})
	if err != nil {
		return err
	}
	tmpls := map[int64][]ordered[Template]{}
	err = db.Scan("templates", func(r sqlite.Row) error {
		t := modernTemplate(r.Str("name"), r.Bytes("config"))
		tmpls[r.Int("ntid")] = append(tmpls[r.Int("ntid")], ordered[Template]{r.Int("ord"), t})
		return nil
	})
	if err != nil {
		return err
	}
	for id, nt := range p.NoteTypes {
		fs := fields[id]
		sort.Slice(fs, func(i, j int) bool { return fs[i].ord < fs[j].ord })
		for _, f := range fs {
			nt.Fields = append(nt.Fields, f.v.name)
			nt.FieldConfigs = append(nt.FieldConfigs, f.v.cfg)
		}
		ts := tmpls[id]
		sort.Slice(ts, func(i, j int) bool { return ts[i].ord < ts[j].ord })
		for _, t := range ts {
			nt.Templates = append(nt.Templates, t.v)
		}
		if len(nt.Fields) == 0 || len(nt.Templates) == 0 {
			return unreadable("apkgEmptyNoteType", "note type %q has no fields or templates; unsupported Anki package version", nt.Name)
		}
		p.NoteTypes[id] = nt
	}
	return db.Scan("decks", func(r sqlite.Row) error {
		p.Decks[r.Int("id")] = modernDeck(r.Str("name"), r.Bytes("kind"))
		return nil
	})
}

// modernFieldConfig reads a Notetype.Field.Config message: sticky = 1,
// rtl = 2, font_name = 3, font_size = 4, plain_text = 6, collapsed = 7,
// exclude_from_search = 8.
func modernFieldConfig(cfg []byte) FieldConfig {
	b := func(field protowire.Number) bool { return protoVarint(cfg, field) != 0 }
	return FieldConfig{Sticky: b(1), RTL: b(2), Font: protoString(cfg, 3), Size: int(protoVarint(cfg, 4)),
		PlainText: b(6), Collapsed: b(7), ExcludeFromSearch: b(8)}
}

// modernDeck reads a row of the anki21b decks table: its name, with
// components separated by \x1f, and its Deck.KindContainer (normal = 1,
// filtered = 2; Deck.Normal: config_id = 1, description = 4).
func modernDeck(name string, kind []byte) Deck {
	d := Deck{
		Name:     strings.ReplaceAll(name, "\x1f", "::"),
		Filtered: protoHas(kind, 2),
	}
	if normal := protoBytesAll(kind, 1); len(normal) > 0 {
		d.PresetID = int64(protoVarint(normal[len(normal)-1], 1))
		d.Description = protoString(normal[len(normal)-1], 4)
	}
	return d
}

// readModernPresets reads the anki21b deck_config table, whose config is a
// DeckConfig.Config protobuf message (field numbers from Anki's
// deck_config.proto).
func readModernPresets(db *sqlite.DB, p *Package) error {
	return db.Scan("deck_config", func(r sqlite.Row) error {
		cfg := r.Bytes("config")
		v := func(field protowire.Number) int { return int(protoVarint(cfg, field)) }
		b := func(field protowire.Number) bool { return protoVarint(cfg, field) != 0 }
		p.Presets[r.Int("id")] = Preset{
			Name:      r.Str("name"),
			NewPerDay: v(9), ReviewsPerDay: v(10),
			LearnSteps: protoFloats(cfg, 1), RelearnSteps: protoFloats(cfg, 2),
			DesiredRetention: protoFloat(cfg, 37),
			MaxInterval:      v(16), MinLapseInterval: v(17),
			LeechAction: v(21), LeechThreshold: v(22),
			InsertOrder: v(20), GatherPriority: v(34), SortOrder: v(32), ReviewOrder: v(33),
			NewMix: v(30), InterdayMix: v(31),
			BuryNew: b(27), BuryReviews: b(28), BuryInterdayLearning: b(29),
			DisableAutoplay: b(23), SkipQuestionWhenReplaying: b(26),
			SecondsToShowQuestion: protoFloat(cfg, 41), SecondsToShowAnswer: protoFloat(cfg, 42),
			QuestionAction: v(36), AnswerAction: v(43),
		}
		return nil
	})
}

// readLegacyPresets reads the presets of the col.dconf JSON. Presets that
// cannot be read are skipped; their decks get the default preset.
func readLegacyPresets(dconf string, p *Package) {
	var cs map[string]json.RawMessage
	if json.Unmarshal([]byte(dconf), &cs) != nil {
		return
	}
	for id, raw := range cs {
		pid, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			continue
		}
		var c struct {
			Name string `json:"name"`
			New  struct {
				PerDay float64   `json:"perDay"`
				Delays []float64 `json:"delays"`
				Order  float64   `json:"order"`
				Bury   bool      `json:"bury"`
			} `json:"new"`
			Rev struct {
				PerDay float64 `json:"perDay"`
				MaxIvl float64 `json:"maxIvl"`
				Bury   bool    `json:"bury"`
			} `json:"rev"`
			Lapse struct {
				Delays      []float64 `json:"delays"`
				MinInt      float64   `json:"minInt"`
				LeechFails  float64   `json:"leechFails"`
				LeechAction float64   `json:"leechAction"`
			} `json:"lapse"`
			Autoplay              *bool   `json:"autoplay"`
			ReplayQ               *bool   `json:"replayq"`
			BuryInterdayLearning  bool    `json:"buryInterdayLearning"`
			DesiredRetention      float64 `json:"desiredRetention"`
			NewMix                float64 `json:"newMix"`
			InterdayLearningMix   float64 `json:"interdayLearningMix"`
			ReviewOrder           float64 `json:"reviewOrder"`
			NewSortOrder          float64 `json:"newSortOrder"`
			NewGatherPriority     float64 `json:"newGatherPriority"`
			SecondsToShowQuestion float64 `json:"secondsToShowQuestion"`
			SecondsToShowAnswer   float64 `json:"secondsToShowAnswer"`
			QuestionAction        float64 `json:"questionAction"`
			AnswerAction          float64 `json:"answerAction"`
		}
		if json.Unmarshal(raw, &c) != nil {
			continue
		}
		pr := Preset{
			Name: c.Name, NewPerDay: int(c.New.PerDay), ReviewsPerDay: int(c.Rev.PerDay),
			LearnSteps: c.New.Delays, RelearnSteps: c.Lapse.Delays, DesiredRetention: c.DesiredRetention,
			MaxInterval: int(c.Rev.MaxIvl), MinLapseInterval: int(c.Lapse.MinInt),
			LeechThreshold: int(c.Lapse.LeechFails), LeechAction: int(c.Lapse.LeechAction),
			GatherPriority: int(c.NewGatherPriority), SortOrder: int(c.NewSortOrder), ReviewOrder: int(c.ReviewOrder),
			NewMix: int(c.NewMix), InterdayMix: int(c.InterdayLearningMix),
			BuryNew: c.New.Bury, BuryReviews: c.Rev.Bury, BuryInterdayLearning: c.BuryInterdayLearning,
			DisableAutoplay:           c.Autoplay != nil && !*c.Autoplay,
			SkipQuestionWhenReplaying: c.ReplayQ != nil && !*c.ReplayQ,
			SecondsToShowQuestion:     c.SecondsToShowQuestion, SecondsToShowAnswer: c.SecondsToShowAnswer,
			QuestionAction: int(c.QuestionAction), AnswerAction: int(c.AnswerAction),
		}
		// Legacy new card order: 0 random, 1 due.
		if c.New.Order == 0 {
			pr.InsertOrder = 1
		}
		p.Presets[pid] = pr
	}
}

func readLegacyTypes(models, decks string, p *Package) error {
	var ms map[string]struct {
		Name  string  `json:"name"`
		Type  int     `json:"type"`
		CSS   string  `json:"css"`
		Mod   float64 `json:"mod"`
		SortF int     `json:"sortf"`
		// Anki reads an invalid value as false.
		LatexSVG json.RawMessage `json:"latexsvg"`
		Flds     []struct {
			Name              string  `json:"name"`
			Ord               int     `json:"ord"`
			Sticky            bool    `json:"sticky"`
			RTL               bool    `json:"rtl"`
			Font              string  `json:"font"`
			Size              float64 `json:"size"`
			PlainText         bool    `json:"plainText"`
			Collapsed         bool    `json:"collapsed"`
			ExcludeFromSearch bool    `json:"excludeFromSearch"`
		} `json:"flds"`
		Tmpls []struct {
			Name  string `json:"name"`
			QFmt  string `json:"qfmt"`
			AFmt  string `json:"afmt"`
			BQFmt string `json:"bqfmt"`
			BAFmt string `json:"bafmt"`
			// DID is the deck override: a deck ID, or null or 0 for none.
			DID json.RawMessage `json:"did"`
			Ord int             `json:"ord"`
		} `json:"tmpls"`
	}
	if err := json.Unmarshal([]byte(models), &ms); err != nil {
		return unreadable("apkgNoteTypesUnreadable", "its note types could not be read (%v)", err)
	}
	for id, m := range ms {
		nid, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			continue
		}
		sort.Slice(m.Flds, func(i, j int) bool { return m.Flds[i].Ord < m.Flds[j].Ord })
		sort.Slice(m.Tmpls, func(i, j int) bool { return m.Tmpls[i].Ord < m.Tmpls[j].Ord })
		nt := NoteType{Name: m.Name, Cloze: m.Type == 1, CSS: m.CSS, Mod: int64(m.Mod), LatexSVG: string(m.LatexSVG) == "true",
			SortField: m.SortF}
		for _, f := range m.Flds {
			nt.Fields = append(nt.Fields, f.Name)
			nt.FieldConfigs = append(nt.FieldConfigs, FieldConfig{Sticky: f.Sticky, RTL: f.RTL, Font: f.Font, Size: int(f.Size),
				PlainText: f.PlainText, Collapsed: f.Collapsed, ExcludeFromSearch: f.ExcludeFromSearch})
		}
		for _, t := range m.Tmpls {
			did, _ := strconv.ParseInt(string(t.DID), 10, 64)
			nt.Templates = append(nt.Templates, Template{Name: t.Name, QFmt: t.QFmt, AFmt: t.AFmt,
				BrowserQFmt: t.BQFmt, BrowserAFmt: t.BAFmt, DeckID: did})
		}
		p.NoteTypes[nid] = nt
	}
	var ds map[string]struct {
		Name string `json:"name"`
		Dyn  int    `json:"dyn"`
		Conf int64  `json:"conf"`
		Desc string `json:"desc"`
	}
	if err := json.Unmarshal([]byte(decks), &ds); err != nil {
		return unreadable("apkgDecksUnreadable", "its decks could not be read (%v)", err)
	}
	for id, d := range ds {
		if did, err := strconv.ParseInt(id, 10, 64); err == nil {
			p.Decks[did] = Deck{Name: d.Name, Filtered: d.Dyn != 0, PresetID: d.Conf, Description: d.Desc}
		}
	}
	return nil
}

var zstdMagic = []byte{0x28, 0xb5, 0x2f, 0xfd}

func readMedia(p *Package, files map[string]*zip.File, read func(string) ([]byte, error), dec *zstd.Decoder, yield func()) error {
	if files["media"] == nil {
		return nil
	}
	raw, err := read("media")
	if err != nil {
		return err
	}
	// name → file name in the zip
	entries := map[string]string{}
	if bytes.HasPrefix(raw, zstdMagic) {
		if raw, err = dec.DecodeAll(raw, nil); err != nil {
			return unreadable("apkgMediaListDamaged", "the media list is damaged (%v)", err)
		}
		// MediaEntries { repeated MediaEntry entries = 1; }
		// MediaEntry { name = 1; size = 2; sha1 = 3; legacy_zip_filename = 255; }
		for i, e := range protoBytesAll(raw, 1) {
			zipName := strconv.Itoa(i)
			if n := protoVarint(e, 255); n != 0 {
				zipName = strconv.FormatUint(n, 10)
			}
			entries[protoString(e, 1)] = zipName
		}
	} else if len(bytes.TrimSpace(raw)) > 0 {
		var m map[string]string
		if err := json.Unmarshal(raw, &m); err != nil {
			return unreadable("apkgMediaListDamaged", "the media list is damaged (%v)", err)
		}
		for zipName, name := range m {
			entries[name] = zipName
		}
	}
	for name, zipName := range entries {
		yield()
		data, err := read(zipName)
		if err != nil {
			return err
		}
		if bytes.HasPrefix(data, zstdMagic) {
			if data, err = dec.DecodeAll(data, nil); err != nil {
				return unreadable("apkgMediaDamaged", "media file %s is damaged (%v)", name, err)
			}
		}
		p.Media[name] = data
	}
	return nil
}

// Minimal protobuf helpers: they scan a message for one field number and
// skip everything else, so unknown fields from newer Anki versions are fine.

func protoEach(b []byte, fn func(num protowire.Number, typ protowire.Type, v []byte, x uint64)) {
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return
		}
		b = b[n:]
		switch typ {
		case protowire.VarintType:
			x, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return
			}
			fn(num, typ, nil, x)
			b = b[n:]
		case protowire.BytesType:
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return
			}
			fn(num, typ, v, 0)
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(num, typ, b)
			if n < 0 {
				return
			}
			b = b[n:]
		}
	}
}

// modernTemplate reads a card type from its name and Anki's
// Notetype.Template.Config: q_format = 1, a_format = 2, q_format_browser =
// 3, a_format_browser = 4, target_deck_id = 5.
func modernTemplate(name string, cfg []byte) Template {
	return Template{Name: name, QFmt: protoString(cfg, 1), AFmt: protoString(cfg, 2),
		BrowserQFmt: protoString(cfg, 3), BrowserAFmt: protoString(cfg, 4), DeckID: int64(protoVarint(cfg, 5))}
}

func protoVarint(b []byte, field protowire.Number) uint64 {
	var out uint64
	protoEach(b, func(num protowire.Number, typ protowire.Type, _ []byte, x uint64) {
		if num == field && typ == protowire.VarintType {
			out = x
		}
	})
	return out
}

func protoString(b []byte, field protowire.Number) string {
	var out string
	protoEach(b, func(num protowire.Number, typ protowire.Type, v []byte, _ uint64) {
		if num == field && typ == protowire.BytesType {
			out = string(v)
		}
	})
	return out
}

func protoBytesAll(b []byte, field protowire.Number) [][]byte {
	var out [][]byte
	protoEach(b, func(num protowire.Number, typ protowire.Type, v []byte, _ uint64) {
		if num == field && typ == protowire.BytesType {
			out = append(out, v)
		}
	})
	return out
}

// protoFloat returns float field number field (fixed32), or 0.
func protoFloat(b []byte, field protowire.Number) float64 {
	fs := protoFloats(b, field)
	if len(fs) == 0 {
		return 0
	}
	return fs[len(fs)-1]
}

// protoFloats returns the values of repeated float field number field,
// packed or not.
func protoFloats(b []byte, field protowire.Number) []float64 {
	var out []float64
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return out
		}
		b = b[n:]
		switch {
		case num == field && typ == protowire.Fixed32Type:
			v, n := protowire.ConsumeFixed32(b)
			if n < 0 {
				return out
			}
			out = append(out, float64(math.Float32frombits(v)))
			b = b[n:]
		case num == field && typ == protowire.BytesType:
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return out
			}
			for len(v) >= 4 {
				x, m := protowire.ConsumeFixed32(v)
				if m < 0 {
					break
				}
				out = append(out, float64(math.Float32frombits(x)))
				v = v[m:]
			}
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(num, typ, b)
			if n < 0 {
				return out
			}
			b = b[n:]
		}
	}
	return out
}

func protoHas(b []byte, field protowire.Number) bool {
	found := false
	protoEach(b, func(num protowire.Number, _ protowire.Type, _ []byte, _ uint64) {
		if num == field {
			found = true
		}
	})
	return found
}
