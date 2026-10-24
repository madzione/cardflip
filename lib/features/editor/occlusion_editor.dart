import 'dart:async';
import 'dart:math' as math;
import 'dart:ui' as ui;

import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:html/parser.dart' as html_parser;

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/widgets/occlusion_view.dart';
import 'package:cardflip/l10n/l10n.dart';

// The Occlusion field holds the shapes as cloze deletions in Anki's syntax,
// coordinates relative to the image:
//
//   {{c1::image-occlusion:rect:left=.1:top=.2:width=.3:height=.1:oi=1}}
//
// Shapes sharing a number are asked together; text labels are c0. oi=1 is
// "Hide all, guess one". A colon in a value is escaped as "\:".

final _deletion = RegExp(
  r'\{\{c(\d+)::image-occlusion:(.*?)\}\}',
  dotAll: true,
);

/// The shapes of an Occlusion field, in their order. Unknown shape kinds
/// and properties are ignored, and so are shapes without a usable size.
List<OcclusionShape> parseOcclusionField(String field) {
  final out = <OcclusionShape>[];
  for (final m in _deletion.allMatches(field)) {
    final ord = int.parse(m[1]!);
    final parts = _splitUnescaped(m[2]!);
    final props = <String, String>{};
    for (final p in parts.skip(1)) {
      final eq = p.indexOf('=');
      if (eq > 0) props[p.substring(0, eq).trim()] = p.substring(eq + 1);
    }
    double n(String key) => double.tryParse(props[key]?.trim() ?? '') ?? 0;
    final kind = parts.first.trim();
    final shape = OcclusionShape(
      kind: kind,
      ord: ord,
      left: n('left'),
      top: n('top'),
      width: n('width'),
      height: n('height'),
      rx: n('rx'),
      ry: n('ry'),
      angle: n('angle'),
      points: _points(props['points'] ?? ''),
      text: props.containsKey('text')
          ? html_parser.parseFragment(props['text']!).text ?? ''
          : '',
      scale: props.containsKey('scale') ? n('scale') : 1,
      fontSize: n('fs'),
      occludeInactive: props['oi']?.trim() == '1',
    );
    final ok = switch (kind) {
      'rect' => shape.width > 0 && shape.height > 0,
      'ellipse' => shape.rx > 0 && shape.ry > 0,
      'polygon' => shape.points.length >= 3,
      'text' => shape.text.trim().isNotEmpty,
      _ => false,
    };
    if (ok) out.add(shape);
  }
  return out;
}

List<String> _splitUnescaped(String s) {
  final parts = <String>[];
  final cur = StringBuffer();
  for (var i = 0; i < s.length; i++) {
    if (s[i] == r'\' && i + 1 < s.length && s[i + 1] == ':') {
      cur.write(':');
      i++;
    } else if (s[i] == ':') {
      parts.add(cur.toString());
      cur.clear();
    } else {
      cur.write(s[i]);
    }
  }
  return parts..add(cur.toString());
}

List<Offset> _points(String value) {
  final out = <Offset>[];
  for (final pair in value.trim().split(RegExp(r'\s+'))) {
    if (pair.isEmpty) continue;
    final xy = pair.split(',');
    final x = xy.length == 2 ? double.tryParse(xy[0]) : null;
    final y = xy.length == 2 ? double.tryParse(xy[1]) : null;
    if (x == null || y == null) return const [];
    out.add(Offset(x, y));
  }
  return out;
}

/// The Occlusion field for [shapes], with their numbers as they are, in
/// "Hide all, guess one" mode when [hideAll].
String formatOcclusionField(
  List<OcclusionShape> shapes, {
  required bool hideAll,
}) {
  return [
    for (final s in shapes)
      '{{c${s.isLabel ? 0 : s.ord}::${_shapeText(s, hideAll)}}}',
  ].join('<br>');
}

String _shapeText(OcclusionShape s, bool hideAll) {
  final props = <String>['image-occlusion', s.kind];
  void add(String key, String value) =>
      props.add('$key=${value.replaceAll(':', r'\:')}');
  void num(String key, double v) => add(key, _number(v));
  switch (s.kind) {
    case 'rect':
      num('left', s.left);
      num('top', s.top);
      num('width', s.width);
      num('height', s.height);
    case 'ellipse':
      num('left', s.left);
      num('top', s.top);
      num('rx', s.rx);
      num('ry', s.ry);
    case 'polygon':
      add(
        'points',
        [for (final p in s.points) '${_number(p.dx)},${_number(p.dy)}']
            .join(' '),
      );
    case 'text':
      num('left', s.left);
      num('top', s.top);
      add('text', _escapeText(s.text));
      num('scale', s.scale);
      num('fs', s.fontSize > 0 ? s.fontSize : OcclusionShape.defaultLabelSize);
  }
  if (s.angle != 0) num('angle', s.angle);
  if (hideAll) add('oi', '1');
  return props.join(':');
}

/// A label's text as field HTML; "}" would end the deletion.
String _escapeText(String text) => text
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('}', '&#125;');

/// A coordinate as Anki writes it: at most four decimals, without a
/// leading zero (".25").
String _number(double v) {
  var s = v.toStringAsFixed(4);
  if (s.contains('.')) {
    s = s.replaceFirst(RegExp(r'0+$'), '').replaceFirst(RegExp(r'\.$'), '');
  }
  if (s == '-0') s = '0';
  if (s.startsWith('0.')) return s.substring(1);
  if (s.startsWith('-0.')) return '-${s.substring(2)}';
  return s;
}

