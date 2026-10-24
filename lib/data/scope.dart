import 'package:flutter/widgets.dart';

import 'package:cardflip/features/import/pick.dart';

import 'backups_model.dart';
import 'changes.dart';
import 'deck_list_model.dart';
import 'profiles_model.dart';
import 'render_environment.dart';
import 'settings_model.dart';
import 'sync_model.dart';
import 'undo_model.dart';

/// The shared state of the opened collection.
class AppState {
  AppState({
    required this.isWeb,
    this.picker = const FilePickerService(),
    this.switchProfile = _cannotSwitch,
    SyncModel? sync,
  }) : sync = sync ?? SyncModel(isWeb: true) {
    changes = ChangeHub();
    settings = SettingsModel(changes);
    deckList = DeckListModel(changes);
    backups = BackupsModel(changes);
    undo = UndoModel(changes);
    renderEnv = RenderEnvironment(isWeb: isWeb);
  }

  final bool isWeb;
  final FilePickerService picker;

  /// Device sync, which outlives the collection: the app keeps one for all
  /// profiles.
  final SyncModel sync;

  /// Closes this collection and opens profile `id`'s, with a new
  /// [AppState]; the app starts over on its deck list.
  final Future<void> Function(String id) switchProfile;

  static Future<void> _cannotSwitch(String id) async =>
      throw UnsupportedError('no profile switching here');

  final profiles = ProfilesModel();
  late final ChangeHub changes;
  late final SettingsModel settings;
  late final DeckListModel deckList;
  late final BackupsModel backups;
  late final UndoModel undo;

  /// Where cards are shown, which Go renders them for.
  late final RenderEnvironment renderEnv;

  /// How many study screens are open; reminders wait while one is.
  final openSessions = ValueNotifier<int>(0);

  /// How many study screens hide the app's navigation ("Hide bars during
  /// review"); the navigation is hidden while one is.
  final barsHidden = ValueNotifier<int>(0);

  Future<void> load() => Future.wait([
    settings.load(),
    deckList.load(),
    backups.load(),
    undo.load(),
    profiles.load(),
  ]);

  void dispose() {
    settings.dispose();
    deckList.dispose();
    backups.dispose();
    undo.dispose();
    profiles.dispose();
    openSessions.dispose();
    barsHidden.dispose();
    changes.dispose();
  }
}

/// Makes [AppState] available to the widget tree.
class AppScope extends InheritedWidget {
  const AppScope({super.key, required this.state, required super.child});

  final AppState state;

  static AppState of(BuildContext context) =>
      context.dependOnInheritedWidgetOfExactType<AppScope>()!.state;

  /// Like [of], without rebuilding the caller when the state changes; can be
  /// used in `initState` and `dispose`.
  static AppState read(BuildContext context) =>
      context.getInheritedWidgetOfExactType<AppScope>()!.state;

  /// Like [read], or null outside an [AppScope] (widget tests).
  static AppState? maybeRead(BuildContext context) =>
      context.getInheritedWidgetOfExactType<AppScope>()?.state;

  @override
  bool updateShouldNotify(AppScope oldWidget) => state != oldWidget.state;
}
