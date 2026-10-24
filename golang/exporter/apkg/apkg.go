// Package apkg exports Anki packages: deck packages (.apkg) and collection
// packages (.colpkg). Both hold an Anki schema 11 collection (design
// decision 1): by default as collection.anki21 next to the dummy
// collection.anki2 Anki writes for old clients, or, to support older Anki
// versions, as collection.anki2 itself.
package apkg

import (
	"archive/zip"
	"bytes"
	"cmp"
	"crypto/sha1"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/exporter/sqlite"
	"cardflip/golang/render"
	"cardflip/golang/scheduler"
)

// Options configure an export.
type Options struct {
	// Scheduling keeps the cards' scheduling and review history; without
	// it every card is new and unflagged, and notes lose the leech tag.
	Scheduling bool
	// Presets exports the decks' presets as Anki deck options presets;
	// without it the decks use Anki's default preset.
	Presets bool
	// Media includes the media files the notes use.
	Media bool
	// Legacy writes the collection as collection.anki2, which Anki 2.1
	// versions before 2.1.50 read too.
	Legacy bool
}

// Input is what an export writes.
type Input struct {
	Set *core.ExportSet
	// Reviews is the review history of the set's cards; it is written
	// only with Options.Scheduling.
	Reviews []core.ReviewEntry
	// Media reads one of the set's media files.
	Media func(name string) ([]byte, error)
	Now   time.Time
}

// Progress reports export progress: done of total steps in a stage
// ("collection", "media" or "saving").
type Progress func(stage string, done, total int)

// Collection file names and package format versions (Anki's
// PackageMetadata.version: 1 for collection.anki2, 2 for
// collection.anki21).
const (
	fileLegacy1 = "collection.anki2"
	fileLegacy2 = "collection.anki21"
	metaLegacy1 = 1
	metaLegacy2 = 2
)

// Write writes a package of in. It is the same for .apkg and .colpkg; the
// caller picks the content (a deck or the whole collection) and the name.
func Write(in Input, o Options, progress Progress, yield func()) ([]byte, error) {
	if progress == nil {
		progress = func(string, int, int) {}
	}
	if yield == nil {
		yield = func() {}
	}
	progress("collection", 0, 1)
	col, err := writeCollection(in, o, yield)
	if err != nil {
		return nil, err
	}
	progress("collection", 1, 1)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	put := func(name string, method uint16, data []byte) error {
		w, err := zw.CreateHeader(&zip.FileHeader{Name: name, Method: method, Modified: in.Now})
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	}
	if o.Legacy {
		err = put(fileLegacy1, zip.Deflate, col)
	} else {
		var dummy []byte
		if dummy, err = dummyCollection(in.Now); err != nil {
			return nil, err
		}
		if err = put(fileLegacy2, zip.Deflate, col); err == nil {
			err = put(fileLegacy1, zip.Deflate, dummy)
		}
	}
	if err != nil {
		return nil, err
	}

	// The media map names each file's zip entry by its index.
	mediaMap := map[string]string{}
	if o.Media {
		for i, name := range in.Set.Media {
			progress("media", i, len(in.Set.Media))
			yield()
			data, err := in.Media(name)
			if err != nil {
				continue // removed since; the card shows it as missing
			}
			key := strconv.Itoa(len(mediaMap))
			// Media is mostly compressed already.
			if err := put(key, zip.Store, data); err != nil {
				return nil, err
			}
			mediaMap[key] = name
		}
		progress("media", len(in.Set.Media), len(in.Set.Media))
	}
	media, _ := json.Marshal(mediaMap)
	if err := put("media", zip.Deflate, media); err != nil {
		return nil, err
	}
	meta := []byte{0x08, metaLegacy2} // PackageMetadata{version: 2}
	if o.Legacy {
		meta[1] = metaLegacy1
	}
	if err := put("meta", zip.Store, meta); err != nil {
		return nil, err
	}
	progress("saving", 0, 1)
	if err := zw.Close(); err != nil {
		return nil, err
	}
	progress("saving", 1, 1)
	return buf.Bytes(), nil
}

