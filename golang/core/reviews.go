package core

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"cardflip/golang/scheduler"
	"cardflip/golang/storage"
)

func monthKey(t time.Time) string { return t.Format("2006-01") }

// loadRevlog loads the review log months that the current study day can
// touch: this month and the previous one (the day may have started there).
func (c *Collection) loadRevlog() error {
	now := c.now()
	for _, m := range []string{monthKey(now.AddDate(0, 0, -2)), monthKey(now)} {
		if err := c.loadRevlogMonth(m); err != nil {
			return err
		}
	}
	return nil
}

// revlogInMemory reports whether c.revlog holds the entries of month: a
// loaded month, or one after them.
func (c *Collection) revlogInMemory(month string) bool {
	if c.revlogMonths[month] {
		return true
	}
	for m := range c.revlogMonths {
		if m >= month {
			return false
		}
	}
	return true
}

func (c *Collection) loadRevlogMonth(month string) error {
	if c.revlogMonths[month] {
		return nil
	}
	entries, err := c.store.ReadRevlog(month)
	if err != nil {
		return err
	}
	for i, raw := range entries {
		var e ReviewEntry
		if err := json.Unmarshal(raw, &e); err != nil {
			return fmt.Errorf("%w: review log %s entry %d: %v", ErrCorrupt, month, i, err)
		}
		c.revlog = append(c.revlog, &e)
	}
	c.revlogMonths[month] = true
	return nil
}

// deckOptions returns the options of a card's home deck.
func (c *Collection) deckOptions(card *Card) DeckOptions { return c.optionsOf(card.HomeDeck()) }

// params returns the scheduling parameters of a card's home deck.
func (c *Collection) params(card *Card) scheduler.Params {
	opts := c.deckOptions(card)
	p := scheduler.DefaultParams(opts.DesiredRetention)
	p.LearningSteps = durations(opts.LearningSteps)
	p.RelearningSteps = durations(opts.RelearningSteps)
	p.MaximumInterval = opts.MaxInterval
	p.MinimumRelearnInterval = opts.MinRelearnInterval
	p.RolloverHour = c.settings.RolloverHour
	p.Weights = opts.Weights
	p.EasyDays = opts.EasyDays
	p.DueLoad = c.load.count
	return p
}

func durations(secs []int) []time.Duration {
	out := make([]time.Duration, len(secs))
	for i, s := range secs {
		out[i] = time.Duration(s) * time.Second
	}
	return out
}

// Preview is what one rating would do to a card.
type Preview struct {
	Rating   scheduler.Rating `json:"rating"`
	Interval time.Duration    `json:"interval"`
	// Label is the interval for a rating button, e.g. "<10m", "3d", "1.2mo".
	Label string `json:"label"`
}

// Previews returns the outcome of each rating for a card, without changing
// it.
func (c *Collection) Previews(cardID ID) ([4]Preview, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	card, ok := c.cards[cardID]
	if !ok {
		return [4]Preview{}, userErr("cardGone", "That card no longer exists.")
	}
	var out [4]Preview
	if delays, ok := c.previewDelays(card); ok {
		for i, secs := range append(delays[:], 0) {
			d := time.Duration(secs) * time.Second
			out[i] = Preview{Rating: scheduler.Rating(i + 1), Interval: d, Label: FormatInterval(d)}
			if secs == 0 {
				out[i].Label = PreviewEndLabel
			}
		}
		return out, nil
	}
	for i, o := range scheduler.Preview(card.CardState, c.now(), c.params(card), int64(card.ID)) {
		r := scheduler.Rating(i + 1)
		out[i] = Preview{Rating: r, Interval: o.Interval, Label: FormatInterval(o.Interval)}
	}
	return out, nil
}

// PreviewEndLabel is the rating button label of a preview answer that
// returns the card to its home deck.
const PreviewEndLabel = "End"

// previewDelays returns the preview delays of the filtered deck a card is
// in, when that deck previews its cards. Callers hold a lock.
func (c *Collection) previewDelays(card *Card) ([3]int, bool) {
	if d, ok := c.decks[card.DeckID]; ok && card.HomeDeckID != 0 && d.Filter != nil && !d.Filter.Reschedule {
		return d.Filter.PreviewDelays, true
	}
	return [3]int{}, false
}

// FormatInterval formats an interval for a rating button.
func FormatInterval(d time.Duration) string {
	const day = 24 * time.Hour
	switch {
	case d < time.Hour:
		m := int(math.Ceil(d.Minutes()))
		if m < 1 {
			m = 1
		}
		return fmt.Sprintf("<%dm", m)
	case d < day:
		return trimZero(d.Hours()) + "h"
	}
	days := math.Round(d.Hours() / 24)
	switch {
	case days < 30:
		return fmt.Sprintf("%dd", int(days))
	case days < 365:
		return trimZero(days/30) + "mo"
	}
	return trimZero(days/365) + "y"
}

func trimZero(v float64) string {
	s := strconv.FormatFloat(math.Round(v*10)/10, 'f', 1, 64)
	return strings.TrimSuffix(s, ".0")
}

