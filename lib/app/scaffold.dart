import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/scheduler.dart';
import 'package:flutter/services.dart' show TextInputClient;

import 'package:cardflip/data/deck_list_model.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/features/browse/browse_screen.dart';
import 'package:cardflip/features/decks/deck_list_screen.dart';
import 'package:cardflip/features/decks/deck_overview_screen.dart';
import 'package:cardflip/features/decks/deck_picker.dart';
import 'package:cardflip/features/editor/math_insert.dart' show MenuEntryLabel;
import 'package:cardflip/features/editor/note_editor_screen.dart';
import 'package:cardflip/features/settings/settings_screen.dart';
import 'package:cardflip/features/study/study_screen.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/media_player.dart';
import 'package:cardflip/l10n/go_text.dart';
import 'package:cardflip/l10n/l10n.dart';

import 'appearance.dart';
import 'errors.dart';
import 'shortcuts.dart';

/// Width from which navigation is a side rail instead of a bottom bar.
const wideLayoutWidth = 840.0;

/// Index of the Settings destination.
const _settingsIndex = 3;

class _Destination {
  const _Destination(this.label, this.icon, this.selectedIcon, this.build);

  final String Function(AppLocalizations l) label;
  final IconData icon;
  final IconData selectedIcon;
  final WidgetBuilder build;
}

final _destinations = [
  _Destination(
    (l) => l.navDecks,
    Icons.style_outlined,
    Icons.style,
    (_) => const DeckListScreen(),
  ),
  _Destination(
    (l) => l.navBrowse,
    Icons.search_outlined,
    Icons.search,
    (_) => const BrowseScreen(),
  ),
  _Destination(
    (l) => l.navAdd,
    Icons.add_circle_outline,
    Icons.add_circle,
    (_) => const NoteEditorScreen(),
  ),
  _Destination(
    (l) => l.navSettings,
    Icons.settings_outlined,
    Icons.settings,
    (_) => const SettingsScreen(),
  ),
];

/// The top-level layout: Decks, Browse, Add and Settings, as a side rail on
/// wide screens and a bottom bar on narrow ones. Each destination keeps its
/// own navigation stack. It also reminds the user when learning cards become
/// due outside a study session.
class AdaptiveScaffold extends StatefulWidget {
  const AdaptiveScaffold({super.key});

  /// Switches to a destination (0 Decks, 1 Browse, 2 Add, 3 Settings).
  static void select(BuildContext context, int index) =>
      context.findAncestorStateOfType<_AdaptiveScaffoldState>()?._select(index);

  /// Asks which deck to study ("Study deck", also `/`) and opens its
  /// overview on the Decks tab.
  static void pickDeck(BuildContext context) =>
      context.findAncestorStateOfType<_AdaptiveScaffoldState>()?._pickDeck();

  @override
  State<AdaptiveScaffold> createState() => _AdaptiveScaffoldState();
}

class _AdaptiveScaffoldState extends State<AdaptiveScaffold> {
  int _index = 0;
  final _navigators = List.generate(
    _destinations.length,
    (_) => GlobalKey<NavigatorState>(),
  );
  // Each destination's focus, kept while it is hidden (see [_go]).
  final _focus = List.generate(
    _destinations.length,
    (i) => FocusScopeNode(debugLabel: _destinations[i].label(english)),
  );
  late final StreamSubscription<LearningDue> _reminders;
  late final ValueNotifier<int> _barsHidden;

  @override
  void initState() {
    super.initState();
    final app = AppScope.read(context);
    _reminders = app.deckList.learningDue.listen(_remind);
    _barsHidden = app.barsHidden..addListener(_barsChanged);
  }

  @override
  void dispose() {
    _reminders.cancel();
    _barsHidden.removeListener(_barsChanged);
    for (final f in _focus) {
      f.dispose();
    }
    super.dispose();
  }

  /// A study screen hides or shows the navigation. A screen that is
  /// disposed reports it while the tree is locked, so the rebuild waits
  /// for the end of the frame then.
  void _barsChanged() {
    if (SchedulerBinding.instance.schedulerPhase ==
        SchedulerPhase.persistentCallbacks) {
      SchedulerBinding.instance.addPostFrameCallback((_) {
        if (mounted) setState(() {});
      });
      return;
    }
    setState(() {});
  }

