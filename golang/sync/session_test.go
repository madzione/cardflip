package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math/rand/v2"
	"os"
	"slices"
	"strconv"
	"strings"
	gosync "sync"
	"sync/atomic"
	"testing"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/scheduler"
	"cardflip/golang/storage"
)

var errBroken = errors.New("connection broken")

// pipe is an in-memory connection. Messages go through JSON, as on the
// network. After failAfter messages (if positive) it breaks.
type pipe struct {
	a2b, b2a  chan []byte
	closed    chan struct{}
	once      gosync.Once
	sent      atomic.Int64
	failAfter int64
}

func newPipe(failAfter int64) *pipe {
	return &pipe{a2b: make(chan []byte, 64), b2a: make(chan []byte, 64), closed: make(chan struct{}), failAfter: failAfter}
}

func (p *pipe) close() { p.once.Do(func() { close(p.closed) }) }

type pipeEnd struct {
	p       *pipe
	in, out chan []byte
}

func (p *pipe) ends() (*pipeEnd, *pipeEnd) {
	return &pipeEnd{p, p.b2a, p.a2b}, &pipeEnd{p, p.a2b, p.b2a}
}

func (e *pipeEnd) Send(m *Message) error {
	if n := e.p.sent.Add(1); e.p.failAfter > 0 && n > e.p.failAfter {
		e.p.close()
		return errBroken
	}
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	select {
	case e.out <- data:
		return nil
	case <-e.p.closed:
		return errBroken
	}
}

func (e *pipeEnd) Recv() (*Message, error) {
	var data []byte
	select {
	case data = <-e.in:
	case <-e.p.closed:
		// As on the network, what was sent before the close arrives.
		select {
		case data = <-e.in:
		default:
			return nil, errBroken
		}
	}
	var m Message
	return &m, json.Unmarshal(data, &m)
}

// clock is the time of all devices; each device's clock is offset from it.
type clock struct {
	mu gosync.Mutex
	t  time.Time
}

func (c *clock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *clock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

// device is a device with one profile.
type device struct {
	t       *testing.T
	id      string
	name    string
	col     *core.Collection
	clk     *clock
	offset  time.Duration
	mu      gosync.Mutex
	links   map[string]Link // by peer device ID, as saved
	backups int
	proto   int
}

func newDevice(t *testing.T, clk *clock, id, name string, offset time.Duration) *device {
	t.Helper()
	d := &device{t: t, id: id, name: name, clk: clk, offset: offset, links: map[string]Link{}}
	col, _, err := core.Open(storage.NewMemoryStore().Open(), core.Options{Now: d.now})
	if err != nil {
		t.Fatal(err)
	}
	col.SetDevice(id)
	d.col = col
	return d
}

func (d *device) now() time.Time { return d.clk.now().Add(d.offset) }

func (d *device) link(peer *device) {
	id := min(d.id, peer.id) + "-" + max(d.id, peer.id)
	d.links[peer.id] = NewLink(id, peer.id, "default")
	peer.links[d.id] = NewLink(id, d.id, "default")
}

func (d *device) opts() Options {
	return Options{Device: d.id, Name: d.name, Now: d.now, MaxItems: 13, Protocol: d.proto,
		Backup: func() error {
			d.mu.Lock()
			defer d.mu.Unlock()
			d.backups++
			return nil
		},
		SaveLink: func(l Link) error {
			d.mu.Lock()
			defer d.mu.Unlock()
			d.links[l.Peer] = l
			return nil
		}}
}

func (d *device) savedLink(peer string) Link {
	d.mu.Lock()
	defer d.mu.Unlock()
	l := d.links[peer]
	l.Echo = slices.Clone(l.Echo)
	return l
}

// runSession runs one session started by a, with b responding.
func runSession(a, b *device, choice Choice, failAfter int64) (sa, sb Summary, ea, eb error) {
	p := newPipe(failAfter)
	ca, cb := p.ends()
	var wg gosync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer p.close()
		l := a.savedLink(b.id)
		sa, ea = Initiate(context.Background(), ca, a.col, &l, choice, a.opts())
	}()
	go func() {
		defer wg.Done()
		defer p.close()
		lookup := func(peer, link string) (Collection, *Link, func(), error) {
			l := b.savedLink(peer)
			if l.ID != link {
				return b.col, nil, nil, nil
			}
			return b.col, &l, nil, nil
		}
		sb, eb = Respond(context.Background(), cb, a.id, lookup, b.opts())
	}()
	wg.Wait()
	return sa, sb, ea, eb
}

