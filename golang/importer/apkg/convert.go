package apkg

import (
	"cardflip/golang/internal/usermsg"
	"encoding/json"
	"fmt"
	"math"
	"path"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/scheduler"
)

// unsupportedMedia are file extensions of audio and video in formats the
// app cannot play on every platform; they are skipped with a warning. MP3,
// OGG, Opus, M4A/AAC, WAV, FLAC, WEBM and MP4 are imported.
var unsupportedMedia = map[string]bool{
	".spx": true, ".amr": true, ".mov": true, ".mkv": true,
	".avi": true, ".mpg": true, ".mpeg": true, ".ogv": true, ".3gp": true,
}

// Convert turns a package into an import batch. With keepScheduling the
// cards keep their Anki state (memory states from Anki or estimated) and
// flag; otherwise every card is New and unflagged, and notes lose the
// leech and marked tags.
func Convert(p *Package, keepScheduling bool) (*core.ImportBatch, error) {
	b := &core.ImportBatch{Media: map[string][]byte{}}
	crt := time.Unix(p.Crt, 0)

	used := map[int64]bool{}
	for _, n := range p.Notes {
		used[n.TypeID] = true
	}
	ids := make([]int64, 0, len(used))
	for id := range used {
		if _, ok := p.NoteTypes[id]; !ok {
			return nil, unreadable("apkgMissingNoteType", "a note uses a missing note type")
		}
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		b.NoteTypes = append(b.NoteTypes, importNoteType(id, p.NoteTypes[id], p.Decks))
	}

	b.Presets, b.DeckPresets = convertPresets(p)
	for _, d := range p.Decks {
		if !d.Filtered && strings.TrimSpace(d.Description) != "" {
			if b.DeckDescriptions == nil {
				b.DeckDescriptions = map[string]string{}
			}
			b.DeckDescriptions[d.Name] = d.Description
		}
	}

	// deckName is the name of a regular deck. withFiltered also names a
	// filtered deck, for its cards without a home deck (design decision 1).
	deckName := func(id int64, withFiltered bool) string {
		if d, ok := p.Decks[id]; ok && (!d.Filtered || withFiltered) {
			return d.Name
		}
		return ""
	}
	cardsByNote := map[int64][]Card{}
	for _, c := range p.Cards {
		cardsByNote[c.NoteID] = append(cardsByNote[c.NoteID], c)
	}
	revlogByCard := map[int64][]Review{}
	if keepScheduling {
		for _, r := range p.Revlog {
			revlogByCard[r.CardID] = append(revlogByCard[r.CardID], r)
		}
	}
	for i, n := range p.Notes {
		in := core.ImportNote{
			TypeKey: strconv.FormatInt(n.TypeID, 10), GUID: n.GUID, Fields: n.Fields, Tags: n.Tags,
			Modified: unixTime(n.Mod), Cards: []core.ImportCard{},
		}
		if !keepScheduling {
			in.Tags = slices.DeleteFunc(slices.Clone(n.Tags), func(t string) bool {
				return strings.EqualFold(t, core.LeechTagName) || strings.EqualFold(t, core.MarkedTag)
			})
		}
		for _, c := range cardsByNote[n.ID] {
			deck := c.DeckID
			due := c.Due
			if c.ODeckID != 0 { // in a filtered deck: use the home deck
				deck, due = c.ODeckID, c.ODue
			}
			// Without a home deck, a card in a filtered deck goes to a regular
			// deck of that name, as in old shared decks.
			ic := core.ImportCard{Ord: c.Ord, DeckName: deckName(deck, c.ODeckID == 0)}
			if ic.DeckName == "" {
				ic.DeckName = core.DefaultDeckName
			}
			if in.DeckName == "" {
				in.DeckName = ic.DeckName
			}
			if keepScheduling {
				ac := scheduler.AnkiCard{Type: c.Type, Queue: c.Queue, Due: due, Ivl: c.Ivl, Factor: c.Factor,
					Reps: c.Reps, Lapses: c.Lapses, Left: c.Left, Memory: memoryFromData(c.Data)}
				for _, r := range revlogByCard[c.ID] {
					ac.History = append(ac.History, scheduler.AnkiReview{Time: time.UnixMilli(r.ID), Ease: r.Ease,
						Ivl: r.Ivl, Factor: r.Factor, Type: r.Type})
				}
				ic.State = scheduler.FromAnki(ac, crt)
				ic.Suspended = c.Queue == -1
				ic.Flag = c.Flag
				ic.Modified = unixTime(c.Mod)
				ic.Order = due*1_000_000 + int64(i)
				if pos, ok := originalPosFromData(c.Data); ok {
					o := pos*1_000_000 + int64(i)
					ic.OriginalOrder = &o
				}
				ic.History = convertHistory(revlogByCard[c.ID])
			} else {
				ic.State = scheduler.CardState{State: scheduler.New}
				ic.Order = int64(i)
			}
			if ic.State.State != scheduler.New {
				ic.Order = 0
			}
			in.Cards = append(in.Cards, ic)
		}
		if len(in.Cards) == 0 {
			in.Cards = nil // generate cards
			in.DeckName = core.DefaultDeckName
		}
		b.Notes = append(b.Notes, in)
	}

	var skipped []string
	for name, data := range p.Media {
		if unsupportedMedia[strings.ToLower(path.Ext(name))] {
			skipped = append(skipped, name)
			continue
		}
		b.Media[name] = data
	}
	if len(skipped) > 0 {
		sort.Strings(skipped)
		noun := "files were"
		if len(skipped) == 1 {
			noun = "file was"
		}
		b.Warnings = append(b.Warnings, usermsg.Tag("apkgMediaSkipped", usermsg.Args(len(skipped), strings.Join(skipped, ", ")),
			fmt.Sprintf("%d audio/video %s skipped, because Cardflip cannot play their format: %s",
				len(skipped), noun, strings.Join(skipped, ", "))))
	}
	return b, nil
}

