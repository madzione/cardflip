package core

import (
	"encoding/json"
	"testing"
)

func TestBackupSettingsDefaults(t *testing.T) {
	c, _, _ := testCollection(t)
	if got := c.Settings().Backup; got != (BackupSettings{IntervalMinutes: 30, Daily: 8, Weekly: 8, Monthly: 8}) {
		t.Fatalf("new collection: %+v", got)
	}
	// Settings saved before backup settings existed get the defaults; zero
	// values that were saved stay zero.
	var old Settings
	if err := json.Unmarshal([]byte(`{"theme":"dark","rolloverHour":4}`), &old); err != nil {
		t.Fatal(err)
	}
	if old.Theme != "dark" || old.Backup != DefaultBackupSettings() {
		t.Fatalf("old settings: %+v", old)
	}
	var off Settings
	if err := json.Unmarshal([]byte(`{"theme":"dark","backup":{"intervalMinutes":0,"daily":8,"weekly":0,"monthly":2}}`), &off); err != nil {
		t.Fatal(err)
	}
	if off.Backup != (BackupSettings{Daily: 8, Monthly: 2}) {
		t.Fatalf("saved zeros: %+v", off.Backup)
	}
}

func TestBackupSettingsValidation(t *testing.T) {
	c, _, _ := testCollection(t)
	for _, b := range []BackupSettings{{IntervalMinutes: -1}, {IntervalMinutes: MaxBackupInterval + 1},
		{Daily: -1}, {Weekly: MaxBackupsKept + 1}} {
		s := c.Settings()
		s.Backup = b
		if err := c.UpdateSettings(s); !IsUserError(err) {
			t.Errorf("%+v: err = %v", b, err)
		}
	}
	s := c.Settings()
	s.Backup = BackupSettings{IntervalMinutes: 0, Daily: 0, Weekly: 3, Monthly: MaxBackupsKept}
	if err := c.UpdateSettings(s); err != nil {
		t.Fatal(err)
	}
	if c.Settings().Backup != s.Backup {
		t.Fatalf("saved %+v", c.Settings().Backup)
	}
}

func TestRevisionCountsChanges(t *testing.T) {
	c, _, _ := testCollection(t)
	r0 := c.Revision()
	c.Settings() // reads do not count
	s := c.Settings()
	if err := c.UpdateSettings(s); err != nil || c.Revision() != r0 {
		t.Fatalf("unchanged settings: revision %d → %d, %v", r0, c.Revision(), err)
	}
	s.RolloverHour = 6
	if err := c.UpdateSettings(s); err != nil || c.Revision() != r0+1 {
		t.Fatalf("changed settings: revision %d → %d, %v", r0, c.Revision(), err)
	}
	snap, _ := c.Snapshot()
	if err := c.RestoreKeepingMedia(snap, nil, nil); err != nil || c.Revision() != r0+2 {
		t.Fatalf("restore: revision %d → %d, %v", r0, c.Revision(), err)
	}
}

func TestRestoreKeepingMedia(t *testing.T) {
	c, store, clk := testCollection(t)
	morning, _ := c.Snapshot()
	if err := c.Restore(morning, nil, nil, map[string][]byte{"cat.png": []byte("meow")}); err != nil {
		t.Fatal(err)
	}
	s := c.Settings()
	s.RolloverHour = 7
	if err := c.UpdateSettings(s); err != nil {
		t.Fatal(err)
	}
	if err := c.RestoreKeepingMedia(morning, nil, nil); err != nil {
		t.Fatal(err)
	}
	if c.Settings().RolloverHour != 4 {
		t.Fatalf("collection not restored: %+v", c.Settings())
	}
	if data, err := c.Media("cat.png"); err != nil || string(data) != "meow" {
		t.Fatalf("media: %q, %v", data, err)
	}
	c2 := reopen(t, c, store, clk)
	if data, err := c2.Media("cat.png"); err != nil || string(data) != "meow" {
		t.Fatalf("media after reopen: %q, %v", data, err)
	}
}
