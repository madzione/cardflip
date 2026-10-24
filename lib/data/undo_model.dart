import 'dart:async';

import 'package:flutter/foundation.dart';

import 'package:cardflip/app/errors.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/go_text.dart';
import 'package:cardflip/l10n/l10n.dart';

import 'changes.dart';

/// The app's undo history: what Undo and Redo act on, kept up to date with
/// every change.
class UndoModel extends ChangeNotifier {
  UndoModel(ChangeHub changes) {
    _sub = changes.collection.listen((_) => load());
  }

  late final StreamSubscription<golang.Change> _sub;

  /// The names of the steps Undo and Redo act on; empty when there is none.
  golang.UndoInfo info = const golang.UndoInfo(undo: '', redo: '');

  Future<void> load() async {
    try {
      info = await golang.app.undoInfo();
      notifyListeners();
    } catch (_) {
      // Shown by the next action that fails.
    }
  }

  /// Reverts the most recent change.
  Future<void> undo() => _run(golang.app.undo, (l, step) => l.undone(step));

  /// Re-applies the most recently undone change.
  Future<void> redo() => _run(golang.app.redo, (l, step) => l.redone(step));

  Future<void> _run(
    Future<String> Function() action,
    String Function(AppLocalizations l, String step) done,
  ) async {
    try {
      final name = await action();
      showInfo(done(appL10n, goText(name)));
    } catch (e) {
      showError(e);
    }
    await load();
  }

  /// Confirms a change the user may regret, with an Undo action.
  void offer(String message) => showUndoable(message, undo);

  @override
  void dispose() {
    _sub.cancel();
    super.dispose();
  }
}