// importNoteType converts a note type of the package, whose decks are
// decks.
func importNoteType(id int64, nt NoteType, decks map[int64]Deck) core.ImportNoteType {
	t := core.NoteType{Name: nt.Name, Kind: core.Standard, Fields: nt.Fields, CSS: nt.CSS, Source: fmt.Sprintf("anki:%d", id),
		Modified: unixTime(nt.Mod), LatexSVG: nt.LatexSVG, FieldOptions: fieldOptions(nt), SortField: nt.SortField}
	if t.SortField < 0 || t.SortField >= len(t.Fields) {
		t.SortField = 0
	}
	if nt.Cloze {
		t.Kind = core.Cloze
		if len(nt.Templates) > 1 {
			nt.Templates = nt.Templates[:1]
		}
	}
	var overrides []string
	for i, tm := range nt.Templates {
		t.Templates = append(t.Templates, core.CardTemplate{Name: tm.Name, QFmt: tm.QFmt, AFmt: tm.AFmt,
			BrowserQFmt: tm.BrowserQFmt, BrowserAFmt: tm.BrowserAFmt})
		if d, ok := decks[tm.DeckID]; ok && tm.DeckID != 0 && !d.Filtered {
			if overrides == nil {
				overrides = make([]string, len(nt.Templates))
			}
			overrides[i] = d.Name
		}
	}
	return core.ImportNoteType{Key: strconv.FormatInt(id, 10), Type: t, TemplateDecks: overrides}
}

// Anki's editing font and size for new fields, which Cardflip's editor
// defaults stand for.
const (
	ankiFont = "Arial"
	ankiSize = 20
)

