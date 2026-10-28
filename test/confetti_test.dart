import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/app/theme.dart';
import 'package:cardflip/features/progress/confetti.dart';

void main() {
  Widget app(Widget child) => MaterialApp(
    debugShowCheckedModeBanner: false,
    theme: lightTheme(),
    home: Scaffold(body: child),
  );

  testWidgets('confetti ignores taps and finishes', (tester) async {
    var taps = 0;
    var ended = false;
    await tester.pumpWidget(
      app(
        Stack(
          children: [
            Center(
              child: TextButton(
                onPressed: () => taps++,
                child: const Text('Under'),
              ),
            ),
            Positioned.fill(child: Confetti(onEnd: () => ended = true)),
          ],
        ),
      ),
    );
    await tester.pump(const Duration(milliseconds: 300));
    await tester.tap(find.text('Under'));
    expect(taps, 1);
    await tester.pump(const Duration(milliseconds: 1000));
    expect(ended, isFalse);
    await tester.pump(const Duration(milliseconds: 600));
    expect(ended, isTrue);
  });

  testWidgets('confetti at mid-burst', (tester) async {
    tester.view
      ..physicalSize = const Size(400, 400)
      ..devicePixelRatio = 1;
    addTearDown(tester.view.reset);
    await tester.pumpWidget(
      app(
        RepaintBoundary(
          key: const Key('confetti'),
          child: Confetti(seed: 3, origin: const Alignment(0, 0.2)),
        ),
      ),
    );
    await tester.pump();
    for (var i = 0; i < 12; i++) {
      await tester.pump(const Duration(milliseconds: 50));
    }
    await expectLater(
      find.byKey(const Key('confetti')),
      matchesGoldenFile('goldens/confetti.png'),
    );
  });
}