/// The tools of the mask editor.
enum OcclusionTool { select, rect, ellipse, polygon, text }

/// The shapes of one Image Occlusion note being edited, in creation order,
/// and the mode. A shape's [OcclusionShape.ord] is its group: shapes with
/// the same one are asked together, on one card. A new note numbers its
/// groups 1 to N in creation order when saved; an existing note
/// ([keepNumbers]) keeps the numbers its shapes have, so each card stays
/// with its shape, and new groups get numbers that neither its shapes nor
/// its [cardNumbers] have: a removed shape leaves an empty card, whose
/// scheduling a new shape must not take over.
class OcclusionController extends ChangeNotifier {
  OcclusionController({
    String field = '',
    this.keepNumbers = false,
    this.cardNumbers = const [],
  }) {
    load(field);
  }

  final bool keepNumbers;

  /// The cloze numbers of the note's cards, when editing one.
  final List<int> cardNumbers;
  List<OcclusionShape> _shapes = [];
  bool _hideAll = true;
  Set<int> _selected = {};
  int _nextGroup = 1;

  List<OcclusionShape> get shapes => List.unmodifiable(_shapes);

  /// "Hide all, guess one"; false is "Hide one, guess one".
  bool get hideAll => _hideAll;
  set hideAll(bool value) {
    if (value == _hideAll) return;
    _hideAll = value;
    notifyListeners();
  }

  /// The indexes of the selected shapes.
  Set<int> get selected => Set.unmodifiable(_selected);

  /// Replaces the shapes with those of an Occlusion field. A field without
  /// shapes starts in "Hide all, guess one" mode.
  void load(String field) {
    _shapes = parseOcclusionField(field);
    final masks = _shapes.where((s) => !s.isLabel);
    _hideAll = masks.isEmpty || masks.any((s) => s.occludeInactive);
    _selected = {};
    _nextGroup =
        [...masks.map((s) => s.ord), ...cardNumbers].fold(0, math.max) + 1;
    notifyListeners();
  }

  /// The Occlusion field for the shapes.
  String get field {
    final numbers = this.numbers;
    return formatOcclusionField([
      for (final s in _shapes) s.isLabel ? s : s.copyWith(ord: numbers[s.ord]),
    ], hideAll: _hideAll);
  }

  /// The card number each group is saved with, by group.
  Map<int, int> get numbers {
    final out = <int, int>{};
    for (final s in _shapes) {
      if (s.isLabel || out.containsKey(s.ord)) continue;
      out[s.ord] = keepNumbers ? s.ord : out.length + 1;
    }
    return out;
  }

  /// How many cards the shapes make.
  int get cardCount => numbers.length;

  /// Adds [shape] as a group of its own (labels have none) and selects it.
  void add(OcclusionShape shape) {
    _shapes = [
      ..._shapes,
      shape.copyWith(ord: shape.isLabel ? 0 : _nextGroup++),
    ];
    _selected = {_shapes.length - 1};
    notifyListeners();
  }

  /// Replaces the shape at [index], keeping its group.
  void replace(int index, OcclusionShape shape) {
    _shapes = [..._shapes]..[index] = shape.copyWith(ord: _shapes[index].ord);
    notifyListeners();
  }

  /// Selects [index] alone, or adds it to or removes it from the selection
  /// when [toggle]; null clears the selection.
  void select(int? index, {bool toggle = false}) {
    if (index == null) {
      _selected = {};
    } else if (toggle) {
      _selected = {..._selected};
      if (!_selected.remove(index)) _selected.add(index);
    } else {
      _selected = {index};
    }
    notifyListeners();
  }

  void deleteSelected() {
    if (_selected.isEmpty) return;
    _shapes = [
      for (final (i, s) in _shapes.indexed)
        if (!_selected.contains(i)) s,
    ];
    _selected = {};
    notifyListeners();
  }

  /// Whether the selection has shapes of more than one group to join.
  bool get canGroup =>
      _selected
          .map((i) => _shapes[i])
          .where((s) => !s.isLabel)
          .map((s) => s.ord)
          .toSet()
          .length >
      1;

  /// Whether the selection has a shape that shares its group.
  bool get canUngroup => _selected.any((i) {
    final s = _shapes[i];
    return !s.isLabel && _shapes.where((o) => o.ord == s.ord).length > 1;
  });

  /// Joins the groups of the selected shapes into the first one's: they
  /// are asked together, on one card.
  void groupSelected() {
    if (!canGroup) return;
    final indexes = _selected.toList()..sort();
    final groups = {
      for (final i in indexes)
        if (!_shapes[i].isLabel) _shapes[i].ord,
    };
    final target = groups.first;
    _shapes = [
      for (final s in _shapes)
        groups.contains(s.ord) && !s.isLabel ? s.copyWith(ord: target) : s,
    ];
    notifyListeners();
  }

