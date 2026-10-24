import 'dart:convert';
import 'dart:math' as math;
import 'dart:typed_data';

import 'package:flutter/material.dart';

import 'css_color.dart';
import 'media_cache.dart';

/// One shape of an image occlusion note, in coordinates relative to the
/// image (0 to 1), as Go sends it in `<cf-occlusion shapes>` and as the
/// mask editor edits it. Which values matter depends on [kind]: a `rect`
/// has [left], [top], [width] and [height]; an `ellipse` [left] and [top]
/// of its bounding box and the radii [rx] and [ry]; a `polygon` [points];
/// a `text` label [left], [top], [text], [scale] and [fontSize] (relative
/// to the image height). [angle] rotates a shape in degrees around its
/// [left] and [top]. [ord] is the card (cloze) number; labels have 0 and
/// no card. [occludeInactive] is set in "Hide all, guess one" mode.
@immutable
class OcclusionShape {
  const OcclusionShape({
    required this.kind,
    this.ord = 0,
    this.left = 0,
    this.top = 0,
    this.width = 0,
    this.height = 0,
    this.rx = 0,
    this.ry = 0,
    this.angle = 0,
    this.points = const [],
    this.text = '',
    this.scale = 1,
    this.fontSize = 0,
    this.occludeInactive = false,
  });

  final String kind;
  final int ord;
  final double left;
  final double top;
  final double width;
  final double height;
  final double rx;
  final double ry;
  final double angle;
  final List<Offset> points;
  final String text;
  final double scale;
  final double fontSize;
  final bool occludeInactive;

  bool get isLabel => kind == 'text';

  factory OcclusionShape.fromJson(Map<String, dynamic> m) {
    double n(String key) => (m[key] as num?)?.toDouble() ?? 0;
    return OcclusionShape(
      kind: m['kind'] as String? ?? '',
      ord: (m['ord'] as num?)?.toInt() ?? 0,
      left: n('left'),
      top: n('top'),
      width: n('width'),
      height: n('height'),
      rx: n('rx'),
      ry: n('ry'),
      angle: n('angle'),
      points: [
        for (final p in (m['points'] as List<dynamic>?) ?? const [])
          Offset(
            ((p as List<dynamic>)[0] as num).toDouble(),
            (p[1] as num).toDouble(),
          ),
      ],
      text: m['text'] as String? ?? '',
      scale: (m['scale'] as num?)?.toDouble() ?? 1,
      fontSize: n('fontSize'),
      occludeInactive: m['oi'] as bool? ?? false,
    );
  }

  /// The shapes of a `<cf-occlusion shapes>` attribute; none when it does
  /// not parse.
  static List<OcclusionShape> listFromJson(String? json) {
    try {
      return [
        for (final m in jsonDecode(json ?? '[]') as List<dynamic>)
          OcclusionShape.fromJson(m as Map<String, dynamic>),
      ];
    } on Object {
      return const [];
    }
  }

  OcclusionShape copyWith({
    int? ord,
    double? left,
    double? top,
    double? width,
    double? height,
    double? rx,
    double? ry,
    double? angle,
    List<Offset>? points,
    String? text,
    double? scale,
    double? fontSize,
    bool? occludeInactive,
  }) => OcclusionShape(
    kind: kind,
    ord: ord ?? this.ord,
    left: left ?? this.left,
    top: top ?? this.top,
    width: width ?? this.width,
    height: height ?? this.height,
    rx: rx ?? this.rx,
    ry: ry ?? this.ry,
    angle: angle ?? this.angle,
    points: points ?? this.points,
    text: text ?? this.text,
    scale: scale ?? this.scale,
    fontSize: fontSize ?? this.fontSize,
    occludeInactive: occludeInactive ?? this.occludeInactive,
  );

  /// The text size in pixels on an image drawn [imageHeight] high.
  double fontPixels(double imageHeight) =>
      (fontSize > 0 ? fontSize : defaultLabelSize) * imageHeight * scale;

  /// The label size of shapes without one, relative to the image height.
  static const defaultLabelSize = 0.05;

  @override
  bool operator ==(Object other) =>
      other is OcclusionShape &&
      other.kind == kind &&
      other.ord == ord &&
      other.left == left &&
      other.top == top &&
      other.width == width &&
      other.height == height &&
      other.rx == rx &&
      other.ry == ry &&
      other.angle == angle &&
      _samePoints(other.points, points) &&
      other.text == text &&
      other.scale == scale &&
      other.fontSize == fontSize &&
      other.occludeInactive == occludeInactive;

  static bool _samePoints(List<Offset> a, List<Offset> b) {
    if (a.length != b.length) return false;
    for (var i = 0; i < a.length; i++) {
      if (a[i] != b[i]) return false;
    }
    return true;
  }

  @override
  int get hashCode => Object.hash(
    kind,
    ord,
    left,
    top,
    width,
    height,
    rx,
    ry,
    angle,
    Object.hashAll(points),
    text,
    scale,
    fontSize,
    occludeInactive,
  );