// Answer rates a card, saves the new scheduling state and the review log
// entry together, and returns the entry.
//
// A card in a filtered deck is scheduled with its home deck's options; one
// reviewed before it is due is scheduled from the actual time since its
// last review. It returns home once it is a review card again. In a
// filtered deck that previews its cards, the answer does not change the
// card's scheduling: Again, Hard and Good show it again after the deck's
// preview delays, and Easy or a delay of 0 returns it home.
func (c *Collection) Answer(cardID ID, rating scheduler.Rating, duration time.Duration) (ReviewEntry, error) {
	if !rating.Valid() {
		return ReviewEntry{}, userErr("invalidRating", "Invalid rating %d.", rating)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	card, ok := c.cards[cardID]
	if !ok {
		return ReviewEntry{}, userErr("cardGone", "That card no longer exists.")
	}
	now := c.now()
	duration = min(max(duration, 0), c.deckOptions(card).MaxAnswerTime())
	entry := &ReviewEntry{
		ID: c.ids.next(), CardID: card.ID, Time: now, Rating: rating,
		StateBefore: card.State, IvlBefore: card.ScheduledDays, IvlAfter: card.ScheduledDays,
		DurationMs: duration.Milliseconds(), Before: card.CardState,
	}
	if card.HomeDeckID != 0 {
		entry.Filtered = &FilteredPlace{Deck: card.DeckID, Home: card.HomeDeckID, Pos: card.FilteredPos, PreviewDue: card.PreviewDue}
	}
	next := *card
	next.BuriedUntil = time.Time{}
	if delays, ok := c.previewDelays(card); ok {
		entry.Preview = true
		if rating == scheduler.Easy || delays[rating-1] == 0 {
			next.returnHome()
		} else {
			next.PreviewDue = now.Add(time.Duration(delays[rating-1]) * time.Second)
		}
		return c.commitAnswer(&op{Kind: "answer", Cards: []*Card{&next}, Review: entry}, entry)
	}
	out := scheduler.Answer(card.CardState, rating, now, c.params(card), int64(card.ID))
	entry.IvlAfter = out.State.ScheduledDays
	next.CardState = out.State
	if card.HomeDeckID != 0 && out.State.State == scheduler.Review {
		next.returnHome()
	}
	if card.State == scheduler.New && next.OriginalPos == nil {
		pos := card.NewPos
		next.OriginalPos = &pos
	}
	o := &op{Kind: "answer", Cards: []*Card{&next}, Review: entry}
	if rating == scheduler.Again && card.State == scheduler.Review && isLeech(out.State.Lapses, c.deckOptions(card).LeechThreshold) {
		c.markLeech(o, &next, entry, now)
	}
	return c.commitAnswer(o, entry)
}

// commitAnswer saves answer o with its review log entry. Callers hold the
// write lock.
func (c *Collection) commitAnswer(o *op, entry *ReviewEntry) (ReviewEntry, error) {
	raw, err := json.Marshal(entry)
	if err != nil {
		return ReviewEntry{}, err
	}
	if err := c.commit(o, storage.RevlogWrite{Month: monthKey(entry.Time), ID: int64(entry.ID), Entry: raw}); err != nil {
		return ReviewEntry{}, err
	}
	return *entry, nil
}

// isLeech reports whether a lapse count makes a card a leech: it reaches
// the threshold, or the threshold plus a multiple of half the threshold
// rounded up (scheduling spec, "Leech detection").
func isLeech(lapses, threshold int) bool {
	if threshold < 1 || lapses < threshold {
		return false
	}
	return (lapses-threshold)%((threshold+1)/2) == 0
}

// markLeech adds to o, the answer that made next a leech, the leech tag on
// its note and, with the suspend action, the card's suspension, and records
// in entry what it changed. Callers hold the write lock.
func (c *Collection) markLeech(o *op, next *Card, entry *ReviewEntry, now time.Time) {
	entry.Leech = true
	if n, ok := c.notes[next.NoteID]; ok && !hasTag(n.Tags, LeechTagName) {
		nn := cloneNote(n)
		nn.Tags = append(nn.Tags, LeechTagName)
		nn.Modified = now
		o.Notes = append(o.Notes, &nn)
		entry.LeechTagged = true
	}
	if c.deckOptions(next).LeechAction == LeechSuspend && !next.Suspended {
		next.Suspended = true
		entry.LeechSuspended = true
	}
}

func hasTag(tags []string, tag string) bool {
	for _, t := range tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}

// withoutTag returns tags without tag (ignoring case).
func withoutTag(tags []string, tag string) []string {
	var out []string
	for _, t := range tags {
		if !strings.EqualFold(t, tag) {
			out = append(out, t)
		}
	}
	return out
}

// UndoReview undoes an answer: the card gets back its exact scheduling
// state from before the answer, a suspension and leech tag the answer added
// are removed, and the review log entry is removed. Only a
// card's most recent answer can be undone. When the answer is the most
// recent step of the undo history, it is undone there (so Redo can bring it
// back); otherwise reverting it is a new step.
func (c *Collection) UndoReview(entryID ID) (cardID ID, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry := c.reviewEntry(entryID)
	if entry != nil && (entry.Source != "" || entry.Rating == 0) {
		return 0, userErr("undoOnlyCardflip", "Only answers given in Cardflip can be undone.")
	}
	if n := len(c.hist.undo); entry != nil && n > 0 && c.hist.undo[n-1].review == entryID {
		if _, err := c.undoLocked(); err != nil {
			return 0, err
		}
		return entry.CardID, nil
	}
	if entry == nil {
		return 0, userErr("nothingToUndo", "There is nothing to undo.")
	}
	for _, e := range c.revlog {
		if e.CardID == entry.CardID && e.ID > entry.ID {
			return 0, userErr("undoOnlyLatest", "Only the most recent answer to a card can be undone.")
		}
	}
	card, ok := c.cards[entry.CardID]
	if !ok {
		return 0, userErr("cardGone", "That card no longer exists.")
	}
	restored := *card
	restored.CardState = entry.Before
	if entry.LeechSuspended {
		restored.Suspended = false
	}
	// Put the card back in the filtered deck it was answered in, if it
	// still exists and the card is in no other filtered deck.
	if f := entry.Filtered; f != nil && (card.HomeDeckID == 0 || card.DeckID == f.Deck) {
		if d, ok := c.decks[f.Deck]; ok && d.Filter != nil && c.decks[f.Home] != nil {
			restored.DeckID, restored.HomeDeckID, restored.FilteredPos, restored.PreviewDue = f.Deck, f.Home, f.Pos, f.PreviewDue
		}
	}
	o := &op{Kind: "undoAnswer", Cards: []*Card{&restored}, ReviewUndone: entry.ID}
	if n, ok := c.notes[card.NoteID]; ok && entry.LeechTagged && hasTag(n.Tags, LeechTagName) {
		nn := cloneNote(n)
		nn.Tags = withoutTag(n.Tags, LeechTagName)
		nn.Modified = c.now()
		o.Notes = append(o.Notes, &nn)
	}
	if err := c.commit(o, storage.RevlogWrite{Month: monthKey(entry.Time), ID: int64(entry.ID)}); err != nil {
		return 0, err
	}
	return card.ID, nil
}

// reviewWrites are the review log writes that store add and remove del.
func reviewWrites(add, del []*ReviewEntry) ([]storage.RevlogWrite, error) {
	out := make([]storage.RevlogWrite, 0, len(add)+len(del))
	for _, e := range del {
		out = append(out, storage.RevlogWrite{Month: monthKey(e.Time), ID: int64(e.ID)})
	}
	for _, e := range add {
		raw, err := json.Marshal(e)
		if err != nil {
			return nil, err
		}
		out = append(out, storage.RevlogWrite{Month: monthKey(e.Time), ID: int64(e.ID), Entry: raw})
	}
	return out, nil
}

// todayEntries returns the review log entries of the current study day.
// Callers hold a lock.
func (c *Collection) todayEntries(now time.Time) []*ReviewEntry {
	// Open loaded the months the day can start in; a month that began while
	// the app was running only has entries made in this session, which are
	// already in memory.
	start := scheduler.DayStart(now, c.settings.RolloverHour)
	var out []*ReviewEntry
	for _, e := range c.revlog {
		if !e.Time.Before(start) && !e.Time.After(now) {
			out = append(out, e)
		}
	}
	return out
}

// ReviewsExist reports which of the review log entries ids exist.
func (c *Collection) ReviewsExist(ids []ID) map[ID]bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	want := make(map[ID]bool, len(ids))
	for _, id := range ids {
		want[id] = true
	}
	out := map[ID]bool{}
	for _, e := range c.revlog {
		if want[e.ID] {
			out[e.ID] = true
		}
	}
	return out
}

