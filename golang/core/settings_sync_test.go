package core

import (
	"reflect"
	"slices"
	"testing"
)

// Every Settings field is either synced or local (sync design decision 6):
// a new field must be added to one of the lists.
func TestSettingsClassified(t *testing.T) {
	typ := reflect.TypeFor[Settings]()
	fields := map[string]bool{}
	for i := range typ.NumField() {
		name := typ.Field(i).Name
		fields[name] = true
		synced, local := slices.Contains(syncedSettings, name), slices.Contains(localSettings, name)
		switch {
		case synced && local:
			t.Errorf("Settings.%s is both synced and local", name)
		case !synced && !local:
			t.Errorf("Settings.%s is neither synced nor local: add it to syncedSettings or localSettings", name)
		}
	}
	for _, name := range append(slices.Clone(syncedSettings), localSettings...) {
		if !fields[name] {
			t.Errorf("%s is listed but is not a Settings field", name)
		}
	}
}

func TestWithSyncedSettings(t *testing.T) {
	ahead := 5
	src := DefaultSettings()
	src.RolloverHour, src.LearnAheadMinutes, src.FlagNames[2], src.RestDaysPerWeek = 6, &ahead, "Check", 1
	src.Theme, src.Language, src.UIScale, src.LastDeckID = ThemeDark, "de", 150, 7
	src.SeenAchievements, src.SeenAchievementIDs = []string{"level:5"}, []string{"level"}
	src.HideStudyRewards = true
	dst := DefaultSettings()
	dst.Theme, dst.Language = ThemeLight, "en"
	got := withSyncedSettings(dst, src)
	if got.RolloverHour != 6 || got.LearnAheadMinutes == nil || *got.LearnAheadMinutes != 5 || got.FlagNames[2] != "Check" ||
		got.RestDaysPerWeek != 1 {
		t.Fatalf("synced fields not copied: %+v", got)
	}
	if got.Theme != ThemeLight || got.Language != "en" || got.UIScale != 0 || got.LastDeckID != 0 ||
		got.SeenAchievements != nil || got.SeenAchievementIDs != nil || got.HideStudyRewards {
		t.Fatalf("local fields copied: %+v", got)
	}
	if !syncedSettingsEqual(got, src) || syncedSettingsEqual(dst, src) {
		t.Fatal("syncedSettingsEqual disagrees")
	}
}