  @override
  String toString() => 'OcclusionShape($kind c$ord)';
}

/// A shape's border: a width in pixels and a colour.
typedef OcclusionBorder = ({double width, Color color});

/// The colours shapes are drawn in. Note types set them with the CSS
/// variables Anki uses on its occlusion canvas, which Go inlines on the
/// `<cf-occlusion>` element; the defaults are Anki's.
@immutable
class OcclusionColors {
  const OcclusionColors({
    this.inactive = const Color(0xFFFFEBA2),
    this.active = const Color(0xFFFF8E8E),
    this.highlight = const Color(0x00FF8E8E),
    this.inactiveBorder = const (width: 1, color: Color(0xFF212121)),
    this.activeBorder = const (width: 1, color: Color(0xFF212121)),
    this.highlightBorder = const (width: 1, color: Color(0xFFFF8E8E)),
  });

  final Color inactive;
  final Color active;
  final Color highlight;
  final OcclusionBorder inactiveBorder;
  final OcclusionBorder activeBorder;
  final OcclusionBorder highlightBorder;

  /// The colours a style attribute sets with `--inactive-shape-color`,
  /// `--active-shape-border` and the like, and the defaults for the rest.
  factory OcclusionColors.fromStyle(String? style) {
    final vars = <String, String>{};
    for (final decl in (style ?? '').split(';')) {
      final i = decl.indexOf(':');
      if (i > 0) vars[decl.substring(0, i).trim()] = decl.substring(i + 1);
    }
    const d = OcclusionColors();
    Color color(String name, Color fallback) =>
        parseCssColor(vars['--$name-shape-color']) ?? fallback;
    OcclusionBorder border(String name, OcclusionBorder fallback) =>
        _border(vars['--$name-shape-border']) ?? fallback;
    return OcclusionColors(
      inactive: color('inactive', d.inactive),
      active: color('active', d.active),
      highlight: color('highlight', d.highlight),
      inactiveBorder: border('inactive', d.inactiveBorder),
      activeBorder: border('active', d.activeBorder),
      highlightBorder: border('highlight', d.highlightBorder),
    );
  }

  /// A border such as `1px #212121` or `2px solid red`.
  static OcclusionBorder? _border(String? value) {
    if (value == null) return null;
    double? width;
    Color? color;
    for (final part in value.trim().split(RegExp(r'\s+'))) {
      final px = RegExp(r'^([\d.]+)(px)?$').firstMatch(part.toLowerCase());
      if (px != null) {
        width = double.tryParse(px[1]!);
      } else {
        color = parseCssColor(part) ?? color;
      }
    }
    if (width == null && color == null) return null;
    return (width: width ?? 1, color: color ?? const Color(0xFF212121));
  }

  @override
  bool operator ==(Object other) =>
      other is OcclusionColors &&
      other.inactive == inactive &&
      other.active == active &&
      other.highlight == highlight &&
      other.inactiveBorder == inactiveBorder &&
      other.activeBorder == activeBorder &&
      other.highlightBorder == highlightBorder;

  @override
  int get hashCode => Object.hash(
    inactive,
    active,
    highlight,
    inactiveBorder,
    activeBorder,
    highlightBorder,
  );
}

/// How a shape is drawn on a card.
enum OcclusionLook {
  /// Not drawn: a shape not being asked, in "Hide one, guess one" mode.
  hidden,

  /// Masked in the inactive colour: a shape not being asked, in "Hide all,
  /// guess one" mode.
  inactive,

  /// Masked in the active colour: the asked shape on the question.
  active,

  /// Revealed and outlined: the asked shape on the answer.
  highlight,

  /// A text label, shown on every card.
  label,
}

/// How [shape] is drawn on the card for cloze number [active], on the
/// question or the answer.
OcclusionLook occlusionLook(OcclusionShape shape, int active, bool question) {
  if (shape.isLabel) return OcclusionLook.label;
  if (shape.ord == active) {
    return question ? OcclusionLook.active : OcclusionLook.highlight;
  }
  return shape.occludeInactive ? OcclusionLook.inactive : OcclusionLook.hidden;
}

/// Paints occlusion shapes over an image filling the canvas, each in the
/// colours of its [look]. Labels go on top.
class OcclusionPainter extends CustomPainter {
  const OcclusionPainter({
    required this.shapes,
    required this.look,
    this.colors = const OcclusionColors(),
  });

  final List<OcclusionShape> shapes;
  final OcclusionLook Function(OcclusionShape shape) look;
  final OcclusionColors colors;

  @override
  void paint(Canvas canvas, Size size) {
    for (final s in [
      ...shapes.where((s) => !s.isLabel),
      ...shapes.where((s) => s.isLabel),
    ]) {
      final l = look(s);
      if (l == OcclusionLook.hidden) continue;
      paintShape(canvas, size, s, l, colors);
    }
  }

