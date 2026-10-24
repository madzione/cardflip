import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/features/editor/occlusion_editor.dart';
import 'package:cardflip/widgets/occlusion_view.dart';

// The Occlusion fields of the fixture deck made with Anki 26.09
// (golang/importer/apkg/testdata, cf-io-all and cf-io-one).
const ankiHideAll =
    '{{c1::image-occlusion:rect:left=.1:top=.2:width=.3:height=.1:oi=1}}<br>'
    '{{c2::image-occlusion:ellipse:left=.5:top=.5:rx=.1:ry=.05:angle=10:oi=1}}<br>'
    '{{c3::image-occlusion:polygon:points=.1,.6 .3,.6 .2,.8:oi=1}}<br>'
    '{{c3::image-occlusion:rect:left=.7:top=.7:width=.1:height=.1:oi=1}}<br>'
    r'{{c0::image-occlusion:text:left=.6:top=.1:text=Label\: heart:scale=1:fs=.05:oi=1}}';
const ankiHideOne =
    '{{c1::image-occlusion:rect:left=.05:top=.05:width=.2:height=.2}}<br>'
    '{{c2::image-occlusion:rect:left=.6:top=.6:width=.3:height=.25}}';

/// A grey PNG of [width] × [height] pixels.
Future<Uint8List> png(int width, int height) async {
  final recorder = ui.PictureRecorder();
  Canvas(recorder).drawRect(
    Rect.fromLTWH(0, 0, width.toDouble(), height.toDouble()),
    Paint()..color = const Color(0xFF888888),
  );
  final image = await recorder.endRecording().toImage(width, height);
  final data = await image.toByteData(format: ui.ImageByteFormat.png);
  return data!.buffer.asUint8List();
}

/// Pumps the editor 400 pixels wide on a 200×100 image, so the canvas is
/// 400×200.
Future<void> pumpEditor(WidgetTester tester, OcclusionController c) async {
  final image = await tester.runAsync(() => png(200, 100));
  await tester.pumpWidget(
    MaterialApp(
      home: Scaffold(
        body: Center(
          child: SizedBox(
            width: 400,
            child: OcclusionEditor(controller: c, image: image!),
          ),
        ),
      ),
    ),
  );
  // Decoding the image runs outside the fake clock.
  for (
    var i = 0;
    i < 10 && find.byKey(const Key('occlusion-canvas')).evaluate().isEmpty;
    i++
  ) {
    await tester.runAsync(
      () => Future<void>.delayed(const Duration(milliseconds: 20)),
    );
    await tester.pump();
  }
  expect(canvasRect(tester).size, const Size(400, 200));
}

Rect canvasRect(WidgetTester tester) =>
    tester.getRect(find.byKey(const Key('occlusion-canvas')));

/// A point on the canvas at relative image coordinates.
Offset at(WidgetTester tester, double x, double y) {
  final r = canvasRect(tester);
  return r.topLeft + Offset(x * r.width, y * r.height);
}

Future<void> tool(WidgetTester tester, OcclusionTool tool) async {
  await tester.tap(find.byKey(Key('occlusion-tool-${tool.name}')));
  await tester.pump();
}

String countText(WidgetTester tester) =>
    tester.widget<Text>(find.byKey(const Key('occlusion-count'))).data!;