// schema creates Anki's schema 11 tables and indexes.
func schema(w *sqlite.Writer) (col, notes, cards, revlog *sqlite.Table, err error) {
	mk := func(name string, defs ...string) *sqlite.Table {
		if err != nil {
			return nil
		}
		var t *sqlite.Table
		t, err = w.CreateTable(name, defs...)
		return t
	}
	col = mk("col", "id integer PRIMARY KEY", "crt integer NOT NULL", "mod integer NOT NULL",
		"scm integer NOT NULL", "ver integer NOT NULL", "dty integer NOT NULL", "usn integer NOT NULL",
		"ls integer NOT NULL", "conf text NOT NULL", "models text NOT NULL", "decks text NOT NULL",
		"dconf text NOT NULL", "tags text NOT NULL")
	notes = mk("notes", "id integer PRIMARY KEY", "guid text NOT NULL", "mid integer NOT NULL",
		"mod integer NOT NULL", "usn integer NOT NULL", "tags text NOT NULL", "flds text NOT NULL",
		"sfld integer NOT NULL", "csum integer NOT NULL", "flags integer NOT NULL", "data text NOT NULL")
	cards = mk("cards", "id integer PRIMARY KEY", "nid integer NOT NULL", "did integer NOT NULL",
		"ord integer NOT NULL", "mod integer NOT NULL", "usn integer NOT NULL", "type integer NOT NULL",
		"queue integer NOT NULL", "due integer NOT NULL", "ivl integer NOT NULL", "factor integer NOT NULL",
		"reps integer NOT NULL", "lapses integer NOT NULL", "left integer NOT NULL", "odue integer NOT NULL",
		"odid integer NOT NULL", "flags integer NOT NULL", "data text NOT NULL")
	revlog = mk("revlog", "id integer PRIMARY KEY", "cid integer NOT NULL", "usn integer NOT NULL",
		"ease integer NOT NULL", "ivl integer NOT NULL", "lastIvl integer NOT NULL", "factor integer NOT NULL",
		"time integer NOT NULL", "type integer NOT NULL")
	mk("graves", "usn integer NOT NULL", "oid integer NOT NULL", "type integer NOT NULL")
	if err != nil {
		return
	}
	for _, ix := range []struct {
		name  string
		table *sqlite.Table
		cols  []string
	}{
		{"ix_notes_usn", notes, []string{"usn"}},
		{"ix_cards_usn", cards, []string{"usn"}},
		{"ix_revlog_usn", revlog, []string{"usn"}},
		{"ix_cards_nid", cards, []string{"nid"}},
		{"ix_cards_sched", cards, []string{"did", "queue", "due"}},
		{"ix_revlog_cid", revlog, []string{"cid"}},
		{"ix_notes_csum", notes, []string{"csum"}},
	} {
		if err = w.CreateIndex(ix.name, ix.table, ix.cols...); err != nil {
			return
		}
	}
	return
}

// exporter holds the ID mappings and day numbers of one export.
type exporter struct {
	in       Input
	o        Options
	rollover int
	crt      time.Time
	crtDay   int
	deckIDs  map[core.ID]int64
	typeIDs  map[core.ID]int64
	decks    map[core.ID]*core.Deck
	presets  map[core.ID]*core.Preset
	// factors are the cards' exported ease factors, for their review log.
	factors map[core.ID]int64
	// reviewed is when each card was last reviewed, by its review log.
	reviewed map[core.ID]time.Time
}

