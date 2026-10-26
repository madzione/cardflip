import 'dart:math' as math;

import 'package:flutter/material.dart';

import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';

import 'achievement_badge.dart';
import 'achievements.dart';
import 'achievements_screen.dart';
import 'celebration.dart';

/// A level and how far into it its XP are, from 0 to 1.
typedef LevelAt = ({int level, double fraction});

/// The session-complete screen's card of what the session earned (progress
/// spec, "Reward card on the summary" and "Celebrations"): the reached goal
/// as a full ring, the new level in a ring filled from where the session
/// started, up to three badges of the earned [tiers] and "+N" for more, and
/// lines saying so. It animates once, about 2.6 seconds, with a particle
/// burst behind it, unless [animate] is off; everything but the card's own
/// tap ignores pointers. Tapping it opens the achievements screen.
class RewardCard extends StatefulWidget {
  const RewardCard({
    super.key,
    required this.goalReached,
    required this.levelFrom,
    required this.levelTo,
    required this.tiers,
    this.animate = true,
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
  final bool animate;
  final bool minimalist;

  static const duration = Duration(milliseconds: 2600);

  /// The most badges shown; more become "+N".
  static const maxBadges = 3;

  bool get levelUp => levelTo.level > levelFrom.level;

  bool get isEmpty => !goalReached && !levelUp && tiers.isEmpty;

  @override
  State<RewardCard> createState() => _RewardCardState();
}

class _RewardCardState extends State<RewardCard>
    with SingleTickerProviderStateMixin {
  late final _controller = AnimationController(
    vsync: this,
    duration: RewardCard.duration,
  );

  /// Whether the particle burst is on show; it starts at 1.2 seconds.
  bool _burst = false;

  static const _ms = 2600.0;

  @override
  void initState() {
    super.initState();
    if (!widget.animate) {
      _controller.value = 1;
      return;
    }
    _controller.addListener(_tick);
    _controller.forward();
  }

  void _tick() {
    if (!_burst && _controller.value >= 1200 / _ms) {
      _controller.removeListener(_tick);
      setState(() => _burst = true);
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  /// How far the part from [startMs] to [endMs] is at [t], from 0 to 1.
  static double _part(double t, double startMs, double endMs) =>
      ((t * _ms - startMs) / (endMs - startMs)).clamp(0, 1);

  @override
  Widget build(BuildContext context) {
    final l = context.l10n;
    final theme = Theme.of(context);
    final colors = theme.colorScheme;
    final tiers = widget.tiers;
    final lines = [
      if (widget.goalReached) l.goalReached,
      if (widget.levelUp) l.levelReached(widget.levelTo.level),
      if (tiers.isNotEmpty) l.newAchievementsLine(newTierLabels(l, tiers)),
    ];
    final shown = tiers.take(RewardCard.maxBadges).toList();
    final more = tiers.length - shown.length;
    final card = Card(
      key: const Key('reward-card'),
      elevation: 0,
      color: widget.minimalist
          ? Colors.transparent
          : colors.surfaceContainerHigh,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: widget.minimalist
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
              child: AnimatedBuilder(
                animation: _controller,
                builder: (context, _) {
                  final t = _controller.value;
                  return Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Wrap(
                        alignment: WrapAlignment.center,
                        crossAxisAlignment: WrapCrossAlignment.center,
                        spacing: 8,
                        runSpacing: 8,
                        children: [
                          if (widget.goalReached)
                            _Ring(
                              key: const Key('reward-goal'),
                              value: Curves.easeOut.transform(
                                _part(t, 400, 1000),
                              ),
                              child: Icon(
                                Icons.flag_outlined,
                                color: colors.primary,
                              ),
                            ),
                          if (widget.levelUp) _level(t),
                          for (final (i, a) in shown.indexed)
                            Transform.scale(
                              scale: Curves.easeOutBack.transform(
                                _part(t, 800 + 200.0 * i, 1100 + 200.0 * i),
                              ),
                              child: AchievementBadge(
                                id: a.id,
                                tier: a.tierName,
                                stars: newTierStars(a),
                                size: _Ring.size,
                                minimalist: widget.minimalist,
                              ),
                            ),
                          if (more > 0)
                            Opacity(
                              opacity: _part(t, 1400, 1700),
                              child: Text(
                                '+$more',
                                key: const Key('reward-more'),
                                style: theme.textTheme.titleMedium,
                              ),
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
                  );
                },
              ),
            ),
          ),
        ),
      ),
    );
    return Stack(
      clipBehavior: Clip.none,
      children: [
        if (_burst)
          Positioned.fill(
            child: Celebration(
              key: const Key('celebration'),
              duration: const Duration(milliseconds: 1400),
              onEnd: () {
                if (mounted) setState(() => _burst = false);
              },
            ),
          ),
        card,
      ],
    );
  }

  /// The level ring: it fills from where the session started to the end of
  /// that level over 1.2 seconds, then shows the new level and how far
  /// into it the XP are.
  Widget _level(double t) {
    final from = widget.levelFrom, to = widget.levelTo;
    final rising = t * _ms < 1200;
    final value = rising
        ? from.fraction +
              (1 - from.fraction) *
                  Curves.easeInOut.transform(_part(t, 0, 1200))
        : to.fraction;
    return _Ring(
      key: const Key('reward-level'),
      value: value,
      child: Text(
        '${rising ? from.level : to.level}',
        style: Theme.of(context).textTheme.titleMedium,
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
