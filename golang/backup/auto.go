package backup

import (
	"sort"
	"strings"
	"time"

	"cardflip/golang/storage"
)

// Automatic backups are collection-only backups (no media) kept in a
// storage.BackupStore and thinned after every new one.

const (
	autoPrefix = "backup-"
	autoSuffix = ".cardflip"
	// autoLayout is UTC with milliseconds, using only characters that are
	// valid backup names; names sort by time.
	autoLayout = "2006-01-02-150405.000"
)

// AutoName is the store name of an automatic backup made at t.
func AutoName(t time.Time) string {
	ts := strings.Replace(t.UTC().Format(autoLayout), ".", "-", 1)
	return autoPrefix + ts + autoSuffix
}

// ParseAutoName returns when the automatic backup name was made; ok is
// false for names AutoName does not produce.
func ParseAutoName(name string) (t time.Time, ok bool) {
	ts, found := strings.CutPrefix(name, autoPrefix)
	if !found {
		return time.Time{}, false
	}
	if ts, found = strings.CutSuffix(ts, autoSuffix); !found || len(ts) != len(autoLayout) {
		return time.Time{}, false
	}
	// Put the dot before the milliseconds back.
	ts = ts[:len(ts)-4] + "." + ts[len(ts)-3:]
	t, err := time.Parse(autoLayout, ts)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// Auto is a stored automatic backup.
type Auto struct {
	Name      string
	CreatedAt time.Time
	Size      int64
}

// ListAuto returns the automatic backups in store, newest first. Other
// files in the store are left out.
func ListAuto(store storage.BackupStore) ([]Auto, error) {
	files, err := store.List()
	if err != nil {
		return nil, err
	}
	var out []Auto
	for _, f := range files {
		if t, ok := ParseAutoName(f.Name); ok {
			out = append(out, Auto{Name: f.Name, CreatedAt: t, Size: f.Size})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

// Limits says how many automatic backups to keep besides today's.
type Limits struct {
	Daily, Weekly, Monthly int
}

// Obsolete returns the backups thinning removes. All backups of today (in
// now's location) are kept. Going from newest to oldest, a backup is then
// kept as a daily backup if it is from an earlier day than the last kept
// one, else as a weekly one if it is from an earlier ISO week, else as a
// monthly one if it is from an earlier month, while that period's limit
// is not used up. Every other backup is obsolete. A kept backup fills one
// period's slot only, so no backup counts twice.
func Obsolete(backups []Auto, now time.Time, limits Limits) []Auto {
	sorted := append([]Auto(nil), backups...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].CreatedAt.After(sorted[j].CreatedAt) })
	loc := now.Location()
	today := dayOf(now)
	const unset = int(^uint(0) >> 1) // later than any period
	lastDay, lastWeek, lastMonth := unset, unset, unset
	var out []Auto
	for _, b := range sorted {
		t := b.CreatedAt.In(loc)
		day, week, month := dayOf(t), weekOf(t), t.Year()*100+int(t.Month())
		keep := func(limit *int, period, last int) bool {
			if *limit <= 0 || period >= last {
				return false
			}
			*limit--
			lastDay, lastWeek, lastMonth = day, week, month
			return true
		}
		if day >= today ||
			keep(&limits.Daily, day, lastDay) ||
			keep(&limits.Weekly, week, lastWeek) ||
			keep(&limits.Monthly, month, lastMonth) {
			continue
		}
		out = append(out, b)
	}
	return out
}

func dayOf(t time.Time) int { return t.Year()*10000 + int(t.Month())*100 + t.Day() }

func weekOf(t time.Time) int {
	y, w := t.ISOWeek()
	return y*100 + w
}

// WriteAuto makes an automatic backup of src in store, then thins the
// automatic backups. A failed thinning does not fail the backup; it is
// repeated after the next one.
func WriteAuto(store storage.BackupStore, src Source, appVersion string, now time.Time, limits Limits, yield func()) (Auto, error) {
	data, err := ExportCollection(src, appVersion, now, yield)
	if err != nil {
		return Auto{}, err
	}
	existing, err := store.List()
	if err != nil {
		return Auto{}, err
	}
	taken := map[string]bool{}
	for _, f := range existing {
		taken[f.Name] = true
	}
	// Never replace a backup, even one made in the same millisecond.
	at := now
	for taken[AutoName(at)] {
		at = at.Add(time.Millisecond)
	}
	name := AutoName(at)
	if err := store.Write(name, data); err != nil {
		return Auto{}, err
	}
	made, _ := ParseAutoName(name)
	_ = Thin(store, now, limits)
	return Auto{Name: name, CreatedAt: made, Size: int64(len(data))}, nil
}

// Thin deletes the automatic backups Obsolete selects.
func Thin(store storage.BackupStore, now time.Time, limits Limits) error {
	all, err := ListAuto(store)
	if err != nil {
		return err
	}
	for _, b := range Obsolete(all, now, limits) {
		if err := store.Delete(b.Name); err != nil {
			return err
		}
	}
	return nil
}
