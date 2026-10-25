import 'dart:math' as math;

import 'package:flutter/material.dart';

import 'package:cardflip/app/theme.dart';

import 'achievements.dart';

/// An achievement's badge, as in the draft designs in assets/badges/: a
/// scalloped rim in the colour of [tier], a disc in the achievement's
/// colour with its icon, and [stars] under the icon for the tier's number.
/// Without a tier it is locked: grey, with a padlock. [minimalist] leaves
/// out the gloss and the shadow.
class AchievementBadge extends StatelessWidget {
  const AchievementBadge({
    super.key,
    required this.id,
    required this.tier,
    this.stars = 0,
    this.size = 96,
    this.minimalist = false,
  });

  /// The achievement, which picks the disc's colour and icon.
  final String id;

  /// "bronze", "silver", "gold" or "platinum"; null for a locked badge.
  final String? tier;
  final int stars;
  final double size;
  final bool minimalist;

  @override
  Widget build(BuildContext context) {
    final colors = BadgeColors.of(context);
    final look = AchievementLook.of(id);
    final tier = this.tier;
    final locked = tier == null;
    final unit = size / _BadgePainter.extent;
    // The padlock's disc is centred 32 units right of and below the
    // middle.
    final padlockCentre = (_BadgePainter.extent / 2 + 32) * unit;
    return ExcludeSemantics(
      child: SizedBox.square(
        dimension: size,
        child: Stack(
          children: [
            Positioned.fill(
              child: CustomPaint(
                painter: _BadgePainter(
                  rim: locked ? colors.locked : colors.rim(tier),
                  discTop: locked ? colors.lockedDiscTop : look.discTop,
                  discBottom: locked
                      ? colors.lockedDiscBottom
                      : look.discBottom,
                  stars: locked ? 0 : stars,
                  padlock: locked ? colors.padlock : null,
                  gloss: !minimalist,
                  shadow: !minimalist,
                ),
              ),
            ),
            Align(
              alignment: const Alignment(0, -0.04),
              child: Icon(
                look.icon,
                size: 52 * unit,
                color: locked
                    ? Colors.white.withValues(alpha: 0.6)
                    : Colors.white,
              ),
            ),
            if (locked)
              Positioned(
                left: padlockCentre - 7 * unit,
                top: padlockCentre - 7.5 * unit,
                child: Icon(Icons.lock, size: 14 * unit, color: Colors.white),
              ),
          ],
        ),
      ),
    );
  }
}

/// Paints a badge in units of the draft, where the seal is [extent] wide.
class _BadgePainter extends CustomPainter {
  _BadgePainter({
    required this.rim,
    required this.discTop,
    required this.discBottom,
    required this.stars,
    required this.padlock,
    required this.gloss,
    required this.shadow,
  });

  static const extent = 122.0;

  final RimColors rim;
  final Color discTop;
  final Color discBottom;
  final int stars;

  /// The padlock disc's colour on a locked badge; null otherwise.
  final Color? padlock;
  final bool gloss;
  final bool shadow;

  /// The seal: 24 points at radius 61, notched to 56 in between.
  static final _seal = () {
    final path = Path();
    for (var k = 0; k < 48; k++) {
      final angle = -math.pi / 2 + k * math.pi / 24;
      final r = k.isEven ? 61.0 : 56.0;
      final p = Offset(r * math.cos(angle), r * math.sin(angle));
      k == 0 ? path.moveTo(p.dx, p.dy) : path.lineTo(p.dx, p.dy);
    }
    return path..close();
  }();

  /// A five-pointed star centred on [c].
  static Path _star(Offset c) {
    final path = Path();
    for (var k = 0; k < 10; k++) {
      final angle = -math.pi / 2 + k * math.pi / 5;
      final r = k.isEven ? 3.8 : 1.6;
      final p = c + Offset(r * math.cos(angle), r * math.sin(angle));
      k == 0 ? path.moveTo(p.dx, p.dy) : path.lineTo(p.dx, p.dy);
    }
    return path..close();
  }

  static Shader _vertical(double r, Color top, Color bottom) => LinearGradient(
    begin: Alignment.topCenter,
    end: Alignment.bottomCenter,
    colors: [top, bottom],
  ).createShader(Rect.fromCircle(center: Offset.zero, radius: r));

  @override
  void paint(Canvas canvas, Size size) {
    canvas
      ..scale(size.width / extent)
      ..translate(extent / 2, extent / 2);
    if (shadow) canvas.drawShadow(_seal, Colors.black, 2, false);
    canvas
      ..drawPath(_seal, Paint()..color = rim.seal)
      ..drawCircle(
        Offset.zero,
        53,
        Paint()..shader = _vertical(53, rim.top, rim.bottom),
      )
      ..drawCircle(
        Offset.zero,
        43,
        Paint()..shader = _vertical(43, discTop, discBottom),
      )
      ..drawCircle(
        Offset.zero,
        43,
        Paint()
          ..color = rim.seal
          ..style = PaintingStyle.stroke
          ..strokeWidth = 1.5,
      );
    if (gloss) {
      canvas.drawOval(
        Rect.fromCenter(center: const Offset(0, -21), width: 68, height: 36),
        Paint()..color = Colors.white.withValues(alpha: 0.14),
      );
    }
    for (var i = 0; i < stars; i++) {
      final star = _star(Offset((i - (stars - 1) / 2) * 11, 47.6));
      canvas
        ..drawPath(star, Paint()..color = Colors.white)
        ..drawPath(
          star,
          Paint()
            ..color = rim.seal
            ..style = PaintingStyle.stroke
            ..strokeWidth = 0.6,
        );
    }
    if (padlock case final color?) {
      const centre = Offset(32, 32);
      canvas
        ..drawCircle(centre, 13, Paint()..color = color)
        ..drawCircle(
          centre,
          13,
          Paint()
            ..color = Colors.white
            ..style = PaintingStyle.stroke
            ..strokeWidth = 2,
        );
    }
  }

  @override
  bool shouldRepaint(_BadgePainter old) =>
      old.rim != rim ||
      old.discTop != discTop ||
      old.discBottom != discBottom ||
      old.stars != stars ||
      old.padlock != padlock ||
      old.gloss != gloss ||
      old.shadow != shadow;
}