  /// Gives each selected shape that shares its group a group of its own;
  /// the first shape of a group keeps it.
  void ungroupSelected() {
    if (!canUngroup) return;
    final firstOf = <int, int>{};
    for (final (i, s) in _shapes.indexed) {
      if (!s.isLabel) firstOf.putIfAbsent(s.ord, () => i);
    }
    _shapes = [
      for (final (i, s) in _shapes.indexed)
        _selected.contains(i) && !s.isLabel && firstOf[s.ord] != i
            ? s.copyWith(ord: _nextGroup++)
            : s,
    ];
    notifyListeners();
  }
}

/// The geometry of a shape on the canvas, in pixels of the image as shown
/// unzoomed: rotated by [angle] (radians) around [origin], it covers
/// [local].
typedef _Frame = ({Offset origin, double angle, Rect local});

Offset _rotate(Offset p, double angle) {
  if (angle == 0) return p;
  final c = math.cos(angle), s = math.sin(angle);
  return Offset(p.dx * c - p.dy * s, p.dx * s + p.dy * c);
}

/// The size of a label's text on an image shown [size] large.
Size _labelSize(OcclusionShape s, Size size) {
  final painter = TextPainter(
    text: TextSpan(
      text: s.text,
      style: TextStyle(
        fontSize: math.max(1, s.fontPixels(size.height)),
        fontFamily: 'Arial',
      ),
    ),
    textDirection: TextDirection.ltr,
  )..layout();
  final out = painter.size;
  painter.dispose();
  return out;
}

_Frame _frameOf(OcclusionShape s, Size size) {
  if (s.kind == 'polygon') {
    final pts = [
      for (final p in s.points) Offset(p.dx * size.width, p.dy * size.height),
    ];
    var r = Rect.fromPoints(pts.first, pts.first);
    for (final p in pts) {
      r = r.expandToInclude(Rect.fromPoints(p, p));
    }
    return (origin: Offset.zero, angle: 0, local: r);
  }
  final origin = Offset(s.left * size.width, s.top * size.height);
  final angle = s.angle * math.pi / 180;
  final Size extent = switch (s.kind) {
    'rect' => Size(s.width * size.width, s.height * size.height),
    'ellipse' => Size(2 * s.rx * size.width, 2 * s.ry * size.height),
    _ => _labelSize(s, size),
  };
  return (origin: origin, angle: angle, local: Offset.zero & extent);
}

Offset _toLocal(_Frame f, Offset p) => _rotate(p - f.origin, -f.angle);
Offset _toCanvas(_Frame f, Offset p) => f.origin + _rotate(p, f.angle);

bool _hits(OcclusionShape s, Size size, Offset p) {
  final f = _frameOf(s, size);
  final local = _toLocal(f, p);
  switch (s.kind) {
    case 'polygon':
      final path = Path()
        ..addPolygon([
          for (final q in s.points)
            Offset(q.dx * size.width, q.dy * size.height),
        ], true);
      return path.contains(p);
    case 'ellipse':
      final r = f.local;
      final dx = (local.dx - r.center.dx) / (r.width / 2);
      final dy = (local.dy - r.center.dy) / (r.height / 2);
      return dx * dx + dy * dy <= 1;
    default:
      return f.local.inflate(2).contains(local);
  }
}

/// The corners of a shape's frame on the canvas: top left, top right,
/// bottom right, bottom left.
List<Offset> _corners(_Frame f) => [
  _toCanvas(f, f.local.topLeft),
  _toCanvas(f, f.local.topRight),
  _toCanvas(f, f.local.bottomRight),
  _toCanvas(f, f.local.bottomLeft),
];

/// [s] moved by [delta] canvas pixels on an image shown [size] large.
OcclusionShape _movedBy(OcclusionShape s, Size size, Offset delta) {
  final d = Offset(delta.dx / size.width, delta.dy / size.height);
  if (s.kind == 'polygon') {
    return s.copyWith(points: [for (final p in s.points) p + d]);
  }
  return s.copyWith(left: s.left + d.dx, top: s.top + d.dy);
}

/// [s] with the frame corner [corner] dragged to [to], the opposite corner
/// staying where it is.
OcclusionShape _resized(OcclusionShape s, Size size, int corner, Offset to) {
  const minSize = 4.0;
  final f = _frameOf(s, size);
  final old = f.local;
  final fixed = [
    old.bottomRight,
    old.bottomLeft,
    old.topLeft,
    old.topRight,
  ][corner];
  var pointer = _toLocal(f, to);
  // At least minSize, on the side the corner started on.
  final signX = corner == 1 || corner == 2 ? 1.0 : -1.0;
  final signY = corner >= 2 ? 1.0 : -1.0;
  pointer = Offset(
    signX > 0
        ? math.max(pointer.dx, fixed.dx + minSize)
        : math.min(pointer.dx, fixed.dx - minSize),
    signY > 0
        ? math.max(pointer.dy, fixed.dy + minSize)
        : math.min(pointer.dy, fixed.dy - minSize),
  );
  final r = Rect.fromPoints(fixed, pointer);
  if (s.kind == 'polygon') {
    double sx(double x) =>
        old.width == 0 ? r.left : r.left + (x - old.left) * r.width / old.width;
    double sy(double y) =>
        old.height == 0 ? r.top : r.top + (y - old.top) * r.height / old.height;
    return s.copyWith(
      points: [
        for (final p in s.points)
          Offset(
            sx(p.dx * size.width) / size.width,
            sy(p.dy * size.height) / size.height,
          ),
      ],
    );
  }
  final origin = _toCanvas(f, r.topLeft);
  final moved = s.copyWith(
    left: origin.dx / size.width,
    top: origin.dy / size.height,
  );
  return switch (s.kind) {
    'rect' => moved.copyWith(
      width: r.width / size.width,
      height: r.height / size.height,
    ),
    'ellipse' => moved.copyWith(
      rx: r.width / 2 / size.width,
      ry: r.height / 2 / size.height,
    ),
    // A label keeps its proportions: its height sets the text size.
    _ => moved.copyWith(scale: s.scale * r.height / old.height),
  };
}

