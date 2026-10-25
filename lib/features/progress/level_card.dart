import 'package:flutter/material.dart';

import 'package:cardflip/features/settings/daily_goal_dialog.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';

/// The achievements screen's card of the level (progress spec, "Level
/// card"): the level, the XP toward the next one, where the XP came from
/// and how they are earned. Without a daily goal it says that one earns
/// more; tapping that sets one.
class LevelCard extends StatelessWidget {
  const LevelCard({
    super.key,
    required this.level,
    required this.hasGoal,
    this.minimalist = false,
  });

  final golang.LevelStatus level;
  final bool hasGoal;
  final bool minimalist;

  @override
  Widget build(BuildContext context) {
    final l = context.l10n;
    final theme = Theme.of(context);
    final colors = theme.colorScheme;
    final v = level;
    final done = v.xp - v.levelStart;
    final need = v.nextLevelXP - v.levelStart;
    final small = theme.textTheme.bodySmall?.copyWith(
      color: colors.onSurfaceVariant,
    );
    final sources = [
      (l.xpFromAnswers, v.fromAnswers),
      (l.xpFromStudyDays, v.fromStudyDays),
      (l.xpFromGoalDays, v.fromGoalDays),
      (l.xpFromLongTerm, v.fromLongTerm),
      (l.xpFromAchievements, v.fromAchievements),
    ];
    return Card(
      key: const Key('level-card'),
      elevation: 0,
      color: minimalist ? Colors.transparent : colors.surfaceContainerHigh,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: minimalist
            ? BorderSide(color: colors.outlineVariant)
            : BorderSide.none,
      ),
      margin: EdgeInsets.zero,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.trending_up, color: colors.primary),
                const SizedBox(width: 8),
                Text(l.levelTitle(v.level), style: theme.textTheme.titleLarge),
              ],
            ),
            const SizedBox(height: 12),
            LinearProgressIndicator(
              key: const Key('level-bar'),
              value: need <= 0 ? 0 : (done / need).clamp(0, 1),
              borderRadius: BorderRadius.circular(2),
            ),
            const SizedBox(height: 6),
            Text(
              l.levelXpProgress(done, need, v.level + 1),
              key: const Key('level-xp'),
              style: theme.textTheme.bodyMedium,
            ),
            const SizedBox(height: 12),
            for (final (label, xp) in sources)
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 2),
                child: Row(
                  children: [
                    Expanded(child: Text(label)),
                    Text(l.xpAmount(xp)),
                  ],
                ),
              ),
            const SizedBox(height: 12),
            Text(l.xpExplanation, style: small),
            if (!hasGoal)
              Padding(
                padding: const EdgeInsets.only(top: 4),
                child: InkWell(
                  key: const Key('level-goal-hint'),
                  onTap: () => editDailyGoal(context),
                  child: Text(
                    l.xpNoGoalHint,
                    style: small?.copyWith(color: colors.primary),
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }
}
