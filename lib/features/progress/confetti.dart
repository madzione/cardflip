import 'dart:math' as math;

import 'package:flutter/material.dart';

/// Confetti for the level and achievement celebration pages
/// (add-celebration-screens design decision 5): about 60 rectangles and
/// circles in [colors] (by default the theme's primary, tertiary and
/// secondary) burst upward from [origin] and fall under gravity with drift
/// and spin, over [duration], then [onEnd] is called. The pieces come from
/// a fixed [seed], so goldens are stable and two bursts can differ. It
/// ignores pointers and may paint outside its box. Callers leave it out
/// when reduce motion is in effect.
class Confetti extends StatelessWidget {
  const Confetti({
    super.key,
    this.seed = 1,
    this.colors,
    this.origin = Alignment.center,
    this.onEnd,
    this.duration = const Duration(milliseconds: 1800),
  });

  final int seed;
  final List<Color>? colors;
  final Alignment origin;
  final VoidCallback? onEnd;
  final Duration duration;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final colors =
        this.colors ?? [scheme.primary, scheme.tertiary, scheme.secondary];
    return IgnorePointer(
      child: ExcludeSemantics(
        child: TweenAnimationBuilder<double>(
          tween: Tween(begin: 0, end: 1),
          duration: duration,
          onEnd: onEnd,
          builder: (context, t, _) => CustomPaint(
            size: Size.infinite,
            painter: ConfettiPainter(
              t: t,
              seed: seed,
              colors: colors,
              origin: origin,
              seconds: duration.inMilliseconds / 1000,
            ),
          ),
        ),
      ),
    );
  }
}

/// The confetti at time [t] from 0 to 1 of a burst lasting [seconds].
class ConfettiPainter extends CustomPainter {
  ConfettiPainter({
    required this.t,
    required this.seed,
    required this.colors,
    this.origin = Alignment.center,
    this.seconds = 1.8,
  }) : _pieces = _piecesOf(seed);

  final double t;
  final int seed;
  final List<Color> colors;
  final Alignment origin;
  final double seconds;
  final List<_Piece> _pieces;

  static const count = 60;

  static final _cache = <int, List<_Piece>>{};

  static List<_Piece> _piecesOf(int seed) => _cache.putIfAbsent(seed, () {
    final r = math.Random(seed);
    return [
      for (var i = 0; i < count; i++)
        _Piece(
          // Upward, up to about 50° either side.
          angle: -math.pi / 2 + (r.nextDouble() * 2 - 1) * 0.9,
          speed: 0.9 + r.nextDouble() * 0.8,
          drift: r.nextDouble() * 2 - 1,
          phase: r.nextDouble() * 2 * math.pi,
          spin: (r.nextDouble() * 2 - 1) * 4 * math.pi,
          scale: 0.7 + r.nextDouble() * 0.6,
          round: r.nextDouble() < 0.3,
          color: r.nextInt(1 << 16),
        ),
    ];
  });

  @override
  void paint(Canvas canvas, Size size) {
    if (t <= 0 || t >= 1) return;
    final start = origin.alongSize(size);
    // Distances scale with the box, so the burst fills it on any screen.
    final unit = math.max(size.shortestSide, 160.0);
    const drag = 1.6;
    final gravity = unit * 1.4;
    final s = t * seconds;
    final travel = (1 - math.exp(-drag * s)) / drag;
    final alpha = t < 0.75 ? 1.0 : (1 - t) * 4;
    final paint = Paint();
    for (final p in _pieces) {
      final v = p.speed * unit * 1.1;
      final position =
          start +
          Offset(
            math.cos(p.angle) * v * travel +
                p.drift * unit * 0.05 * math.sin(p.phase + s * 6),
            math.sin(p.angle) * v * travel + 0.5 * gravity * s * s,
          );
      paint.color = colors[p.color % colors.length].withValues(
        alpha: alpha.clamp(0.0, 1.0),
      );
      canvas
        ..save()
        ..translate(position.dx, position.dy)
        ..rotate(p.phase + p.spin * s);
      if (p.round) {
        canvas.drawCircle(Offset.zero, 3.5 * p.scale, paint);
      } else {
        // Flipping: the width follows the spin, as a piece of paper turns.
        final w = 7 * p.scale * (0.35 + 0.65 * math.cos(p.phase + s * 9).abs());
        canvas.drawRect(
          Rect.fromCenter(center: Offset.zero, width: w, height: 11 * p.scale),
          paint,
        );
      }
      canvas.restore();
    }
  }

  @override
  bool shouldRepaint(ConfettiPainter old) =>
      old.t != t ||
      old.seed != seed ||
      old.colors != colors ||
      old.origin != origin ||
      old.seconds != seconds;
}

/// One piece of confetti: its launch [angle] and [speed] (relative), how it
/// drifts and spins, its size, shape and colour index.
class _Piece {
  const _Piece({
    required this.angle,
    required this.speed,
    required this.drift,
    required this.phase,
    required this.spin,
    required this.scale,
    required this.round,
    required this.color,
  });

  final double angle;
  final double speed;
  final double drift;
  final double phase;
  final double spin;
  final double scale;
  final bool round;
  final int color;
}