  /// Paints one shape on an image of [size]: moved to its left and top and
  /// rotated there, so a rectangle is drawn from the origin.
  static void paintShape(
    Canvas canvas,
    Size size,
    OcclusionShape s,
    OcclusionLook look,
    OcclusionColors colors,
  ) {
    final (fill, border) = switch (look) {
      OcclusionLook.active => (colors.active, colors.activeBorder),
      OcclusionLook.highlight => (colors.highlight, colors.highlightBorder),
      _ => (colors.inactive, colors.inactiveBorder),
    };
    final fillPaint = Paint()..color = fill;
    final borderPaint = Paint()
      ..style = PaintingStyle.stroke
      ..strokeWidth = border.width
      ..color = border.color;
    canvas.save();
    if (s.kind == 'polygon') {
      final path = Path()
        ..addPolygon([
          for (final p in s.points)
            Offset(p.dx * size.width, p.dy * size.height),
        ], true);
      canvas
        ..drawPath(path, fillPaint)
        ..drawPath(path, borderPaint);
      canvas.restore();
      return;
    }
    canvas.translate(s.left * size.width, s.top * size.height);
    if (s.angle != 0) canvas.rotate(s.angle * math.pi / 180);
    switch (s.kind) {
      case 'rect':
        final r = Rect.fromLTWH(
          0,
          0,
          s.width * size.width,
          s.height * size.height,
        );
        canvas
          ..drawRect(r, fillPaint)
          ..drawRect(r, borderPaint);
      case 'ellipse':
        final r = Rect.fromLTWH(
          0,
          0,
          2 * s.rx * size.width,
          2 * s.ry * size.height,
        );
        canvas
          ..drawOval(r, fillPaint)
          ..drawOval(r, borderPaint);
      case 'text':
        _paintLabel(canvas, size, s);
    }
    canvas.restore();
  }

  /// Draws a label as Anki does: black text on a white box.
  static void _paintLabel(Canvas canvas, Size size, OcclusionShape s) {
    final painter = TextPainter(
      text: TextSpan(
        text: s.text,
        style: TextStyle(
          color: Colors.black,
          fontSize: math.max(1, s.fontPixels(size.height)),
          fontFamily: 'Arial',
        ),
      ),
      textDirection: TextDirection.ltr,
    )..layout();
    const pad = 2.0;
    canvas.drawRect(
      Rect.fromLTWH(
        -pad,
        -pad,
        painter.width + 2 * pad,
        painter.height + 2 * pad,
      ),
      Paint()..color = const Color(0xFFFFFFFF),
    );
    painter.paint(canvas, Offset.zero);
    painter.dispose();
  }

  @override
  bool shouldRepaint(OcclusionPainter old) =>
      old.shapes != shapes || old.look != look || old.colors != colors;
}

/// An image occlusion card side, for `<cf-occlusion>`: the note's [image]
/// from the media at its own size (never wider than the card), with the
/// [shapes] drawn over it for card [active] on the [question] or the
/// answer. Without its image only a placeholder, [missing], shows.
class OcclusionView extends StatefulWidget {
  const OcclusionView({
    super.key,
    required this.image,
    required this.shapes,
    required this.active,
    required this.question,
    required this.cache,
    required this.missing,
    this.colors = const OcclusionColors(),
  });

  final String image;
  final List<OcclusionShape> shapes;
  final int active;
  final bool question;
  final MediaCache cache;
  final Widget missing;
  final OcclusionColors colors;

  @override
  State<OcclusionView> createState() => _OcclusionViewState();
}

class _OcclusionViewState extends State<OcclusionView> {
  late Future<Uint8List> _bytes = _load();

  Future<Uint8List> _load() => loadMediaBytes(widget.cache, widget.image);

  @override
  void didUpdateWidget(OcclusionView old) {
    super.didUpdateWidget(old);
    if (old.image != widget.image || old.cache != widget.cache) {
      _bytes = _load();
    }
  }

  @override
  Widget build(BuildContext context) {
    return Center(
      child: FutureBuilder<Uint8List>(
        future: _bytes,
        builder: (context, snap) {
          if (snap.hasError || widget.image.isEmpty) return widget.missing;
          if (!snap.hasData) return const SizedBox(width: 48, height: 48);
          return Image.memory(
            snap.data!,
            key: const Key('occlusion-image'),
            fit: BoxFit.contain,
            gaplessPlayback: true,
            errorBuilder: (_, _, _) => widget.missing,
            // The masks go on once the image is there, so its area is
            // never shown unmasked or the masks without it.
            frameBuilder: (context, image, frame, _) {
              if (frame == null) return image;
              return Stack(
                children: [
                  image,
                  Positioned.fill(
                    child: CustomPaint(
                      key: const Key('occlusion-masks'),
                      painter: OcclusionPainter(
                        shapes: widget.shapes,
                        look: (s) =>
                            occlusionLook(s, widget.active, widget.question),
                        colors: widget.colors,
                      ),
                    ),
                  ),
                ],
              );
            },
          );
        },
      ),
    );
  }
}
