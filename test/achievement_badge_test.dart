import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/app/theme.dart';
import 'package:cardflip/features/progress/achievement_badge.dart';

void main() {
  setUpAll(() async {
    // The icons on the discs; without the font they are boxes.
    await (FontLoader(
      'MaterialIcons',
    )..addFont(rootBundle.load('fonts/MaterialIcons-Regular.otf'))).load();
  });

  /// Every achievement at a tier, the Streak in each tier and locked, and
  /// a minimalist badge, as in assets/badges/achievement-badges.png.
  Widget sheet(ThemeData theme) => MaterialApp(
    debugShowCheckedModeBanner: false,
    theme: theme,
    home: Scaffold(
      body: Center(
        child: RepaintBoundary(
          key: const Key('sheet'),
          child: ColoredBox(
            color: theme.colorScheme.surface,
            child: const Padding(
              padding: EdgeInsets.all(8),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      AchievementBadge(id: 'answers', tier: 'bronze', stars: 1),
                      AchievementBadge(id: 'streak', tier: 'gold', stars: 3),
                      AchievementBadge(
                        id: 'study-days',
                        tier: 'silver',
                        stars: 2,
                      ),
                      AchievementBadge(
                        id: 'daily-goal',
                        tier: 'gold',
                        stars: 3,
                      ),
                      AchievementBadge(
                        id: 'long-term',
                        tier: 'silver',
                        stars: 2,
                      ),
                      AchievementBadge(id: 'comeback', tier: 'gold'),
                    ],
                  ),
                  Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      AchievementBadge(id: 'streak', tier: 'bronze', stars: 1),
                      AchievementBadge(id: 'streak', tier: 'silver', stars: 2),
                      AchievementBadge(id: 'streak', tier: 'gold', stars: 3),
                      AchievementBadge(
                        id: 'streak',
                        tier: 'platinum',
                        stars: 4,
                      ),
                      AchievementBadge(id: 'streak', tier: null),
                      AchievementBadge(
                        id: 'answers',
                        tier: 'silver',
                        stars: 2,
                        minimalist: true,
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    ),
  );

  testWidgets('badges in the light theme', (tester) async {
    tester.view
      ..physicalSize = const Size(1200, 600)
      ..devicePixelRatio = 1;
    addTearDown(tester.view.reset);
    await tester.pumpWidget(sheet(lightTheme()));
    await expectLater(
      find.byKey(const Key('sheet')),
      matchesGoldenFile('goldens/achievement_badges_light.png'),
    );
  });

  testWidgets('badges in the dark theme', (tester) async {
    tester.view
      ..physicalSize = const Size(1200, 600)
      ..devicePixelRatio = 1;
    addTearDown(tester.view.reset);
    await tester.pumpWidget(sheet(darkTheme()));
    await expectLater(
      find.byKey(const Key('sheet')),
      matchesGoldenFile('goldens/achievement_badges_dark.png'),
    );
  });

  testWidgets('a locked badge has a padlock and no stars', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        theme: lightTheme(),
        home: const AchievementBadge(id: 'streak', tier: null, stars: 3),
      ),
    );
    expect(find.byIcon(Icons.lock), findsOneWidget);
    expect(find.byIcon(Icons.local_fire_department), findsOneWidget);
  });
}