func mustSync(t *testing.T, a, b *device, choice Choice) (Summary, Summary) {
	t.Helper()
	sa, sb, ea, eb := runSession(a, b, choice, 0)
	if ea != nil || eb != nil {
		t.Fatalf("sync %s → %s: %v / %v", a.name, b.name, ea, eb)
	}
	return sa, sb
}

// content returns what must be the same on all devices: the snapshot
// without sync metadata and device settings, the review log and the media
// files notes refer to.
func content(t *testing.T, d *device) string {
	t.Helper()
	snap, err := d.col.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	var s map[string]any
	if err := json.Unmarshal(snap, &s); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"epoch", "seq", "meta", "tombs", "revSeq", "nextId", "mediaBuilt"} {
		delete(s, k)
	}
	settings := s["settings"].(map[string]any)
	s["settings"] = map[string]any{"rolloverHour": settings["rolloverHour"], "flagNames": settings["flagNames"]}
	if m, ok := s["media"].(map[string]any); ok {
		for _, e := range m {
			delete(e.(map[string]any), "m") // Missing and the sequence number are local
			delete(e.(map[string]any), "s")
		}
	}
	var log []string
	months, _ := d.col.RevlogMonths()
	for _, m := range months {
		entries, _ := d.col.ReadRevlog(m)
		for _, e := range entries {
			log = append(log, string(e))
		}
	}
	slices.Sort(log)
	s["revlog"] = log
	media := map[string]string{}
	for _, id := range d.col.AllNoteIDs() {
		n, _ := d.col.Note(id)
		for _, name := range core.MediaNames(n.Fields...) {
			data, err := d.col.Media(name)
			if err != nil {
				media[name] = "MISSING"
				continue
			}
			media[name] = fmt.Sprintf("%x", data)
		}
	}
	s["files"] = media
	out, _ := json.MarshalIndent(s, "", " ")
	return string(out)
}

func diff(a, b string) string {
	la, lb := strings.Split(a, "\n"), strings.Split(b, "\n")
	for i := range min(len(la), len(lb)) {
		if la[i] != lb[i] {
			lo, hi := max(0, i-5), min(len(la), i+5)
			return fmt.Sprintf("line %d:\n%s\n--- vs ---\n%s", i, strings.Join(la[lo:hi], "\n"), strings.Join(lb[lo:min(len(lb), hi)], "\n"))
		}
	}
	return fmt.Sprintf("lengths %d and %d", len(la), len(lb))
}

func sameContent(t *testing.T, devs ...*device) {
	t.Helper()
	want := content(t, devs[0])
	for _, d := range devs[1:] {
		if got := content(t, d); got != want {
			t.Fatalf("%s and %s differ: %s", devs[0].name, d.name, diff(want, got))
		}
	}
}

func basic(t *testing.T, c *core.Collection) core.NoteType {
	for _, nt := range c.NoteTypes() {
		if nt.Name == "Basic" {
			return nt
		}
	}
	t.Fatal("no Basic note type")
	return core.NoteType{}
}