/// The mask editor of an Image Occlusion note: the [image] with the
/// [controller]'s shapes over it, and tools to draw rectangles, ellipses,
/// polygons and text labels, to select, move, resize, delete, group and
/// ungroup shapes, to choose the mode and to zoom. It works with the mouse
/// (Shift or Ctrl and click adds to the selection, Ctrl and the wheel
/// zooms, Delete removes, Enter finishes a polygon, Escape cancels) and with
/// touch (a long press adds to the selection, two fingers zoom and move the
/// image). Large images are decoded at most [previewWidth] pixels wide.
class OcclusionEditor extends StatefulWidget {
  const OcclusionEditor({
    super.key,
    required this.controller,
    required this.image,
    this.previewWidth = 2048,
    this.maxHeight = 480,
  });

  final OcclusionController controller;
  final Uint8List image;
  final int previewWidth;

  /// The canvas is at most this high, unzoomed; the image keeps its
  /// proportions.
  final double maxHeight;

  @override
  State<OcclusionEditor> createState() => _OcclusionEditorState();
}

// Kept alive in a scrolling list, so scrolling to the other fields keeps
// the tool, the zoom and the decoded image.
class _OcclusionEditorState extends State<OcclusionEditor>
    with AutomaticKeepAliveClientMixin {
  @override
  bool get wantKeepAlive => true;

  static const _maxZoom = 8.0;

  /// Half the side of a handle's touch target, in screen pixels.
  static const handleReach = 16.0;

  ui.Image? _image;
  Object? _imageError;
  var _tool = OcclusionTool.select;
  final _focus = FocusNode(debugLabel: 'occlusion canvas');

  double _zoom = 1;
  Offset _pan = Offset.zero;

  // The canvas size of the current layout, unzoomed.
  Size _size = Size.zero;

  // The gesture in progress.
  _Drag? _drag;
  Offset _downAt = Offset.zero;
  bool _moved = false;
  Timer? _longPress;
  bool _longPressed = false;

  // The polygon being drawn, in canvas pixels.
  List<Offset> _polygon = [];
  Offset? _hover;

  OcclusionController get _c => widget.controller;

  @override
  void initState() {
    super.initState();
    _c.addListener(_changed);
    _decode();
  }

  @override
  void didUpdateWidget(OcclusionEditor old) {
    super.didUpdateWidget(old);
    if (old.controller != widget.controller) {
      old.controller.removeListener(_changed);
      _c.addListener(_changed);
    }
    if (old.image != widget.image) _decode();
  }

  @override
  void dispose() {
    _c.removeListener(_changed);
    _longPress?.cancel();
    _focus.dispose();
    _image?.dispose();
    super.dispose();
  }

  void _changed() => setState(() {});

  Future<void> _decode() async {
    final bytes = widget.image;
    try {
      final buffer = await ui.ImmutableBuffer.fromUint8List(bytes);
      final descriptor = await ui.ImageDescriptor.encoded(buffer);
      final width = descriptor.width > widget.previewWidth
          ? widget.previewWidth
          : null;
      final codec = await descriptor.instantiateCodec(targetWidth: width);
      final frame = await codec.getNextFrame();
      codec.dispose();
      descriptor.dispose();
      buffer.dispose();
      if (!mounted || bytes != widget.image) {
        frame.image.dispose();
        return;
      }
      setState(() {
        _image?.dispose();
        _image = frame.image;
        _imageError = null;
        _zoom = 1;
        _pan = Offset.zero;
      });
    } catch (e) {
      if (mounted && bytes == widget.image) setState(() => _imageError = e);
    }
  }

  /// A screen point on the canvas in canvas pixels.
  Offset _canvasPoint(Offset p) => (p - _pan) / _zoom;

  bool get _multiKey {
    final k = HardwareKeyboard.instance;
    return k.isShiftPressed || k.isControlPressed || k.isMetaPressed;
  }

  int? _shapeAt(Offset canvas) {
    final shapes = _c.shapes;
    // Topmost first: labels are drawn last.
    final order = [
      for (final (i, s) in shapes.indexed)
        if (s.isLabel) i,
      for (final (i, s) in shapes.indexed)
        if (!s.isLabel) i,
    ].reversed;
    for (final i in order) {
      if (_hits(shapes[i], _size, canvas)) return i;
    }
    return null;
  }

  /// The resize handle of the one selected shape under screen point [p].
  int? _handleAt(Offset p) {
    if (_c.selected.length != 1) return null;
    final s = _c.shapes[_c.selected.single];
    final corners = _corners(_frameOf(s, _size));
    for (final (i, c) in corners.indexed) {
      final screen = c * _zoom + _pan;
      if ((screen - p).distance <= handleReach) return i;
    }
    return null;
  }

  void _onStart(ScaleStartDetails d) {
    _focus.requestFocus();
    if (d.pointerCount >= 2) {
      _cancelDrag();
      _drag = _Zoom(_zoom, _pan, d.localFocalPoint);
      return;
    }
    _downAt = d.localFocalPoint;
    _moved = false;
    _longPressed = false;
    _longPress?.cancel();
    _longPress = Timer(kLongPressTimeout, () {
      if (_moved || _drag is _Zoom) return;
      _longPressed = true;
      _cancelDrag();
      final i = _shapeAt(_canvasPoint(_downAt));
      if (i != null) {
        HapticFeedback.selectionClick();
        _c.select(i, toggle: true);
      }
    });
    final p = _canvasPoint(d.localFocalPoint);
    switch (_tool) {
      case OcclusionTool.select:
        final handle = _handleAt(d.localFocalPoint);
        if (handle != null) {
          final i = _c.selected.single;
          _drag = _Resize(i, handle, _c.shapes[i]);
          return;
        }
        final i = _shapeAt(p);
        if (i == null) {
          _drag = null;
          return;
        }
        // The selection changes once it is clear what the press is: a tap,
        // a long press or a move.
        _drag = _Press(i, p);
      case OcclusionTool.rect || OcclusionTool.ellipse:
        _drag = _Draw(p, p);
      case OcclusionTool.polygon || OcclusionTool.text:
        _drag = null;
    }
  }

  void _onUpdate(ScaleUpdateDetails d) {
    if (d.pointerCount >= 2 && _drag is! _Zoom) {
      _cancelDrag();
      _drag = _Zoom(_zoom, _pan, d.localFocalPoint, d.scale);
    }
    final drag = _drag;
    if (drag is _Zoom) {
      final zoom = (drag.zoom * d.scale / drag.scaleAtStart).clamp(
        1.0,
        _maxZoom,
      );
      final content = (drag.focal - drag.pan) / drag.zoom;
      setState(() {
        _zoom = zoom;
        _pan = _clampPan(d.localFocalPoint - content * zoom, zoom);
      });
      return;
    }
    if ((d.localFocalPoint - _downAt).distance > kTouchSlop) {
      _moved = true;
      _longPress?.cancel();
    }
    if (!_moved || _longPressed) return;
    final p = _canvasPoint(d.localFocalPoint);
    switch (drag) {
      case _Press(:final index, :final start):
        // Moving an unselected shape moves it (with the selection when
        // adding to it).
        if (!_c.selected.contains(index)) _c.select(index, toggle: _multiKey);
        final originals = {for (final j in _c.selected) j: _c.shapes[j]};
        _drag = _Move(start, originals);
        for (final MapEntry(key: i, value: s) in originals.entries) {
          _c.replace(i, _movedBy(s, _size, p - start));
        }
      case _Move(:final start, :final originals):
        for (final MapEntry(key: i, value: s) in originals.entries) {
          _c.replace(i, _movedBy(s, _size, p - start));
        }
      case _Resize(:final index, :final corner, :final original):
        _c.replace(index, _resized(original, _size, corner, p));
      case _Draw(:final start):
        setState(() => _drag = _Draw(start, p));
      case _Zoom() || null:
    }
  }

  void _onEnd(ScaleEndDetails d) {
    _longPress?.cancel();
    final drag = _drag;
    if (d.pointerCount > 0) {
      // A finger was added or lifted while others stay: no tap, and a
      // drawing in progress is dropped.
      if (drag is! _Zoom) _cancelDrag();
      return;
    }
    _drag = null;
    if (drag is _Zoom) {
      setState(() {});
      return;
    }
    if (_longPressed) return;
    if (!_moved) {
      _tap(_canvasPoint(_downAt));
      return;
    }
    if (drag case _Draw(:final start, :final end)) {
      final r = Rect.fromPoints(start, end);
      setState(() {});
      if (r.width < 4 || r.height < 4) return;
      final w = _size.width, h = _size.height;
      _c.add(switch (_tool) {
        OcclusionTool.ellipse => OcclusionShape(
          kind: 'ellipse',
          left: r.left / w,
          top: r.top / h,
          rx: r.width / 2 / w,
          ry: r.height / 2 / h,
        ),
        _ => OcclusionShape(
          kind: 'rect',
          left: r.left / w,
          top: r.top / h,
          width: r.width / w,
          height: r.height / h,
        ),
      });
    }
  }

  /// Puts back the shapes a move or resize in progress changed.
  void _cancelDrag() {
    switch (_drag) {
      case _Move(:final originals):
        for (final MapEntry(key: i, value: s) in originals.entries) {
          _c.replace(i, s);
        }
      case _Resize(:final index, :final original):
        _c.replace(index, original);
      default:
    }
    setState(() => _drag = null);
  }

  Future<void> _tap(Offset p) async {
    switch (_tool) {
      case OcclusionTool.select:
        _c.select(_shapeAt(p), toggle: _multiKey && _shapeAt(p) != null);
      case OcclusionTool.polygon:
        final close =
            _polygon.length >= 3 &&
            (_polygon.first - p).distance * _zoom <= handleReach;
        if (close) {
          _finishPolygon();
        } else {
          setState(() => _polygon = [..._polygon, p]);
        }
      case OcclusionTool.text:
        final text = await _askLabel(context);
        if (text == null || text.trim().isEmpty || !mounted) return;
        _c.add(
          OcclusionShape(
            kind: 'text',
            left: p.dx / _size.width,
            top: p.dy / _size.height,
            text: text.trim(),
            fontSize: OcclusionShape.defaultLabelSize,
          ),
        );
      case OcclusionTool.rect || OcclusionTool.ellipse:
        _c.select(null);
    }
  }

  void _finishPolygon() {
    if (_polygon.length < 3) return;
    final pts = [
      for (final p in _polygon) Offset(p.dx / _size.width, p.dy / _size.height),
    ];
    setState(() => _polygon = []);
    _c.add(OcclusionShape(kind: 'polygon', points: pts));
  }

  void _setTool(OcclusionTool tool) {
    setState(() {
      _tool = tool;
      _polygon = [];
    });
  }

  Offset _clampPan(Offset pan, double zoom) {
    // The zoomed image keeps covering the canvas.
    final minX = _size.width - _size.width * zoom;
    final minY = _size.height - _size.height * zoom;
    return Offset(pan.dx.clamp(minX, 0), pan.dy.clamp(minY, 0));
  }

  void _zoomBy(double factor, [Offset? focal]) {
    final f = focal ?? _size.center(Offset.zero);
    final zoom = (_zoom * factor).clamp(1.0, _maxZoom);
    final content = (f - _pan) / _zoom;
    setState(() {
      _zoom = zoom;
      _pan = _clampPan(f - content * zoom, zoom);
    });
  }

  KeyEventResult _onKey(FocusNode node, KeyEvent e) {
    if (e is! KeyDownEvent) return KeyEventResult.ignored;
    switch (e.logicalKey) {
      case LogicalKeyboardKey.delete || LogicalKeyboardKey.backspace:
        if (_c.selected.isEmpty) return KeyEventResult.ignored;
        _c.deleteSelected();
      case LogicalKeyboardKey.enter || LogicalKeyboardKey.numpadEnter:
        if (_polygon.isEmpty) return KeyEventResult.ignored;
        _finishPolygon();
      case LogicalKeyboardKey.escape:
        if (_polygon.isNotEmpty) {
          setState(() => _polygon = []);
        } else if (_c.selected.isNotEmpty) {
          _c.select(null);
        } else {
          return KeyEventResult.ignored;
        }
      default:
        return KeyEventResult.ignored;
    }
    return KeyEventResult.handled;
  }

  @override
  Widget build(BuildContext context) {
    super.build(context);
    final scheme = Theme.of(context).colorScheme;
    final count = _c.cardCount;
    final l = context.l10n;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        _toolbar(context),
        const SizedBox(height: 8),
        _canvas(context),
        const SizedBox(height: 8),
        Wrap(
          spacing: 12,
          runSpacing: 8,
          crossAxisAlignment: WrapCrossAlignment.center,
          children: [
            SegmentedButton<bool>(
              key: const Key('occlusion-mode'),
              segments: [
                ButtonSegment(value: true, label: Text(l.hideAllGuessOne)),
                ButtonSegment(value: false, label: Text(l.hideOneGuessOne)),
              ],
              selected: {_c.hideAll},
              onSelectionChanged: (v) => _c.hideAll = v.single,
            ),
            Text(
              count == 0 ? l.drawShapes : l.countCards(count),
              key: const Key('occlusion-count'),
              style: TextStyle(color: scheme.onSurfaceVariant),
            ),
          ],
        ),
      ],
    );
  }

  Widget _toolbar(BuildContext context) {
    final sel = _c.selected;
    final l = context.l10n;
    return Wrap(
      spacing: 4,
      runSpacing: 4,
      crossAxisAlignment: WrapCrossAlignment.center,
      children: [
        for (final (tool, icon, label) in [
          (OcclusionTool.select, Icons.near_me_outlined, l.selectAndMove),
          (OcclusionTool.rect, Icons.crop_square, l.rectangle),
          (OcclusionTool.ellipse, Icons.circle_outlined, l.ellipse),
          (OcclusionTool.polygon, Icons.pentagon_outlined, l.polygonTool),
          (OcclusionTool.text, Icons.text_fields, l.textLabel),
        ])
          IconButton(
            key: Key('occlusion-tool-${tool.name}'),
            tooltip: label,
            isSelected: _tool == tool,
            icon: Icon(icon),
            onPressed: () => _setTool(tool),
          ),
        const SizedBox(height: 24, child: VerticalDivider()),
        if (_polygon.isNotEmpty)
          IconButton(
            key: const Key('occlusion-finish-polygon'),
            tooltip: l.finishPolygon,
            icon: const Icon(Icons.check),
            onPressed: _polygon.length >= 3 ? _finishPolygon : null,
          ),
        IconButton(
          key: const Key('occlusion-group'),
          tooltip: l.groupShapes,
          icon: const Icon(Icons.link),
          onPressed: _c.canGroup ? _c.groupSelected : null,
        ),
        IconButton(
          key: const Key('occlusion-ungroup'),
          tooltip: l.ungroup,
          icon: const Icon(Icons.link_off),
          onPressed: _c.canUngroup ? _c.ungroupSelected : null,
        ),
        IconButton(
          key: const Key('occlusion-delete'),
          tooltip: l.deleteSelectedShapes,
          icon: const Icon(Icons.delete_outline),
          onPressed: sel.isEmpty ? null : _c.deleteSelected,
        ),
        const SizedBox(height: 24, child: VerticalDivider()),
        IconButton(
          key: const Key('occlusion-zoom-out'),
          tooltip: l.zoomOut,
          icon: const Icon(Icons.zoom_out),
          onPressed: _zoom > 1 ? () => _zoomBy(1 / 1.5) : null,
        ),
        IconButton(
          key: const Key('occlusion-zoom-in'),
          tooltip: l.zoomIn,
          icon: const Icon(Icons.zoom_in),
          onPressed: _zoom < _maxZoom ? () => _zoomBy(1.5) : null,
        ),
        IconButton(
          key: const Key('occlusion-zoom-fit'),
          tooltip: l.fitImage,
          icon: const Icon(Icons.fit_screen_outlined),
          onPressed: _zoom > 1
              ? () => setState(() {
                  _zoom = 1;
                  _pan = Offset.zero;
                })
              : null,
        ),
      ],
    );
  }

  Widget _canvas(BuildContext context) {
    final image = _image;
    if (_imageError != null) {
      return Padding(
        key: const Key('occlusion-image-error'),
        padding: const EdgeInsets.all(16),
        child: Text(context.l10n.imageCannotShow),
      );
    }
    if (image == null) {
      return SizedBox(
        height: 120,
        child: Center(child: Text(context.l10n.loadingImage)),
      );
    }
    return LayoutBuilder(
      builder: (context, constraints) {
        final aspect = image.width / image.height;
        var w = constraints.maxWidth;
        var h = w / aspect;
        if (h > widget.maxHeight) {
          h = widget.maxHeight;
          w = h * aspect;
        }
        final size = Size(w, h);
        if (size != _size) {
          if (_size != Size.zero) {
            _pan = Offset(
              _pan.dx * w / _size.width,
              _pan.dy * h / _size.height,
            );
          }
          _size = size;
        }
        final scheme = Theme.of(context).colorScheme;
        return Center(
          child: Focus(
            focusNode: _focus,
            onKeyEvent: _onKey,
            child: MouseRegion(
              cursor: _tool == OcclusionTool.select
                  ? SystemMouseCursors.basic
                  : SystemMouseCursors.precise,
              onHover: (e) {
                if (_polygon.isNotEmpty) {
                  setState(() => _hover = _canvasPoint(e.localPosition));
                }
              },
              child: Listener(
                onPointerSignal: (e) {
                  if (e is PointerScrollEvent &&
                      HardwareKeyboard.instance.isControlPressed) {
                    GestureBinding.instance.pointerSignalResolver.register(e, (
                      e,
                    ) {
                      final dy = (e as PointerScrollEvent).scrollDelta.dy;
                      _zoomBy(dy < 0 ? 1.2 : 1 / 1.2, e.localPosition);
                    });
                  }
                },
                child: RawGestureDetector(
                  key: const Key('occlusion-canvas'),
                  gestures: {
                    _EagerScaleRecognizer:
                        GestureRecognizerFactoryWithHandlers<
                          _EagerScaleRecognizer
                        >(
                          _EagerScaleRecognizer.new,
                          (r) => r
                            ..onStart = _onStart
                            ..onUpdate = _onUpdate
                            ..onEnd = _onEnd,
                        ),
                  },
                  child: ClipRect(
                    child: SizedBox.fromSize(
                      size: size,
                      child: CustomPaint(
                        painter: _EditorPainter(
                          image: image,
                          shapes: _c.shapes,
                          selected: _c.selected,
                          numbers: _c.numbers,
                          zoom: _zoom,
                          pan: _pan,
                          draft: switch (_drag) {
                            _Draw(:final start, :final end) when _moved => (
                              _tool,
                              Rect.fromPoints(start, end),
                            ),
                            _ => null,
                          },
                          polygon: _polygon,
                          hover: _hover,
                          accent: scheme.primary,
                        ),
                      ),
                    ),
                  ),
                ),
              ),
            ),
          ),
        );
      },
    );
  }
}

