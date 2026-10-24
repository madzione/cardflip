import 'dart:async';

import 'package:flutter/widgets.dart';

import 'package:cardflip/golang/golang.dart' as golang;

import 'changes.dart';

/// Learning cards became due in these top-level decks since the last load.
class LearningDue {
  const LearningDue(this.decks);

  final List<golang.DeckNode> decks;
}

/// The deck tree with due counts. It reloads whenever the collection
/// changes, so counts stay current after studying, editing or importing. It
/// also reloads when counts change with time (a learning card becomes due,
/// a new study day starts) and when the app returns to the foreground.
class DeckListModel extends ChangeNotifier {
  DeckListModel(ChangeHub changes) {
    _sub = changes.collection.listen((_) => load());
    _lifecycle = AppLifecycleListener(onResume: load);
  }

  /// The longest the model waits between time-driven reloads, so clock
  /// changes and late timers (desktop sleep) are caught up.
  static const maxWait = Duration(minutes: 15);

  late final StreamSubscription<golang.Change> _sub;
  late final AppLifecycleListener _lifecycle;
  final _learningDue = StreamController<LearningDue>.broadcast();
  Timer? _timer;
  bool _disposed = false;

  /// When the counts next change on their own, from the last load.
  DateTime? _nextChange;

  List<golang.DeckNode> decks = const [];
  bool loading = true;
  Object? error;

  /// Top-level decks in which learning cards became due as time passed.
  Stream<LearningDue> get learningDue => _learningDue.stream;

  Future<void> load() async {
    final started = DateTime.now();
    try {
      final tree = await golang.decks.tree();
      if (_disposed) return;
      _reportLearningDue(tree, started);
      decks = tree;
      error = null;
    } catch (e) {
      if (_disposed) return;
      error = e;
    }
    loading = false;
    notifyListeners();
    await _schedule();
  }

  /// Reports top-level decks whose learning count went up, when the load
  /// started after the previous next-change moment: answering or importing
  /// before that moment never produces a reminder.
  void _reportLearningDue(List<golang.DeckNode> tree, DateTime started) {
    final due = _nextChange;
    if (due == null || started.isBefore(due)) return;
    final before = {
      for (final d in decks)
        if (d.depth == 0) d.id: d.counts.learning,
    };
    final increased = [
      for (final d in tree)
        if (before[d.id] case final b? when d.counts.learning > b) d,
    ];
    if (increased.isNotEmpty) _learningDue.add(LearningDue(increased));
  }

  /// Sets the one timer for the next time-driven reload.
  Future<void> _schedule() async {
    DateTime? next;
    try {
      next = await golang.decks.nextChange();
    } catch (_) {
      // Retry at the cap.
    }
    if (_disposed) return;
    if (next != null) _nextChange = next;
    var wait = maxWait;
    if (next != null) {
      // Go counts a card as due at Due <= now; the slack absorbs rounding
      // and timer jitter.
      final until = next
          .add(const Duration(seconds: 1))
          .difference(DateTime.now());
      if (until < wait) wait = until;
      if (wait < const Duration(seconds: 1)) wait = const Duration(seconds: 1);
    }
    _timer?.cancel();
    _timer = Timer(wait, load);
  }

  /// The decks to show: children of collapsed parents are hidden.
  List<golang.DeckNode> get visible {
    final out = <golang.DeckNode>[];
    final hidden = <int>{};
    for (final d in decks) {
      if (hidden.contains(d.parentID)) {
        hidden.add(d.id);
        continue;
      }
      out.add(d);
      if (d.collapsed) hidden.add(d.id);
    }
    return out;
  }

  @override
  void dispose() {
    _disposed = true;
    _timer?.cancel();
    _lifecycle.dispose();
    _sub.cancel();
    _learningDue.close();
    super.dispose();
  }
}
