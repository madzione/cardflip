//go:build !js

package main

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"
	gosync "sync"
	"testing"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/scheduler"
	"cardflip/golang/storage"
)

// e2eDevice is a whole app on its own data directory, with real storage.
type e2eDevice struct {
	name     string
	a        *app
	app      *AppService
	notes    *NotesService
	study    *StudyService
	decks    *DecksService
	media    *MediaService
	sync     *SyncService
	progress *ProgressService
	mu       gosync.Mutex
	asked    int // first-sync questions
}

// e2eUI is a user who accepts pairing, confirms codes and merges.
type e2eUI struct {
	fakeSyncUI
	d *e2eDevice
}

func (u *e2eUI) ChooseFirstSync(ctx context.Context, l LinkInfo) (string, error) {
	u.d.mu.Lock()
	defer u.d.mu.Unlock()
	u.d.asked++
	return "merge", nil
}

func newE2EDevice(t *testing.T, name string) *e2eDevice {
	t.Helper()
	return newE2EDeviceAt(t, name, time.Now)
}

// newE2EDeviceAt is newE2EDevice with the clock now.
func newE2EDeviceAt(t *testing.T, name string, now func() time.Time) *e2eDevice {
	t.Helper()
	a := newApp()
	a.now = now
	a.backups.tick = 0
	d := &e2eDevice{name: name, a: a, app: &AppService{a: a}, notes: &NotesService{a: a}, study: &StudyService{a: a},
		decks: &DecksService{a: a}, media: &MediaService{a: a}, progress: &ProgressService{a: a}}
	ui := &e2eUI{fakeSyncUI: fakeSyncUI{accept: true, confirm: func(context.Context, string) bool { return true }}, d: d}
	d.sync = newSyncService(a, ui)
	if _, err := d.app.Open(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	if _, err := d.sync.Init(name); err != nil {
		t.Fatal(err)
	}
	d.sync.SetNetwork(NetworkWiFi, "Home", "192.168.1.0/24")
	check(t, d.sync.TrustNetwork())
	t.Cleanup(func() {
		_ = d.sync.UntrustNetwork(d.sync.Status().Network.ID)
		a.mu.Lock()
		a.closeLocked()
		a.mu.Unlock()
	})
	return d
}

func (d *e2eDevice) addNote(t *testing.T, front string) {
	t.Helper()
	def := must(d.notes.Defaults(0))
	must(d.notes.Add(NewNote{TypeID: def.TypeID, DeckID: def.DeckID, Fields: []string{front, "back"}}))
}

// picture is the state both devices must agree on: decks with their due
// counts, notes, reviews and media.
func (d *e2eDevice) picture(t *testing.T) string {
	t.Helper()
	var b strings.Builder
	for _, n := range must(d.decks.Tree()) {
		fmt.Fprintf(&b, "deck %s: %d cards, due %+v\n", n.FullName, n.Cards, n.Counts)
	}
	c := must(d.a.collection())
	var fronts []string
	for _, id := range c.AllNoteIDs() {
		n := must(c.Note(id))
		fronts = append(fronts, n.Fields[0])
		for _, name := range core.MediaNames(n.Fields...) {
			data, err := d.media.Get(name)
			fmt.Fprintf(&b, "media %s: %q %v\n", name, data, err)
		}
	}
	reviews := 0
	for _, m := range must(c.RevlogMonths()) {
		reviews += len(must(c.ReadRevlog(m)))
	}
	fmt.Fprintf(&b, "notes %q\nreviews %d\n", fronts, reviews)
	return b.String()
}

func samePicture(t *testing.T, what string, a, b *e2eDevice) string {
	t.Helper()
	pa, pb := a.picture(t), b.picture(t)
	if pa != pb {
		t.Fatalf("%s: %s and %s differ:\n%s\n---\n%s", what, a.name, b.name, pa, pb)
	}
	return pa
}

func syncNow(t *testing.T, d *e2eDevice) SyncSummary {
	t.Helper()
	sum, err := d.sync.SyncNow(context.Background())
	if err != nil || sum.Error != "" || !sum.Reachable {
		t.Fatalf("%s: sync now = %+v, %v", d.name, sum, err)
	}
	return sum
}

// Two apps pair, link and merge their profiles, study on one side and
// sync, and agree on due counts, reviews and media; a restore makes the
// next sync ask again.
func TestSyncEndToEnd(t *testing.T) {
	laptop, phone := newE2EDevice(t, "Laptop"), newE2EDevice(t, "Phone")
	cat := must(laptop.media.Add("cat.png", []byte("meow")))
	laptop.addNote(t, fmt.Sprintf(`uno <img src="%s">`, cat))
	laptop.addNote(t, "dos")
	laptop.addNote(t, "tres")
	dog := must(phone.media.Add("dog.png", []byte("woof")))
	phone.addNote(t, fmt.Sprintf(`cuatro <img src="%s">`, dog))

	// Merge.
	link(t, laptop, phone)
	if laptop.asked != 1 {
		t.Fatalf("asked %d times", laptop.asked)
	}
	p := samePicture(t, "after merging", laptop, phone)
	if !strings.Contains(p, `media cat.png: "meow"`) || !strings.Contains(p, `media dog.png: "woof"`) ||
		!strings.Contains(p, "4 cards") {
		t.Fatalf("merged:\n%s", p)
	}

	// Study on the phone, then sync from the phone.
	deck := must(phone.decks.Tree())[0].ID
	st := must(phone.study.Start(deck))
	for range 3 {
		if st.Card == nil {
			t.Fatal("nothing to study")
		}
		st = must(phone.study.Answer(st.Card.CardID, 3, 2000))
	}
	if sum := syncNow(t, phone); sum.Sent == 0 {
		t.Fatalf("the phone sent nothing: %+v", sum)
	}
	p = samePicture(t, "after studying", laptop, phone)
	if !strings.Contains(p, "reviews 3") {
		t.Fatalf("after studying:\n%s", p)
	}

	// Restoring the backup made before the merge asks again.
	backups := must(laptop.app.Backups()).Backups
	if len(backups) == 0 {
		t.Fatal("no backup before the merge")
	}
	check(t, laptop.app.RestoreBackup(backups[len(backups)-1].Name))
	if strings.Contains(laptop.picture(t), "cuatro") {
		t.Fatal("the restored backup has the phone's note")
	}
	syncNow(t, laptop)
	if laptop.asked != 2 {
		t.Fatalf("asked %d times after the restore", laptop.asked)
	}
	p = samePicture(t, "after the restore", laptop, phone)
	if !strings.Contains(p, "reviews 3") || !strings.Contains(p, "cuatro") {
		t.Fatalf("after the restore:\n%s", p)
	}
	if !bytes.Contains([]byte(p), []byte("tres")) {
		t.Fatal("the laptop's note is gone")
	}
}

// link pairs and links two devices' default profiles and merges them.
func link(t *testing.T, laptop, phone *e2eDevice) {
	t.Helper()
	check(t, phone.sync.StartPairing())
	check(t, laptop.sync.PairWith(context.Background(), "127.0.0.1", phone.sync.Status().Port))
	phone.sync.StopPairing()
	laptop.sync.Seen(phone.sync.Status().DeviceID, "Phone", "127.0.0.1", phone.sync.Status().Port, false)
	phone.sync.Seen(laptop.sync.Status().DeviceID, "Laptop", "127.0.0.1", laptop.sync.Status().Port, false)
	if _, err := laptop.sync.Link(context.Background(), phone.sync.Status().DeviceID, storage.DefaultProfile); err != nil {
		t.Fatal(err)
	}
	syncNow(t, laptop)
}

// studyNew adds n notes and answers their cards Good.
func (d *e2eDevice) studyNew(t *testing.T, n int) {
	t.Helper()
	for i := range n {
		d.addNote(t, fmt.Sprintf("%s %d %d", d.name, d.a.now().UnixNano(), i))
	}
	st := must(d.study.Start(must(d.decks.Tree())[0].ID))
	for range n {
		if st.Card == nil {
			t.Fatalf("%s: nothing to study", d.name)
		}
		st = must(d.study.Answer(st.Card.CardID, 3, 2000))
	}
}

// Two devices (spec): the phone studies yesterday, both study today, and
// after syncing both report the same study days and streak.
func TestSyncStreaks(t *testing.T) {
	var mu gosync.Mutex
	offset := time.Duration(0)
	now := func() time.Time {
		mu.Lock()
		defer mu.Unlock()
		return time.Now().Add(offset)
	}
	laptop, phone := newE2EDeviceAt(t, "Laptop", now), newE2EDeviceAt(t, "Phone", now)
	link(t, laptop, phone)

	phone.studyNew(t, 2) // yesterday, on the phone only
	// An hour into the next study day.
	mu.Lock()
	t0 := time.Now()
	offset = scheduler.NextDayStart(t0, core.DefaultRolloverHour).Add(time.Hour).Sub(t0)
	mu.Unlock()
	phone.studyNew(t, 1) // today on both
	laptop.studyNew(t, 2)
	// The laptop asked before the sync, so its cached days must follow the
	// synced reviews.
	if p := must(laptop.progress.Today()); p.Current != 1 || p.TodayCards != 2 {
		t.Fatalf("laptop before syncing: %+v", p)
	}
	syncNow(t, phone)
	syncNow(t, laptop)
	samePicture(t, "after studying", laptop, phone)
	lp, pp := must(laptop.progress.Today()), must(phone.progress.Today())
	if !reflect.DeepEqual(lp, pp) {
		t.Fatalf("laptop %+v, phone %+v", lp, pp)
	}
	if lp.Current != 2 || !lp.TodayStudied || lp.Longest != 2 || lp.TodayCards != 3 {
		t.Fatalf("after syncing: %+v", lp)
	}
	ld := must(must(laptop.a.collection()).StudyDays(core.DefaultRolloverHour, time.Local))
	pd := must(must(phone.a.collection()).StudyDays(core.DefaultRolloverHour, time.Local))
	if fmt.Sprint(ld.Days) != fmt.Sprint(pd.Days) || len(ld.Days) != 2 {
		t.Fatalf("study days: laptop %v, phone %v", ld.Days, pd.Days)
	}
}

// One device earns a tier: each device tells about it once (spec: earned
// on another device; achievements shown per device).
func TestSyncAchievements(t *testing.T) {
	laptop, phone := newE2EDevice(t, "Laptop"), newE2EDevice(t, "Phone")
	link(t, laptop, phone)
	for _, d := range []*e2eDevice{laptop, phone} {
		if p := must(d.progress.Today()); len(p.NewAchievements) != 0 { // the first evaluation
			t.Fatalf("%s: %+v", d.name, p.NewAchievements)
		}
	}
	// 99 answers yesterday on the phone, then the 100th today.
	c := must(phone.a.collection())
	yesterday := scheduler.DayStart(c.Now(), core.DefaultRolloverHour).Add(-12 * time.Hour)
	var history []core.ImportReview
	for i := range 99 {
		history = append(history, core.ImportReview{Time: yesterday.Add(time.Duration(i) * time.Second), Rating: scheduler.Good,
			StateBefore: scheduler.Review, IvlBefore: 1, IvlAfter: 1, DurationMs: 1000})
	}
	basic, _ := c.BasicNoteType()
	if _, err := c.Import(&core.ImportBatch{
		NoteTypes: []core.ImportNoteType{{Key: "basic", ExistingID: basic}},
		Notes: []core.ImportNote{{TypeKey: "basic", GUID: "old", DeckName: "Default", Fields: []string{"old", "a"},
			Cards: []core.ImportCard{{DeckName: "Default", History: history,
				State: scheduler.CardState{State: scheduler.Review, Due: c.Now().AddDate(0, 0, 5), ScheduledDays: 5}}}}},
	}, nil); err != nil {
		t.Fatal(err)
	}
	phone.studyNew(t, 1)
	shows := func(d *e2eDevice) []string {
		t.Helper()
		var keys []string
		for _, a := range must(d.progress.Today()).NewAchievements {
			keys = append(keys, a.Key)
		}
		check(t, d.progress.MarkSeen(keys))
		return keys
	}
	if got := shows(phone); fmt.Sprint(got) != "[answers:100]" {
		t.Fatalf("phone: %v", got)
	}
	syncNow(t, phone)
	syncNow(t, laptop)
	if got := shows(laptop); fmt.Sprint(got) != "[answers:100]" {
		t.Fatalf("laptop after syncing: %v", got)
	}
	date := must(laptop.progress.Achievements())[0].Earned
	if want := must(phone.progress.Achievements())[0].Earned; fmt.Sprint(date) != fmt.Sprint(want) || len(date) != 1 {
		t.Fatalf("earned: laptop %v, phone %v", date, want)
	}
	// Shown once: neither device tells again after another sync.
	syncNow(t, laptop)
	syncNow(t, phone)
	for _, d := range []*e2eDevice{laptop, phone} {
		if got := shows(d); len(got) != 0 {
			t.Fatalf("%s again: %v", d.name, got)
		}
	}
}