  /// Shows a reminder for newly due learning cards, unless a study session
  /// is open (it handles them itself).
  void _remind(LearningDue due) {
    if (AppScope.read(context).openSessions.value > 0) return;
    final deck = due.decks.length == 1 ? due.decks.single : null;
    messengerKey.currentState
      ?..hideCurrentSnackBar()
      ..showSnackBar(
        SnackBar(
          content: Text(
            deck != null
                ? context.l10n.learningDueInDeck(deck.fullName)
                : context.l10n.learningDueInDecks(due.decks.length),
          ),
          behavior: SnackBarBehavior.floating,
          duration: const Duration(seconds: 10),
          action: SnackBarAction(
            label: deck != null ? context.l10n.study : context.l10n.showDecks,
            onPressed: () => deck != null ? _study(deck) : _showDecks(),
          ),
        ),
      );
  }

  /// Shows the deck list on the Decks tab.
  void _showDecks() {
    _go(0);
    _navigators[0].currentState?.popUntil((r) => r.isFirst);
  }

  /// Opens a deck's study session on the Decks tab.
  void _study(golang.DeckNode deck) {
    _showDecks();
    _navigators[0].currentState?.push(
      MaterialPageRoute<void>(
        builder: (_) => StudyScreen(deckId: deck.id, deckName: deck.fullName),
      ),
    );
  }

  /// Opens the overview of a deck chosen in the deck picker on the Decks
  /// tab, leaving whatever it showed (a study session too).
  Future<void> _pickDeck() async {
    final deck = await showDeckPicker(context);
    if (deck == null || !mounted) return;
    _showDecks();
    _navigators[0].currentState?.push(
      MaterialPageRoute<void>(builder: (_) => DeckOverviewScreen(deck: deck)),
    );
  }

  void _select(int index) {
    if (index == _index) {
      // Tapping the current destination goes back to its first screen.
      _navigators[index].currentState?.popUntil((r) => r.isFirst);
      return;
    }
    // Periodic backups may have been made since Settings was last shown.
    if (index == _settingsIndex) AppScope.read(context).backups.load();
    _go(index);
  }

  /// Shows another destination. The one left stays mounted, so whatever it
  /// plays is stopped: the user can no longer see it.
  void _go(int index) {
    if (index == _index) return;
    MediaPlayer.stopAll();
    setState(() => _index = index);
    // A hidden destination cannot hold the focus, so it went elsewhere.
    // Give it back to what had it, e.g. the study screen, whose keys (R to
    // replay) would not work otherwise.
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted && _index == index) _focus[index].requestFocus();
    });
  }

  @override
  Widget build(BuildContext context) {
    final body = IndexedStack(
      index: _index,
      children: [
        for (var i = 0; i < _destinations.length; i++)
          HeroControllerScope.none(
            child: FocusScope(
              node: _focus[i],
              child: Navigator(
                key: _navigators[i],
                onGenerateRoute: (settings) => MaterialPageRoute(
                  settings: settings,
                  builder: _destinations[i].build,
                ),
              ),
            ),
          ),
      ],
    );
    final undo = AppScope.read(context).undo;
    // "Hide bars during review": the study screen on show hides them.
    final navigation = _barsHidden.value == 0;
    return UndoShortcuts(
      onUndo: undo.undo,
      onRedo: undo.redo,
      onPickDeck: _pickDeck,
      child: NavigatorPopHandler(
        onPopWithResult: (_) => _navigators[_index].currentState?.maybePop(),
        child: LayoutBuilder(
          builder: (context, constraints) {
            if (constraints.maxWidth >= wideLayoutWidth) {
              return Scaffold(
                body: Row(
                  children: [
                    if (navigation) ...[
                      NavigationRail(
                        selectedIndex: _index,
                        onDestinationSelected: _select,
                        labelType: NavigationRailLabelType.all,
                        destinations: [
                          for (final d in _destinations)
                            NavigationRailDestination(
                              icon: Icon(d.icon),
                              selectedIcon: Icon(d.selectedIcon),
                              label: Text(d.label(context.l10n)),
                            ),
                        ],
                      ),
                      const VerticalDivider(width: 1),
                    ],
                    Expanded(child: body),
                  ],
                ),
              );
            }
            return Scaffold(
              body: body,
              bottomNavigationBar: !navigation
                  ? null
                  : NavigationBar(
                      selectedIndex: _index,
                      onDestinationSelected: _select,
                      destinations: [
                        for (final d in _destinations)
                          NavigationDestination(
                            icon: Icon(d.icon),
                            selectedIcon: Icon(d.selectedIcon),
                            label: d.label(context.l10n),
                          ),
                      ],
                    ),
            );
          },
        ),
      ),
    );
  }
}

