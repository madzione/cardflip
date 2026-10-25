import 'package:flutter/foundation.dart';

import 'package:cardflip/golang/golang.dart' as golang;

import 'deck_list_model.dart';
import 'settings_model.dart';

/// The streak and today's progress toward the daily goal, for the deck list
/// header and the study summary. It reloads with the deck list, which
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

  Future<void> load() async {
    _shown = _settings.showProgress;
    if (!_shown) return;
    try {
      final p = await golang.progress.today();
      if (_disposed) return;
      progress = p;
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