/// Asks for a label's text.
Future<String?> _askLabel(BuildContext context) => showDialog<String>(
  context: context,
  animationStyle: popupAnimation(context),
  builder: (_) => const _LabelDialog(),
);

class _LabelDialog extends StatefulWidget {
  const _LabelDialog();

  @override
  State<_LabelDialog> createState() => _LabelDialogState();
}

class _LabelDialogState extends State<_LabelDialog> {
  final _text = TextEditingController();

  @override
  void dispose() {
    _text.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text(context.l10n.textLabel),
      content: TextField(
        key: const Key('occlusion-label-text'),
        controller: _text,
        autofocus: true,
        decoration: InputDecoration(hintText: context.l10n.shownOnEveryCard),
        onSubmitted: (v) => Navigator.of(context).pop(v),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: Text(context.l10n.cancel),
        ),
        FilledButton(
          key: const Key('occlusion-label-ok'),
          onPressed: () => Navigator.of(context).pop(_text.text),
          child: Text(context.l10n.add),
        ),
      ],
    );
  }
}

/// A gesture in progress on the canvas.
sealed class _Drag {}

/// A press on a shape, before it is a tap, long press or move.
class _Press extends _Drag {
  _Press(this.index, this.start);
  final int index;
  final Offset start;
}

class _Move extends _Drag {
  _Move(this.start, this.originals);
  final Offset start;