// fieldOptions converts the field options of a note type of the package;
// nil while every field has the defaults.
func fieldOptions(nt NoteType) []core.FieldOptions {
	var out []core.FieldOptions
	for i := range nt.Fields {
		var f FieldConfig
		if i < len(nt.FieldConfigs) {
			f = nt.FieldConfigs[i]
		}
		o := core.FieldOptions{HTMLByDefault: f.PlainText, Sticky: f.Sticky, RTL: f.RTL, Collapsed: f.Collapsed,
			ExcludeFromSearch: f.ExcludeFromSearch, Font: strings.TrimSpace(f.Font), Size: max(f.Size, 0)}
		if o.Font == ankiFont {
			o.Font = ""
		}
		if o.Size == ankiSize {
			o.Size = 0
		}
		out = append(out, o)
	}
	for len(out) > 0 && out[len(out)-1] == (core.FieldOptions{}) {
		out = out[:len(out)-1]
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// unixTime converts epoch seconds to a time; 0 (unknown) is the zero time.
func unixTime(sec int64) time.Time {
	if sec <= 0 {
		return time.Time{}
	}
	return time.Unix(sec, 0)
}

// ConvertCollection turns a collection package into the batch of a new
// collection: like Convert with scheduling, plus the note types no note
// uses and every deck, empty ones included. Filtered decks are left out;
// their cards are in their home decks.
func ConvertCollection(p *Package) (*core.ImportBatch, error) {
	b, err := Convert(p, true)
	if err != nil {
		return nil, err
	}
	used := map[string]bool{}
	for _, nt := range b.NoteTypes {
		used[nt.Key] = true
	}
	ids := make([]int64, 0, len(p.NoteTypes))
	for id := range p.NoteTypes {
		if !used[strconv.FormatInt(id, 10)] {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		if nt := p.NoteTypes[id]; len(nt.Fields) > 0 && len(nt.Templates) > 0 {
			b.NoteTypes = append(b.NoteTypes, importNoteType(id, nt, p.Decks))
		}
	}
	for _, d := range p.Decks {
		if !d.Filtered {
			b.Decks = append(b.Decks, d.Name)
		}
	}
	sort.Strings(b.Decks)
	return b, nil
}

// convertPresets maps the presets of the package's decks (design decision
// 3). Decks using Anki's default preset get Cardflip's Default, as Anki
// itself maps them, so they get no entry.
func convertPresets(p *Package) ([]core.ImportPreset, map[string]string) {
	var presets []core.ImportPreset
	deckPresets := map[string]string{}
	added := map[int64]bool{}
	for _, d := range p.Decks {
		pr, ok := p.Presets[d.PresetID]
		if d.Filtered || d.PresetID == AnkiDefaultPreset || !ok {
			continue
		}
		key := strconv.FormatInt(d.PresetID, 10)
		deckPresets[d.Name] = key
		if !added[d.PresetID] {
			added[d.PresetID] = true
			presets = append(presets, core.ImportPreset{Key: key, Name: pr.Name, Options: presetOptions(pr)})
		}
	}
	sort.Slice(presets, func(i, j int) bool { return presets[i].Key < presets[j].Key })
	return presets, deckPresets
}

// Anki's enum values (deck_config.proto) and the Cardflip options they map
// to; values without a counterpart keep Cardflip's default.
var (
	ankiGatherPriority = map[int]string{0: core.GatherDeck, 5: core.GatherDeckRandomNotes, 1: core.GatherAscending,
		2: core.GatherDescending, 3: core.GatherRandomNotes, 4: core.GatherRandomCards}
	ankiSortOrder = map[int]string{0: core.SortTypeThenGathered, 1: core.SortGathered, 2: core.SortTypeThenRandom,
		3: core.SortRandomNoteThenType, 4: core.SortRandom}
	// Ascending ease is descending difficulty, as Anki labels it with FSRS.
	ankiReviewOrder = map[int]string{0: core.ReviewDueRandom, 1: core.ReviewDueDeck, 2: core.ReviewDeckDue,
		3: core.ReviewIntervalsAscending, 4: core.ReviewIntervalsDescending, 5: core.ReviewDifficultyDescending,
		6: core.ReviewDifficultyAscending, 7: core.ReviewRetrievabilityAscending, 12: core.ReviewRelativeOverdueness}
	ankiMix            = map[int]string{0: core.NewMix, 1: core.NewAfterReviews, 2: core.NewBeforeReviews}
	ankiQuestionAction = map[int]string{0: core.QuestionShowAnswer, 1: core.QuestionShowReminder}
	ankiAnswerAction   = map[int]string{0: core.AnswerBuryCard, 1: core.AnswerAgain, 2: core.AnswerGood,
		3: core.AnswerHard, 4: core.AnswerShowReminder}
)

// presetOptions maps an Anki preset's options to Cardflip's. Values outside
// Cardflip's ranges are clamped; zero values Anki never writes (retention,
// intervals, leech threshold) keep the defaults.
func presetOptions(p Preset) core.DeckOptions {
	o := core.DefaultDeckOptions()
	o.NewPerDay = min(max(p.NewPerDay, 0), 9999)
	o.ReviewsPerDay = min(max(p.ReviewsPerDay, 0), 99999)
	o.LearningSteps, o.RelearningSteps = stepSeconds(p.LearnSteps), stepSeconds(p.RelearnSteps)
	if p.DesiredRetention > 0 {
		r := math.Round(p.DesiredRetention*1e4) / 1e4
		o.DesiredRetention = min(max(r, core.MinRetention), core.MaxRetention)
	}
	if p.MaxInterval > 0 {
		o.MaxInterval = min(p.MaxInterval, core.MaxIntervalDays)
	}
	o.MinRelearnInterval = min(max(p.MinLapseInterval, 1), o.MaxInterval)
	if p.LeechThreshold > 0 {
		o.LeechThreshold = min(p.LeechThreshold, core.MaxLeechThreshold)
	}
	if p.LeechAction == 0 {
		o.LeechAction = core.LeechSuspend
	}
	if p.InsertOrder == 1 {
		o.InsertionOrder = core.InsertRandom
	}
	set := func(dst *string, m map[int]string, v int) {
		if s, ok := m[v]; ok {
			*dst = s
		}
	}
	set(&o.NewGatherOrder, ankiGatherPriority, p.GatherPriority)
	set(&o.NewSortOrder, ankiSortOrder, p.SortOrder)
	set(&o.ReviewSortOrder, ankiReviewOrder, p.ReviewOrder)
	set(&o.NewReviewOrder, ankiMix, p.NewMix)
	set(&o.InterdayLearningOrder, ankiMix, p.InterdayMix)
	o.BuryNew, o.BuryReview, o.BuryInterdayLearning = p.BuryNew, p.BuryReviews, p.BuryInterdayLearning
	o.DisableAutoplay, o.SkipQuestionOnReplay = p.DisableAutoplay, p.SkipQuestionWhenReplaying
	o.SecondsToShowQuestion, o.SecondsToShowAnswer = autoAdvanceSeconds(p.SecondsToShowQuestion), autoAdvanceSeconds(p.SecondsToShowAnswer)
	set(&o.QuestionAction, ankiQuestionAction, p.QuestionAction)
	set(&o.AnswerAction, ankiAnswerAction, p.AnswerAction)
	return o
}

// autoAdvanceSeconds rounds Anki's auto advance time, which may have a
// fraction, to whole seconds: a time under half a second stays on as 1.
func autoAdvanceSeconds(s float64) int {
	if s <= 0 {
		return 0
	}
	return int(min(max(math.Round(s), 1), core.MaxAutoAdvanceSeconds))
}

// stepSeconds converts Anki's step delays in minutes to seconds, keeping
// at most core.MaxSteps valid ones.
func stepSeconds(minutes []float64) []int {
	out := []int{}
	for _, m := range minutes {
		if s := int(math.Round(m * 60)); s > 0 && s <= core.MaxStepSeconds && len(out) < core.MaxSteps {
			out = append(out, s)
		}
	}
	return out
}

// convertHistory maps a card's Anki review log, oldest first. Anki's manual
// and rescheduled entries become manual entries (rating 0), and reviews in
// filtered decks become reviews. The state before learning entries is
// tracked through the history, so that the first one (or the first after a
// reset) is a New card's.
func convertHistory(rs []Review) []core.ImportReview {
	var out []core.ImportReview
	cur := scheduler.New
	for _, r := range rs {
		if r.Ease < 0 || r.Ease > 4 {
			continue
		}
		e := core.ImportReview{
			Time: time.UnixMilli(r.ID), Rating: scheduler.Rating(r.Ease),
			IvlBefore: ankiDays(r.LastIvl), IvlAfter: ankiDays(r.Ivl), DurationMs: max(r.Time, 0),
			Source: fmt.Sprintf("anki:%d", r.ID),
		}
		if r.Ease == 0 || r.Type == 4 || r.Type == 5 {
			// A reset to new leaves the interval at 0; a set due date or a
			// reschedule makes the card a review card.
			e.Rating = 0
			e.StateBefore = cur
			cur = scheduler.Review
			if r.Ivl == 0 {
				cur = scheduler.New
			}
			out = append(out, e)
			continue
		}
		switch r.Type {
		case 0:
			e.StateBefore = scheduler.Learning
			if cur == scheduler.New {
				e.StateBefore = scheduler.New
			}
		case 2:
			e.StateBefore = scheduler.Relearning
		default: // 1 review, 3 filtered
			e.StateBefore = scheduler.Review
		}
		switch {
		case r.Type == 3 && r.Ivl == 0:
			// A cram review that did not reschedule the card.
		case r.Ivl > 0:
			cur = scheduler.Review
		case e.StateBefore == scheduler.Review || e.StateBefore == scheduler.Relearning:
			cur = scheduler.Relearning
		default:
			cur = scheduler.Learning
		}
		out = append(out, e)
	}
	return out
}

// ankiDays converts an Anki interval to days: positive values are days,
// negative ones seconds.
func ankiDays(ivl int64) int {
	if ivl < 0 {
		return int(-ivl / 86400)
	}
	return int(ivl)
}

func memoryFromData(data string) *scheduler.Memory {
	if data == "" {
		return nil
	}
	var d struct {
		S *float64 `json:"s"`
		D *float64 `json:"d"`
	}
	if json.Unmarshal([]byte(data), &d) != nil || d.S == nil || d.D == nil || *d.S <= 0 {
		return nil
	}
	return &scheduler.Memory{Stability: *d.S, Difficulty: *d.D}
}

// originalPosFromData returns the original new-card position that Anki
// keeps in a card's data ("pos") once the card has been studied.
func originalPosFromData(data string) (int64, bool) {
	if data == "" {
		return 0, false
	}
	var d struct {
		Pos *int64 `json:"pos"`
	}
	if json.Unmarshal([]byte(data), &d) != nil || d.Pos == nil || *d.Pos < 0 {
		return 0, false
	}
	return *d.Pos, true
}

// ReadCollection reads a .colpkg file into the batch of a new collection
// (see ConvertCollection), for core.Collection.ReplaceWithImport.
func ReadCollection(data []byte, yield func()) (*core.ImportBatch, error) {
	p, err := Read(data, yield)
	if err != nil {
		return nil, err
	}
	return ConvertCollection(p)
}

// Options configure an .apkg import.
type Options struct {
	// KeepScheduling keeps the cards' Anki scheduling (see Convert).
	KeepScheduling bool
	// UpdateNotes, UpdateNoteTypes and MergeNoteTypes decide what happens
	// to notes and note types that exist already (see core.ImportBatch).
	UpdateNotes     core.UpdatePolicy
	UpdateNoteTypes core.UpdatePolicy
	MergeNoteTypes  bool
}

// Import reads an .apkg file and adds it to the collection in one step.
func Import(c *core.Collection, data []byte, opts Options, progress core.Progress, yield func()) (core.ImportSummary, error) {
	if progress != nil {
		progress("reading", 0, 1)
	}
	p, err := Read(data, yield)
	if err != nil {
		return core.ImportSummary{}, err
	}
	b, err := Convert(p, opts.KeepScheduling)
	if err != nil {
		return core.ImportSummary{}, err
	}
	b.UpdateNotes, b.UpdateNoteTypes, b.MergeNoteTypes = opts.UpdateNotes, opts.UpdateNoteTypes, opts.MergeNoteTypes
	return c.Import(b, progress)
}
