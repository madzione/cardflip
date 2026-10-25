import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/data/scope.dart';
import 'package:cardflip/features/progress/celebration.dart';

import 'study_test.dart' show FakeStudy, pumpStudy, studyCard;
import 'support/fake_go.dart';

void main() {
  /// Studies one card of a session, with [before] as the progress when the
  /// session starts and [after] once the card is answered.
  Future<AppState> studyOne(
    WidgetTester tester,
    FakeGo go, {
    required Map<String, Object?> before,
    required Map<String, Object?> after,
  }) async {
    FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
    go.progress = before;
    final app = await pumpStudy(tester, loadSettings: true);
    go.progress = after;
    await tester.tap(find.text('Show answer'));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('rate-3')));
    await tester.pumpAndSettle();
    expect(find.text('Session complete'), findsOneWidget);
    return app;
  }

  FakeGo withGoal(String kind, int value) {
    final go = FakeGo();
    go.settings = {...go.settings, 'dailyGoalKind': kind, 'dailyGoal': value};
    return go;
  }

  const suggestion = 'A daily goal shows how far you are each day.';

  testWidgets('streak extended', (tester) async {
    // Streak extended, and Finish with streak and goal (specs): the first
    // session of the day makes the streak 13 days and reaches the goal.
    await studyOne(
      tester,
      withGoal('cards', 1),
      before: progressJson(current: 12, goalKind: 'cards', goal: 1),
      after: progressJson(current: 13, cards: 1, goalKind: 'cards', goal: 1),
    );
    expect(find.text('Streak extended to 13 days'), findsOneWidget);
    expect(find.text('Daily goal reached'), findsOneWidget);
  });

  testWidgets('goal left', (tester) async {
    // Goal left (spec): 18 of 30 cards; today was already a study day.
    await studyOne(
      tester,
      withGoal('cards', 30),
      before: progressJson(current: 12, cards: 17, goalKind: 'cards', goal: 30),
      after: progressJson(current: 12, cards: 18, goalKind: 'cards', goal: 30),
    );
    expect(find.text('Current streak: 12 days'), findsOneWidget);
    expect(find.text("12 more cards to reach today's goal"), findsOneWidget);
    expect(find.textContaining('extended'), findsNothing);
  });

  testWidgets('goal reached', (tester) async {
    // Goal reached (spec): 34 of 30 cards.
    await studyOne(
      tester,
      withGoal('cards', 30),
      before: progressJson(current: 12, cards: 33, goalKind: 'cards', goal: 30),
      after: progressJson(current: 12, cards: 34, goalKind: 'cards', goal: 30),
    );
    expect(find.text('Daily goal reached'), findsOneWidget);
    expect(find.textContaining('more cards'), findsNothing);
  });

  testWidgets('minutes left', (tester) async {
    await studyOne(
      tester,
      withGoal('minutes', 15),
      before: progressJson(current: 2, cards: 3, goalKind: 'minutes', goal: 15),
      after: progressJson(
        current: 2,
        cards: 4,
        minutes: 14,
        goalKind: 'minutes',
        goal: 15,
      ),
    );
    expect(find.text("1 more minute to reach today's goal"), findsOneWidget);
  });

  testWidgets('suggestion dismissed', (tester) async {
    // Dismissed (spec): no goal; once dismissed, the suggestion is gone.
    final go = FakeGo();
    await studyOne(
      tester,
      go,
      before: progressJson(current: 12),
      after: progressJson(current: 13, cards: 1),
    );
    expect(find.text('Streak extended to 13 days'), findsOneWidget);
    expect(find.text(suggestion), findsOneWidget);
    await tester.tap(find.text('No thanks'));
    await tester.pumpAndSettle();
    expect(go.callsTo('app.GoalSuggestionShown'), hasLength(1));
    expect(go.settings['goalSuggestionShown'], isTrue);
    expect(find.text(suggestion), findsNothing);
  });

  testWidgets('the suggestion opens the goal setting', (tester) async {
    final go = FakeGo();
    await studyOne(
      tester,
      go,
      before: progressJson(current: 1, cards: 1),
      after: progressJson(current: 1, cards: 2),
    );
    await tester.tap(find.text('Set a daily goal'));
    await tester.pumpAndSettle();
    expect(go.callsTo('app.GoalSuggestionShown'), hasLength(1));
    await tester.tap(find.text('Cards'));
    await tester.pumpAndSettle();
    await tester.enterText(find.byKey(const Key('daily-goal-value')), '25');
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    expect(go.callsTo('app.SetDailyGoal'), [
      ['cards', 25],
    ]);
    expect(find.text(suggestion), findsNothing);
  });

  testWidgets('a suggestion seen before is not shown', (tester) async {
    final go = FakeGo();
    go.settings = {...go.settings, 'goalSuggestionShown': true};
    await studyOne(
      tester,
      go,
      before: progressJson(current: 1, cards: 1),
      after: progressJson(current: 1, cards: 2),
    );
    expect(find.text('Current streak: 1 day'), findsOneWidget);
    expect(find.text(suggestion), findsNothing);
  });

  testWidgets('hidden', (tester) async {
    // Hidden (spec): only cards studied and time spent.
    final go = FakeGo();
    go.settings = {...go.settings, 'showProgress': false};
    await studyOne(
      tester,
      go,
      before: progressJson(current: 12),
      after: progressJson(current: 13, cards: 1),
    );
    expect(find.byKey(const Key('progress-lines')), findsNothing);
    expect(
      find.textContaining(RegExp('streak|goal', caseSensitive: false)),
      findsNothing,
    );
    expect(go.callsTo('progress.Today'), isEmpty);
  });

  testWidgets('nothing due with a goal', (tester) async {
    // Nothing due with a goal (study-session spec).
    final go = withGoal('cards', 30);
    FakeStudy(go, []);
    go.progress = progressJson(
      current: 12,
      cards: 18,
      goalKind: 'cards',
      goal: 30,
    );
    await pumpStudy(tester, loadSettings: true);
    expect(
      find.text('Congratulations! Nothing is due right now.'),
      findsOneWidget,
    );
    expect(find.text('Current streak: 12 days'), findsOneWidget);
    expect(find.text("12 more cards to reach today's goal"), findsOneWidget);
  });

  testWidgets('new achievement on the summary', (tester) async {
    // Earned during the session (study-session spec): the 1,000th answer
    // is named on the summary, nothing appeared while studying, and the
    // tier is marked shown, so the deck list does not tell again.
    final go = FakeGo();
    final app = await studyOne(
      tester,
      go,
      before: progressJson(current: 12, cards: 5),
      after: progressJson(
        current: 12,
        cards: 6,
        newAchievements: [newAchievementJson('answers', 1000)],
      ),
    );
    expect(find.text('New: 1,000 answers'), findsOneWidget);
    expect(find.byType(SnackBar), findsNothing);
    expect(go.callsTo('progress.MarkSeen'), [
      [
        ['answers:1000'],
      ],
    ]);
    expect(app.progress.newAchievements, isEmpty);
    // The line stays after the progress reloads without the tier.
    await app.progress.load();
    await tester.pumpAndSettle();
    expect(find.text('New: 1,000 answers'), findsOneWidget);
    expect(go.callsTo('progress.MarkSeen'), hasLength(1));
  });

  testWidgets('several new achievements on the summary', (tester) async {
    // Several earned (study-session spec): one line names both.
    final go = FakeGo();
    await studyOne(
      tester,
      go,
      before: progressJson(current: 99, cards: 5),
      after: progressJson(
        current: 100,
        cards: 6,
        newAchievements: [
          newAchievementJson('answers', 1000),
          newAchievementJson('study-days', 100),
        ],
      ),
    );
    expect(find.text('New: 1,000 answers, 100 study days'), findsOneWidget);
  });

  testWidgets('level reached', (tester) async {
    // Level reached (progress spec): from level 12 to 13 in the session.
    await studyOne(
      tester,
      FakeGo(),
      before: progressJson(current: 12, cards: 5, xp: 3890),
      after: progressJson(current: 12, cards: 6, xp: 3901),
    );
    expect(find.text('Reached level 13'), findsOneWidget);
    expect(find.textContaining('XP to level'), findsNothing);
  });

  testWidgets('XP left', (tester) async {
    // XP left (progress spec): the session did not raise the level.
    await studyOne(
      tester,
      FakeGo(),
      before: progressJson(current: 12, cards: 5, xp: 3749),
      after: progressJson(current: 12, cards: 6, xp: 3760),
    );
    expect(find.text('Level 12 · 140 XP to level 13'), findsOneWidget);
    expect(find.textContaining('Reached'), findsNothing);
  });

  testWidgets('level tier with the level line', (tester) async {
    // Level tier with the level line (study-session spec): the Level 10
    // tier is left out of the new line but still marked shown.
    final go = FakeGo();
    await studyOne(
      tester,
      go,
      before: progressJson(current: 12, cards: 5, xp: 2240),
      after: progressJson(
        current: 12,
        cards: 6,
        xp: 2276,
        newAchievements: [
          newAchievementJson('answers', 1000),
          newAchievementJson('level', 10),
        ],
      ),
    );
    expect(find.text('Reached level 10'), findsOneWidget);
    expect(find.text('New: 1,000 answers'), findsOneWidget);
    expect(go.callsTo('progress.MarkSeen'), [
      [
        ['answers:1000', 'level:10'],
      ],
    ]);
  });

  testWidgets('a level tier alone is named with the level line', (
    tester,
  ) async {
    final go = FakeGo();
    await studyOne(
      tester,
      go,
      before: progressJson(current: 12, cards: 5, xp: 2240),
      after: progressJson(
        current: 12,
        cards: 6,
        xp: 2251,
        newAchievements: [newAchievementJson('level', 10)],
      ),
    );
    expect(find.text('Reached level 10'), findsOneWidget);
    expect(find.textContaining('New:'), findsNothing);
    expect(go.callsTo('progress.MarkSeen'), hasLength(1));
  });

  testWidgets('nothing due shows the XP left', (tester) async {
    final go = FakeGo();
    FakeStudy(go, []);
    go.progress = progressJson(current: 12, xp: 3760);
    await pumpStudy(tester, loadSettings: true);
    expect(find.text('Level 12 · 140 XP to level 13'), findsOneWidget);
  });

  testWidgets('no new achievements while hidden', (tester) async {
    final go = FakeGo();
    go.settings = {...go.settings, 'showProgress': false};
    await studyOne(
      tester,
      go,
      before: progressJson(current: 12),
      after: progressJson(
        current: 13,
        cards: 1,
        newAchievements: [newAchievementJson('answers', 1000)],
      ),
    );
    expect(find.textContaining('New:'), findsNothing);
    expect(go.callsTo('progress.MarkSeen'), isEmpty);
  });

  group('celebration', () {
    final celebration = find.byType(Celebration);

    /// Answers the session's only card, with [before] as the progress at
    /// the start and [after] once answered, and shows the first frames of
    /// the summary.
    Future<AppState> finish(
      WidgetTester tester,
      FakeGo go, {
      required Map<String, Object?> before,
      required Map<String, Object?> after,
    }) async {
      FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
      go.progress = before;
      final app = await pumpStudy(tester, loadSettings: true);
      go.progress = after;
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('rate-3')));
      for (var i = 0; i < 5; i++) {
        await tester.pump(const Duration(milliseconds: 20));
      }
      expect(find.text('Session complete'), findsOneWidget);
      return app;
    }

    final reached = (
      before: progressJson(current: 3, cards: 29, goalKind: 'cards', goal: 30),
      after: progressJson(current: 3, cards: 30, goalKind: 'cards', goal: 30),
    );

    testWidgets('the goal reached in the session plays once', (tester) async {
      // Goal reached (progress spec): a short celebration, once.
      final app = await finish(
        tester,
        withGoal('cards', 30),
        before: reached.before,
        after: reached.after,
      );
      expect(find.text('Daily goal reached'), findsOneWidget);
      expect(celebration, findsOneWidget);
      expect(
        tester
            .widget<Icon>(
              find.descendant(of: celebration, matching: find.byType(Icon)),
            )
            .icon,
        Icons.flag_outlined,
      );
      // At most two seconds.
      await tester.pump(const Duration(seconds: 2));
      expect(celebration, findsNothing);
      // Not again when the progress reloads.
      await app.progress.load();
      await tester.pumpAndSettle();
      expect(celebration, findsNothing);
      expect(find.text('Daily goal reached'), findsOneWidget);
    });

    testWidgets('a new achievement plays it, and buttons stay usable', (
      tester,
    ) async {
      final go = FakeGo();
      await finish(
        tester,
        go,
        before: progressJson(current: 12, cards: 5),
        after: progressJson(
          current: 12,
          cards: 6,
          newAchievements: [newAchievementJson('answers', 1000)],
        ),
      );
      expect(celebration, findsOneWidget);
      expect(
        tester
            .widget<Icon>(
              find.descendant(of: celebration, matching: find.byType(Icon)),
            )
            .icon,
        Icons.military_tech_outlined,
      );
      // The goal suggestion's buttons lie under the celebration.
      expect(
        tester
            .getRect(celebration)
            .overlaps(tester.getRect(find.text('No thanks'))),
        isTrue,
      );
      await tester.tap(find.text('No thanks'));
      await tester.pump();
      expect(go.callsTo('app.GoalSuggestionShown'), hasLength(1));
      expect(celebration, findsOneWidget);
      await tester.pumpAndSettle();
    });

    testWidgets('a level-up plays it', (tester) async {
      // Level reached (progress spec): a short celebration, once.
      final app = await finish(
        tester,
        FakeGo(),
        before: progressJson(current: 12, cards: 5, xp: 3890),
        after: progressJson(current: 12, cards: 6, xp: 3901),
      );
      expect(celebration, findsOneWidget);
      expect(
        tester
            .widget<Icon>(
              find.descendant(of: celebration, matching: find.byType(Icon)),
            )
            .icon,
        Icons.trending_up,
      );
      await tester.pump(const Duration(seconds: 2));
      expect(celebration, findsNothing);
      await app.progress.load();
      await tester.pumpAndSettle();
      expect(celebration, findsNothing);
      expect(find.text('Reached level 13'), findsOneWidget);
    });

    testWidgets('the goal comes before a level-up', (tester) async {
      await finish(
        tester,
        withGoal('cards', 30),
        before: progressJson(
          current: 3,
          cards: 29,
          goalKind: 'cards',
          goal: 30,
          xp: 140,
        ),
        after: progressJson(
          current: 3,
          cards: 30,
          goalKind: 'cards',
          goal: 30,
          xp: 161,
        ),
      );
      expect(find.text('Reached level 3'), findsOneWidget);
      expect(
        tester
            .widget<Icon>(
              find.descendant(of: celebration, matching: find.byType(Icon)),
            )
            .icon,
        Icons.flag_outlined,
      );
      await tester.pumpAndSettle();
    });

    testWidgets('reduce motion: a level-up without an animation', (
      tester,
    ) async {
      tester.platformDispatcher.accessibilityFeaturesTestValue =
          const FakeAccessibilityFeatures(disableAnimations: true);
      addTearDown(
        tester.platformDispatcher.clearAccessibilityFeaturesTestValue,
      );
      await finish(
        tester,
        FakeGo(),
        before: progressJson(current: 12, cards: 5, xp: 3890),
        after: progressJson(current: 12, cards: 6, xp: 3901),
      );
      expect(find.text('Reached level 13'), findsOneWidget);
      expect(celebration, findsNothing);
    });

    testWidgets('reduce motion: the line without an animation', (tester) async {
      // Reduce motion (progress spec).
      tester.platformDispatcher.accessibilityFeaturesTestValue =
          const FakeAccessibilityFeatures(disableAnimations: true);
      addTearDown(
        tester.platformDispatcher.clearAccessibilityFeaturesTestValue,
      );
      await finish(
        tester,
        withGoal('cards', 30),
        before: reached.before,
        after: reached.after,
      );
      expect(find.text('Daily goal reached'), findsOneWidget);
      expect(celebration, findsNothing);
    });

    testWidgets('not when the goal was reached earlier', (tester) async {
      // Goal already reached earlier (progress spec): the evening session.
      await finish(
        tester,
        withGoal('cards', 30),
        before: progressJson(
          current: 3,
          cards: 31,
          goalKind: 'cards',
          goal: 30,
        ),
        after: progressJson(current: 3, cards: 32, goalKind: 'cards', goal: 30),
      );
      expect(find.text('Daily goal reached'), findsOneWidget);
      expect(celebration, findsNothing);
    });

    testWidgets('not while streaks and goals are hidden', (tester) async {
      final go = withGoal('cards', 30);
      go.settings = {...go.settings, 'showProgress': false};
      await finish(tester, go, before: reached.before, after: reached.after);
      expect(find.text('Daily goal reached'), findsNothing);
      expect(celebration, findsNothing);
    });

    testWidgets('not during studying', (tester) async {
      final go = withGoal('cards', 2);
      FakeStudy(go, [
        studyCard(1, 'hola', 'hello'),
        studyCard(2, 'uno', 'one'),
      ]);
      go.progress = progressJson(
        current: 3,
        cards: 1,
        goalKind: 'cards',
        goal: 2,
      );
      await pumpStudy(tester, loadSettings: true);
      go.progress = progressJson(
        current: 3,
        cards: 2,
        goalKind: 'cards',
        goal: 2,
      );
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('rate-3')));
      await tester.pump();
      expect(find.text('Session complete'), findsNothing);
      expect(celebration, findsNothing);
      await tester.pumpAndSettle();
    });
  });
}
