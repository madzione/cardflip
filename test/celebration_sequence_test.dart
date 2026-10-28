import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/app/theme.dart';
import 'package:cardflip/data/changes.dart';
import 'package:cardflip/data/settings_model.dart';
import 'package:cardflip/features/progress/celebration_feedback.dart';
import 'package:cardflip/features/progress/celebration_sequence.dart';
import 'package:cardflip/features/progress/session_rewards.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';

import 'support/fake_celebration_feedback.dart';
import 'support/fake_go.dart';

/// Thursday, 29 October 2026.
final thursday = DateTime(2026, 10, 29, 12);

/// The 7 days ending on [thursday]: Monday and Wednesday studied, Tuesday
/// a rest day, Thursday studied in the session; the week before has no
/// study.
const restWeek = ['none', 'none', 'none', 'study', 'rest', 'study', 'study'];

golang.NewAchievement tier(String id, int threshold) =>
    golang.NewAchievement.fromJson(newAchievementJson(id, threshold));

const streak = StreakCelebration(from: 12, to: 13, week: restWeek, best: false);
const goal = GoalCelebration(progress: 34, goal: 30, kind: 'cards', xp: 20);
const level = LevelCelebration(
  fromLevel: 12,
  fromFraction: 0.8,
  toLevel: 13,
  xpLeft: 640,
);

/// Loads settings from the fake, with [minimalist] mode.
Future<SettingsModel> loadSettings(
  WidgetTester tester, {
  bool minimalist = false,
}) async {
  final go = FakeGo();
  go.settings = {...go.settings, 'minimalist': minimalist};
  final hub = ChangeHub();
  final settings = SettingsModel(hub);
  addTearDown(() {
    settings.dispose();
    hub.dispose();
  });
  await settings.load();
  return settings;
}

/// The page of [event] in an app of [theme] and [locale], with [reduce]
/// motion.
Widget pageApp(
  CelebrationEvent event,
  SettingsModel settings, {
  ThemeData? theme,
  Locale locale = const Locale('en'),
  bool reduce = false,
  VoidCallback? onNext,
  VoidCallback? onSkip,
}) => MaterialApp(
  debugShowCheckedModeBanner: false,
  theme: theme ?? lightTheme(minimalist: settings.minimalist),
  locale: locale,
  localizationsDelegates: AppLocalizations.localizationsDelegates,
  supportedLocales: const [Locale('en'), Locale('en', 'GB'), Locale('de')],
  home: Builder(
    builder: (context) => MediaQuery(
      data: MediaQuery.of(context).copyWith(disableAnimations: reduce),
      child: Scaffold(
        body: CelebrationPage(
          event: event,
          settings: settings,
          today: thursday,
          onNext: onNext ?? () {},
          onSkip: onSkip ?? () {},
        ),
      ),
    ),
  ),
);

