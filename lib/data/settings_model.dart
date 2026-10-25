import 'dart:async';

import 'package:flutter/material.dart';

import 'package:cardflip/app/errors.dart';
import 'package:cardflip/app/theme.dart';
import 'package:cardflip/golang/golang.dart' as golang;

import 'changes.dart';

/// The app settings, stored in the Go collection (so backups include them).
/// The most rest days per week a streak can allow.
const maxRestDaysPerWeek = 2;

class SettingsModel extends ChangeNotifier {
  SettingsModel(ChangeHub changes) {
    _sub = changes.stream.listen((c) {
      if (c.settings || c.all) load();
    });
  }

  late final StreamSubscription<golang.Change> _sub;

  ThemeMode themeMode = ThemeMode.system;

  /// The interface language as a language code, such as "de"; empty
  /// follows the device's languages.
  String language = '';
  int rolloverHour = 4;

  /// "UI size" in percent, 75 to 200 in steps of 5: scales text, icons and
  /// spacing.
  int uiScale = 100;

  /// "Minimalist mode": no shadows, elevation tints or decorative fills.
  bool minimalist = false;

  /// "Reduce motion": "system" follows the operating system, "on" or "off".
  String reduceMotion = 'system';

  /// "Show streaks and goals" on the deck list and the study summary; it
  /// belongs to the device.
  bool showProgress = true;

  /// The daily goal: "cards" or "minutes" with [dailyGoal] of them, or ""
  /// for none.
  String dailyGoalKind = '';
  int dailyGoal = 0;

  /// How many days without study, 0 to [maxRestDaysPerWeek], a streak
  /// survives in any 7 days in a row; it belongs to the collection.
  int restDaysPerWeek = 0;

  /// Whether the user opened or dismissed the one-time suggestion to set a
  /// daily goal.
  bool goalSuggestionShown = false;

  /// Whether a waiting study session shows learning cards as soon as they
  /// are due, instead of asking first.
  bool autoContinueLearning = false;

  /// "Learn ahead limit" in minutes: when nothing else is due, a session
  /// shows learning cards due within this time; 0 waits until they are due.
  int learnAheadMinutes = 20;

  /// "Timebox time limit" in minutes: each time it passes in a study
  /// session, the study screen reports the cards studied; 0 turns it off.
  int timeboxMinutes = 0;

  /// "Show remaining card count" on the study screen.
  bool showRemainingCounts = true;

  /// "Show next review time above answer buttons".
  bool showIntervals = true;

  /// "Space/Enter also answers Good": when off, they only reveal the answer.
  bool spaceAnswersGood = true;

  /// "Hide bars during review": the app bar and navigation are hidden while
  /// studying.
  bool hideBarsDuringReview = false;

  /// "Question fade" in milliseconds; 0 shows each question at once.
  int questionFadeMs = 100;

  /// "Show play buttons on cards with audio".
  bool showPlayButtons = true;

  /// "Interrupt current audio when answering".
  bool interruptAudio = true;

  /// The custom names of flags 1 to 7; an empty name means the colour name.
  List<String> flagNames = List.filled(flagCount, '');

  /// The name of flag 1 to 7: its custom name, else its colour name.
  String flagName(int flag) {
    final custom = flag - 1 < flagNames.length ? flagNames[flag - 1] : '';
    return custom.isEmpty ? flagColorName(flag) : custom;
  }

  /// "Strip formatting when pasting": a normal paste inserts plain text.
  bool stripPasteFormatting = false;

  /// "Paste images as PNG": pasted images are stored as PNG, not JPEG.
  bool pasteImagesAsPng = false;

  /// "Deck for new cards: Change deck depending on note type": the add
  /// screen picks the deck last used with the chosen note type. Off is
  /// "Current deck", the last used deck.
  bool deckByNoteType = false;

  /// "Ignore accents in search": text and field searches ignore accents.
  bool ignoreAccents = false;

