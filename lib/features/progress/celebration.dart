import 'dart:math' as math;

import 'package:flutter/material.dart';

/// A short, one-shot celebration (progress spec, "Celebrations"): [icon]
/// grows in and fades out while a few particles in the theme's colours
/// burst from it. It takes [duration], under two seconds, then calls
/// [onEnd]. It ignores pointers, so whatever it covers stays usable, and it
/// may paint outside its box. Callers leave it out when reduce motion is in
/// effect.
class Celebration extends StatelessWidget {
  const Celebration({
    super.key,
    required this.icon,
    this.onEnd,
    this.duration = const Duration(milliseconds: 1600),
  });

  final IconData icon;
  final VoidCallback? onEnd;
  final Duration duration;

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    final particles = [colors.primary, colors.tertiary, colors.secondary];
    return IgnorePointer(
      child: ExcludeSemantics(
        child: TweenAnimationBuilder<double>(
          tween: Tween(begin: 0, end: 1),
          duration: duration,
          onEnd: onEnd,
          builder: (context, t, _) {
            // In over the first quarter, out over the last third.
            final scale = Curves.easeOutBack.transform(math.min(t * 4, 1));
            final opacity = t < 2 / 3 ? 1.0 : (1 - t) * 3;
            return CustomPaint(
              painter: _Particles(t: t, colors: particles),
              // The lines it covers may be lower than the icon.
              child: OverflowBox(
                maxWidth: double.infinity,
                maxHeight: double.infinity,
                child: Opacity(
                  opacity: opacity.clamp(0, 1),
                  child: Transform.scale(
                    scale: 0.6 + 0.6 * scale,
                    child: Icon(icon, size: 48, color: colors.primary),
                  ),
                ),
              ),
            );
          },
        ),
      ),
    );
  }
}

/// Twelve dots flying out from the centre and fading, at time [t] from 0
/// to 1.
class _Particles extends CustomPainter {
  _Particles({required this.t, required this.colors});

  final double t;
  final List<Color> colors;

  static const _count = 12;

  @override
  void paint(Canvas canvas, Size size) {
    final center = size.center(Offset.zero);
    final travel = Curves.easeOutCubic.transform(t) * 80;
    final alpha = (1 - t).clamp(0.0, 1.0);
    final paint = Paint();
    for (var i = 0; i < _count; i++) {
      // Fixed, uneven angles and distances, so the burst looks loose.
      final angle = i * 2 * math.pi / _count + (i.isEven ? 0.2 : -0.1);
      final distance = travel * (0.7 + 0.3 * ((i * 7) % 5) / 4);
      paint.color = colors[i % colors.length].withValues(alpha: alpha);
      canvas.drawCircle(
        center + Offset(math.cos(angle), math.sin(angle)) * distance,
        3 + (i % 3).toDouble(),
        paint,
      );
    }
  }

  @override
  bool shouldRepaint(_Particles old) => old.t != t || old.colors != colors;
}