  /// The moved shapes as they were, by index.
  final Map<int, OcclusionShape> originals;
}

class _Resize extends _Drag {
  _Resize(this.index, this.corner, this.original);
  final int index;
  final int corner;
  final OcclusionShape original;
}

class _Draw extends _Drag {
  _Draw(this.start, this.end);
  final Offset start;
  final Offset end;
}

class _Zoom extends _Drag {
  _Zoom(this.zoom, this.pan, this.focal, [this.scaleAtStart = 1]);
  final double zoom;
  final Offset pan;
  final Offset focal;
  final double scaleAtStart;
}

/// A scale recognizer that takes the pointer at once, so drawing on the
/// canvas never scrolls the page around it.
class _EagerScaleRecognizer extends ScaleGestureRecognizer {
  @override
  void addAllowedPointer(PointerDownEvent event) {
    super.addAllowedPointer(event);
    resolve(GestureDisposition.accepted);
  }
}

class _EditorPainter extends CustomPainter {
  _EditorPainter({
    required this.image,
    required this.shapes,
    required this.selected,
    required this.numbers,
    required this.zoom,
    required this.pan,
    required this.draft,
    required this.polygon,
    required this.hover,
    required this.accent,
  });

  final ui.Image image;
  final List<OcclusionShape> shapes;
  final Set<int> selected;
  final Map<int, int> numbers;
  final double zoom;
  final Offset pan;
  final (OcclusionTool, Rect)? draft;
  final List<Offset> polygon;
  final Offset? hover;
  final Color accent;

