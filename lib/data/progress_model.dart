import 'package:flutter/foundation.dart';

import 'package:cardflip/golang/golang.dart' as golang;

import 'deck_list_model.dart';
import 'settings_model.dart';

/// The streak, today's progress toward the daily goal and the achievement
/// tiers not shown yet, for the deck list header, the study summary and the
/// deck list's new achievement message. It reloads with the deck list, which
/// reloads after every change of the collection (answers, undo, imports,
/// restores, sync) and when a new study day starts. While streaks and goals
/// are hidden it does not load: the first load reads the whole review log.
class ProgressModel extends ChangeNotifier {
  ProgressModel(this._deckList, this._settings) {
    _deckList.addListener(load);
    _settings.addListener(_settingsChanged);
  }

  final DeckListModel _deckList;
  final SettingsModel _settings;
  bool _disposed = false;
  bool _shown = false;

  /// The last progress Go reported; null until the first load.
  golang.Progress? progress;

  /// The tiers [markSeen] marked, which Go may still report as new until
  /// the next load.
  final _marked = <String>{};

  /// The earned achievement tiers this device has not shown yet.
  List<golang.NewAchievement> get newAchievements => [
    for (final a
        in progress?.newAchievements ?? const <golang.NewAchievement>[])
      if (!_marked.contains(a.key)) a,
  ];

  /// Records [tiers] as shown, so [newAchievements] leaves them out from
  /// now on. It does not notify, so a build can call it.
  void markSeen(List<golang.NewAchievement> tiers) {
    final keys = [for (final a in tiers) a.key];
    _marked.addAll(keys);
    golang.progress.markSeen(keys).catchError((Object _) {
      // Shown again on the next start at worst.
    });
  }

  Future<void> load() async {
    _shown = _settings.showProgress;
    if (!_shown) return;
    try {
      final p = await golang.progress.today();
      if (_disposed) return;
      progress = p;
      // Once Go has them as seen they need no mark here.
      final reported = {for (final a in p.newAchievements) a.key};
      _marked.retainWhere(reported.contains);
      notifyListeners();
    } catch (_) {
      // The header and summary lines stay as they were.
    }
  }

  /// Loads when streaks and goals are shown again.
  void _settingsChanged() {
    if (_settings.showProgress && !_shown) load();
    _shown = _settings.showProgress;
  }

  @override
  void dispose() {
    _disposed = true;
    _deckList.removeListener(load);
    _settings.removeListener(_settingsChanged);
    super.dispose();
  }
}
