package backup

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"cardflip/golang/storage"
)

func TestAutoNameRoundTrip(t *testing.T) {
	at := time.Date(2026, 10, 13, 9, 5, 7, 123_456_789, time.UTC)
	name := AutoName(at)
	if name != "backup-2026-10-13-090507-123.cardflip" {
		t.Fatalf("AutoName = %q", name)
	}
	if err := storage.ValidBackupName(name); err != nil {
		t.Fatal(err)
	}
	got, ok := ParseAutoName(name)
	if !ok || !got.Equal(at.Truncate(time.Millisecond)) {
		t.Fatalf("ParseAutoName = %v, %v", got, ok)
	}
	for _, other := range []string{"cardflip-2026-10-13.cardflip", "backup-2026-10-13.cardflip", "backup-x.cardflip", ""} {
		if _, ok := ParseAutoName(other); ok {
			t.Errorf("ParseAutoName(%q) accepted", other)
		}
	}
}

// daily returns one backup at 10:00 on each of the days days back from now.
func daily(days int) []Auto {
	var out []Auto
	for i := 1; i <= days; i++ {
		at := time.Date(now.Year(), now.Month(), now.Day()-i, 10, 0, 0, 0, now.Location())
		out = append(out, Auto{Name: AutoName(at), CreatedAt: at})
	}
	return out
}

func todays(hours ...int) []Auto {
	var out []Auto
	for _, h := range hours {
		at := time.Date(now.Year(), now.Month(), now.Day(), h, 0, 0, 0, now.Location())
		out = append(out, Auto{Name: AutoName(at), CreatedAt: at})
	}
	return out
}

// kept returns the local dates of the backups Obsolete keeps, newest first.
func kept(all []Auto, limits Limits) []string {
	gone := map[string]bool{}
	for _, b := range Obsolete(all, now, limits) {
		gone[b.Name] = true
	}
	var out []string
	for _, b := range all {
		if !gone[b.Name] {
			out = append(out, b.CreatedAt.In(now.Location()).Format("01-02 15h"))
		}
	}
	return out
}

func TestThinningSixtyDays(t *testing.T) {
	// now is Tuesday 2026-10-13; one backup a day for 60 days, three today.
	all := append(todays(11, 9, 8), daily(60)...)
	got := kept(all, Limits{Daily: 8, Weekly: 8, Monthly: 8})
	want := []string{
		"10-13 11h", "10-13 09h", "10-13 08h", // today: all
		"10-12 10h", "10-11 10h", "10-10 10h", "10-09 10h", "10-08 10h", "10-07 10h", "10-06 10h", "10-05 10h", // 8 daily
		// 8 weekly (Sundays, W40–W33) and, in between, the latest of
		// September and of August as monthly ones. Aug 14 and 15 share
		// week and month with kept backups.
		"10-04 10h", "09-30 10h" /* monthly */, "09-27 10h", "09-20 10h", "09-13 10h", "09-06 10h",
		"08-31 10h" /* monthly */, "08-30 10h", "08-23 10h", "08-16 10h",
	}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("kept\n%v\nwant\n%v", got, want)
	}
}

func TestThinningMonthly(t *testing.T) {
	got := kept(daily(400), Limits{Daily: 8, Weekly: 8, Monthly: 8})
	if len(got) != 24 {
		t.Fatalf("kept %d: %v", len(got), got)
	}
	// After 8 daily, 8 weekly backups with September and August as monthly
	// ones in between (see TestThinningSixtyDays), six more months.
	older := strings.Join(got[18:], ",")
	if older != "07-31 10h,06-30 10h,05-31 10h,04-30 10h,03-31 10h,02-28 10h" {
		t.Fatalf("oldest monthly = %v", older)
	}
}

func TestThinningLimits(t *testing.T) {
	all := append(todays(9), daily(30)...)
	if got := kept(all, Limits{}); len(got) != 1 {
		t.Fatalf("all limits 0: kept %v, want only today's", got)
	}
	got := kept(all, Limits{Daily: 2, Weekly: 1})
	if strings.Join(got, ",") != "10-13 09h,10-12 10h,10-11 10h,10-04 10h" {
		t.Fatalf("kept %v", got)
	}
	// Gaps: the latest backup of each day that has one counts, however old.
	sparse := []Auto{daily(5)[4], daily(40)[39]}
	if got := kept(sparse, Limits{Daily: 8}); len(got) != 2 {
		t.Fatalf("sparse: kept %v", got)
	}
}

func TestWriteAutoThinsAndIsCollectionOnly(t *testing.T) {
	c := populated(t)
	if len(mediaOf(t, c)) == 0 {
		t.Fatal("test collection has no media")
	}
	bs := storage.NewMemory().Backups()
	// Something that is not an automatic backup is never touched.
	if err := bs.Write("notes.txt", []byte("mine")); err != nil {
		t.Fatal(err)
	}
	limits := Limits{Daily: 2, Weekly: 0, Monthly: 0}
	var last Auto
	for i := 5; i >= 0; i-- {
		at := now.AddDate(0, 0, -i)
		var err error
		if last, err = WriteAuto(bs, c, "test", at, limits, nil); err != nil {
			t.Fatal(err)
		}
	}
	list, err := ListAuto(bs)
	if err != nil {
		t.Fatal(err)
	}
	var days []string
	for _, b := range list {
		days = append(days, b.CreatedAt.In(now.Location()).Format("01-02"))
	}
	if fmt.Sprint(days) != "[10-13 10-12 10-11]" {
		t.Fatalf("after thinning: %v", days)
	}
	if list[0].Name != last.Name || list[0].Size != last.Size || last.Size == 0 {
		t.Fatalf("newest = %+v, WriteAuto returned %+v", list[0], last)
	}
	if _, err := bs.Read("notes.txt"); err != nil {
		t.Fatalf("other file: %v", err)
	}

	data, err := bs.Read(last.Name)
	if err != nil {
		t.Fatal(err)
	}
	a, err := Read(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Media) != 0 {
		t.Fatalf("automatic backup has %d media files", len(a.Media))
	}
	snap, _ := c.Snapshot()
	if !jsonEqual(t, a.Collection, snap) || len(a.Revlog) != len(revlogOf(t, c)) {
		t.Fatal("automatic backup does not hold the collection and review log")
	}
}

func TestWriteAutoNeverReplaces(t *testing.T) {
	c := populated(t)
	bs := storage.NewMemory().Backups()
	limits := Limits{Daily: 8}
	first, err := WriteAuto(bs, c, "test", now, limits, nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := WriteAuto(bs, c, "test", now, limits, nil)
	if err != nil {
		t.Fatal(err)
	}
	if first.Name == second.Name || !second.CreatedAt.After(first.CreatedAt) {
		t.Fatalf("same-millisecond backups: %+v, %+v", first, second)
	}
	if list, _ := ListAuto(bs); len(list) != 2 || list[0].Name != second.Name {
		t.Fatalf("list = %+v", list)
	}
}