// CardHistory returns a card's whole stored review log, oldest first.
func (c *Collection) CardHistory(cardID ID) ([]ReviewEntry, error) {
	months, err := c.store.RevlogMonths()
	if err != nil {
		return nil, err
	}
	// Only entries mentioning the card's ID are decoded.
	needle := []byte(`"card":` + strconv.FormatInt(int64(cardID), 10))
	var out []ReviewEntry
	for _, m := range months {
		c.yield()
		entries, err := c.store.ReadRevlog(m)
		if err != nil {
			return nil, err
		}
		for i, raw := range entries {
			if !bytes.Contains(raw, needle) {
				continue
			}
			var e ReviewEntry
			if err := json.Unmarshal(raw, &e); err != nil {
				return nil, fmt.Errorf("%w: review log %s entry %d: %v", ErrCorrupt, m, i, err)
			}
			if e.CardID == cardID {
				out = append(out, e)
			}
		}
	}
	slices.SortStableFunc(out, func(a, b ReviewEntry) int {
		if c := a.Time.Compare(b.Time); c != 0 {
			return c
		}
		return cmp.Compare(a.ID, b.ID)
	})
	return out, nil
}

// ReviewsOfCard returns a card's review log entries in the loaded months.
func (c *Collection) ReviewsOfCard(cardID ID) []ReviewEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []ReviewEntry
	for _, e := range c.revlog {
		if e.CardID == cardID {
			out = append(out, *e)
		}
	}
	return out
}