class _UndoIntent extends Intent {
  const _UndoIntent();
}

class _RedoIntent extends Intent {
  const _RedoIntent();
}

class _PickDeckIntent extends Intent {
  const _PickDeckIntent();
}

/// Whether the focus is in a text field, plain or rich (a note field): an
/// ancestor of the focused widget takes text input.
bool get textInputHasFocus {
  final focused = FocusManager.instance.primaryFocus?.context;
  if (focused == null) return false;
  var found = false;
  focused.visitAncestorElements((e) {
    found = e is StatefulElement && e.state is TextInputClient;
    return !found;
  });
  return found;
}

/// Runs a callback, unless a text field has focus: the key then goes on to
/// the text field, which undoes its own typing or types the character.
class _OutsideTextAction<T extends Intent> extends Action<T> {
  _OutsideTextAction(this.callback);

  final VoidCallback callback;

  @override
  bool isEnabled(T intent) => !textInputHasFocus;

  @override
  Object? invoke(T intent) {
    callback();
    return null;
  }
}

/// The app-wide undo and redo shortcuts, and `/` for the deck picker,
/// none of which apply while a text field has focus. Screens with their own
/// undo (the study screen) bind the same keys closer to the focus, which
/// wins.
class UndoShortcuts extends StatelessWidget {
  const UndoShortcuts({
    super.key,
    required this.onUndo,
    required this.onRedo,
    required this.onPickDeck,
    required this.child,
  });

  final VoidCallback onUndo;
  final VoidCallback onRedo;
  final VoidCallback onPickDeck;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return Shortcuts(
      shortcuts: {
        for (final a in undoShortcuts) a: const _UndoIntent(),
        for (final a in redoShortcuts) a: const _RedoIntent(),
        pickDeckShortcut: const _PickDeckIntent(),
      },
      child: Actions(
        actions: {
          _UndoIntent: _OutsideTextAction<_UndoIntent>(onUndo),
          _RedoIntent: _OutsideTextAction<_RedoIntent>(onRedo),
          _PickDeckIntent: _OutsideTextAction<_PickDeckIntent>(onPickDeck),
        },
        child: child,
      ),
    );
  }
}

enum _UndoAction { undo, redo }

/// The app bar menu with Undo and Redo, naming the step each acts on,
/// below the screen's own [actions].
class UndoMenuButton extends StatelessWidget {
  const UndoMenuButton({super.key, this.actions = const []});

  /// Menu entries above Undo and Redo: a label and what choosing it does.
  final List<(String, VoidCallback)> actions;

  @override
  Widget build(BuildContext context) {
    final undo = AppScope.read(context).undo;
    final l = context.l10n;
    PopupMenuItem<Object> item(
      _UndoAction value,
      String verb,
      String Function(String step) withStep,
      String step,
      String shortcut,
    ) => PopupMenuItem(
      value: value,
      enabled: step.isNotEmpty,
      child: MenuEntryLabel(
        step.isEmpty ? verb : withStep(goText(step)),
        shortcut,
      ),
    );
    return PopupMenuButton<Object>(
      popUpAnimationStyle: popupAnimation(context),
      tooltip: l.more,
      onSelected: (a) => switch (a) {
        _UndoAction.undo => undo.undo(),
        _UndoAction.redo => undo.redo(),
        _ => actions[a as int].$2(),
      },
      itemBuilder: (_) => [
        for (final (i, (label, _)) in actions.indexed)
          PopupMenuItem(value: i, child: Text(label)),
        if (actions.isNotEmpty) const PopupMenuDivider(),
        item(
          _UndoAction.undo,
          l.undo,
          l.undoStep,
          undo.info.undo,
          undoShortcutLabel,
        ),
        item(
          _UndoAction.redo,
          l.redo,
          l.redoStep,
          undo.info.redo,
          redoShortcutLabel,
        ),
      ],
    );
  }
}
