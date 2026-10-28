import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/app/theme.dart';
import 'package:cardflip/features/progress/reward_banner.dart';
import 'package:cardflip/features/progress/reward_card.dart';
import 'package:cardflip/features/progress/session_rewards.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';

import 'support/fake_go.dart';

void main() {
  setUpAll(() async {
    // The icons on the badges; without the font they are boxes.
    await (FontLoader(
      'MaterialIcons',
    )..addFont(rootBundle.load('fonts/MaterialIcons-Regular.otf'))).load();
  });

  final tier = golang.NewAchievement.fromJson(
    newAchievementJson('answers', 1000),
  );

  /// [child] in [theme] and its minimalist variant, one above the other.
  Widget sheet(
    ThemeData Function({bool minimalist}) theme,
    Widget Function(bool minimalist) child,
  ) => MaterialApp(
    debugShowCheckedModeBanner: false,
    theme: theme(),
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    home: Scaffold(
      body: Center(
        child: RepaintBoundary(
          key: const Key('sheet'),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              for (final minimalist in [false, true])
                Theme(
                  data: theme(minimalist: minimalist),
                  child: Builder(
                    builder: (context) => ColoredBox(
                      color: Theme.of(context).colorScheme.surface,
                      child: Padding(
                        padding: const EdgeInsets.all(16),
                        child: child(minimalist),
                      ),
                    ),
                  ),
                ),
            ],
          ),
        ),
      ),
    ),
  );

  /// Compares the sheet of [child] in [theme] with golden [name], [after]
  /// its start.
  Future<void> golden(
    WidgetTester tester,
    ThemeData Function({bool minimalist}) theme,
    Widget Function(bool minimalist) child,
    String name, {
    Duration after = const Duration(seconds: 1),
  }) async {
    tester.view
      ..physicalSize = const Size(800, 600)
      ..devicePixelRatio = 1;
    addTearDown(tester.view.reset);
    await tester.pumpWidget(sheet(theme, child));
    // Frame by frame, as the animations run.
    const frame = Duration(milliseconds: 50);
    for (var t = Duration.zero; t < after; t += frame) {
      await tester.pump(frame);
    }
    await expectLater(
      find.byKey(const Key('sheet')),
      matchesGoldenFile('goldens/$name.png'),
    );
  }

  Widget banner(bool minimalist) => RewardBanner(
    events: RewardEvents(level: 13, tiers: [tier]),
    minimalist: minimalist,
  );

  // Past the banner's slide in, before it fades out.
  testWidgets('the banner in the light theme', (tester) async {
    await golden(tester, lightTheme, banner, 'reward_banner_light');
  });

  testWidgets('the banner in the dark theme', (tester) async {
    await golden(tester, darkTheme, banner, 'reward_banner_dark');
  });

  Widget rewardCard(bool minimalist) => SizedBox(
    width: 360,
    child: RewardCard(
      goalReached: true,
      levelFrom: (level: 12, fraction: 0.8),
      levelTo: (level: 13, fraction: 0.25),
      tiers: [
        tier,
        golang.NewAchievement.fromJson(newAchievementJson('streak', 30)),
        golang.NewAchievement.fromJson(newAchievementJson('study-days', 100)),
        golang.NewAchievement.fromJson(newAchievementJson('comeback', 1)),
      ],
      minimalist: minimalist,
    ),
  );

  // A static recap: as it first appears.
  testWidgets('the card in the light theme', (tester) async {
    await golden(
      tester,
      lightTheme,
      rewardCard,
      'reward_card_light',
      after: Duration.zero,
    );
  });

  testWidgets('the card in the dark theme', (tester) async {
    await golden(
      tester,
      darkTheme,
      rewardCard,
      'reward_card_dark',
      after: Duration.zero,
    );
  });
}