func writeCollection(in Input, o Options, yield func()) ([]byte, error) {
	e := &exporter{in: in, o: o, rollover: in.Set.Settings.RolloverHour,
		deckIDs: map[core.ID]int64{}, typeIDs: map[core.ID]int64{}, decks: map[core.ID]*core.Deck{},
		presets: map[core.ID]*core.Preset{}, factors: map[core.ID]int64{}, reviewed: map[core.ID]time.Time{}}
	for _, r := range in.Reviews {
		if r.Rating > 0 && !isCram(r) {
			e.reviewed[r.CardID] = r.Time
		}
	}
	e.crt = e.creationTime()
	e.crtDay = scheduler.DayNumber(e.crt, e.rollover)
	for i := range in.Set.Presets {
		e.presets[in.Set.Presets[i].ID] = &in.Set.Presets[i]
	}
	e.mapIDs()

	w := sqlite.NewWriter()
	colT, notesT, cardsT, revlogT, err := schema(w)
	if err != nil {
		return nil, err
	}
	conf, models, decks, dconf, err := e.colJSON()
	if err != nil {
		return nil, err
	}
	mod := in.Now.UnixMilli()
	if err := colT.Insert(int64(1), e.crt.Unix(), mod, mod, 11, 0, 0, 0, conf, models, decks, dconf, "{}"); err != nil {
		return nil, err
	}
	for i, n := range in.Set.Notes {
		if i%500 == 0 {
			yield()
		}
		if err := e.writeNote(notesT, n); err != nil {
			return nil, err
		}
	}
	if err := e.writeCards(cardsT, yield); err != nil {
		return nil, err
	}
	if o.Scheduling {
		if err := e.writeRevlog(revlogT, yield); err != nil {
			return nil, err
		}
	}
	return w.Bytes(yield)
}

// creationTime picks the collection's creation time, from which Anki counts
// review due days: the start of the study day of the oldest note, card,
// deck or review, so that day numbers are never negative.
func (e *exporter) creationTime() time.Time {
	t := e.in.Now
	older := func(u time.Time) {
		if !u.IsZero() && u.Before(t) {
			t = u
		}
	}
	idTime := func(id core.ID) time.Time { return time.UnixMilli(int64(id)).In(e.in.Now.Location()) }
	for _, n := range e.in.Set.Notes {
		older(idTime(n.ID))
	}
	for _, c := range e.in.Set.Cards {
		older(idTime(c.ID))
		older(c.Due)
		older(c.LastReview)
	}
	for _, r := range e.in.Reviews {
		older(r.Time)
	}
	return scheduler.DayStart(t.In(e.in.Now.Location()), e.rollover)
}

// Anki's default deck and preset have ID 1.
const ankiDefaultID = 1

// mapIDs assigns Anki IDs. Cardflip's IDs are epoch milliseconds like
// Anki's and are kept, except that the top-level Default deck becomes
// Anki's default deck, and note types imported from Anki get their Anki ID
// back (unless two exported note types came from the same one).
func (e *exporter) mapIDs() {
	for i := range e.in.Set.Decks {
		d := &e.in.Set.Decks[i]
		e.decks[d.ID] = d
		e.deckIDs[d.ID] = int64(d.ID)
		if strings.EqualFold(d.Name, core.DefaultDeckName) {
			e.deckIDs[d.ID] = ankiDefaultID
		}
	}
	used := map[int64]bool{}
	for _, nt := range e.in.Set.NoteTypes {
		id := int64(nt.ID)
		if src, ok := strings.CutPrefix(nt.Source, "anki:"); ok {
			if n, err := strconv.ParseInt(src, 10, 64); err == nil && n > 0 && !used[n] {
				id = n
			}
		}
		used[id] = true
		e.typeIDs[nt.ID] = id
	}
}

// options returns the options of a card's deck.
func (e *exporter) options(deckID core.ID) core.DeckOptions {
	d, ok := e.decks[deckID]
	if !ok {
		return core.DefaultDeckOptions()
	}
	p, ok := e.presets[d.PresetID]
	if !ok {
		return core.DefaultDeckOptions()
	}
	return p.Options
}

// imgSrc matches <img> elements, whose file names Anki keeps in the sort
// field and checksum.
var imgSrc = regexp.MustCompile(`(?i)<img[^>]*?\ssrc\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s>]+))[^>]*>`)

// stripKeepingMedia strips HTML as Anki's strip_html_preserving_media_filenames.
func stripKeepingMedia(s string) string {
	s = imgSrc.ReplaceAllStringFunc(s, func(m string) string {
		g := imgSrc.FindStringSubmatch(m)
		return " " + g[1] + g[2] + g[3] + " "
	})
	return render.StripHTML(s)
}

// checksum is Anki's field checksum: the first 4 bytes of the SHA-1 of the
// stripped first field.
func checksum(field string) int64 {
	h := sha1.Sum([]byte(stripKeepingMedia(field)))
	return int64(binary.BigEndian.Uint32(h[:4]))
}

