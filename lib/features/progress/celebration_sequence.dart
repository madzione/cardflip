import 'dart:async';
import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/app/theme.dart';
import 'package:cardflip/data/settings_model.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';

import 'achievement_badge.dart';
import 'achievements.dart';
import 'celebration.dart';
import 'celebration_feedback.dart';
import 'confetti.dart';
import 'session_rewards.dart';

/// The study day of now with rollover hour [rolloverHour]: the date whose
/// weekday the streak page's week strip ends on.
DateTime studyToday(int rolloverHour) =>
    DateTime.now().subtract(Duration(hours: rolloverHour));

/// One full-screen page of the celebration sequence after a session
/// (progress spec, "Celebration sequence" and add-celebration-screens
/// design decision 4): a large centrepiece animated once, a headline,
/// supporting lines, the XP it earned and Continue. Continue, Enter and
/// Space call [onNext], Escape [onSkip], each only after the first half
/// second. It plays its sound and vibrates once when it appears, as
/// [settings] allow. Under reduce motion it shows its final state at once.
/// [today] is the study day, for the streak's week strip.
class CelebrationPage extends StatefulWidget {
  const CelebrationPage({
    super.key,
    required this.event,
    required this.settings,
    required this.today,
    required this.onNext,
    required this.onSkip,
  });

  final CelebrationEvent event;
  final SettingsModel settings;
  final DateTime today;
  final VoidCallback onNext;
  final VoidCallback onSkip;

  static const duration = Duration(milliseconds: 2400);

  /// How long input is ignored after the page appears, so that a key
  /// pressed for the last card does not skip it.
  static const inputGuard = Duration(milliseconds: 500);

  /// The size of the centrepiece.
  static const size = 160.0;

  @override
  State<CelebrationPage> createState() => _CelebrationPageState();
}

