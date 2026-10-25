import 'package:flutter/material.dart';
import 'package:flutter/scheduler.dart';

import 'package:cardflip/data/scope.dart';
import 'package:cardflip/l10n/l10n.dart';

import 'achievements.dart';
import 'achievements_screen.dart';

/// Wraps the deck list to tell about achievement tiers this device has not
/// shown yet with a short message (progress spec, "New achievement
/// notice"), for tiers earned through sync or import: tiers earned in a
/// session are on its summary. It waits while a study session is open or
/// the deck list is covered or on a hidden tab, and says nothing while
/// streaks and goals are off.
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
    final tiers = _app.progress.newAchievements;
    if (tiers.isEmpty) return;
    _app.progress.markSeen(tiers);
    final l = context.l10n;
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(
        SnackBar(
          key: const Key('new-achievements'),
          content: Text(newAchievementsMessage(l, tiers)),
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

  @override
  Widget build(BuildContext context) => widget.child;
}