func addNote(t *testing.T, c *core.Collection, deck core.ID, fields ...string) core.ID {
	t.Helper()
	id, err := c.AddNote(core.NoteInput{TypeID: basic(t, c).ID, DeckID: deck, Fields: fields})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// randomOp changes the collection in a random way. Failures of the
// operation itself (a name taken, nothing to undo) are fine.
func randomOp(t *testing.T, r *rand.Rand, d *device, n int) {
	t.Helper()
	c := d.col
	notes := c.AllNoteIDs()
	decks := c.Decks()
	pickNote := func() (core.ID, bool) {
		if len(notes) == 0 {
			return 0, false
		}
		return notes[r.IntN(len(notes))], true
	}
	pickCard := func() (core.ID, bool) {
		id, ok := pickNote()
		if !ok {
			return 0, false
		}
		cards := c.CardsOfNote(id)
		if len(cards) == 0 {
			return 0, false
		}
		return cards[r.IntN(len(cards))].ID, true
	}
	deck := decks[r.IntN(len(decks))]
	for deck.Filter != nil {
		deck = decks[r.IntN(len(decks))]
	}
	switch op := r.IntN(15); op {
	case 0, 1, 2:
		addNote(t, c, deck.ID, fmt.Sprintf("%s q%d", d.name, n), "a")
	case 3:
		if id, ok := pickNote(); ok {
			_ = c.UpdateNote(id, []string{fmt.Sprintf("%s edit %d", d.name, n), "b"}, nil)
		}
	case 4:
		if id, ok := pickNote(); ok {
			_ = c.DeleteNotes([]core.ID{id})
		}
	case 5, 6:
		if id, ok := pickCard(); ok {
			_, _ = c.Answer(id, scheduler.Rating(1+r.IntN(4)), time.Second)
		}
	case 7:
		if id, ok := pickCard(); ok {
			_ = c.SetFlag([]core.ID{id}, 1+r.IntN(7))
		}
	case 8:
		if id, ok := pickCard(); ok {
			_ = c.SetSuspended([]core.ID{id}, r.IntN(2) == 0)
		}
	case 9:
		_, _ = c.CreateDeck(fmt.Sprintf("Deck %d", r.IntN(4)))
	case 10:
		if id, ok := pickCard(); ok {
			_ = c.MoveCards([]core.ID{id}, deck.ID)
		}
	case 11:
		name, err := c.AddMedia(fmt.Sprintf("img%d.png", r.IntN(3)), []byte(fmt.Sprintf("%s image %d", d.name, n)))
		if err == nil {
			addNote(t, c, deck.ID, fmt.Sprintf(`<img src="%s">`, name), "pic")
		}
	case 12:
		_, _ = c.Undo()
	case 13:
		s := c.Settings()
		s.RolloverHour = r.IntN(24)
		_ = c.UpdateSettings(s)
	case 14:
		if len(decks) > 2 && r.IntN(3) == 0 {
			d := decks[r.IntN(len(decks))]
			if d.Name != core.DefaultDeckName {
				_ = c.DeleteDeck(d.ID)
			}
		} else {
			_, _ = c.CloneNoteType(basic(t, c).ID, fmt.Sprintf("Type %d", r.IntN(3)))
		}
	}
}

// Three devices with separately created collections merge, change their
// collections at random and sync in random pairs, some sessions breaking
// off; once they have synced enough, all three are the same.
// CARDFLIP_SYNC_SEEDS sets the number of runs, for stress testing.
func TestConvergence(t *testing.T) {
	seeds := uint64(16)
	if n, err := strconv.ParseUint(os.Getenv("CARDFLIP_SYNC_SEEDS"), 10, 64); err == nil {
		seeds = n
	}
	for seed := range seeds {
		t.Run(fmt.Sprint(seed), func(t *testing.T) {
			r := rand.New(rand.NewPCG(seed, 99))
			clk := &clock{t: time.Date(2026, 10, 5, 12, 0, 0, 0, time.Local)}
			// The same moment on two devices makes colliding IDs.
			devs := []*device{newDevice(t, clk, "AAAA", "Laptop", 0), newDevice(t, clk, "BBBB", "Phone", 0),
				newDevice(t, clk, "CCCC", "Tablet", 3*time.Minute)}
			for i, a := range devs {
				for _, b := range devs[i+1:] {
					a.link(b)
				}
			}
			for _, d := range devs {
				for i := range 5 {
					randomOp(t, r, d, i)
				}
			}
			// First syncs: merge.
			mustSync(t, devs[0], devs[1], ChoiceMerge)
			mustSync(t, devs[2], devs[1], ChoiceMerge)
			mustSync(t, devs[0], devs[2], ChoiceMerge)
			for step := range 250 {
				clk.advance(time.Duration(r.IntN(90)) * time.Second)
				randomOp(t, r, devs[r.IntN(3)], step+10)
				if r.IntN(4) == 0 {
					i := r.IntN(3)
					j := (i + 1 + r.IntN(2)) % 3
					fail := int64(0)
					if r.IntN(3) == 0 {
						fail = 1 + r.Int64N(40)
					}
					_, _, ea, eb := runSession(devs[i], devs[j], ChoiceNone, fail)
					if fail == 0 && (ea != nil || eb != nil) {
						t.Fatalf("step %d: sync %s → %s: %v / %v", step, devs[i].name, devs[j].name, ea, eb)
					}
				}
			}
			for range 3 {
				for i, a := range devs {
					for _, b := range devs[i+1:] {
						mustSync(t, a, b, ChoiceNone)
					}
				}
			}
			sameContent(t, devs...)
			for _, d := range devs {
				if missing := d.col.MissingMedia(); len(missing) > 0 {
					t.Errorf("%s misses media %v", d.name, missing)
				}
			}
		})
	}
}

// A first sync without the user's choice asks for it and changes nothing.
func TestFirstSyncNeedsChoice(t *testing.T) {
	clk := &clock{t: time.Date(2026, 10, 5, 12, 0, 0, 0, time.Local)}
	a, b := newDevice(t, clk, "AAAA", "Laptop", 0), newDevice(t, clk, "BBBB", "Phone", time.Second)
	a.link(b)
	addNote(t, a.col, a.col.Decks()[0].ID, "only here", "")
	revA, revB := a.col.Revision(), b.col.Revision()
	_, _, ea, eb := runSession(a, b, ChoiceNone, 0)
	var r *Refusal
	if !errors.As(ea, &r) || r.Code != RefuseNeedsChoice || r.Device != "Phone" {
		t.Fatalf("initiator: %v", ea)
	}
	if eb == nil {
		t.Fatal("responder did not refuse")
	}
	if !a.savedLink(b.id).NeedsChoice {
		t.Fatal("link not marked as needing the choice")
	}
	// Building the media manifest is the only change.
	if a.col.Revision() > revA+1 || b.col.Revision() > revB+1 || len(b.col.AllNoteIDs()) != 0 {
		t.Fatal("something changed before the choice")
	}
	if a.backups+b.backups != 0 {
		t.Fatal("backup made before the choice")
	}
	sa, sb := mustSync(t, a, b, ChoiceMerge)
	if sa.Mode != ModeMerge || sb.Mode != ModeMerge || a.backups != 1 || b.backups != 1 {
		t.Fatalf("summaries %+v / %+v, backups %d / %d", sa, sb, a.backups, b.backups)
	}
	if l := a.savedLink(b.id); l.NeedsChoice || l.Epoch != a.col.Epoch() || l.PeerEpoch != b.col.Epoch() {
		t.Fatalf("link after merging: %+v", l)
	}
	sameContent(t, a, b)
	// Then incremental syncs.
	addNote(t, b.col, b.col.Decks()[0].ID, "from the phone", "")
	if sa, _ := mustSync(t, a, b, ChoiceNone); sa.Mode != ModeSync || sa.Received == 0 {
		t.Fatalf("summary %+v", sa)
	}
	sameContent(t, a, b)
}

// Replacing the other device's profile, or this one, makes it exactly the
// sender's, after a backup of it; a restore afterwards asks again.
func TestReplace(t *testing.T) {
	for _, choice := range []Choice{ChoiceReplaceOther, ChoiceReplaceThis} {
		t.Run(string(choice), func(t *testing.T) {
			clk := &clock{t: time.Date(2026, 10, 5, 12, 0, 0, 0, time.Local)}
			a, b := newDevice(t, clk, "AAAA", "Laptop", 0), newDevice(t, clk, "BBBB", "Phone", time.Second)
			a.link(b)
			sender, receiver := a, b
			if choice == ChoiceReplaceThis {
				sender, receiver = b, a
			}
			n := addNote(t, sender.col, sender.col.Decks()[0].ID, `<img src="cat.png">`, "sender")
			if _, err := sender.col.AddMedia("cat.png", []byte("meow")); err != nil {
				t.Fatal(err)
			}
			if _, err := sender.col.Answer(sender.col.CardsOfNote(n)[0].ID, scheduler.Good, time.Second); err != nil {
				t.Fatal(err)
			}
			addNote(t, receiver.col, receiver.col.Decks()[0].ID, "receiver's own", "")
			if _, err := receiver.col.AddMedia("cat.png", []byte("a different cat")); err != nil {
				t.Fatal(err)
			}
			sa, sb := mustSync(t, a, b, choice)
			sr := sb
			if receiver == a {
				sr = sa
			}
			if sr.Mode != ModeReplaced || sr.Media != 1 {
				t.Fatalf("receiver's summary %+v", sr)
			}
			if receiver.backups != 1 || sender.backups != 0 {
				t.Fatalf("backups: receiver %d, sender %d", receiver.backups, sender.backups)
			}
			sameContent(t, a, b)
			if data, err := receiver.col.Media("cat.png"); err != nil || !bytes.Equal(data, []byte("meow")) {
				t.Fatalf("receiver's cat.png = %q, %v", data, err)
			}
			if a.col.Epoch() != b.col.Epoch() {
				t.Fatal("epochs differ after replacing")
			}
			// Incremental syncs both ways afterwards.
			addNote(t, receiver.col, receiver.col.Decks()[0].ID, "after", "")
			addNote(t, sender.col, sender.col.Decks()[0].ID, "after too", "")
			mustSync(t, a, b, ChoiceNone)
			sameContent(t, a, b)
			if len(a.col.AllNoteIDs()) != 3 {
				t.Fatalf("%d notes", len(a.col.AllNoteIDs()))
			}
			// A restore of the receiver makes the next sync ask again.
			snap, _ := receiver.col.Snapshot()
			if err := receiver.col.Restore(snap, nil, nil, nil); err != nil {
				t.Fatal(err)
			}
			if _, _, ea, _ := runSession(a, b, ChoiceNone, 0); !isRefusal(ea, RefuseNeedsChoice) {
				t.Fatalf("sync after restore: %v", ea)
			}
		})
	}
}

func isRefusal(err error, code string) bool {
	var r *Refusal
	return errors.As(err, &r) && r.Code == code
}

// Devices whose clocks differ by more than 10 minutes refuse to sync,
// naming the other device and the difference.
func TestClocksMustAgree(t *testing.T) {
	clk := &clock{t: time.Date(2026, 10, 5, 12, 0, 0, 0, time.Local)}
	a, b := newDevice(t, clk, "AAAA", "Laptop", 0), newDevice(t, clk, "BBBB", "Phone", 2*time.Hour)
	a.link(b)
	_, _, ea, eb := runSession(a, b, ChoiceMerge, 0)
	var ra, rb *Refusal
	if !errors.As(ea, &ra) || ra.Code != RefuseClock || ra.Device != "Phone" || !strings.Contains(ra.Error(), "differ by 2 hours") {
		t.Fatalf("laptop: %v", ea)
	}
	if !errors.As(eb, &rb) || rb.Code != RefuseClock || rb.Device != "Laptop" || !strings.Contains(rb.Error(), "differ by 2 hours") {
		t.Fatalf("phone: %v", eb)
	}
	if len(b.col.AllNoteIDs())+len(a.col.AllNoteIDs()) != 0 || a.backups+b.backups != 0 {
		t.Fatal("something was exchanged")
	}
	// 9 minutes are fine.
	b.offset = 9 * time.Minute
	mustSync(t, a, b, ChoiceMerge)
}

// A device refuses to sync with one whose sync format it does not support,
// saying which device to update; newer devices still sync with older ones
// whose format they support.
func TestCompatibleVersions(t *testing.T) {
	clk := &clock{t: time.Date(2026, 10, 5, 12, 0, 0, 0, time.Local)}
	a, b := newDevice(t, clk, "AAAA", "Laptop", 0), newDevice(t, clk, "BBBB", "Phone", time.Second)
	a.link(b)
	a.proto = ProtocolVersion + 1
	optsA := a.opts()
	optsA.MinProtocol = ProtocolVersion + 1
	p := newPipe(0)
	ca, cb := p.ends()
	var ea, eb error
	var wg gosync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer p.close()
		l := a.savedLink(b.id)
		_, ea = Initiate(context.Background(), ca, a.col, &l, ChoiceMerge, optsA)
	}()
	go func() {
		defer wg.Done()
		defer p.close()
		_, eb = Respond(context.Background(), cb, a.id, func(peer, link string) (Collection, *Link, func(), error) {
			l := b.savedLink(peer)
			return b.col, &l, nil, nil
		}, b.opts())
	}()
	wg.Wait()
	var ra, rb *Refusal
	if !errors.As(eb, &rb) || rb.Code != RefuseVersion || !rb.Mine || !strings.Contains(rb.Error(), "on this device") {
		t.Fatalf("phone: %v", eb)
	}
	if !errors.As(ea, &ra) || ra.Code != RefuseVersion || ra.Mine || ra.Device != "Phone" || !strings.Contains(ra.Error(), "on Phone") {
		t.Fatalf("laptop: %v", ea)
	}
	if a.backups+b.backups != 0 {
		t.Fatal("something was exchanged")
	}
	// A newer laptop that still speaks the phone's version syncs.
	mustSync(t, a, b, ChoiceMerge)
}

func TestMergeSlots(t *testing.T) {
	got := MergeSlots(map[string]int{"A": 0, "C": 1}, map[string]int{"B": 1, "C": 2}, "A", "D")
	// C keeps 1 (the smaller of its two), so B, which has the larger ID
	// than... B < C, so C moves; D is new.
	want := map[string]int{"A": 0, "B": 1, "C": 2, "D": 3}
	if !maps.Equal(got, want) {
		t.Fatalf("MergeSlots = %v, want %v", got, want)
	}
	if again := MergeSlots(map[string]int{"B": 1, "C": 2}, map[string]int{"A": 0, "C": 1}, "D", "A"); !maps.Equal(again, got) {
		t.Fatalf("not symmetric: %v vs %v", again, got)
	}
}