class _CelebrationPageState extends State<CelebrationPage>
    with SingleTickerProviderStateMixin {
  late final _controller = AnimationController(
    vsync: this,
    duration: CelebrationPage.duration,
  );
  final _continue = FocusNode(debugLabel: 'celebration continue');
  late final Timer _guard;
  bool _ready = false;
  bool _started = false;

  /// Whether the burst or confetti is on show; it starts at 60 %.
  bool _burst = false;

  @override
  void initState() {
    super.initState();
    _guard = Timer(CelebrationPage.inputGuard, () => _ready = true);
    final feedback = CelebrationFeedback.instance;
    feedback.play(widget.settings, switch (widget.event) {
      StreakCelebration() => CelebrationSound.streak,
      GoalCelebration() => CelebrationSound.goal,
      LevelCelebration() => CelebrationSound.level,
      AchievementCelebration() ||
      MoreAchievementsCelebration() => CelebrationSound.achievement,
    });
    feedback.vibrate(widget.settings, CelebrationHaptic.medium);
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) _continue.requestFocus();
    });
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (_started) return;
    _started = true;
    if (reduceMotion(context)) {
      _controller.value = 1;
    } else {
      _controller
        ..addListener(_tick)
        ..forward();
    }
  }

  void _tick() {
    if (!_burst && _controller.value >= 0.6) {
      _controller.removeListener(_tick);
      setState(() => _burst = true);
    }
  }

  @override
  void dispose() {
    _guard.cancel();
    _controller.dispose();
    _continue.dispose();
    super.dispose();
  }

  void _next() {
    if (_ready) widget.onNext();
  }

  void _skip() {
    if (_ready) widget.onSkip();
  }

  @override
  Widget build(BuildContext context) {
    final l = context.l10n;
    final theme = Theme.of(context);
    final colors = theme.colorScheme;
    final minimalist = widget.settings.minimalist;
    final event = widget.event;
    final (headline, lines, xp) = _texts(context);
    final confetti = switch (event) {
      LevelCelebration() => [colors.primary, colors.tertiary, colors.secondary],
      AchievementCelebration(:final tier) => [
        BadgeColors.of(context).rim(tier.tierName).bottom,
        AchievementLook.of(tier.id).discTop,
        colors.primary,
        colors.tertiary,
      ],
      MoreAchievementsCelebration() => [
        colors.primary,
        colors.tertiary,
        colors.secondary,
        BadgeColors.of(context).gold.bottom,
      ],
      _ => null,
    };
    final centrepiece = SizedBox.square(
      dimension: CelebrationPage.size,
      child: Stack(
        clipBehavior: Clip.none,
        alignment: Alignment.center,
        children: [
          if (!minimalist)
            Positioned.fill(
              child: OverflowBox(
                maxWidth: CelebrationPage.size * 1.8,
                maxHeight: CelebrationPage.size * 1.8,
                child: DecoratedBox(
                  key: const Key('celebration-glow'),
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    gradient: RadialGradient(
                      colors: [
                        colors.primaryContainer,
                        colors.primaryContainer.withValues(alpha: 0),
                      ],
                    ),
                  ),
                  child: const SizedBox.expand(),
                ),
              ),
            ),
          AnimatedBuilder(
            animation: _controller,
            builder: (context, _) => _centrepiece(context, _controller.value),
          ),
          if (_burst && confetti == null)
            Positioned.fill(
              child: Celebration(
                key: const Key('celebration'),
                duration: const Duration(milliseconds: 1000),
                onEnd: () {
                  if (mounted) setState(() => _burst = false);
                },
              ),
            ),
        ],
      ),
    );
    final content = Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        centrepiece,
        const SizedBox(height: 32),
        Text(
          headline,
          key: const Key('celebration-headline'),
          textAlign: TextAlign.center,
          style: theme.textTheme.headlineMedium?.copyWith(
            fontWeight: FontWeight.bold,
          ),
        ),
        for (final line in lines)
          Padding(
            padding: const EdgeInsets.only(top: 8),
            child: Text(
              line,
              textAlign: TextAlign.center,
              style: theme.textTheme.titleMedium,
            ),
          ),
        if (xp > 0)
          Padding(
            padding: const EdgeInsets.only(top: 16),
            child: _XPChip(text: l.xpGained(xp), minimalist: minimalist),
          ),
        if (event case StreakCelebration(:final week))
          Padding(
            padding: const EdgeInsets.only(top: 24),
            child: AnimatedBuilder(
              animation: _controller,
              builder: (context, _) => WeekStrip(
                week: week,
                today: widget.today,
                stamp: _part(_controller.value, 0.6, 0.8),
                minimalist: minimalist,
              ),
            ),
          ),
      ],
    );
    return CallbackShortcuts(
      bindings: {
        const SingleActivator(LogicalKeyboardKey.enter): _next,
        const SingleActivator(LogicalKeyboardKey.numpadEnter): _next,
        const SingleActivator(LogicalKeyboardKey.space): _next,
        const SingleActivator(LogicalKeyboardKey.escape): _skip,
      },
      child: ColoredBox(
        color: colors.surface,
        child: Stack(
          children: [
            SafeArea(
              child: Column(
                children: [
                  Expanded(
                    child: Center(
                      child: SingleChildScrollView(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 24,
                          vertical: 32,
                        ),
                        child: Semantics(
                          container: true,
                          liveRegion: true,
                          label: [headline, ...lines].join('\n'),
                          child: ExcludeSemantics(child: content),
                        ),
                      ),
                    ),
                  ),
                  Padding(
                    padding: const EdgeInsets.fromLTRB(24, 0, 24, 24),
                    child: ConstrainedBox(
                      constraints: const BoxConstraints(maxWidth: 400),
                      child: SizedBox(
                        width: double.infinity,
                        child: FilledButton(
                          key: const Key('celebration-continue'),
                          focusNode: _continue,
                          onPressed: _next,
                          child: Text(l.continueAction),
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
            if (_burst && confetti != null)
              Positioned.fill(
                child: Confetti(
                  key: const Key('confetti'),
                  seed: switch (event) {
                    AchievementCelebration(:final tier) => tier.threshold,
                    _ => 1,
                  },
                  colors: confetti,
                  origin: const Alignment(0, -0.35),
                  onEnd: () {
                    if (mounted) setState(() => _burst = false);
                  },
                ),
              ),
          ],
        ),
      ),
    );
  }

  /// The headline, the supporting lines and the XP earned (0 for none).
  (String, List<String>, int) _texts(BuildContext context) {
    final l = context.l10n;
    return switch (widget.event) {
      StreakCelebration(:final to, :final best) => (
        l.celebrationStreak(to),
        [if (best) l.celebrationBest],
        0,
      ),
      GoalCelebration(:final progress, :final goal, :final kind, :final xp) => (
        l.goalReached,
        [
          kind == 'minutes'
              ? l.goalProgressMinutes(progress, goal)
              : l.goalProgressCards(progress, goal),
        ],
        xp,
      ),
      LevelCelebration(:final toLevel, :final xpLeft) => (
        l.levelBanner(toLevel),
        [l.xpToLevel(xpLeft, toLevel + 1)],
        0,
      ),
      AchievementCelebration(:final tier) => (
        l.celebrationNewAchievement,
        [_tierLine(l, tier)],
        tier.xp,
      ),
      MoreAchievementsCelebration(:final tiers) => (
        l.celebrationNewAchievements,
        [newTierLabels(l, tiers)],
        tiers.fold(0, (sum, a) => sum + a.xp),
      ),
    };
  }

  static String _tierLine(AppLocalizations l, golang.NewAchievement a) {
    final label = newTierLabel(l, a.id, a.threshold);
    return AchievementLook.of(a.id).oneTier
        ? label
        : l.celebrationTier(label, tierName(l, a.tierName));
  }

  Widget _centrepiece(BuildContext context, double t) {
    final minimalist = widget.settings.minimalist;
    return switch (widget.event) {
      StreakCelebration(:final from, :final to) => _StreakFlame(
        t: t,
        from: from,
        to: to,
      ),
      GoalCelebration() => _GoalRing(t: t),
      LevelCelebration(:final fromLevel, :final fromFraction, :final toLevel) =>
        _LevelRing(
          t: t,
          from: fromLevel,
          fraction: fromFraction,
          to: toLevel,
          minimalist: minimalist,
        ),
      AchievementCelebration(:final tier) => _AchievementDrop(
        t: t,
        tier: tier,
        minimalist: minimalist,
      ),
      MoreAchievementsCelebration(:final tiers) => _MoreBadges(
        t: t,
        tiers: tiers,
        minimalist: minimalist,
      ),
    };
  }
}

/// How far the part from [start] to [end] (0 to 1) is at [t], from 0 to 1.
double _part(double t, double start, double end) =>
    ((t - start) / (end - start)).clamp(0, 1);

/// The XP an event earned, "+50 XP", as a pill.
class _XPChip extends StatelessWidget {
  const _XPChip({required this.text, required this.minimalist});

  final String text;
  final bool minimalist;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colors = theme.colorScheme;
    return DecoratedBox(
      decoration: BoxDecoration(
        color: minimalist ? null : colors.tertiaryContainer,
        border: minimalist ? Border.all(color: colors.outline) : null,
        borderRadius: BorderRadius.circular(999),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
        child: Text(
          text,
          key: const Key('celebration-xp'),
          style: theme.textTheme.titleMedium?.copyWith(
            fontWeight: FontWeight.bold,
            color: minimalist ? colors.onSurface : colors.onTertiaryContainer,
          ),
        ),
      ),
    );
  }
}

/// The streak's flame: it scales in, then the day count rolls from [from]
/// to [to].
class _StreakFlame extends StatelessWidget {
  const _StreakFlame({required this.t, required this.from, required this.to});

  final double t;
  final int from;
  final int to;

  @override
  Widget build(BuildContext context) {
    final look = AchievementLook.of('streak');
    final format = MaterialLocalizations.of(context).formatDecimal;
    final scale = Curves.easeOutBack.transform(_part(t, 0, 0.4));
    final roll = Curves.easeInOutCubic.transform(_part(t, 0.4, 0.6));
    const size = CelebrationPage.size;
    final style = Theme.of(context).textTheme.displaySmall?.copyWith(
      color: Colors.white,
      fontWeight: FontWeight.w900,
      shadows: const [Shadow(color: Color(0x66000000), blurRadius: 4)],
    );
    Widget number(int n, double offset) => FractionalTranslation(
      translation: Offset(0, offset),
      child: Opacity(
        opacity: (1 - offset.abs()).clamp(0, 1),
        child: Text(format(n), style: style),
      ),
    );
    return Transform.scale(
      scale: scale,
      child: Stack(
        alignment: Alignment.center,
        children: [
          ShaderMask(
            blendMode: BlendMode.srcIn,
            shaderCallback: (rect) => LinearGradient(
              begin: Alignment.topCenter,
              end: Alignment.bottomCenter,
              colors: [look.discTop, look.discBottom],
            ).createShader(rect),
            child: const Icon(Icons.local_fire_department, size: size),
          ),
          Positioned(
            left: 0,
            right: 0,
            bottom: size * 0.16,
            child: ClipRect(
              child: Center(
                child: Stack(
                  alignment: Alignment.center,
                  children: [
                    if (roll < 1) number(from, -roll),
                    if (roll > 0) number(to, 1 - roll),
                  ],
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

/// The week strip of the streak page (progress spec, "Streak
/// celebration"): the days of the current week from the first day of the
/// week of the interface language, studied days ticked, rest days marked,
/// other past days empty and later days blank. Today's tick stamps in as
/// [stamp] goes from 0 to 1. [week] is `Progress.week`, ending on [today].
class WeekStrip extends StatelessWidget {
  const WeekStrip({
    super.key,
    required this.week,
    required this.today,
    this.stamp = 1,
    this.minimalist = false,
  });

  final List<String> week;
  final DateTime today;
  final double stamp;
  final bool minimalist;

  @override
  Widget build(BuildContext context) {
    final material = MaterialLocalizations.of(context);
    final first = material.firstDayOfWeekIndex;
    // Today's place in the week, from 0; DateTime counts Sunday as 7.
    final todayAt = (today.weekday % 7 - first + 7) % 7;
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        for (var i = 0; i < 7; i++)
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 4),
            child: _WeekDay(
              index: i,
              label: material.narrowWeekdays[(first + i) % 7],
              state: i > todayAt
                  ? 'later'
                  : week.elementAtOrNull(week.length - 1 - (todayAt - i)) ??
                        'none',
              today: i == todayAt,
              stamp: stamp,
              minimalist: minimalist,
            ),
          ),
      ],
    );
  }
}

/// One day of the [WeekStrip]: [state] "study", "rest", "none" or "later".
class _WeekDay extends StatelessWidget {
  const _WeekDay({
    required this.index,
    required this.label,
    required this.state,
    required this.today,
    required this.stamp,
    required this.minimalist,
  });

  final int index;
  final String label;
  final String state;
  final bool today;
  final double stamp;
  final bool minimalist;

  static const size = 32.0;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colors = theme.colorScheme;
    final flame = AchievementLook.of('streak').discBottom;
    final studied = state == 'study' && (!today || stamp > 0);
    final empty = BoxDecoration(
      shape: BoxShape.circle,
      border: Border.all(
        color: state == 'later' ? colors.outlineVariant : colors.outline,
        width: 1.5,
      ),
    );
    Widget mark = switch (state) {
      _ when studied => Container(
        decoration: BoxDecoration(shape: BoxShape.circle, color: flame),
        child: const Icon(Icons.check, size: 20, color: Colors.white),
      ),
      'rest' => Container(
        decoration: BoxDecoration(
          shape: BoxShape.circle,
          color: minimalist ? null : colors.secondaryContainer,
          border: minimalist ? Border.all(color: colors.outline) : null,
        ),
        child: Icon(
          Icons.bedtime_outlined,
          size: 18,
          color: minimalist ? colors.onSurface : colors.onSecondaryContainer,
        ),
      ),
      _ => DecoratedBox(decoration: empty),
    };
    if (today && studied && stamp < 1) {
      final s = Curves.easeOut.transform(stamp);
      mark = Stack(
        fit: StackFit.expand,
        children: [
          DecoratedBox(decoration: empty),
          Opacity(
            opacity: s,
            child: Transform.scale(scale: 1.4 - 0.4 * s, child: mark),
          ),
        ],
      );
    }
    return Column(
      key: ValueKey('week-$index-$state'),
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(
          label,
          style: theme.textTheme.labelMedium?.copyWith(
            color: state == 'later' ? colors.outline : colors.onSurfaceVariant,
            fontWeight: today ? FontWeight.bold : null,
          ),
        ),
        const SizedBox(height: 6),
        SizedBox.square(dimension: size, child: mark),
      ],
    );
  }
}

/// The goal ring: it fills up, then its flag turns into a check.
class _GoalRing extends StatelessWidget {
  const _GoalRing({required this.t});

  final double t;

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    final fill = Curves.easeInOut.transform(_part(t, 0, 0.4));
    final turn = _part(t, 0.4, 0.6);
    return SizedBox.square(
      dimension: CelebrationPage.size,
      child: CustomPaint(
        painter: _RingPainter(
          value: fill,
          track: colors.surfaceContainerHighest,
          fill: colors.primary,
          stroke: 12,
          disc: turn > 0 ? colors.primary.withValues(alpha: turn) : null,
        ),
        child: Center(
          child: Stack(
            alignment: Alignment.center,
            children: [
              if (turn < 1)
                Opacity(
                  opacity: 1 - turn,
                  child: Icon(
                    Icons.flag_outlined,
                    size: 64,
                    color: colors.primary,
                  ),
                ),
              if (turn > 0)
                Transform.scale(
                  scale: Curves.easeOutBack.transform(turn),
                  child: Icon(
                    Icons.check_rounded,
                    key: const Key('celebration-check'),
                    size: 88,
                    color: colors.onPrimary,
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}

/// The level ring: it fills from [fraction] into level [from] to the end
/// of that level, then the number flips to [to] with a flash, and rays
/// turn and fade.
class _LevelRing extends StatelessWidget {
  const _LevelRing({
    required this.t,
    required this.from,
    required this.fraction,
    required this.to,
    required this.minimalist,
  });

  final double t;
  final int from;
  final double fraction;
  final int to;
  final bool minimalist;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colors = theme.colorScheme;
    final format = MaterialLocalizations.of(context).formatDecimal;
    final fill =
        fraction +
        (1 - fraction) * Curves.easeInOut.transform(_part(t, 0, 0.4));
    final flip = _part(t, 0.4, 0.6);
    final rays = _part(t, 0.6, 1);
    // The old number turns away in the first half of the flip, the new one
    // in from the other side in the second.
    final first = flip < 0.5;
    final angle = first ? flip * math.pi : (flip - 1) * math.pi;
    final flash = math.sin(flip * math.pi);
    return SizedBox.square(
      dimension: CelebrationPage.size,
      child: Stack(
        clipBehavior: Clip.none,
        alignment: Alignment.center,
        children: [
          if (!minimalist && rays > 0 && rays < 1)
            Positioned.fill(
              child: OverflowBox(
                maxWidth: CelebrationPage.size * 2,
                maxHeight: CelebrationPage.size * 2,
                child: CustomPaint(
                  key: const Key('celebration-rays'),
                  size: const Size.square(CelebrationPage.size * 2),
                  painter: _RaysPainter(
                    turn: rays,
                    color: colors.primary.withValues(alpha: 0.35 * (1 - rays)),
                  ),
                ),
              ),
            ),
          CustomPaint(
            size: const Size.square(CelebrationPage.size),
            painter: _RingPainter(
              value: fill,
              track: colors.surfaceContainerHighest,
              fill: colors.primary,
              stroke: 12,
              disc: flash > 0
                  ? colors.primaryContainer.withValues(alpha: flash)
                  : null,
            ),
          ),
          Transform(
            alignment: Alignment.center,
            transform: Matrix4.identity()
              ..setEntry(3, 2, 0.002)
              ..rotateX(angle),
            child: Text(
              format(first ? from : to),
              key: const Key('celebration-level'),
              style: theme.textTheme.displayLarge?.copyWith(
                fontWeight: FontWeight.w800,
                color: colors.primary,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

/// A ring of [stroke] filled to [value] (0 to 1), with a [disc] inside it
/// if given.
class _RingPainter extends CustomPainter {
  _RingPainter({
    required this.value,
    required this.track,
    required this.fill,
    required this.stroke,
    this.disc,
  });

  final double value;
  final Color track;
  final Color fill;
  final double stroke;
  final Color? disc;

  @override
  void paint(Canvas canvas, Size size) {
    final rect = (Offset.zero & size).deflate(stroke / 2);
    final paint = Paint()
      ..style = PaintingStyle.stroke
      ..strokeWidth = stroke
      ..strokeCap = StrokeCap.round;
    canvas.drawArc(rect, 0, 2 * math.pi, false, paint..color = track);
    // Over the track's inner edge and under the fill, so no seam shows.
    if (disc case final disc?) {
      canvas.drawCircle(rect.center, rect.width / 2, Paint()..color = disc);
    }
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
      old.value != value ||
      old.track != track ||
      old.fill != fill ||
      old.stroke != stroke ||
      old.disc != disc;
}

/// Twelve rays around the centre, turned by [turn] (0 to 1).
class _RaysPainter extends CustomPainter {
  _RaysPainter({required this.turn, required this.color});

  final double turn;
  final Color color;

  static const _count = 12;

  @override
  void paint(Canvas canvas, Size size) {
    final center = size.center(Offset.zero);
    final radius = size.shortestSide / 2;
    final paint = Paint()..color = color;
    const half = math.pi / _count / 2.5;
    for (var i = 0; i < _count; i++) {
      final a = i * 2 * math.pi / _count + turn * math.pi / 3;
      canvas.drawPath(
        Path()
          ..moveTo(center.dx, center.dy)
          ..lineTo(
            center.dx + radius * math.cos(a - half),
            center.dy + radius * math.sin(a - half),
          )
          ..lineTo(
            center.dx + radius * math.cos(a + half),
            center.dy + radius * math.sin(a + half),
          )
          ..close(),
        paint,
      );
    }
  }

  @override
  bool shouldRepaint(_RaysPainter old) =>
      old.turn != turn || old.color != color;
}

/// An achievement tier's badge: it drops in from above, a shine sweeps
/// over it, and its stars appear one by one.
class _AchievementDrop extends StatelessWidget {
  const _AchievementDrop({
    required this.t,
    required this.tier,
    required this.minimalist,
  });

  final double t;
  final golang.NewAchievement tier;
  final bool minimalist;

  @override
  Widget build(BuildContext context) {
    final drop = _part(t, 0, 0.4);
    final shine = _part(t, 0.4, 0.6);
    final stars = newTierStars(tier);
    final shown = (stars * _part(t, 0.6, 0.9)).ceil();
    Widget badge = AchievementBadge(
      key: ValueKey('celebration-badge-$shown'),
      id: tier.id,
      tier: tier.tierName,
      stars: shown,
      size: CelebrationPage.size,
      minimalist: minimalist,
    );
    if (!minimalist && shine > 0 && shine < 1) {
      badge = ShaderMask(
        blendMode: BlendMode.srcATop,
        shaderCallback: (rect) => LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [
            Colors.white.withValues(alpha: 0),
            Colors.white.withValues(alpha: 0.55),
            Colors.white.withValues(alpha: 0),
          ],
          stops: [math.max(0, shine - 0.15), shine, math.min(1, shine + 0.15)],
        ).createShader(rect),
        child: badge,
      );
    }
    return Opacity(
      opacity: _part(t, 0, 0.15),
      child: Transform.translate(
        offset: Offset(
          0,
          -CelebrationPage.size * (1 - Curves.easeOutBack.transform(drop)),
        ),
        child: badge,
      ),
    );
  }
}

/// The badges of the tiers beyond the first three, popping in one after
/// the other.
class _MoreBadges extends StatelessWidget {
  const _MoreBadges({
    required this.t,
    required this.tiers,
    required this.minimalist,
  });

  final double t;
  final List<golang.NewAchievement> tiers;
  final bool minimalist;

  @override
  Widget build(BuildContext context) {
    final size = tiers.length <= 2 ? 96.0 : 72.0;
    final step = 0.4 / tiers.length;
    return OverflowBox(
      maxWidth: CelebrationPage.size * 2.4,
      child: Wrap(
        alignment: WrapAlignment.center,
        crossAxisAlignment: WrapCrossAlignment.center,
        spacing: 8,
        runSpacing: 8,
        children: [
          for (final (i, a) in tiers.indexed)
            Transform.scale(
              scale: Curves.easeOutBack.transform(
                _part(t, i * step, i * step + 0.3),
              ),
              child: AchievementBadge(
                key: ValueKey('celebration-badge-${a.key}'),
                id: a.id,
                tier: a.tierName,
                stars: newTierStars(a),
                size: size,
                minimalist: minimalist,
              ),
            ),
        ],
      ),
    );
  }
}
