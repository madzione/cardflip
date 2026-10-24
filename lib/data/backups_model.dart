import 'dart:async';

import 'package:flutter/foundation.dart';

import 'package:cardflip/app/errors.dart';
import 'package:cardflip/golang/golang.dart' as golang;

import 'changes.dart';

/// The automatic backups Go keeps, newest first, and the space they use.
/// Shared by Settings and the app-wide "create backup now" shortcut.
class BackupsModel extends ChangeNotifier {
  BackupsModel(ChangeHub changes) {
    // A restore makes a backup first; other changes only lead to periodic
    // backups later, which [load] picks up when Settings is shown.
    _sub = changes.stream.listen((c) {
      if (c.all) load();
    });
  }

  late final StreamSubscription<golang.Change> _sub;

  List<golang.AutoBackup> backups = const [];

  /// Storage space the backups use, in bytes.
  int bytes = 0;

  /// Whether a backup is being made or restored.
  bool busy = false;

  Future<void> load() async {
    try {
      final list = await golang.app.backups();
      backups = list.backups;
      bytes = list.bytes;
      notifyListeners();
    } catch (e) {
      showError(e);
    }
  }

  Future<void> _run(Future<void> Function() action, String done) async {
    if (busy) return;
    busy = true;
    notifyListeners();
    final ok = await guard(action);
    busy = false;
    await load();
    if (ok) showInfo(done);
  }

  /// Makes a backup now.
  Future<void> create() =>
      _run(() => golang.app.createBackup(), 'Backup created');

  /// Replaces the collection with the backup [name]; media stay as they
  /// are. Ask the user first.
  Future<void> restore(String name) =>
      _run(() => golang.app.restoreBackup(name), 'Backup restored');

  @override
  void dispose() {
    _sub.cancel();
    super.dispose();
  }
}
