import 'dart:math' as math;

import 'package:flutter/material.dart';

import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';

import 'achievement_badge.dart';
import 'achievements.dart';
import 'achievements_screen.dart';

/// A level and how far into it its XP are, from 0 to 1.
typedef LevelAt = ({int level, double fraction});

/// The session-complete screen's card of what the session earned (progress
/// spec, "Reward card on the summary"): the reached goal as a full ring,
/// the new level in its ring, up to three badges of the earned [tiers] and
/// "+N" for more, and lines saying so. It is a static recap after the
/// celebration pages, without an animation of its own. Tapping it opens
/// the achievements screen.
class RewardCard extends StatelessWidget {
  const RewardCard({
    super.key,
    required this.goalReached,
    required this.levelFrom,
    required this.levelTo,
    required this.tiers,
    this.minimalist = false,
  });

  final bool goalReached;

  /// The level when the session started and now; the card shows the level
  /// reached when [levelTo] is higher.
  final LevelAt levelFrom;
  final LevelAt levelTo;

  /// The tiers to show and name: the summary leaves out Level tiers when
  /// the level was reached.
  final List<golang.NewAchievement> tiers;
  final bool minimalist;

  /// The most badges shown; more become "+N".
  static const maxBadges = 3;

  bool get levelUp => levelTo.level > levelFrom.level;

  bool get isEmpty => !goalReached && !levelUp && tiers.isEmpty;

  @override
  Widget build(BuildContext context) {
    final l = context.l10n;
    final theme = Theme.of(context);
    final colors = theme.colorScheme;
    final lines = [
      if (goalReached) l.goalReached,
      if (levelUp) l.levelReached(levelTo.level),
      if (tiers.isNotEmpty) l.newAchievementsLine(newTierLabels(l, tiers)),
    ];
    final shown = tiers.take(maxBadges).toList();
    final more = tiers.length - shown.length;
    return Card(
      key: const Key('reward-card'),
      elevation: 0,
      color: minimalist ? Colors.transparent : colors.surfaceContainerHigh,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: minimalist
            ? BorderSide(color: colors.outlineVariant)
            : BorderSide.none,
      ),
      margin: EdgeInsets.zero,
      clipBehavior: Clip.antiAlias,
      child: Semantics(
        button: true,
        label: lines.join('\n'),
        child: InkWell(
          onTap: () => showAchievements(context),
          child: ExcludeSemantics(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Wrap(
                    alignment: WrapAlignment.center,
                    crossAxisAlignment: WrapCrossAlignment.center,
                    spacing: 8,
                    runSpacing: 8,
                    children: [
                      if (goalReached)
                        _Ring(
                          key: const Key('reward-goal'),
                          value: 1,
                          child: Icon(
                            Icons.flag_outlined,
                            color: colors.primary,
                          ),
                        ),
                      if (levelUp)
                        _Ring(
                          key: const Key('reward-level'),
                          value: levelTo.fraction,
                          child: Text(
                            '${levelTo.level}',
                            style: theme.textTheme.titleMedium,
                          ),
                        ),
                      for (final a in shown)
                        AchievementBadge(
                          id: a.id,
                          tier: a.tierName,
                          stars: newTierStars(a),
                          size: _Ring.size,
                          minimalist: minimalist,
                        ),
                      if (more > 0)
                        Text(
                          '+$more',
                          key: const Key('reward-more'),
                          style: theme.textTheme.titleMedium,
                        ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  for (final line in lines)
                    Padding(
                      padding: const EdgeInsets.only(top: 2),
                      child: Text(line, textAlign: TextAlign.center),
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

/// A ring filled to [value] (0 to 1) around [child].
class _Ring extends StatelessWidget {
  const _Ring({super.key, required this.value, required this.child});

  final double value;
  final Widget child;

  static const size = 48.0;

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    return SizedBox.square(
      dimension: size,
      child: CustomPaint(
        painter: _RingPainter(
          value: value,
          track: colors.surfaceContainerHighest,
          fill: colors.primary,
        ),
        child: Center(child: child),
      ),
    );
  }
}

class _RingPainter extends CustomPainter {
  _RingPainter({required this.value, required this.track, required this.fill});

  final double value;
  final Color track;
  final Color fill;

  static const stroke = 4.0;

  @override
  void paint(Canvas canvas, Size size) {
    final rect = (Offset.zero & size).deflate(stroke / 2);
    final paint = Paint()
      ..style = PaintingStyle.stroke
      ..strokeWidth = stroke
      ..strokeCap = StrokeCap.round;
    canvas.drawArc(rect, 0, 2 * math.pi, false, paint..color = track);
    if (value > 0) {
      canvas.drawArc(
        rect,
        -math.pi / 2,
        2 * math.pi * value.clamp(0, 1),
        false,
        paint..color = fill,
      );
    }
  }

  @override
  bool shouldRepaint(_RingPainter old) =>
      old.value != value || old.track != track || old.fill != fill;
}