void main() {
  group('Occlusion field', () {
    test("round-trips Anki's fields", () {
      for (final field in [ankiHideAll, ankiHideOne]) {
        expect(
          OcclusionController(field: field, keepNumbers: true).field,
          field,
        );
      }
    });

    test('parses each shape kind', () {
      final shapes = parseOcclusionField(ankiHideAll);
      expect(shapes.map((s) => (s.kind, s.ord)), [
        ('rect', 1),
        ('ellipse', 2),
        ('polygon', 3),
        ('rect', 3),
        ('text', 0),
      ]);
      expect(shapes[1].angle, 10);
      expect(shapes[2].points, const [
        Offset(.1, .6),
        Offset(.3, .6),
        Offset(.2, .8),
      ]);
      expect(shapes[4].text, 'Label: heart');
      expect(shapes.every((s) => s.occludeInactive), isTrue);
    });

    test('drops unknown kinds and properties', () {
      final shapes = parseOcclusionField(
        '{{c1::image-occlusion:rect:left=.1:top=.1:width=.2:height=.2:colour=red:oi=1}}'
        '{{c2::image-occlusion:star:points=5}}{{c3::Paris}}'
        '{{c4::image-occlusion:rect:left=.1:top=.1:width=0:height=.2}}',
      );
      expect(shapes, const [
        OcclusionShape(
          kind: 'rect',
          ord: 1,
          left: .1,
          top: .1,
          width: .2,
          height: .2,
          occludeInactive: true,
        ),
      ]);
    });

    test('label text is kept as field HTML', () {
      final c = OcclusionController()
        ..add(
          const OcclusionShape(
            kind: 'text',
            left: .5,
            top: .5,
            text: 'a < b: {x}',
          ),
        );
      expect(
        c.field,
        r'{{c0::image-occlusion:text:left=.5:top=.5:text=a &lt; b\: {x&#125;:scale=1:fs=.05:oi=1}}',
      );
      expect(parseOcclusionField(c.field).single.text, 'a < b: {x}');
    });

    test('a new note numbers its groups 1 to N in creation order', () {
      final c = OcclusionController();
      for (var i = 0; i < 3; i++) {
        c.add(
          OcclusionShape(
            kind: 'rect',
            left: i / 10,
            top: 0,
            width: .05,
            height: .05,
          ),
        );
      }
      c
        ..select(0)
        ..deleteSelected();
      expect(parseOcclusionField(c.field).map((s) => s.ord), [1, 2]);
    });

    test('nor the numbers of its empty cards', () {
      // Shape 3 was removed before; its card is an empty card.
      final c = OcclusionController(
        field: ankiHideOne,
        keepNumbers: true,
        cardNumbers: const [1, 2, 3],
      )..add(const OcclusionShape(kind: 'rect', width: .1, height: .1));
      expect(parseOcclusionField(c.field).map((s) => s.ord), [1, 2, 4]);
    });

    test('an edited note keeps its numbers and does not reuse them', () {
      final c = OcclusionController(field: ankiHideOne, keepNumbers: true)
        ..select(0)
        ..deleteSelected();
      // Card 2 stays with its shape; the new shape gets a new card.
      c.add(
        const OcclusionShape(
          kind: 'rect',
          left: .1,
          top: .1,
          width: .1,
          height: .1,
        ),
      );
      expect(parseOcclusionField(c.field).map((s) => s.ord), [2, 3]);
    });
  });

  testWidgets('drawing a rectangle', (tester) async {
    final c = OcclusionController();
    await pumpEditor(tester, c);
    expect(countText(tester), startsWith('Draw a shape'));
    await tool(tester, OcclusionTool.rect);
    await tester.dragFrom(at(tester, .1, .2), const Offset(200, 100));
    await tester.pump();
    expect(
      c.field,
      '{{c1::image-occlusion:rect:left=.1:top=.2:width=.5:height=.5:oi=1}}',
    );
    expect(countText(tester), '1 card');

    // A tap draws nothing.
    await tester.tapAt(at(tester, .8, .8));
    await tester.pump();
    expect(c.shapes, hasLength(1));
  });

  testWidgets('drawing an ellipse, a polygon and a label', (tester) async {
    final c = OcclusionController();
    await pumpEditor(tester, c);
    await tool(tester, OcclusionTool.ellipse);
    await tester.dragFrom(at(tester, .5, .5), const Offset(80, 40));
    await tester.pump();

    await tool(tester, OcclusionTool.polygon);
    for (final (x, y) in [(.1, .6), (.3, .6), (.2, .8)]) {
      await tester.tapAt(at(tester, x, y));
      await tester.pump();
    }
    // Tapping the first corner again closes it.
    await tester.tapAt(at(tester, .1, .6));
    await tester.pump();

    await tool(tester, OcclusionTool.text);
    await tester.tapAt(at(tester, .6, .1));
    await tester.pumpAndSettle();
    await tester.enterText(
      find.byKey(const Key('occlusion-label-text')),
      'Aorta',
    );
    await tester.tap(find.byKey(const Key('occlusion-label-ok')));
    await tester.pumpAndSettle();

    expect(
      c.field,
      '{{c1::image-occlusion:ellipse:left=.5:top=.5:rx=.1:ry=.1:oi=1}}<br>'
      '{{c2::image-occlusion:polygon:points=.1,.6 .3,.6 .2,.8:oi=1}}<br>'
      '{{c0::image-occlusion:text:left=.6:top=.1:text=Aorta:scale=1:fs=.05:oi=1}}',
    );
    // The label makes no card.
    expect(countText(tester), '2 cards');
  });

  testWidgets('grouping two shapes', (tester) async {
    final c = OcclusionController(field: ankiHideOne);
    await pumpEditor(tester, c);
    expect(countText(tester), '2 cards');
    final group = find.byKey(const Key('occlusion-group'));
    await tester.tapAt(at(tester, .1, .1));
    await tester.sendKeyDownEvent(LogicalKeyboardKey.shiftLeft);
    await tester.tapAt(at(tester, .7, .7));
    await tester.sendKeyUpEvent(LogicalKeyboardKey.shiftLeft);
    await tester.pump();
    expect(c.selected, {0, 1});
    await tester.tap(group);
    await tester.pump();
    expect(
      c.field,
      '{{c1::image-occlusion:rect:left=.05:top=.05:width=.2:height=.2}}<br>'
      '{{c1::image-occlusion:rect:left=.6:top=.6:width=.3:height=.25}}',
    );
    expect(countText(tester), '1 card');

    await tester.tap(find.byKey(const Key('occlusion-ungroup')));
    await tester.pump();
    expect(parseOcclusionField(c.field).map((s) => s.ord), [1, 2]);
  });

  testWidgets('a long press adds to the selection (touch)', (tester) async {
    final c = OcclusionController(field: ankiHideOne);
    await pumpEditor(tester, c);
    await tester.tapAt(at(tester, .1, .1));
    await tester.longPressAt(at(tester, .7, .7));
    await tester.pump();
    expect(c.selected, {0, 1});
    expect(c.canGroup, isTrue);
  });

  testWidgets('switching the mode', (tester) async {
    final c = OcclusionController(field: ankiHideOne);
    await pumpEditor(tester, c);
    expect(c.hideAll, isFalse);
    await tester.tap(find.text('Hide all, guess one'));
    await tester.pump();
    expect(c.field, contains(':oi=1}}'));
    await tester.tap(find.text('Hide one, guess one'));
    await tester.pump();
    expect(c.field, ankiHideOne);
  });

  testWidgets('moving, resizing and deleting a shape', (tester) async {
    final c = OcclusionController(
      field:
          '{{c1::image-occlusion:rect:left=.25:top=.25:width=.25:height=.25}}',
    );
    await pumpEditor(tester, c);
    // Move by a quarter of the width.
    await tester.dragFrom(at(tester, .3, .3), const Offset(100, 0));
    await tester.pump();
    expect(
      c.shapes.single,
      isA<OcclusionShape>().having((s) => s.left, 'left', .5),
    );
    expect(c.selected, {0});

    // Drag the bottom right handle (canvas .75, .5) to (.875, .75).
    await tester.dragFrom(at(tester, .75, .5), const Offset(50, 50));
    await tester.pump();
    final s = c.shapes.single;
    expect((s.left, s.top, s.width, s.height), (.5, .25, .375, .5));

    await tester.sendKeyEvent(LogicalKeyboardKey.delete);
    await tester.pump();
    expect(c.shapes, isEmpty);
    expect(c.field, '');
  });

  testWidgets('drawing while zoomed in', (tester) async {
    final c = OcclusionController();
    await pumpEditor(tester, c);
    await tester.tap(find.byKey(const Key('occlusion-zoom-in')));
    await tester.pump();
    // 1.5 times around the centre: the canvas shows x from 1/6 to 5/6.
    await tool(tester, OcclusionTool.rect);
    await tester.dragFrom(at(tester, .25, .25), const Offset(60, 30));
    await tester.pump();
    final s = c.shapes.single;
    expect(s.left, moreOrLessEquals(1 / 6 + .25 / 1.5));
    expect(s.width, moreOrLessEquals(.1));
    expect(s.height, moreOrLessEquals(.1));
  });
}