// sortValue is the sort field as SQLite stores it in a column with integer
// affinity: numbers become numbers.
func sortValue(s string) any {
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return n
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil && !strings.ContainsAny(s, "xXnN") {
		if f == float64(int64(f)) {
			return int64(f)
		}
		return f
	}
	return s
}

func (e *exporter) writeNote(t *sqlite.Table, n core.Note) error {
	tags := n.Tags
	if !e.o.Scheduling {
		tags = nil
		for _, tag := range n.Tags {
			if !strings.EqualFold(tag, core.LeechTagName) {
				tags = append(tags, tag)
			}
		}
	}
	tagStr := ""
	if len(tags) > 0 {
		tagStr = " " + strings.Join(tags, " ") + " "
	}
	first := ""
	if len(n.Fields) > 0 {
		first = n.Fields[0]
	}
	return t.Insert(int64(n.ID), n.GUID, e.typeIDs[n.TypeID], n.Modified.Unix(), -1, tagStr,
		strings.Join(n.Fields, "\x1f"), sortValue(stripKeepingMedia(first)), checksum(first), 0, "")
}

// ankiCard is a card's row in Anki's cards table.
type ankiCard struct {
	typ, queue                                int
	due, ivl, factor, reps, lapses, left, flg int64
	data                                      string
}

// difficultyFactor maps an FSRS difficulty to an ease factor in permille,
// the inverse of the importer's estimate (difficulty 10–1 for ease
// 130%–350%).
func difficultyFactor(d float64) int64 {
	f := 1300 + (10-d)*2200/9
	return int64(min(max(f, 1300), 3500)+5) / 10 * 10
}

// convertCard maps a card's scheduling to Anki's (design decision 3).
func (e *exporter) convertCard(c core.Card, newDue int64) ankiCard {
	if !e.o.Scheduling {
		return ankiCard{due: newDue, data: "{}"}
	}
	a := ankiCard{reps: int64(c.Reps), lapses: int64(c.Lapses), flg: int64(c.Flag), data: "{}"}
	day := func(t time.Time) int64 { return int64(scheduler.DayNumber(t, e.rollover) - e.crtDay) }
	opts := e.options(c.DeckID)
	if c.Memory != nil && c.State != scheduler.New {
		a.factor = difficultyFactor(c.Memory.Difficulty)
		data := map[string]any{"s": round(c.Memory.Stability, 4), "d": round(c.Memory.Difficulty, 4),
			"dr": round(opts.DesiredRetention, 2)}
		// Anki takes the last review time from the review log, where
		// there is one.
		if t, ok := e.reviewed[c.ID]; ok {
			data["lrt"] = t.Unix()
		} else if !c.LastReview.IsZero() {
			data["lrt"] = c.LastReview.Unix()
		}
		b, _ := json.Marshal(data)
		a.data = string(b)
	}
	switch c.State {
	case scheduler.New:
		a.due = newDue
		a.factor = 0
	case scheduler.Review:
		a.typ, a.queue = 2, 2
		a.due = day(c.Due)
		a.ivl = int64(max(c.ScheduledDays, 1))
		if a.factor == 0 {
			a.factor = 2500
		}
	default: // learning, relearning
		steps := opts.LearningSteps
		a.typ = 1
		if c.State == scheduler.Relearning {
			steps = opts.RelearningSteps
			a.typ = 3
			a.ivl = 1
			if c.Memory != nil {
				a.ivl = int64(max(int(c.Memory.Stability+0.5), 1))
			}
		} else {
			a.factor = 0
		}
		remaining := int64(max(len(steps)-c.Step, 1))
		a.left = remaining*1000 + remaining
		if c.Interday {
			a.queue = 3
			a.due = day(c.Due)
		} else {
			a.queue = 1
			a.due = c.Due.Unix()
		}
	}
	if c.Suspended {
		a.queue = -1
	} else if c.BuriedUntil.After(e.in.Now) {
		a.queue = -3
	}
	return a
}

func round(v float64, digits int) float64 {
	p, _ := strconv.ParseFloat(strconv.FormatFloat(v, 'f', digits, 64), 64)
	return p
}