  // Anki's colours, see-through so the image shows under the masks.
  static const _colors = OcclusionColors(
    inactive: Color(0xC0FFEBA2),
    active: Color(0xC0FF8E8E),
  );

  @override
  void paint(Canvas canvas, Size size) {
    canvas
      ..save()
      ..translate(pan.dx, pan.dy)
      ..scale(zoom);
    canvas.drawImageRect(
      image,
      Offset.zero & Size(image.width.toDouble(), image.height.toDouble()),
      Offset.zero & size,
      Paint()..filterQuality = FilterQuality.medium,
    );
    final order = [
      for (final (i, s) in shapes.indexed)
        if (!s.isLabel) i,
      for (final (i, s) in shapes.indexed)
        if (s.isLabel) i,
    ];
    for (final i in order) {
      final s = shapes[i];
      OcclusionPainter.paintShape(
        canvas,
        size,
        s,
        s.isLabel
            ? OcclusionLook.label
            : selected.contains(i)
            ? OcclusionLook.active
            : OcclusionLook.inactive,
        _colors,
      );
    }
    for (final (i, s) in shapes.indexed) {
      if (!s.isLabel) _number(canvas, size, s, numbers[s.ord] ?? 0);
      if (selected.contains(i)) {
        _selection(canvas, size, s, single: selected.length == 1);
      }
    }
    if (draft case (final tool, final r)) {
      final stroke = Paint()
        ..style = PaintingStyle.stroke
        ..strokeWidth = 2 / zoom
        ..color = accent;
      if (tool == OcclusionTool.ellipse) {
        canvas.drawOval(r, stroke);
      } else {
        canvas.drawRect(r, stroke);
      }
    }
    if (polygon.isNotEmpty) {
      final line = Paint()
        ..style = PaintingStyle.stroke
        ..strokeWidth = 2 / zoom
        ..color = accent;
      final path = Path()..addPolygon([...polygon, ?hover], false);
      canvas.drawPath(path, line);
      for (final p in polygon) {
        canvas.drawCircle(p, 4 / zoom, Paint()..color = accent);
      }
    }
    canvas.restore();
  }