  /// The browser's columns, in order (see `golang.browse.rows`).
  List<String> browserColumns = defaultBrowserColumns;

  /// "notes" when the browser lists one row per note, else "cards".
  String browserMode = 'cards';

  /// The share of the width the browser's list takes next to the details
  /// on wide screens.
  double browserSplit = 0.6;

  /// "Default browser search": the search the browser starts with; empty
  /// lists all cards.
  String defaultSearch = '';

  /// The last input of Set due date, which its dialog starts with. Go
  /// remembers it when a due date is set.
  String lastDueDate = '';

  /// The IDs of the decks most recently studied, latest first, which the
  /// deck picker lists first. Go records them when a session starts.
  List<int> recentDecks = const [];

  /// The browser's columns until the user changes them.
  static const defaultBrowserColumns = ['question', 'deck', 'due', 'noteType'];

  /// Backup interval and how many backups to keep.
  golang.BackupSettings backup = const golang.BackupSettings(
    intervalMinutes: 30,
    daily: 8,
    weekly: 8,
    monthly: 8,
  );
  bool loaded = false;

  Future<void> load() async {
    try {
      final s = await golang.app.settings();
      themeMode = themeModeOf(s.theme);
      language = s.language;
      rolloverHour = s.rolloverHour;
      uiScale = s.uiScale;
      minimalist = s.minimalist;
      reduceMotion = s.reduceMotion;
      showProgress = s.showProgress;
      dailyGoalKind = s.dailyGoalKind;
      dailyGoal = s.dailyGoal;
      restDaysPerWeek = s.restDaysPerWeek;
      goalSuggestionShown = s.goalSuggestionShown;
      autoContinueLearning = s.autoContinueLearning;
      learnAheadMinutes = s.learnAheadMinutes;
      timeboxMinutes = s.timeboxMinutes;
      showRemainingCounts = s.showRemainingCounts;
      showIntervals = s.showIntervals;
      spaceAnswersGood = s.spaceAnswersGood;
      hideBarsDuringReview = s.hideBarsDuringReview;
      questionFadeMs = s.questionFadeMs;
      showPlayButtons = s.showPlayButtons;
      interruptAudio = s.interruptAudio;
      flagNames = s.flagNames;
      stripPasteFormatting = s.stripPasteFormatting;
      pasteImagesAsPng = s.pasteImagesAsPng;
      deckByNoteType = s.deckByNoteType;
      ignoreAccents = s.ignoreAccents;
      if (s.browserColumns.isNotEmpty) browserColumns = s.browserColumns;
      browserMode = s.browserMode;
      browserSplit = s.browserSplit;
      defaultSearch = s.defaultSearch;
      lastDueDate = s.lastDueDate;
      recentDecks = s.recentDecks;
      backup = s.backup;
      loaded = true;
      notifyListeners();
    } catch (e) {
      showError(e);
    }
  }

  Future<void> setThemeMode(ThemeMode mode) async {
    final previous = themeMode;
    themeMode = mode;
    notifyListeners();
    final ok = await guard(() => golang.app.setTheme(themeSettingOf(mode)));
    if (!ok) {
      themeMode = previous;
      notifyListeners();
    }
  }

  Future<void> setLanguage(String code) async {
    final previous = language;
    language = code;
    notifyListeners();
    if (!await guard(() => golang.app.setLanguage(code))) {
      language = previous;
      notifyListeners();
    }
  }

  Future<void> setUiScale(int percent) async {
    final previous = uiScale;
    uiScale = percent;
    notifyListeners();
    if (!await guard(() => golang.app.setUIScale(percent))) {
      uiScale = previous;
      notifyListeners();
    }
  }

  Future<void> setMinimalist(bool on) async {
    minimalist = on;
    notifyListeners();
    if (!await guard(() => golang.app.setMinimalist(on))) {
      minimalist = !on;
      notifyListeners();
    }
  }