void main() {
  group('texts', () {
    testWidgets('streak', (tester) async {
      // Extended (spec): "13 day streak".
      await tester.pumpWidget(pageApp(streak, await loadSettings(tester)));
      await tester.pump(const Duration(seconds: 3));
      expect(find.text('13 day streak'), findsOneWidget);
      expect(find.text('13'), findsOneWidget); // the rolled number
      expect(find.text('New personal best'), findsNothing);
      expect(find.bySemanticsLabel('13 day streak'), findsOneWidget);
    });

    testWidgets('personal best and the first day', (tester) async {
      final settings = await loadSettings(tester);
      await tester.pumpWidget(
        pageApp(
          const StreakCelebration(
            from: 40,
            to: 41,
            week: restWeek,
            best: true,
          ),
          settings,
        ),
      );
      await tester.pump(const Duration(seconds: 3));
      expect(find.text('41 day streak'), findsOneWidget);
      expect(find.text('New personal best'), findsOneWidget);
      await tester.pumpWidget(
        pageApp(
          const StreakCelebration(from: 0, to: 1, week: restWeek, best: false),
          settings,
          locale: const Locale('de'),
        ),
      );
      await tester.pump(const Duration(seconds: 3));
      expect(find.text('1 Tag in Folge'), findsOneWidget);
    });

    testWidgets('goal', (tester) async {
      // Card goal (spec).
      await tester.pumpWidget(pageApp(goal, await loadSettings(tester)));
      await tester.pump(const Duration(seconds: 3));
      expect(find.text('Daily goal reached'), findsOneWidget);
      expect(find.text('34 / 30 cards'), findsOneWidget);
      expect(find.text('+20 XP'), findsOneWidget);
      expect(find.byKey(const Key('celebration-check')), findsOneWidget);
    });

    testWidgets('level', (tester) async {
      // One level (spec).
      await tester.pumpWidget(pageApp(level, await loadSettings(tester)));
      await tester.pump(const Duration(milliseconds: 200));
      expect(
        tester.widget<Text>(find.byKey(const Key('celebration-level'))).data,
        '12',
      );
      await tester.pump(const Duration(seconds: 3));
      expect(find.text('Level 13'), findsOneWidget);
      expect(find.text('640 XP to level 14'), findsOneWidget);
      expect(
        tester.widget<Text>(find.byKey(const Key('celebration-level'))).data,
        '13',
      );
      expect(find.byKey(const Key('celebration-xp')), findsNothing);
    });

    testWidgets('achievement', (tester) async {
      // One tier (spec).
      await tester.pumpWidget(
        pageApp(
          AchievementCelebration(tier('answers', 1000)),
          await loadSettings(tester),
        ),
      );
      await tester.pump(const Duration(milliseconds: 1600));
      expect(find.byKey(const Key('confetti')), findsOneWidget);
      await tester.pump(const Duration(seconds: 3));
      expect(find.text('New achievement'), findsOneWidget);
      expect(find.text('1,000 answers · Silver'), findsOneWidget);
      expect(find.text('+50 XP'), findsOneWidget);
      // Both of the silver tier's stars.
      expect(find.byKey(const ValueKey('celebration-badge-2')), findsOneWidget);
      expect(find.byKey(const Key('confetti')), findsNothing);
    });

    testWidgets('the tiers beyond three', (tester) async {
      // Five tiers (spec): the fourth page names the other two.
      await tester.pumpWidget(
        pageApp(
          MoreAchievementsCelebration([
            tier('study-time', 10),
            tier('long-term', 100),
          ]),
          await loadSettings(tester),
        ),
      );
      await tester.pump(const Duration(seconds: 5));
      expect(find.text('New achievements'), findsOneWidget);
      expect(
        find.text('10 hours of study, 100 cards in long-term memory'),
        findsOneWidget,
      );
      expect(find.text('+50 XP'), findsOneWidget);
      expect(
        find.byKey(const ValueKey('celebration-badge-study-time:10')),
        findsOneWidget,
      );
      expect(
        find.byKey(const ValueKey('celebration-badge-long-term:100')),
        findsOneWidget,
      );
    });
  });

  group('week strip', () {
    /// The states of the strip's days, from its first.
    List<String> states(WidgetTester tester) => [
      for (var i = 0; i < 7; i++)
        for (final s in ['study', 'rest', 'none', 'later'])
          if (find.byKey(ValueKey('week-$i-$s')).evaluate().isNotEmpty) s,
    ];

    testWidgets('from Monday, with a rest day', (tester) async {
      // Week with a rest day (spec).
      await tester.pumpWidget(
        pageApp(
          streak,
          await loadSettings(tester),
          locale: const Locale('en', 'GB'),
        ),
      );
      await tester.pump(const Duration(seconds: 3));
      expect(states(tester), [
        'study', // Monday
        'rest', // Tuesday
        'study', // Wednesday
        'study', // Thursday, today
        'later',
        'later',
        'later',
      ]);
    });

    testWidgets('from Sunday', (tester) async {
      await tester.pumpWidget(pageApp(streak, await loadSettings(tester)));
      await tester.pump(const Duration(seconds: 3));
      expect(states(tester), [
        'none', // Sunday
        'study',
        'rest',
        'study',
        'study', // Thursday
        'later',
        'later',
      ]);
    });

    testWidgets("today's tick stamps in last", (tester) async {
      await tester.pumpWidget(
        pageApp(
          streak,
          await loadSettings(tester),
          locale: const Locale('en', 'GB'),
        ),
      );
      await tester.pump(const Duration(milliseconds: 500));
      Iterable<double> todayOpacity() => tester
          .widgetList<Opacity>(
            find.descendant(
              of: find.byKey(const ValueKey('week-3-study')),
              matching: find.byType(Opacity),
            ),
          )
          .map((o) => o.opacity);
      expect(todayOpacity(), isEmpty); // not stamped yet
      await tester.pump(const Duration(milliseconds: 1000)); // 1.5 s
      expect(todayOpacity().single, inExclusiveRange(0, 1));
      await tester.pump(const Duration(seconds: 1));
      expect(todayOpacity(), isEmpty); // stamped
    });

    testWidgets('on the first day of the week only today', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          locale: const Locale('en', 'GB'),
          localizationsDelegates: AppLocalizations.localizationsDelegates,
          supportedLocales: const [Locale('en', 'GB')],
          home: WeekStrip(week: restWeek, today: DateTime(2026, 10, 26, 12)),
        ),
      );
      expect(states(tester), [
        'study',
        'later',
        'later',
        'later',
        'later',
        'later',
        'later',
      ]);
    });
  });

  group('moving on', () {
    testWidgets('input waits half a second', (tester) async {
      // Space right after the last card (spec).
      var next = 0, skip = 0;
      await tester.pumpWidget(
        pageApp(
          streak,
          await loadSettings(tester),
          onNext: () => next++,
          onSkip: () => skip++,
        ),
      );
      await tester.pump(const Duration(milliseconds: 200));
      await tester.sendKeyEvent(LogicalKeyboardKey.space);
      await tester.sendKeyEvent(LogicalKeyboardKey.escape);
      await tester.tap(find.byKey(const Key('celebration-continue')));
      expect((next, skip), (0, 0));
      await tester.pump(const Duration(milliseconds: 400)); // 600 ms
      await tester.sendKeyEvent(LogicalKeyboardKey.space);
      expect(next, 1);
      await tester.sendKeyEvent(LogicalKeyboardKey.enter);
      expect(next, 2);
      await tester.sendKeyEvent(LogicalKeyboardKey.escape);
      expect(skip, 1);
      // Continue during the animation (spec).
      await tester.tap(find.byKey(const Key('celebration-continue')));
      expect(next, 3);
      await tester.pump(const Duration(seconds: 3));
    });

    testWidgets('reduce motion shows the final state at once', (
      tester,
    ) async {
      // Reduce motion (spec).
      final settings = await loadSettings(tester);
      await tester.pumpWidget(pageApp(goal, settings, reduce: true));
      await tester.pump();
      expect(find.byKey(const Key('celebration-check')), findsOneWidget);
      expect(find.text('Daily goal reached'), findsOneWidget);
      await tester.pump(const Duration(seconds: 1));
      expect(find.byKey(const Key('celebration')), findsNothing);

      await tester.pumpWidget(pageApp(level, settings, reduce: true));
      await tester.pump();
      expect(
        tester.widget<Text>(find.byKey(const Key('celebration-level'))).data,
        '13',
      );
      await tester.pump(const Duration(seconds: 2));
      expect(find.byKey(const Key('confetti')), findsNothing);
      expect(find.byKey(const Key('celebration-rays')), findsNothing);
    });

    testWidgets('one sound and vibration per page', (tester) async {
      // Level page (spec): on Android the level sound and one vibration,
      // also under reduce motion.
      debugDefaultTargetPlatformOverride = TargetPlatform.android;
      final fake = useFakeCelebrationFeedback();
      final settings = await loadSettings(tester);
      for (final (event, sound) in [
        (streak, CelebrationSound.streak),
        (goal, CelebrationSound.goal),
        (level, CelebrationSound.level),
        (
          AchievementCelebration(tier('answers', 1000)),
          CelebrationSound.achievement,
        ),
      ]) {
        fake.sounds.clear();
        fake.haptics.clear();
        await tester.pumpWidget(
          pageApp(event, settings, reduce: event == level),
        );
        await tester.pump(const Duration(seconds: 4));
        expect(fake.sounds, [sound]);
        expect(fake.haptics, [CelebrationHaptic.medium]);
        await tester.pumpWidget(const SizedBox());
      }
      // Off (spec): nothing.
      await settings.setCelebrationSounds(false);
      fake.sounds.clear();
      fake.haptics.clear();
      await tester.pumpWidget(pageApp(level, settings));
      await tester.pump(const Duration(seconds: 4));
      expect(fake.sounds, isEmpty);
      expect(fake.haptics, isEmpty);
      debugDefaultTargetPlatformOverride = null;
    });
  });
}