  /// The card number of a shape, at its top left corner.
  void _number(Canvas canvas, Size size, OcclusionShape s, int n) {
    final at = _corners(_frameOf(s, size)).first;
    final painter = TextPainter(
      text: TextSpan(
        text: '$n',
        style: TextStyle(
          fontSize: 11 / zoom,
          color: Colors.white,
          fontWeight: FontWeight.bold,
        ),
      ),
      textDirection: TextDirection.ltr,
    )..layout();
    final r = 9 / zoom;
    canvas.drawCircle(at, r, Paint()..color = const Color(0xFF212121));
    painter.paint(canvas, at - Offset(painter.width / 2, painter.height / 2));
    painter.dispose();
  }

  void _selection(
    Canvas canvas,
    Size size,
    OcclusionShape s, {
    required bool single,
  }) {
    final corners = _corners(_frameOf(s, size));
    final outline = Paint()
      ..style = PaintingStyle.stroke
      ..strokeWidth = 1.5 / zoom
      ..color = accent;
    canvas.drawPath(Path()..addPolygon(corners, true), outline);
    if (!single) return;
    // Drawn 12 pixels wide; the touch target is larger.
    final half = 6 / zoom;
    for (final c in corners) {
      final r = Rect.fromCenter(center: c, width: 2 * half, height: 2 * half);
      canvas
        ..drawRect(r, Paint()..color = Colors.white)
        ..drawRect(r, outline);
    }
  }

  @override
  bool shouldRepaint(_EditorPainter old) => true;
}
