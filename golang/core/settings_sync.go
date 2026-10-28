package core

import "reflect"

// syncedSettings names the Settings fields that device sync carries
// (sync design decision 6): those that affect studying and card content.
// They form one register. Every other field belongs to the device and is
// listed in localSettings; a test checks that each field is in one list.
var syncedSettings = []string{
	"RolloverHour", "LearnAheadMinutes", "AutoContinueLearning", "FlagNames",
	// Review display and timer.
	"HidePlayButtons", "KeepAudioOnAnswer", "HideRemainingCounts", "HideIntervals", "SpaceOnlyReveals",
	"HideBarsDuringReview", "QuestionFadeMs", "TimeboxMinutes",
	"IgnoreAccents", "StripPasteFormatting", "PasteImagesAsPng", "DeckByNoteType",
	"DailyGoalKind", "DailyGoal", "RestDaysPerWeek",
}

// localSettings names the Settings fields that stay on the device: the
// look, the language, the last-used pickers, the browser's layout, current
// and recent decks, backups and one-time notices.
var localSettings = []string{
	"Theme", "Language", "UIScale", "Minimalist", "ReduceMotion",
	"LastNoteTypeID", "LastDeckID", "LastDeckByNoteType", "LastDueDate", "CurrentDeckID", "RecentDecks",
	"BrowserColumns", "BrowserMode", "BrowserSplit", "SavedSearches", "DefaultSearch",
	"Backup", "PersistenceNoticeShown", "HideProgress", "HideStudyRewards", "MuteCelebrations",
	"DeferLevelUps", "GoalSuggestionShown",
	"SeenAchievements", "SeenAchievementIDs", "DeckProgressInList", "RecapDismissedWeek",
}

// syncedSettingsEqual reports whether a and b agree on every synced field.
func syncedSettingsEqual(a, b Settings) bool {
	va, vb := reflect.ValueOf(a), reflect.ValueOf(b)
	for _, name := range syncedSettings {
		if !reflect.DeepEqual(va.FieldByName(name).Interface(), vb.FieldByName(name).Interface()) {
			return false
		}
	}
	return true
}

// withSyncedSettings returns dst with the synced fields of src.
func withSyncedSettings(dst, src Settings) Settings {
	vd, vs := reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src)
	for _, name := range syncedSettings {
		vd.FieldByName(name).Set(vs.FieldByName(name))
	}
	return dst
}
