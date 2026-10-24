import 'dart:ui' as ui;

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

/// A PNG of 200 × 100 pixels in [color], grey by default.
Future<Uint8List> occlusionImage(
  WidgetTester tester, {
  Color color = const Color(0xFF888888),
}) async {
  final bytes = await tester.runAsync(() async {
    final recorder = ui.PictureRecorder();
    Canvas(recorder)
        .drawRect(const Rect.fromLTWH(0, 0, 200, 100), Paint()..color = color);
    final image = await recorder.endRecording().toImage(200, 100);
    final data = await image.toByteData(format: ui.ImageByteFormat.png);
    return data!.buffer.asUint8List();
  });
  return bytes!;
}

/// Waits for the mask editor to decode its image.
Future<void> waitForCanvas(WidgetTester tester) async {
  final canvas = find.byKey(const Key('occlusion-canvas'));
  for (var i = 0; i < 20 && canvas.evaluate().isEmpty; i++) {
    await tester.runAsync(
      () => Future<void>.delayed(const Duration(milliseconds: 20)),
    );
    await tester.pump();
  }
  expect(canvas, findsOneWidget);
}

/// Scrolls the mask editor's canvas into view.
Future<void> showCanvas(WidgetTester tester) async {
  await tester.ensureVisible(find.byKey(const Key('occlusion-canvas')));
  await tester.pumpAndSettle();
}

/// A point on the mask editor's canvas, at relative image coordinates.
Offset onCanvas(WidgetTester tester, double x, double y) {
  final r = tester.getRect(find.byKey(const Key('occlusion-canvas')));
  return r.topLeft + Offset(x * r.width, y * r.height);
}
