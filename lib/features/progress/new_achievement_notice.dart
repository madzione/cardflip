import 'package:flutter/material.dart';
import 'package:flutter/scheduler.dart';

import 'package:cardflip/data/scope.dart';
import 'package:cardflip/l10n/l10n.dart';

import 'achievements.dart';
import 'achievements_screen.dart';
import 'session_rewards.dart';

/// Wraps the deck list to tell about achievement tiers this device has not
/// shown yet with a short message (progress spec, "New achievement
/// notice"), for tiers earned through sync or import, and about what a
/// session left before its summary earned and no reward banner showed:
/// the level and goal reached and its tiers (progress spec, "Rewards after
/// leaving a session"). It waits while a study session is open or the deck
/// list is covered or on a hidden tab, and says nothing while streaks and
/// goals are off.
class NewAchievementNotice extends StatefulWidget {
  const NewAchievementNotice({super.key, required this.child});

  final Widget child;

  @override
  State<NewAchievementNotice> createState() => _NewAchievementNoticeState();
}

class _NewAchievementNoticeState extends State<NewAchievementNotice> {
  late final _app = AppScope.read(context);

  /// Whether the deck list is the visible screen of the visible tab.
  bool _visible = false;
  bool _scheduled = false;

  @override
  void initState() {
    super.initState();
    _app.progress.addListener(_schedule);
    _app.openSessions.addListener(_schedule);
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _visible =
        (ModalRoute.of(context)?.isCurrent ?? true) &&
        TickerMode.valuesOf(context).enabled;
    _schedule();
  }

  @override
  void dispose() {
    _app.progress.removeListener(_schedule);
    _app.openSessions.removeListener(_schedule);
    super.dispose();
  }

  /// Checks after the frame: the changes come during builds and while a
  /// study screen is disposed.
  void _schedule() {
    if (_scheduled) return;
    _scheduled = true;
    SchedulerBinding.instance
      ..addPostFrameCallback((_) {
        _scheduled = false;
        _check();
      })
      ..ensureVisualUpdate();
  }

  void _check() {
    if (!mounted ||
        !_visible ||
        _app.openSessions.value > 0 ||
        !_app.settings.showProgress) {
      return;
    }
    final progress = _app.progress;
    final p = progress.progress;
    if (p == null) return;
    // The session's tiers were marked shown when it found them. What is
    // told counts as shown: the session's last answer may load later.
    final pending = progress.pendingSession;
    final session = pending?.unshown(p) ?? const RewardEvents();
    final unseen = progress.newAchievements;
    final known = session.keys;
    final events = RewardEvents(
      goal: session.goal,
      level: session.level,
      tiers: [
        ...session.tiers,
        for (final a in unseen)
          if (!known.contains(a.key)) a,
      ],
    );
    if (events.isEmpty) return;
    pending?.shown.addAll(session.keys);
    if (unseen.isNotEmpty) progress.markSeen(unseen);
    final l = context.l10n;
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(
        SnackBar(
          key: const Key('new-achievements'),
          content: Text(_message(l, events)),
          behavior: SnackBarBehavior.floating,
          // A short message: it leaves by itself despite its action.
          persist: false,
          duration: const Duration(seconds: 8),
          action: SnackBarAction(
            label: l.showAchievements,
            onPressed: () => showAchievements(context),
          ),
        ),
      );
  }

  /// "Reached level 13 · Daily goal reached · New: 1,000 answers"; tiers
  /// alone as before, counted when more than three. A Level tier is left
  /// out when the level is named.
  String _message(AppLocalizations l, RewardEvents events) {
    final named = events.named;
    final parts = [
      if (events.level case final level?) l.levelReached(level),
      if (events.goal) l.goalReached,
    ];
    if (parts.isEmpty) return newAchievementsMessage(l, named);
    return [
      ...parts,
      if (named.isNotEmpty) newAchievementsLine(l, named),
    ].join(' · ');
  }

  @override
  Widget build(BuildContext context) => widget.child;
}