  Future<void> setReduceMotion(String mode) async {
    final previous = reduceMotion;
    reduceMotion = mode;
    notifyListeners();
    if (!await guard(() => golang.app.setReduceMotion(mode))) {
      reduceMotion = previous;
      notifyListeners();
    }
  }

  Future<void> setShowProgress(bool on) async {
    showProgress = on;
    notifyListeners();
    if (!await guard(() => golang.app.setShowProgress(on))) {
      showProgress = !on;
      notifyListeners();
    }
  }

  /// Sets the daily goal: [kind] "cards" or "minutes" with [value] from 1
  /// to 9,999, or "" for none. Reports whether it was saved.
  Future<bool> setDailyGoal(String kind, int value) async {
    final ok = await guard(() => golang.app.setDailyGoal(kind, value));
    if (ok) {
      dailyGoalKind = kind;
      dailyGoal = kind.isEmpty ? 0 : value;
      notifyListeners();
    }
    return ok;
  }

  Future<void> setRestDaysPerWeek(int n) async {
    if (await guard(() => golang.app.setRestDaysPerWeek(n))) {
      restDaysPerWeek = n;
      notifyListeners();
    }
  }

  /// Records that the user opened or dismissed the goal suggestion.
  Future<void> markGoalSuggestionShown() async {
    goalSuggestionShown = true;
    notifyListeners();
    await guard(golang.app.goalSuggestionShown);
  }

  Future<bool> setRolloverHour(int hour) async {
    final ok = await guard(() => golang.app.setRolloverHour(hour));
    if (ok) {
      rolloverHour = hour;
      notifyListeners();
    }
    return ok;
  }

  Future<void> setAutoContinueLearning(bool on) async {
    autoContinueLearning = on;
    notifyListeners();
    final ok = await guard(() => golang.app.setAutoContinueLearning(on));
    if (!ok) {
      autoContinueLearning = !on;
      notifyListeners();
    }
  }

  Future<void> setLearnAheadMinutes(int minutes) async {
    final previous = learnAheadMinutes;
    learnAheadMinutes = minutes;
    notifyListeners();
    final ok = await guard(() => golang.app.setLearnAheadMinutes(minutes));
    if (!ok) {
      learnAheadMinutes = previous;
      notifyListeners();
    }
  }

  Future<void> setTimeboxMinutes(int minutes) async {
    final previous = timeboxMinutes;
    timeboxMinutes = minutes;
    notifyListeners();
    final ok = await guard(() => golang.app.setTimeboxMinutes(minutes));
    if (!ok) {
      timeboxMinutes = previous;
      notifyListeners();
    }
  }

  Future<void> setShowRemainingCounts(bool on) async {
    showRemainingCounts = on;
    notifyListeners();
    final ok = await guard(() => golang.app.setShowRemainingCounts(on));
    if (!ok) {
      showRemainingCounts = !on;
      notifyListeners();
    }
  }

  Future<void> setShowIntervals(bool on) async {
    showIntervals = on;
    notifyListeners();
    final ok = await guard(() => golang.app.setShowIntervals(on));
    if (!ok) {
      showIntervals = !on;
      notifyListeners();
    }
  }

  Future<void> setSpaceAnswersGood(bool on) async {
    spaceAnswersGood = on;
    notifyListeners();
    final ok = await guard(() => golang.app.setSpaceAnswersGood(on));
    if (!ok) {
      spaceAnswersGood = !on;
      notifyListeners();
    }
  }

  Future<void> setHideBarsDuringReview(bool on) async {
    hideBarsDuringReview = on;
    notifyListeners();
    final ok = await guard(() => golang.app.setHideBarsDuringReview(on));
    if (!ok) {
      hideBarsDuringReview = !on;
      notifyListeners();
    }
  }

