import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/app/theme.dart';
import 'package:cardflip/features/progress/session_rewards.dart';

import 'celebration_sequence_test.dart'
    show goal, level, loadSettings, pageApp, streak, tier;

void main() {
  setUpAll(() async {
    // The icons on the badges and the flame; without the font they are
    // boxes.
    await (FontLoader(
      'MaterialIcons',
    )..addFont(rootBundle.load('fonts/MaterialIcons-Regular.otf'))).load();
  });

  final pages = <String, CelebrationEvent>{
    'streak': streak,
    'goal': goal,
    'level': level,
    'achievement': AchievementCelebration(tier('answers', 1000)),
    'more': MoreAchievementsCelebration([
      tier('study-time', 10),
      tier('long-term', 100),
    ]),
  };

  /// Compares the page of [event] at its final state with golden [name].
  Future<void> golden(
    WidgetTester tester,
    CelebrationEvent event,
    String name, {
    bool dark = false,
    bool minimalist = false,
  }) async {
    tester.view
      ..physicalSize = const Size(400, 720)
      ..devicePixelRatio = 1;
    addTearDown(tester.view.reset);
    final settings = await loadSettings(tester, minimalist: minimalist);
    await tester.pumpWidget(
      pageApp(
        event,
        settings,
        locale: const Locale('en', 'GB'),
        theme: (dark ? darkTheme : lightTheme)(minimalist: minimalist),
      ),
    );
    // Past the animation and the confetti.
    for (var i = 0; i < 90; i++) {
      await tester.pump(const Duration(milliseconds: 50));
    }
    await expectLater(
      find.byType(MaterialApp),
      matchesGoldenFile('goldens/celebration_$name.png'),
    );
  }

  for (final MapEntry(key: name, value: event) in pages.entries) {
    testWidgets('the $name page in the light theme', (tester) async {
      await golden(tester, event, '${name}_light');
    });
    testWidgets('the $name page in the dark theme', (tester) async {
      await golden(tester, event, '${name}_dark', dark: true);
    });
  }

  testWidgets('the level page in minimalist mode', (tester) async {
    await golden(tester, level, 'level_minimalist', minimalist: true);
  });
}