func (e *exporter) writeCards(t *sqlite.Table, yield func()) error {
	// New cards keep their position; without scheduling the others follow
	// them in creation order.
	var maxPos int64
	for _, c := range e.in.Set.Cards {
		if c.State == scheduler.New {
			maxPos = max(maxPos, c.NewPos)
		}
	}
	for i, c := range e.in.Set.Cards {
		if i%500 == 0 {
			yield()
		}
		newDue := c.NewPos
		if c.State != scheduler.New {
			maxPos++
			newDue = maxPos
		}
		a := e.convertCard(c, newDue)
		e.factors[c.ID] = a.factor
		did, ok := e.deckIDs[c.DeckID]
		if !ok {
			did = ankiDefaultID
		}
		mod := e.in.Now
		if !c.Modified.IsZero() {
			mod = c.Modified
		}
		err := t.Insert(int64(c.ID), int64(c.NoteID), did, c.Ord, mod.Unix(), -1, a.typ, a.queue,
			a.due, a.ivl, a.factor, a.reps, a.lapses, a.left, 0, 0, a.flg, a.data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *exporter) writeRevlog(t *sqlite.Table, yield func()) error {
	exported := map[core.ID]bool{}
	for _, c := range e.in.Set.Cards {
		exported[c.ID] = true
	}
	used := map[int64]bool{}
	for i, r := range e.in.Reviews {
		if i%500 == 0 {
			yield()
		}
		if !exported[r.CardID] {
			continue
		}
		// Entries imported from Anki keep their ID; Anki's IDs are the
		// entries' times in milliseconds, as are new ones'.
		id := r.Time.UnixMilli()
		if src, ok := strings.CutPrefix(r.Source, "anki:"); ok {
			if n, err := strconv.ParseInt(src, 10, 64); err == nil && n > 0 {
				id = n
			}
		}
		for used[id] {
			id++
		}
		used[id] = true
		typ := 1
		switch r.StateBefore {
		case scheduler.New, scheduler.Learning:
			typ = 0
		case scheduler.Relearning:
			typ = 2
		}
		factor := e.factors[r.CardID]
		if isCram(r) {
			typ, factor = 3, 0 // as Anki wrote it
		}
		if r.Rating == 0 {
			typ = 4 // manual: a reset or a set due date
			if r.IvlAfter == 0 {
				factor = 0
			}
		}
		err := t.Insert(id, int64(r.CardID), -1, int(r.Rating), r.IvlAfter, r.IvlBefore, factor, r.DurationMs, typ)
		if err != nil {
			return err
		}
	}
	return nil
}

// isCram reports whether a review log entry is a cram review in an Anki
// filtered deck, which did not reschedule the card: only those pass a
// review card without an interval. Preview answers in Cardflip's filtered
// decks are exported as such.
func isCram(r core.ReviewEntry) bool {
	return r.Preview || r.StateBefore == scheduler.Review && r.Rating > scheduler.Again && r.IvlAfter == 0
}

// colJSON returns the col row's conf, models, decks and dconf.
func (e *exporter) colJSON() (conf, models, decks, dconf string, err error) {
	var curModel int64
	ms := map[string]any{}
	for _, nt := range e.in.Set.NoteTypes {
		id := e.typeIDs[nt.ID]
		if curModel == 0 {
			curModel = id
		}
		ms[strconv.FormatInt(id, 10)] = model(id, nt, e.deckIDs, e.in.Now)
	}
	ds := map[string]any{strconv.Itoa(ankiDefaultID): deck(ankiDefaultID, core.DefaultDeckName, ankiDefaultID, false, core.DeckLimits{}, 0)}
	usedPresets := map[core.ID]bool{}
	for _, d := range e.in.Set.Decks {
		conf := int64(ankiDefaultID)
		limits := core.DeckLimits{}
		if e.o.Presets {
			conf = int64(d.PresetID)
			usedPresets[d.PresetID] = true
			limits = d.Limits
		}
		id := e.deckIDs[d.ID]
		ds[strconv.FormatInt(id, 10)] = deck(id, d.Name, conf, d.Collapsed, limits, e.in.Now.Unix())
	}
	cs := map[string]any{strconv.Itoa(ankiDefaultID): preset(ankiDefaultID, "Default", ankiDefaultOptions(), 0)}
	if e.o.Presets {
		for _, p := range e.in.Set.Presets {
			if usedPresets[p.ID] || p.ID == core.DefaultPresetID {
				cs[strconv.FormatInt(int64(p.ID), 10)] = preset(int64(p.ID), p.Name, p.Options, e.in.Now.Unix())
			}
		}
	}
	_, offset := e.crt.Zone()
	c := map[string]any{
		"activeDecks": []int{ankiDefaultID}, "curDeck": ankiDefaultID, "newSpread": 0, "collapseTime": 1200,
		"timeLim": 0, "estTimes": true, "dueCounts": true, "curModel": curModel, "nextPos": e.nextPos(),
		"sortType": "noteFld", "sortBackwards": false, "addToCur": true, "dayLearnFirst": false,
		"schedVer": 2, "sched2021": true, "creationOffset": -offset / 60, "rollover": e.rollover,
		// Cardflip schedules with FSRS.
		"fsrs": true,
	}
	for _, v := range []struct {
		dst *string
		v   any
	}{{&conf, c}, {&models, ms}, {&decks, ds}, {&dconf, cs}} {
		b, merr := json.Marshal(v.v)
		if merr != nil {
			return "", "", "", "", merr
		}
		*v.dst = string(b)
	}
	return
}

func (e *exporter) nextPos() int64 {
	var n int64
	for _, c := range e.in.Set.Cards {
		n = max(n, c.NewPos)
	}
	return n + 1
}

// LaTeX header and footer of Anki's stock note types.
const (
	latexPre = "\\documentclass[12pt]{article}\n\\special{papersize=3in,5in}\n\\usepackage[utf8]{inputenc}\n" +
		"\\usepackage{amssymb,amsmath}\n\\pagestyle{empty}\n\\setlength{\\parindent}{0in}\n\\begin{document}\n"
	latexPost = "\\end{document}"
)

// fieldRef matches the field names a template refers to.
var fieldRef = regexp.MustCompile(`\{\{[#^/]?([^}]*)\}\}`)

// model returns a note type's JSON. A card type's deck override names the
// package's copy of that deck, or is null when the deck is not exported.
func model(id int64, nt core.NoteType, deckIDs map[core.ID]int64, now time.Time) map[string]any {
	typ := 0
	if nt.Kind == core.Cloze {
		typ = 1
	}
	var flds []map[string]any
	for i, name := range nt.Fields {
		fo := nt.FieldOption(i)
		// The editor's defaults are Anki's font and size for new fields.
		font, size := cmp.Or(fo.Font, "Arial"), cmp.Or(fo.Size, 20)
		flds = append(flds, map[string]any{
			"name": name, "ord": i, "sticky": fo.Sticky, "rtl": fo.RTL, "font": font, "size": size,
			"description": "", "plainText": fo.HTMLByDefault, "collapsed": fo.Collapsed, "excludeFromSearch": fo.ExcludeFromSearch,
		})
	}
	var tmpls []map[string]any
	var req []any
	for i, t := range nt.Templates {
		var did any
		if t.DeckOverride != 0 {
			if ankiID, ok := deckIDs[t.DeckOverride]; ok {
				did = ankiID
			}
		}
		tmpls = append(tmpls, map[string]any{
			"name": t.Name, "ord": i, "qfmt": t.QFmt, "afmt": t.AFmt, "bqfmt": t.BrowserQFmt, "bafmt": t.BrowserAFmt,
			"did": did, "bfont": "", "bsize": 0,
		})
		// Anki 2.1 before 2.1.28 generates cards from req: a card needs
		// any of the fields its question uses.
		var need []int
		for j, name := range nt.Fields {
			for _, m := range fieldRef.FindAllStringSubmatch(t.QFmt, -1) {
				ref := m[1]
				if k := strings.LastIndexByte(ref, ':'); k >= 0 {
					ref = ref[k+1:]
				}
				if strings.TrimSpace(ref) == name {
					need = append(need, j)
					break
				}
			}
		}
		if len(need) == 0 {
			need = []int{0}
		}
		req = append(req, []any{i, "any", need})
	}
	if typ == 1 {
		req = []any{}
	}
	m := map[string]any{
		"id": id, "name": nt.Name, "type": typ, "mod": now.Unix(), "usn": -1, "sortf": nt.SortFieldIndex(), "did": nil,
		"tmpls": tmpls, "flds": flds, "css": nt.CSS, "latexPre": latexPre, "latexPost": latexPost,
		"latexsvg": nt.LatexSVG, "req": req,
	}
	if nt.IsImageOcclusion() {
		// Anki edits notes of this stock kind with its mask editor.
		m["originalStockKind"] = ankiStockImageOcclusion
	}
	return m
}

// ankiStockImageOcclusion is Anki's StockNotetype.Kind of Image Occlusion.
const ankiStockImageOcclusion = 6

func deck(id int64, name string, conf int64, collapsed bool, limits core.DeckLimits, mod int64) map[string]any {
	limit := func(n *int) any {
		if n == nil {
			return nil
		}
		return *n
	}
	usn := -1
	if id == ankiDefaultID {
		usn = 0
	}
	return map[string]any{
		"id": id, "mod": mod, "name": name, "usn": usn, "lrnToday": []int{0, 0}, "revToday": []int{0, 0},
		"newToday": []int{0, 0}, "timeToday": []int{0, 0}, "collapsed": collapsed, "browserCollapsed": collapsed,
		"desc": "", "dyn": 0, "conf": conf, "extendNew": 0, "extendRev": 0,
		"reviewLimit": limit(limits.Review), "newLimit": limit(limits.New),
		"reviewLimitToday": nil, "newLimitToday": nil, "desiredRetention": nil,
	}
}

// ankiDefaultOptions are the options of Anki's default preset, which the
// decks use when presets are not exported.
func ankiDefaultOptions() core.DeckOptions {
	o := core.DefaultDeckOptions()
	o.NewGatherOrder, o.NewSortOrder = core.GatherDeck, core.SortTypeThenGathered
	return o
}

// Anki's enum values (deck_config.proto) for Cardflip's options, the
// inverse of the importer's mapping.
var (
	gatherPriority = map[string]int{core.GatherDeck: 0, core.GatherDeckRandomNotes: 5, core.GatherAscending: 1,
		core.GatherDescending: 2, core.GatherRandomNotes: 3, core.GatherRandomCards: 4}
	sortOrder = map[string]int{core.SortTypeThenGathered: 0, core.SortGathered: 1, core.SortTypeThenRandom: 2,
		core.SortRandomNoteThenType: 3, core.SortRandom: 4}
	reviewOrder = map[string]int{core.ReviewDueRandom: 0, core.ReviewDueDeck: 1, core.ReviewDeckDue: 2,
		core.ReviewIntervalsAscending: 3, core.ReviewIntervalsDescending: 4, core.ReviewDifficultyDescending: 5,
		core.ReviewDifficultyAscending: 6, core.ReviewRetrievabilityAscending: 7, core.ReviewRelativeOverdueness: 12}
	mix = map[string]int{core.NewMix: 0, core.NewAfterReviews: 1, core.NewBeforeReviews: 2}
	// Interday learning orders use the same enum (ReviewMix).
	interdayMix    = map[string]int{core.InterdayMix: 0, core.InterdayAfter: 1, core.InterdayBefore: 2}
	questionAction = map[string]int{core.QuestionShowAnswer: 0, core.QuestionShowReminder: 1}
	answerAction   = map[string]int{core.AnswerBuryCard: 0, core.AnswerAgain: 1, core.AnswerGood: 2,
		core.AnswerHard: 3, core.AnswerShowReminder: 4}
)

func minutes(seconds []int) []float64 {
	out := []float64{}
	for _, s := range seconds {
		out = append(out, float64(s)/60)
	}
	return out
}

// preset returns a schema 11 deck config (Anki's legacy "dconf" JSON).
func preset(id int64, name string, o core.DeckOptions, mod int64) map[string]any {
	order := 1 // due
	if o.InsertionOrder == core.InsertRandom {
		order = 0
	}
	leech := 1 // tag only
	if o.LeechAction == core.LeechSuspend {
		leech = 0
	}
	usn := -1
	if id == ankiDefaultID {
		usn = 0
	}
	timer := 0
	if o.ShowTimer {
		timer = 1
	}
	return map[string]any{
		"id": id, "mod": mod, "name": name, "usn": usn, "maxTaken": o.MaxAnswerSeconds, "autoplay": !o.DisableAutoplay,
		"timer": timer, "replayq": !o.SkipQuestionOnReplay, "dyn": false,
		"new": map[string]any{
			"bury": o.BuryNew, "delays": minutes(o.LearningSteps), "initialFactor": 2500,
			"ints": []int{1, 4, 0}, "order": order, "perDay": o.NewPerDay,
		},
		"rev": map[string]any{
			"bury": o.BuryReview, "ease4": 1.3, "ivlFct": 1.0, "maxIvl": o.MaxInterval,
			"perDay": o.ReviewsPerDay, "hardFactor": 1.2,
		},
		"lapse": map[string]any{
			"delays": minutes(o.RelearningSteps), "leechAction": leech, "leechFails": o.LeechThreshold,
			"minInt": o.MinRelearnInterval, "mult": 0.0,
		},
		"newMix": mix[o.NewReviewOrder], "newPerDayMinimum": 0, "interdayLearningMix": interdayMix[o.InterdayLearningOrder],
		"reviewOrder": reviewOrder[o.ReviewSortOrder], "newSortOrder": sortOrder[o.NewSortOrder],
		"newGatherPriority": gatherPriority[o.NewGatherOrder], "buryInterdayLearning": o.BuryInterdayLearning,
		"fsrsWeights": []float64{}, "fsrsParams5": []float64{}, "fsrsParams6": []float64{},
		"desiredRetention": o.DesiredRetention, "ignoreRevlogsBeforeDate": "",
		"easyDaysPercentages": []float64{1, 1, 1, 1, 1, 1, 1}, "stopTimerOnAnswer": o.StopTimerOnAnswer,
		"secondsToShowQuestion": float64(o.SecondsToShowQuestion), "secondsToShowAnswer": float64(o.SecondsToShowAnswer),
		"questionAction": questionAction[o.QuestionAction], "answerAction": answerAction[o.AnswerAction],
		"waitForAudio": true, "sm2Retention": 0.9, "weightSearch": "",
	}
}

// dummyMessage is the note of the dummy collection.anki2, as Anki writes it.
const dummyMessage = "Please update to the latest Anki version, then import the .colpkg/.apkg file again."

// dummyCollection is the collection.anki2 that clients which cannot read
// collection.anki21 see: one note asking to update Anki.
func dummyCollection(now time.Time) ([]byte, error) {
	set := &core.ExportSet{
		Settings:  core.DefaultSettings(),
		Decks:     []core.Deck{{ID: ankiDefaultID, Name: core.DefaultDeckName, PresetID: core.DefaultPresetID}},
		NoteTypes: []core.NoteType{{ID: core.ID(now.UnixMilli()), Name: "Basic", Kind: core.Standard, Fields: []string{"Front", "Back"}, Templates: []core.CardTemplate{{Name: "Card 1", QFmt: "{{Front}}", AFmt: "{{FrontSide}}\n\n<hr id=answer>\n\n{{Back}}"}}}},
	}
	id := core.ID(now.UnixMilli())
	set.Notes = []core.Note{{ID: id, TypeID: set.NoteTypes[0].ID, GUID: "cardflip-dummy", Fields: []string{dummyMessage, ""}, Modified: now}}
	set.Cards = []core.Card{{ID: id, NoteID: id, DeckID: ankiDefaultID, NewPos: 1}}
	return writeCollection(Input{Set: set, Now: now}, Options{}, func() {})
}

// FileName returns the file name of an export: the deck's name (its last
// component) or "collection", with characters file systems refuse
// replaced.
func FileName(name, ext string) string {
	if i := strings.LastIndex(name, core.DeckSeparator); i >= 0 {
		name = name[i+len(core.DeckSeparator):]
	}
	name = strings.Map(func(r rune) rune {
		if r < 0x20 || strings.ContainsRune(`<>:"/\|?*`, r) {
			return '_'
		}
		return r
	}, name)
	name = strings.TrimRight(strings.TrimSpace(name), ". ")
	if name == "" {
		name = "export"
	}
	return fmt.Sprintf("%s.%s", name, ext)
}