  Future<void> setQuestionFadeMs(int ms) async {
    final previous = questionFadeMs;
    questionFadeMs = ms;
    notifyListeners();
    final ok = await guard(() => golang.app.setQuestionFadeMs(ms));
    if (!ok) {
      questionFadeMs = previous;
      notifyListeners();
    }
  }

  Future<void> setShowPlayButtons(bool on) async {
    showPlayButtons = on;
    notifyListeners();
    final ok = await guard(() => golang.app.setShowPlayButtons(on));
    if (!ok) {
      showPlayButtons = !on;
      notifyListeners();
    }
  }

  Future<void> setInterruptAudio(bool on) async {
    interruptAudio = on;
    notifyListeners();
    final ok = await guard(() => golang.app.setInterruptAudio(on));
    if (!ok) {
      interruptAudio = !on;
      notifyListeners();
    }
  }

  Future<void> setStripPasteFormatting(bool on) async {
    stripPasteFormatting = on;
    notifyListeners();
    final ok = await guard(() => golang.app.setStripPasteFormatting(on));
    if (!ok) {
      stripPasteFormatting = !on;
      notifyListeners();
    }
  }

  Future<void> setPasteImagesAsPng(bool on) async {
    pasteImagesAsPng = on;
    notifyListeners();
    final ok = await guard(() => golang.app.setPasteImagesAsPng(on));
    if (!ok) {
      pasteImagesAsPng = !on;
      notifyListeners();
    }
  }

  Future<void> setIgnoreAccents(bool on) async {
    ignoreAccents = on;
    notifyListeners();
    final ok = await guard(() => golang.app.setIgnoreAccents(on));
    if (!ok) {
      ignoreAccents = !on;
      notifyListeners();
    }
  }

  Future<void> setBrowserColumns(List<String> columns) async {
    final previous = browserColumns;
    browserColumns = List.unmodifiable(columns);
    notifyListeners();
    final ok = await guard(() => golang.app.setBrowserColumns(columns));
    if (!ok) {
      browserColumns = previous;
      notifyListeners();
    }
  }

  /// Lists cards ("cards") or notes ("notes") in the browser.
  Future<void> setBrowserMode(String mode) async {
    final previous = browserMode;
    browserMode = mode;
    notifyListeners();
    final ok = await guard(() => golang.app.setBrowserMode(mode));
    if (!ok) {
      browserMode = previous;
      notifyListeners();
    }
  }

  /// Sets the share of the width the browser's list takes next to the
  /// details. Only Go is told: the browser shows the new split already.
  Future<void> setBrowserSplit(double split) async {
    browserSplit = split;
    await guard(() => golang.app.setBrowserSplit(split));
  }

  /// Sets the search the browser starts with. An invalid search throws
  /// Go's error, naming the term, and changes nothing.
  Future<void> setDefaultSearch(String query) async {
    await golang.app.setDefaultSearch(query);
    defaultSearch = query.trim();
    notifyListeners();
  }

  Future<void> setDeckByNoteType(bool on) async {
    deckByNoteType = on;
    notifyListeners();
    final ok = await guard(() => golang.app.setDeckByNoteType(on));
    if (!ok) {
      deckByNoteType = !on;
      notifyListeners();
    }
  }

  /// Renames flag 1 to 7; an empty name restores its colour name.
  Future<void> setFlagName(int flag, String name) async {
    final ok = await guard(() => golang.app.setFlagName(flag, name));
    if (ok) {
      flagNames = [
        for (var i = 0; i < flagCount; i++)
          i == flag - 1
              ? name.trim()
              : (i < flagNames.length ? flagNames[i] : ''),
      ];
      notifyListeners();
    }
  }

  Future<void> setBackupSettings(golang.BackupSettings value) async {
    final previous = backup;
    backup = value;
    notifyListeners();
    final ok = await guard(() => golang.app.setBackupSettings(value));
    if (!ok) {
      backup = previous;
      notifyListeners();
    }
  }

  @override
  void dispose() {
    _sub.cancel();
    super.dispose();
  }
}
