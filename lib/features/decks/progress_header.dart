import 'dart:math' as math;

import 'package:flutter/material.dart';

import 'package:cardflip/data/scope.dart';
import 'package:cardflip/features/progress/achievements_screen.dart';
import 'package:cardflip/features/stats/stats_screen.dart';
import 'package:cardflip/features/stats/stats_sections.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';

/// The deck list's row above the decks (progress spec, "Deck list
/// header"): the current streak, with a daily goal today's progress toward
/// it, and the level. Tapping it opens the collection statistics at
/// Streaks, tapping the level the achievements. It is hidden while streaks
/// and goals are turned off, and until the first progress is known.
class ProgressHeader extends StatelessWidget {
  const ProgressHeader({super.key});

  @override
  Widget build(BuildContext context) {
    final app = AppScope.of(context);
    return ListenableBuilder(
      listenable: Listenable.merge([app.progress, app.settings]),
      builder: (context, _) {
        final p = app.progress.progress;
        if (!app.settings.showProgress || p == null) {
          return const SizedBox.shrink();
        }
        return _Row(progress: p, minimalist: app.settings.minimalist);
      },
    );
  }
}

class _Row extends StatelessWidget {
  const _Row({required this.progress, required this.minimalist});

  final golang.Progress progress;
  final bool minimalist;

  @override
  Widget build(BuildContext context) {
    final l = context.l10n;
    final theme = Theme.of(context);
    final colors = theme.colorScheme;
    final p = progress;
    final goal = goalProgressLabel(l, p);
    final shape = RoundedRectangleBorder(
      borderRadius: BorderRadius.circular(12),
      side: minimalist
          ? BorderSide(color: colors.outlineVariant)
          : BorderSide.none,
    );
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 4),
      child: Material(
        key: const Key('progress-header'),
        color: minimalist ? Colors.transparent : colors.surfaceContainerHigh,
        shape: shape,
        child: InkWell(
          customBorder: shape,
          onTap: () => showStats(
            context,
            collectionScope,
            section: StatsSection.streaks,
          ),
          child: Tooltip(
            message: l.streakHeaderTooltip,
            excludeFromSemantics: true,
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              child: Row(
                children: [
                  Icon(
                    Icons.local_fire_department_outlined,
                    size: 20,
                    color: p.todayStudied ? colors.primary : colors.outline,
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      p.atRisk
                          ? l.streakAtRisk(p.current)
                          : l.streakDaysLabel(p.current),
                      style: theme.textTheme.bodyMedium,
                    ),
                  ),
                  if (goal != null) ...[
                    const SizedBox(width: 12),
                    SizedBox.square(
                      dimension: 18,
                      child: CircularProgressIndicator(
                        key: const Key('goal-ring'),
                        value: goalShare(p),
                        strokeWidth: 3,
                        backgroundColor: colors.outlineVariant,
                      ),
                    ),
                    const SizedBox(width: 8),
                    Text(goal, style: theme.textTheme.bodyMedium),
                  ],
                  const SizedBox(width: 8),
                  _Level(progress: p, minimalist: minimalist),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

/// The level in a ring of its progress toward the next one, which opens
/// the achievements.
class _Level extends StatelessWidget {
  const _Level({required this.progress, required this.minimalist});

  final golang.Progress progress;
  final bool minimalist;

  @override
  Widget build(BuildContext context) {
    final l = context.l10n;
    final theme = Theme.of(context);
    final colors = theme.colorScheme;
    final p = progress;
    // A node of its own: the row around it is a button too.
    return Semantics(
      container: true,
      label: l.levelHeaderSemantics(
        p.level,
        p.xp - p.levelStart,
        p.nextLevelXP - p.levelStart,
        p.level + 1,
      ),
      button: true,
      excludeSemantics: true,
      child: Tooltip(
        message: l.levelHeaderTooltip,
        excludeFromSemantics: true,
        child: InkWell(
          key: const Key('progress-level'),
          customBorder: const CircleBorder(),
          onTap: () => showAchievements(context),
          child: Padding(
            padding: const EdgeInsets.all(4),
            child: SizedBox.square(
              dimension: 24,
              child: Stack(
                fit: StackFit.expand,
                children: [
                  if (!minimalist)
                    DecoratedBox(
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        color: colors.surfaceContainerHighest,
                      ),
                    ),
                  CircularProgressIndicator(
                    key: const Key('level-ring'),
                    value: levelShare(p),
                    strokeWidth: 2.5,
                    color: colors.primary,
                    backgroundColor: colors.surfaceContainerHighest,
                  ),
                  Padding(
                    padding: const EdgeInsets.all(4),
                    child: FittedBox(
                      child: Text(
                        '${p.level}',
                        key: const Key('level-number'),
                        style: theme.textTheme.labelMedium?.copyWith(
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

/// How much of the way to the next level is done, from 0 to 1.
double levelShare(golang.Progress p) {
  final need = p.nextLevelXP - p.levelStart;
  return need <= 0 ? 0 : ((p.xp - p.levelStart) / need).clamp(0, 1);
}

/// "18 / 30 cards", or null without a daily goal.
String? goalProgressLabel(AppLocalizations l, golang.Progress p) =>
    switch (p.goalKind) {
      'cards' => l.goalProgressCards(p.todayCards, p.goal),
      'minutes' => l.goalProgressMinutes(p.todayMinutes, p.goal),
      _ => null,
    };

/// How much of the daily goal is done, from 0 to 1.
double goalShare(golang.Progress p) {
  final done = p.goalKind == 'minutes' ? p.todayMinutes : p.todayCards;
  return p.goal <= 0 ? 0 : math.min(done / p.goal, 1);
}
