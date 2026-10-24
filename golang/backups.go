package main

import (
	"cardflip/golang/internal/usermsg"
	"errors"
	"sync"
	"time"

	"cardflip/golang/backup"
	"cardflip/golang/core"
	"cardflip/golang/internal/yield"
	"cardflip/golang/storage"
)

// autoBackups tracks automatic backups: collection-only backups made
// periodically while the collection changes, before risky operations and
// on request, kept in the backend's backup store and thinned.
type autoBackups struct {
	// tick is how often the periodic backup is checked; 0 disables the
	// timer (tests call periodicBackup themselves).
	tick  time.Duration
	timer sync.Once

	// mu serialises making backups. Lock order: app.mu before mu, and app.mu
	// is never taken while mu is held.
	mu sync.Mutex
	// col and rev are the collection and revision of the last backup, or of
	// the opened collection when none was made since.
	col *core.Collection
	rev uint64
	// last is when the newest backup was made.
	last time.Time
}

// opened starts tracking a newly opened collection: what was stored is
// the baseline, so only later changes call for a periodic backup.
func (ab *autoBackups) opened(c *core.Collection, b storage.Backend) {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	ab.col, ab.rev, ab.last = c, c.Revision(), time.Time{}
	if list, err := backup.ListAuto(b.Backups()); err == nil && len(list) > 0 {
		ab.last = list[0].CreatedAt
	}
}

func limitsOf(s core.BackupSettings) backup.Limits {
	return backup.Limits{Daily: s.Daily, Weekly: s.Weekly, Monthly: s.Monthly}
}

// backupTarget returns the open collection and its backend, or nil when
// the collection could not be read.
func (a *app) backupTarget() (*core.Collection, storage.Backend) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.col, a.backend
}

// writeBackup makes an automatic backup now. Callers hold backups.mu.
func (a *app) writeBackup(c *core.Collection, b storage.Backend) (backup.Auto, error) {
	rev := c.Revision() // before the export: a change during it is backed up next time
	now := a.now()
	made, err := backup.WriteAuto(b.Backups(), c, appVersion, now, limitsOf(c.Settings().Backup), yield.Yield)
	if err != nil {
		return backup.Auto{}, err
	}
	a.backups.col, a.backups.rev, a.backups.last = c, rev, now
	return made, nil
}

// periodicBackup makes a backup when the collection changed since the last
// one and the backup interval has passed since then. It reports whether it
// made one.
func (a *app) periodicBackup() (bool, error) {
	c, b := a.backupTarget()
	if c == nil {
		return false, nil
	}
	interval := time.Duration(c.Settings().Backup.IntervalMinutes) * time.Minute
	if interval == 0 {
		return false, nil
	}
	ab := &a.backups
	ab.mu.Lock()
	defer ab.mu.Unlock()
	if ab.col == c && ab.rev == c.Revision() {
		return false, nil // nothing changed
	}
	if a.now().Sub(ab.last) < interval {
		return false, nil
	}
	if _, err := a.writeBackup(c, b); err != nil {
		return false, err
	}
	return true, nil
}

// startBackupTimer checks for a periodic backup every tick while the app
// runs. A failed backup is retried at the next tick.
func (a *app) startBackupTimer() {
	if a.backups.tick <= 0 {
		return
	}
	a.backups.timer.Do(func() {
		go func() {
			t := time.NewTicker(a.backups.tick)
			defer t.Stop()
			for range t.C {
				_, _ = a.periodicBackup()
			}
		}()
	})
}

// backupBeforeOperation makes a backup before an operation that replaces or
// changes much of the collection (restoring a backup, importing a
// collection package, checking the database), whatever the backup interval.
// Such operations call it first and do not start when it fails. Without a
// readable collection there is nothing to back up.
func (a *app) backupBeforeOperation() error {
	c, b := a.backupTarget()
	if c == nil {
		return nil
	}
	a.backups.mu.Lock()
	defer a.backups.mu.Unlock()
	if _, err := a.writeBackup(c, b); err != nil {
		return errors.New(usermsg.Text("backupFirstFailed", "the collection could not be backed up first, so nothing was changed: %v", err))
	}
	return nil
}
