import 'package:flutter/material.dart';

import 'package:cardflip/app/errors.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/features/settings/daily_goal_dialog.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';

import 'achievement_badge.dart';
import 'achievements.dart';

/// Opens the achievements screen.
Future<void> showAchievements(BuildContext context) => Navigator.of(context)
    .push(MaterialPageRoute<void>(builder: (_) => const AchievementsScreen()));

/// Every achievement as a badge (progress spec, "Achievements screen"):
/// its highest earned tier with the date, and the progress toward the next
/// tier; achievements without an earned tier are locked. It follows the
/// collection like the deck list header.
class AchievementsScreen extends StatefulWidget {
  const AchievementsScreen({super.key});

  @override
  State<AchievementsScreen> createState() => _AchievementsScreenState();
}

class _AchievementsScreenState extends State<AchievementsScreen> {
  late final _progress = AppScope.read(context).progress;
  late Future<List<golang.AchievementStatus>> _achievements;

  @override
  void initState() {
    super.initState();
    _achievements = golang.progress.achievements();
    _progress.addListener(_reload);
  }

  @override
  void dispose() {
    _progress.removeListener(_reload);
    super.dispose();
  }

  void _reload() =>
      setState(() => _achievements = golang.progress.achievements());

  @override
  Widget build(BuildContext context) {
    final l = context.l10n;
    final minimalist = AppScope.of(context).settings.minimalist;
    return Scaffold(
      appBar: AppBar(title: Text(l.achievements)),
      body: FutureBuilder(
        future: _achievements,
        builder: (context, snapshot) {
          if (snapshot.hasError) {
            return Center(child: Text(errorMessage(snapshot.error!)));
          }
          final all = snapshot.data;
          if (all == null) {
            return const Center(child: CircularProgressIndicator());
          }
          return GridView.builder(
            padding: const EdgeInsets.all(16),
            gridDelegate: const SliverGridDelegateWithMaxCrossAxisExtent(
              maxCrossAxisExtent: 220,
              mainAxisExtent: 232,
              crossAxisSpacing: 12,
              mainAxisSpacing: 12,
            ),
            itemCount: all.length,
            itemBuilder: (context, i) =>
                _Tile(achievement: all[i], minimalist: minimalist),
          );
        },
      ),
    );
  }
}

class _Tile extends StatelessWidget {
  const _Tile({required this.achievement, required this.minimalist});

  final golang.AchievementStatus achievement;
  final bool minimalist;

  @override
  Widget build(BuildContext context) {
    final l = context.l10n;
    final theme = Theme.of(context);
    final a = achievement;
    final earned = a.earned.length;
    final top = earned > 0 ? earned - 1 : null;
    final name = achievementName(l, a.id);
    // The highest earned tier, or the next one while locked.
    final shown = top ?? 0;
    final tier = l.tierWithThreshold(
      tierName(l, a.tierNames[shown]),
      tierThreshold(l, a.id, a.tiers[shown]),
    );
    final next = earned < a.tiers.length ? a.tiers[earned] : null;
    final status = switch (a) {
      _ when a.needsGoal => l.achievementNeedsGoal,
      _ when next == null => l.achievementAllEarned,
      _ when a.id == 'comeback' => l.achievementComebackHint,
      _ => l.achievementProgress(a.progress, next),
    };
    final small = theme.textTheme.bodySmall?.copyWith(
      color: theme.colorScheme.onSurfaceVariant,
    );
    return MergeSemantics(
      child: InkWell(
        key: Key('achievement-${a.id}'),
        borderRadius: BorderRadius.circular(12),
        onTap: a.needsGoal ? () => editDailyGoal(context) : null,
        child: Column(
          children: [
            Semantics(
              label: top == null
                  ? l.achievementLocked(name)
                  : l.achievementEarned(name, tierName(l, a.tierNames[top])),
              child: AchievementBadge(
                id: a.id,
                tier: top == null ? null : a.tierNames[top],
                stars: a.tiers.length > 1 ? earned : 0,
                minimalist: minimalist,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              name,
              style: theme.textTheme.titleSmall,
              textAlign: TextAlign.center,
            ),
            Text(tier, style: small, textAlign: TextAlign.center),
            if (top != null)
              Text(
                l.achievementEarnedOn(earnedDate(a.earned[top])),
                style: small,
                textAlign: TextAlign.center,
              ),
            const SizedBox(height: 4),
            Text(
              status,
              key: Key('achievement-${a.id}-status'),
              style: small,
              textAlign: TextAlign.center,
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
            ),
            if (next != null && !a.needsGoal && a.id != 'comeback')
              Padding(
                padding: const EdgeInsets.fromLTRB(24, 6, 24, 0),
                child: LinearProgressIndicator(
                  value: (a.progress / next).clamp(0, 1),
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
          ],
        ),
      ),
    );
  }
}
