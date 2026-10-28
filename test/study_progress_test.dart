import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter/semantics.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/data/scope.dart';
import 'package:cardflip/features/progress/achievement_badge.dart';
import 'package:cardflip/features/progress/achievements_screen.dart';
import 'package:cardflip/features/progress/celebration_feedback.dart';
import 'package:cardflip/features/study/study_screen.dart';

import 'study_test.dart'
    show
        FakeStudy,
        audioCard,
        fakeMedia,
        inAMinute,
        noLearnAhead,
        pumpStudy,
        studyCard,
        waitForLearningCard;
import 'support/fake_celebration_feedback.dart';
import 'support/fake_go.dart';

/// The headlines of the celebration pages, which Continue goes through
/// to the summary.
Future<List<String>> continueToSummary(WidgetTester tester) async {
  final headlines = <String>[];
  final next = find.byKey(const Key('celebration-continue'));
  while (next.evaluate().isNotEmpty) {
    headlines.add(
      tester.widget<Text>(find.byKey(const Key('celebration-headline'))).data!,
    );
    await tester.pump(const Duration(milliseconds: 600));
    await tester.tap(next);
    await tester.pumpAndSettle();
  }
  return headlines;
}

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
    await continueToSummary(tester);
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

  group('reward card', () {
    final card = find.byKey(const Key('reward-card'));
    final celebration = find.byKey(const Key('celebration'));

    /// Answers the session's only card, with [before] as the progress at
    /// the start and [after] once answered, goes through the celebration
    /// pages and shows the summary.
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
      await continueToSummary(tester);
      expect(find.text('Session complete'), findsOneWidget);
      return app;
    }

    Finder inCard(Finder f) => find.descendant(of: card, matching: f);

    /// The number in the level ring.
    String ringLevel(WidgetTester tester) => tester
        .widget<Text>(
          find.descendant(
            of: find.byKey(const Key('reward-level')),
            matching: find.byType(Text),
          ),
        )
        .data!;

    final reached = (
      before: progressJson(current: 3, cards: 29, goalKind: 'cards', goal: 30),
      after: progressJson(current: 3, cards: 30, goalKind: 'cards', goal: 30),
    );

    testWidgets('level and badge', (tester) async {
      // Level and badge (progress spec): from level 12 to 13 with Answers
      // 1,000.
      await finish(
        tester,
        FakeGo(),
        before: progressJson(current: 12, cards: 5, xp: 3890),
        after: progressJson(
          current: 12,
          cards: 6,
          xp: 3901,
          newAchievements: [newAchievementJson('answers', 1000)],
        ),
      );
      await tester.pumpAndSettle();
      expect(ringLevel(tester), '13');
      expect(inCard(find.text('Reached level 13')), findsOneWidget);
      expect(inCard(find.text('New: 1,000 answers')), findsOneWidget);
      final badge = tester.widget<AchievementBadge>(
        inCard(find.byType(AchievementBadge)),
      );
      expect((badge.id, badge.tier, badge.stars), ('answers', 'silver', 2));
      // Above the lines, which no longer say it.
      expect(find.text('Reached level 13'), findsOneWidget);
      expect(
        tester.getRect(card).bottom,
        lessThanOrEqualTo(
          tester.getRect(find.text('Current streak: 12 days')).top,
        ),
      );
    });

    testWidgets('many tiers', (tester) async {
      // Many tiers (progress spec): three badges and "+2", all five named.
      await finish(
        tester,
        FakeGo(),
        before: progressJson(current: 12, cards: 5),
        after: progressJson(
          current: 12,
          cards: 6,
          newAchievements: [
            newAchievementJson('answers', 1000),
            newAchievementJson('streak', 30),
            newAchievementJson('study-days', 100),
            newAchievementJson('study-time', 10),
            newAchievementJson('comeback', 1),
          ],
        ),
      );
      await tester.pumpAndSettle();
      expect(inCard(find.byType(AchievementBadge)), findsNWidgets(3));
      expect(inCard(find.text('+2')), findsOneWidget);
      expect(
        inCard(
          find.text(
            'New: 1,000 answers, 30-day streak, 100 study days, '
            '10 hours of study, Comeback',
          ),
        ),
        findsOneWidget,
      );
    });

    testWidgets('tapping opens the achievements', (tester) async {
      // Open achievements (progress spec).
      await finish(
        tester,
        withGoal('cards', 30),
        before: reached.before,
        after: reached.after,
      );
      await tester.pumpAndSettle();
      expect(
        tester.getSemantics(card),
        isSemantics(
          isButton: true,
          hasTapAction: true,
          label: 'Daily goal reached',
        ),
      );
      await tester.tap(card);
      await tester.pumpAndSettle();
      expect(find.byType(AchievementsScreen), findsOneWidget);
    });

    testWidgets('nothing earned', (tester) async {
      // Nothing earned (progress spec).
      await finish(
        tester,
        FakeGo(),
        before: progressJson(current: 12, cards: 5, xp: 3749),
        after: progressJson(current: 12, cards: 6, xp: 3760),
      );
      await tester.pumpAndSettle();
      expect(card, findsNothing);
      expect(celebration, findsNothing);
    });

    testWidgets('the goal reached in an earlier session is a line', (
      tester,
    ) async {
      // Goal reached and Goal already reached earlier (progress spec).
      await finish(
        tester,
        withGoal('cards', 30),
        before: progressJson(
          current: 3,
          cards: 33,
          goalKind: 'cards',
          goal: 30,
        ),
        after: progressJson(current: 3, cards: 34, goalKind: 'cards', goal: 30),
      );
      await tester.pumpAndSettle();
      expect(card, findsNothing);
      expect(find.text('Daily goal reached'), findsOneWidget);
    });

    testWidgets('the goal reached in the session is not a line too', (
      tester,
    ) async {
      // Goal reached in the session (progress spec).
      await finish(
        tester,
        withGoal('cards', 30),
        before: progressJson(
          current: 3,
          cards: 25,
          goalKind: 'cards',
          goal: 30,
        ),
        after: progressJson(current: 3, cards: 34, goalKind: 'cards', goal: 30),
      );
      await tester.pumpAndSettle();
      expect(inCard(find.text('Daily goal reached')), findsOneWidget);
      expect(find.text('Daily goal reached'), findsOneWidget);
      expect(inCard(find.byKey(const Key('reward-goal'))), findsOneWidget);
    });

    testWidgets('not on the nothing-due screen', (tester) async {
      final go = withGoal('cards', 30);
      FakeStudy(go, []);
      go.progress = progressJson(
        current: 3,
        cards: 34,
        goalKind: 'cards',
        goal: 30,
        newAchievements: [newAchievementJson('answers', 1000)],
      );
      await pumpStudy(tester, loadSettings: true);
      expect(
        find.text('Congratulations! Nothing is due right now.'),
        findsOneWidget,
      );
      expect(card, findsNothing);
      expect(find.text('Daily goal reached'), findsOneWidget);
      expect(go.callsTo('progress.MarkSeen'), isEmpty);
    });

    testWidgets('a static recap after the sequence', (tester) async {
      // After the sequence (progress spec): the card in its final state,
      // without filling rings or a particle burst.
      final go = withGoal('cards', 30);
      final app = await finish(
        tester,
        go,
        before: progressJson(
          current: 3,
          cards: 29,
          goalKind: 'cards',
          goal: 30,
          xp: 3890,
        ),
        after: progressJson(
          current: 3,
          cards: 30,
          goalKind: 'cards',
          goal: 30,
          xp: 3901,
        ),
      );
      expect(card, findsOneWidget);
      expect(ringLevel(tester), '13');
      expect(inCard(find.text('Daily goal reached')), findsOneWidget);
      expect(celebration, findsNothing);
      expect(tester.binding.hasScheduledFrame, isFalse);
      // Not again when the progress reloads.
      await app.progress.load();
      await tester.pumpAndSettle();
      expect(celebration, findsNothing);
      expect(find.byKey(const Key('celebration-continue')), findsNothing);
    });

    testWidgets('the goal suggestion stays usable', (tester) async {
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
      await tester.tap(find.text('No thanks'));
      await tester.pump();
      expect(go.callsTo('app.GoalSuggestionShown'), hasLength(1));
      await tester.pumpAndSettle();
    });

    testWidgets('reduce motion: the final state', (tester) async {
      // Reduce motion (progress spec): the full goal ring, the new level
      // and "Daily goal reached", without an animation.
      tester.platformDispatcher.accessibilityFeaturesTestValue =
          const FakeAccessibilityFeatures(disableAnimations: true);
      addTearDown(
        tester.platformDispatcher.clearAccessibilityFeaturesTestValue,
      );
      await finish(
        tester,
        withGoal('cards', 30),
        before: progressJson(
          current: 3,
          cards: 29,
          goalKind: 'cards',
          goal: 30,
          xp: 3890,
        ),
        after: progressJson(
          current: 3,
          cards: 30,
          goalKind: 'cards',
          goal: 30,
          xp: 3901,
        ),
      );
      expect(inCard(find.text('Daily goal reached')), findsOneWidget);
      expect(ringLevel(tester), '13');
      expect(celebration, findsNothing);
      expect(tester.binding.hasScheduledFrame, isFalse);
    });

    testWidgets('not while streaks and goals are hidden', (tester) async {
      final go = withGoal('cards', 30);
      go.settings = {...go.settings, 'showProgress': false};
      await finish(tester, go, before: reached.before, after: reached.after);
      expect(card, findsNothing);
      expect(celebration, findsNothing);
      expect(find.text('Daily goal reached'), findsNothing);
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
      expect(card, findsNothing);
      expect(celebration, findsNothing);
      await tester.pumpAndSettle();
    });
  });

  group('reward banner', () {
    final banner = find.byKey(const Key('reward-banner'));

    /// Starts a session of three cards with [before] as the progress.
    Future<AppState> start(
      WidgetTester tester,
      FakeGo go,
      Map<String, Object?> before,
    ) {
      FakeStudy(go, [
        studyCard(1, 'hola', 'hello'),
        studyCard(2, 'uno', 'one'),
        studyCard(3, 'dos', 'two'),
      ]);
      go.progress = before;
      return pumpStudy(tester, loadSettings: true);
    }

    /// Reloads the progress, as the deck list does after the collection
    /// changed with an answer or undo, without letting a banner finish.
    Future<void> changed(WidgetTester tester) async {
      AppScope.read(tester.element(find.byType(Scaffold).first)).progress
          .load();
      for (var i = 0; i < 5; i++) {
        await tester.pump(const Duration(milliseconds: 20));
      }
    }

    /// Answers the card on screen Good, with [after] as the progress then.
    Future<void> answer(
      WidgetTester tester,
      FakeGo go,
      Map<String, Object?> after,
    ) async {
      go.progress = after;
      await tester.tap(find.text('Show answer'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 20));
      await tester.tap(find.byKey(const Key('rate-3')));
      await tester.pump();
      await changed(tester);
    }

    Map<String, Object?> goal(int cards, {int xp = 0}) => progressJson(
      current: 3,
      cards: cards,
      goalKind: 'cards',
      goal: 30,
      xp: xp,
    );

    String bannerText(WidgetTester tester) => tester
        .widget<Text>(find.descendant(of: banner, matching: find.byType(Text)))
        .data!;

    testWidgets('the goal reached', (tester) async {
      // Goal reached (progress spec): "Daily goal reached" for about 2.5
      // seconds, while the next card is on screen.
      final go = withGoal('cards', 30);
      await start(tester, go, goal(29));
      await answer(tester, go, goal(30));
      expect(banner, findsOneWidget);
      expect(bannerText(tester), 'Daily goal reached');
      expect(find.text('Show answer'), findsOneWidget);
      // Above the answer buttons, inside the card area.
      expect(
        tester.getRect(banner).bottom,
        lessThan(tester.getRect(find.text('Show answer')).top),
      );
      await tester.pump(const Duration(seconds: 2));
      expect(banner, findsOneWidget);
      await tester.pump(const Duration(milliseconds: 600));
      expect(banner, findsNothing);
    });

    testWidgets('level and tier from one answer', (tester) async {
      // Level and tier from one answer (progress spec): with Show level-ups
      // while studying off, one banner, the Level 10 tier left out, the
      // Answers badge.
      final go = FakeGo();
      go.settings = {...go.settings, 'showLevelUpsWhileStudying': false};
      await start(tester, go, progressJson(current: 3, cards: 5, xp: 2240));
      await answer(
        tester,
        go,
        progressJson(
          current: 3,
          cards: 6,
          xp: 2276,
          newAchievements: [
            newAchievementJson('level', 10),
            newAchievementJson('answers', 1000),
          ],
        ),
      );
      expect(banner, findsOneWidget);
      expect(bannerText(tester), 'Level 10 · New: 1,000 answers');
      expect(
        tester
            .widget<AchievementBadge>(
              find.descendant(
                of: banner,
                matching: find.byType(AchievementBadge),
              ),
            )
            .id,
        'answers',
      );
      // The tiers belong to the session: marked shown at once.
      expect(go.callsTo('progress.MarkSeen'), [
        [
          ['level:10', 'answers:1000'],
        ],
      ]);
      await tester.pumpAndSettle();
    });

    testWidgets('a tier shown by a banner is named on the card', (
      tester,
    ) async {
      // Already shown by a banner (study-session spec): the summary's
      // reward card still names it, and the deck list does not.
      final go = FakeGo();
      FakeStudy(go, [
        studyCard(1, 'hola', 'hello'),
        studyCard(2, 'uno', 'one'),
      ]);
      go.progress = progressJson(current: 3, cards: 5);
      final app = await pumpStudy(tester, loadSettings: true);
      await answer(
        tester,
        go,
        progressJson(
          current: 3,
          cards: 6,
          newAchievements: [newAchievementJson('answers', 1000)],
        ),
      );
      expect(bannerText(tester), 'New: 1,000 answers');
      await tester.pumpAndSettle();
      await answer(tester, go, progressJson(current: 3, cards: 7));
      await tester.pumpAndSettle();
      // Banner already shown (progress spec): celebrated all the same.
      expect(await continueToSummary(tester), ['New achievement']);
      expect(find.text('Session complete'), findsOneWidget);
      expect(
        find.descendant(
          of: find.byKey(const Key('reward-card')),
          matching: find.text('New: 1,000 answers'),
        ),
        findsOneWidget,
      );
      expect(go.callsTo('progress.MarkSeen'), hasLength(1));
      expect(app.progress.newAchievements, isEmpty);
    });

    testWidgets('undone and answered again', (tester) async {
      // Undone and answered again (progress spec): no second banner.
      final go = withGoal('cards', 30);
      await start(tester, go, goal(29));
      await answer(tester, go, goal(30));
      expect(banner, findsOneWidget);
      await tester.pumpAndSettle();
      expect(banner, findsNothing);
      go.progress = goal(29);
      await tester.tap(find.widgetWithIcon(IconButton, Icons.undo));
      await tester.pump();
      await changed(tester);
      expect(banner, findsNothing);
      await answer(tester, go, goal(30));
      expect(banner, findsNothing);
      await tester.pumpAndSettle();
    });

    testWidgets('nothing new', (tester) async {
      // Nothing new (progress spec): the goal was reached before.
      final go = withGoal('cards', 30);
      await start(tester, go, goal(31));
      await answer(tester, go, goal(32));
      expect(banner, findsNothing);
      await tester.pumpAndSettle();
    });

    testWidgets('none while rewards while studying are off', (tester) async {
      // Rewards while studying off (progress spec): no banner; the session
      // still ends with the celebration sequence and the reward card.
      final go = withGoal('cards', 30);
      go.settings = {...go.settings, 'showStudyRewards': false};
      FakeStudy(go, [
        studyCard(1, 'hola', 'hello'),
        studyCard(2, 'uno', 'one'),
      ]);
      go.progress = goal(29);
      await pumpStudy(tester, loadSettings: true);
      await answer(tester, go, goal(30));
      expect(banner, findsNothing);
      await answer(tester, go, goal(31));
      await tester.pumpAndSettle();
      expect(await continueToSummary(tester), ['Daily goal reached']);
      expect(find.text('Session complete'), findsOneWidget);
      expect(find.text('Daily goal reached'), findsOneWidget);
    });

    testWidgets('none while streaks and goals are hidden', (tester) async {
      final go = withGoal('cards', 30);
      go.settings = {...go.settings, 'showProgress': false};
      await start(tester, go, goal(29));
      await answer(tester, go, goal(30));
      expect(banner, findsNothing);
      await tester.pumpAndSettle();
    });

    testWidgets('the rating key while shown', (tester) async {
      // Rating key while shown (progress spec): 3 rates Good, and the
      // banner stays until its time is up.
      final go = withGoal('cards', 30);
      final study = FakeStudy(go, [
        studyCard(1, 'hola', 'hello'),
        studyCard(2, 'uno', 'one'),
        studyCard(3, 'dos', 'two'),
      ]);
      go.progress = goal(29);
      await pumpStudy(tester, loadSettings: true);
      await answer(tester, go, goal(30));
      expect(banner, findsOneWidget);
      expect(
        find.descendant(of: banner, matching: find.byType(Focus)),
        findsNothing,
      );
      await tester.sendKeyEvent(LogicalKeyboardKey.space);
      await tester.pump();
      await tester.sendKeyEvent(LogicalKeyboardKey.digit3);
      await tester.pump();
      expect(study.answers.last.take(2), [2, 3]);
      expect(banner, findsOneWidget);
      await tester.pump(const Duration(milliseconds: 500));
      expect(banner, findsOneWidget);
      await tester.pumpAndSettle();
      expect(banner, findsNothing);
    });

    testWidgets('a tap passes through', (tester) async {
      // Tap on the banner (progress spec): the tap reaches the card under
      // it, and the banner stays.
      final go = withGoal('cards', 30);
      await start(tester, go, goal(29));
      await answer(tester, go, goal(30));
      final center = tester.getCenter(banner);
      final path = tester.hitTestOnBinding(center).path.map((e) => e.target);
      final scroll = find.byKey(const ValueKey('scroll-2'));
      expect(path, contains(tester.renderObject(scroll)));
      expect(path, isNot(contains(tester.renderObject(banner))));
      await tester.tapAt(center);
      await tester.pump();
      expect(banner, findsOneWidget);
      await tester.pumpAndSettle();
    });

    testWidgets('screen readers hear it politely', (tester) async {
      final go = withGoal('cards', 30);
      await start(tester, go, goal(29));
      tester.takeAnnouncements();
      await answer(tester, go, goal(30));
      final announcement = tester.takeAnnouncements().single;
      expect(announcement.message, 'Daily goal reached');
      expect(announcement.assertiveness, Assertiveness.polite);
      await tester.pumpAndSettle();
    });

    testWidgets('reduce motion', (tester) async {
      // Reduce motion (progress spec): no sliding or fading, the same time.
      tester.platformDispatcher.accessibilityFeaturesTestValue =
          const FakeAccessibilityFeatures(disableAnimations: true);
      addTearDown(
        tester.platformDispatcher.clearAccessibilityFeaturesTestValue,
      );
      final go = withGoal('cards', 30);
      await start(tester, go, goal(29));
      await answer(tester, go, goal(30));
      expect(banner, findsOneWidget);
      expect(find.byKey(const Key('reward-banner-motion')), findsNothing);
      expect(find.byKey(const Key('celebration')), findsNothing);
      await tester.pump(const Duration(seconds: 2));
      expect(banner, findsOneWidget);
      await tester.pump(const Duration(milliseconds: 600));
      expect(banner, findsNothing);
    });

    testWidgets('the pop-in and the burst', (tester) async {
      // Goal reached (progress spec): the icon pops in with a burst.
      final go = withGoal('cards', 30);
      await start(tester, go, goal(29));
      await answer(tester, go, goal(30));
      final burst = find.descendant(
        of: banner,
        matching: find.byKey(const Key('celebration')),
      );
      expect(burst, findsOneWidget);
      await tester.pumpAndSettle();
    });

    testWidgets('a chime and a light vibration per banner', (tester) async {
      // Banner (progress spec): the chime with the banner, on Android a
      // light vibration too.
      debugDefaultTargetPlatformOverride = TargetPlatform.android;
      final feedback = useFakeCelebrationFeedback();
      final go = withGoal('cards', 30);
      await start(tester, go, goal(29));
      expect(feedback.prepared, 1);
      await answer(tester, go, goal(30));
      expect(feedback.sounds, [CelebrationSound.chime]);
      expect(feedback.haptics, [CelebrationHaptic.light]);
      await tester.pumpAndSettle();
      // Another banner, another sound and vibration.
      await answer(
        tester,
        go,
        progressJson(
          current: 3,
          cards: 31,
          goalKind: 'cards',
          goal: 30,
          newAchievements: [newAchievementJson('answers', 1000)],
        ),
      );
      expect(feedback.sounds, hasLength(2));
      expect(feedback.haptics, [
        CelebrationHaptic.light,
        CelebrationHaptic.light,
      ]);
      await tester.pumpAndSettle();
      debugDefaultTargetPlatformOverride = null;
    });

    testWidgets('the sound of the biggest event', (tester) async {
      // Reward banner sound (progress spec): goal only, goal and tier, and
      // a level in a banner.
      final feedback = useFakeCelebrationFeedback();
      final go = withGoal('cards', 30);
      go.settings = {...go.settings, 'showLevelUpsWhileStudying': false};
      await start(tester, go, goal(29, xp: 3880));
      await answer(
        tester,
        go,
        progressJson(
          current: 3,
          cards: 30,
          goalKind: 'cards',
          goal: 30,
          xp: 3895,
          newAchievements: [newAchievementJson('answers', 1000)],
        ),
      );
      expect(bannerText(tester), 'Daily goal reached · New: 1,000 answers');
      expect(feedback.sounds, [CelebrationSound.achievement]);
      await tester.pumpAndSettle();
      await answer(tester, go, goal(31, xp: 3901));
      expect(bannerText(tester), 'Level 13');
      expect(feedback.sounds.last, CelebrationSound.level);
      await tester.pumpAndSettle();
    });

    testWidgets('the chime for the goal alone', (tester) async {
      final feedback = useFakeCelebrationFeedback();
      final go = withGoal('cards', 30);
      await start(tester, go, goal(29));
      await answer(tester, go, goal(30));
      expect(feedback.sounds, [CelebrationSound.chime]);
      await tester.pumpAndSettle();
    });

    testWidgets('no chime with the setting off', (tester) async {
      final feedback = useFakeCelebrationFeedback();
      final go = withGoal('cards', 30);
      go.settings = {...go.settings, 'celebrationSounds': false};
      await start(tester, go, goal(29));
      await answer(tester, go, goal(30));
      expect(banner, findsOneWidget);
      expect(feedback.sounds, isEmpty);
      expect(feedback.haptics, isEmpty);
      expect(feedback.prepared, 0);
      await tester.pumpAndSettle();
    });
  });

  group('level page during study', () {
    final next = find.byKey(const Key('celebration-continue'));
    final banner = find.byKey(const Key('reward-banner'));
    final timer = find.byKey(const Key('answerTimer'));
    String headline(WidgetTester tester) => tester
        .widget<Text>(find.byKey(const Key('celebration-headline')))
        .data!;

    /// Starts a session of [cards] (three by default) at [xp] XP, level 12
    /// by default.
    Future<(AppState, FakeStudy)> start(
      WidgetTester tester,
      FakeGo go, {
      List<Map<String, Object?>>? cards,
      int xp = 3890,
    }) async {
      final study = FakeStudy(
        go,
        cards ??
            [
              studyCard(1, 'hola', 'hello'),
              studyCard(2, 'uno', 'one'),
              studyCard(3, 'dos', 'two'),
            ],
      );
      go.progress = progressJson(current: 3, cards: 5, xp: xp);
      return (await pumpStudy(tester, loadSettings: true), study);
    }

    /// Answers the card on screen Good with [after] as the progress, which
    /// then reloads.
    Future<void> answer(
      WidgetTester tester,
      AppState app,
      FakeGo go,
      Map<String, Object?> after,
    ) async {
      go.progress = after;
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('rate-3')));
      await tester.pump();
      await app.progress.load();
      await tester.pump();
    }

    /// The progress at level 13, 3,901 XP.
    Map<String, Object?> level13({
      List<Map<String, Object?>> newAchievements = const [],
    }) => progressJson(
      current: 3,
      cards: 6,
      xp: 3901,
      newAchievements: newAchievements,
    );

    testWidgets('the level page before the next card', (tester) async {
      // Level-up mid-session (progress spec).
      final go = FakeGo();
      final (app, _) = await start(tester, go);
      await answer(tester, app, go, level13());
      expect(headline(tester), 'Level 13');
      expect(find.text('Show answer'), findsNothing);
      expect(banner, findsNothing);
      await tester.pump(const Duration(milliseconds: 600));
      await tester.tap(next);
      await tester.pumpAndSettle();
      expect(next, findsNothing);
      expect(find.textContaining('uno', findRichText: true), findsOneWidget);
      expect(find.text('Show answer'), findsOneWidget);
      // No banner names the level either.
      expect(banner, findsNothing);
    });

    testWidgets('the timer starts after the page', (tester) async {
      // Timer starts after the page (progress spec).
      final go = FakeGo();
      final (app, study) = await start(
        tester,
        go,
        cards: [
          studyCard(1, 'hola', 'hello', showTimer: true),
          studyCard(2, 'uno', 'one', showTimer: true),
        ],
      );
      await answer(tester, app, go, level13());
      await tester.pump(const Duration(seconds: 5));
      await tester.tap(next);
      await tester.pumpAndSettle();
      expect(tester.widget<Text>(timer).data, '0:00');
      await tester.pump(const Duration(seconds: 2));
      expect(tester.widget<Text>(timer).data, '0:02');
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('rate-3')));
      await tester.pumpAndSettle();
      expect(study.answers.last[2] as int, lessThan(5000));
    });

    testWidgets('Escape returns to studying', (tester) async {
      // Escape returns to studying (progress spec) and Escape on the level
      // page (study-session spec).
      final go = FakeGo();
      final (app, _) = await start(tester, go);
      await answer(tester, app, go, level13());
      await tester.pump(const Duration(seconds: 1));
      await tester.sendKeyEvent(LogicalKeyboardKey.escape);
      await tester.pumpAndSettle();
      expect(next, findsNothing);
      expect(find.text('Show answer'), findsOneWidget);
      // The study keys work again.
      await tester.sendKeyEvent(LogicalKeyboardKey.space);
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('rate-3')), findsOneWidget);
    });

    testWidgets('system Back returns to studying', (tester) async {
      final go = FakeGo();
      final (app, _) = await start(tester, go);
      await answer(tester, app, go, level13());
      await tester.pump(const Duration(seconds: 1));
      await tester.binding.handlePopRoute();
      await tester.pumpAndSettle();
      expect(next, findsNothing);
      expect(find.text('Show answer'), findsOneWidget);
    });

    testWidgets('undo and rating keys do nothing', (tester) async {
      // Undo on the page (progress spec).
      final go = FakeGo();
      final (app, study) = await start(tester, go);
      await answer(tester, app, go, level13());
      await tester.pump(const Duration(seconds: 1));
      await tester.sendKeyDownEvent(LogicalKeyboardKey.controlLeft);
      await tester.sendKeyEvent(LogicalKeyboardKey.keyZ);
      await tester.sendKeyUpEvent(LogicalKeyboardKey.controlLeft);
      await tester.pumpAndSettle();
      await tester.sendKeyEvent(LogicalKeyboardKey.digit3);
      await tester.pumpAndSettle();
      // A card action key too: Ctrl+N flags.
      await tester.sendKeyDownEvent(LogicalKeyboardKey.controlLeft);
      await tester.sendKeyEvent(LogicalKeyboardKey.digit1);
      await tester.sendKeyUpEvent(LogicalKeyboardKey.controlLeft);
      await tester.pumpAndSettle();
      expect(headline(tester), 'Level 13');
      expect(go.callsTo('study.Undo'), isEmpty);
      expect(go.callsTo('study.SetFlag'), isEmpty);
      expect(study.answers, hasLength(1));
    });

    testWidgets('the answer\'s tier gets a banner after the page', (
      tester,
    ) async {
      // Level and tier from one answer (progress spec): the Level 10 tier
      // left out.
      final go = FakeGo();
      final (app, _) = await start(tester, go, xp: 2240);
      await answer(
        tester,
        app,
        go,
        progressJson(
          current: 3,
          cards: 6,
          xp: 2276,
          newAchievements: [
            newAchievementJson('level', 10),
            newAchievementJson('answers', 1000),
          ],
        ),
      );
      expect(headline(tester), 'Level 10');
      expect(banner, findsNothing);
      await tester.pump(const Duration(milliseconds: 600));
      await tester.tap(next);
      await tester.pump();
      expect(
        tester
            .widget<Text>(
              find.descendant(of: banner, matching: find.byType(Text)),
            )
            .data,
        'New: 1,000 answers',
      );
      await tester.pumpAndSettle();
    });

    testWidgets('not again at the end', (tester) async {
      // Not again at the end (progress spec).
      final go = FakeGo();
      final (app, _) = await start(
        tester,
        go,
        cards: [studyCard(1, 'hola', 'hello'), studyCard(2, 'uno', 'one')],
      );
      await answer(tester, app, go, level13());
      await tester.pump(const Duration(milliseconds: 600));
      await tester.tap(next);
      await tester.pumpAndSettle();
      go.progress = progressJson(current: 3, cards: 7, xp: 3902);
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('rate-3')));
      await tester.pumpAndSettle();
      expect(next, findsNothing);
      expect(find.text('Session complete'), findsOneWidget);
    });

    testWidgets('the last answer: the level page at the end', (tester) async {
      // No next card: the level is part of the celebration sequence.
      final go = FakeGo();
      final (app, _) = await start(
        tester,
        go,
        cards: [studyCard(1, 'hola', 'hello')],
      );
      await answer(tester, app, go, level13());
      await tester.pumpAndSettle();
      expect(await continueToSummary(tester), ['Level 13']);
      expect(find.text('Session complete'), findsOneWidget);
    });

    testWidgets('the last answer\'s progress before its result', (
      tester,
    ) async {
      // The progress reloads while the answer's call still runs: the level
      // is still part of the celebration sequence.
      final go = FakeGo();
      final (app, study) = await start(
        tester,
        go,
        cards: [studyCard(1, 'hola', 'hello')],
      );
      final gate = study.answerGate = Completer<void>();
      go.progress = level13();
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('rate-3')));
      await tester.pump();
      await app.progress.load();
      await tester.pump();
      expect(find.byKey(const Key('celebration-headline')), findsNothing);
      gate.complete();
      await tester.pumpAndSettle();
      expect(await continueToSummary(tester), ['Level 13']);
    });

    testWidgets('the progress before the next card', (tester) async {
      // The level page still comes, and the next card waits for it.
      final go = FakeGo();
      final (app, study) = await start(tester, go);
      final gate = study.answerGate = Completer<void>();
      await answer(tester, app, go, level13());
      expect(find.byKey(const Key('celebration-headline')), findsNothing);
      gate.complete();
      await tester.pump();
      await tester.pump();
      expect(headline(tester), 'Level 13');
      await tester.pump(const Duration(milliseconds: 600));
      await tester.tap(next);
      await tester.pumpAndSettle();
      expect(find.textContaining('uno', findRichText: true), findsOneWidget);
    });

    testWidgets('setting off: a banner, the page at the end', (tester) async {
      // Setting off (progress spec).
      final go = FakeGo();
      go.settings = {...go.settings, 'showLevelUpsWhileStudying': false};
      final (app, _) = await start(
        tester,
        go,
        cards: [studyCard(1, 'hola', 'hello'), studyCard(2, 'uno', 'one')],
      );
      await answer(tester, app, go, level13());
      expect(next, findsNothing);
      expect(
        find.descendant(of: banner, matching: find.text('Level 13')),
        findsOneWidget,
      );
      await tester.pumpAndSettle();
      go.progress = progressJson(current: 3, cards: 7, xp: 3902);
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('rate-3')));
      await tester.pumpAndSettle();
      expect(await continueToSummary(tester), ['Level 13']);
    });

    testWidgets('the level sound and one medium vibration', (tester) async {
      // Level page (progress spec, "Celebration sounds and vibration").
      debugDefaultTargetPlatformOverride = TargetPlatform.android;
      final feedback = useFakeCelebrationFeedback();
      final go = FakeGo();
      final (app, _) = await start(tester, go);
      await answer(tester, app, go, level13());
      await tester.pump(const Duration(seconds: 1));
      expect(feedback.sounds, [CelebrationSound.level]);
      expect(feedback.haptics, [CelebrationHaptic.medium]);
      await tester.tap(next);
      await tester.pumpAndSettle();
      expect(feedback.sounds, [CelebrationSound.level]);
      debugDefaultTargetPlatformOverride = null;
    });

    testWidgets('the card audio waits for the page', (tester) async {
      final backend = fakeMedia();
      final go = FakeGo();
      final (app, _) = await start(
        tester,
        go,
        cards: [audioCard(1), audioCard(2)],
      );
      expect(backend.played, ['hola.mp3']);
      await answer(tester, app, go, level13());
      final played = backend.played.length;
      expect(backend.playing, isNull);
      await tester.pump(const Duration(seconds: 1));
      expect(backend.played, hasLength(played));
      await tester.tap(next);
      await tester.pumpAndSettle();
      expect(backend.played, hasLength(played + 1));
      expect(backend.played.last, 'hola.mp3');
    });

    testWidgets('a timebox message waits for the page', (tester) async {
      final go = FakeGo();
      go.settings = {...go.settings, 'timeboxMinutes': 1};
      final (app, _) = await start(tester, go);
      await tester.pump(const Duration(seconds: 61));
      await answer(tester, app, go, level13());
      await tester.pumpAndSettle();
      expect(headline(tester), 'Level 13');
      expect(find.byKey(const Key('timeboxMessage')), findsNothing);
      await tester.tap(next);
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('timeboxMessage')), findsOneWidget);
    });
  });

  group('leaving with Back', () {
    final next = find.byKey(const Key('celebration-continue'));
    final under = find.byKey(const Key('under-study'));
    String headline(WidgetTester tester) => tester
        .widget<Text>(find.byKey(const Key('celebration-headline')))
        .data!;

    /// Starts a session of two cards over another screen, at [before], and
    /// answers the first Good with [after] as the progress.
    Future<AppState> answerOne(
      WidgetTester tester,
      FakeGo go, {
      required Map<String, Object?> before,
      required Map<String, Object?> after,
    }) async {
      FakeStudy(go, [
        studyCard(1, 'hola', 'hello'),
        studyCard(2, 'uno', 'one'),
      ]);
      go.progress = before;
      final app = await pumpStudy(tester, loadSettings: true, pushed: true);
      go.progress = after;
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('rate-3')));
      await tester.pump();
      await app.progress.load();
      await tester.pumpAndSettle();
      return app;
    }

    Map<String, Object?> goal(int cards, {int xp = 3700}) => progressJson(
      current: 3,
      cards: cards,
      goalKind: 'cards',
      goal: 30,
      xp: xp,
    );

    testWidgets('the goal page, then the deck list', (tester) async {
      // Goal reached, left with Back (progress spec).
      final go = withGoal('cards', 30);
      await answerOne(tester, go, before: goal(29), after: goal(30));
      await tester.binding.handlePopRoute();
      await tester.pumpAndSettle();
      expect(headline(tester), 'Daily goal reached');
      expect(under, findsNothing);
      await tester.tap(next);
      await tester.pumpAndSettle();
      expect(under, findsOneWidget);
      expect(find.byType(StudyScreen), findsNothing);
    });

    testWidgets('Escape on the first of two pages leaves', (tester) async {
      // Skip after Back (progress spec): the goal and level pages.
      final go = withGoal('cards', 30);
      go.settings = {...go.settings, 'showLevelUpsWhileStudying': false};
      await answerOne(
        tester,
        go,
        before: goal(29, xp: 3890),
        after: goal(30, xp: 3901),
      );
      await tester.sendKeyEvent(LogicalKeyboardKey.escape);
      await tester.pumpAndSettle();
      expect(headline(tester), 'Daily goal reached');
      await tester.pump(const Duration(milliseconds: 600));
      await tester.sendKeyEvent(LogicalKeyboardKey.escape);
      await tester.pumpAndSettle();
      expect(under, findsOneWidget);
      expect(find.byType(StudyScreen), findsNothing);
    });

    testWidgets('nothing earned: left at once', (tester) async {
      // Nothing earned (progress spec, "Celebrations when leaving with
      // Back").
      final go = withGoal('cards', 30);
      await answerOne(tester, go, before: goal(10), after: goal(11));
      await tester.binding.handlePopRoute();
      await tester.pump();
      expect(next, findsNothing);
      await tester.pumpAndSettle();
      expect(under, findsOneWidget);
    });

    testWidgets('a level a banner showed gets its page', (tester) async {
      // Banner shown, left with Back (progress spec).
      final go = FakeGo();
      go.settings = {...go.settings, 'showLevelUpsWhileStudying': false};
      await answerOne(
        tester,
        go,
        before: progressJson(current: 3, cards: 5, xp: 3890),
        after: progressJson(current: 3, cards: 6, xp: 3901),
      );
      await tester.binding.handlePopRoute();
      await tester.pumpAndSettle();
      expect(await continueToSummary(tester), ['Level 13']);
      expect(under, findsOneWidget);
    });

    testWidgets('the pages wait for the progress after the answer', (
      tester,
    ) async {
      // Back right after an answer: the empty surface, then the page.
      final go = withGoal('cards', 30);
      FakeStudy(go, [
        studyCard(1, 'hola', 'hello'),
        studyCard(2, 'uno', 'one'),
      ]);
      go.progress = goal(29);
      await pumpStudy(tester, loadSettings: true, pushed: true);
      go.progress = goal(30);
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('rate-3')));
      await tester.pump();
      await tester.binding.handlePopRoute();
      await tester.pump();
      expect(find.text('Show answer'), findsNothing);
      await tester.pumpAndSettle();
      expect(headline(tester), 'Daily goal reached');
      await tester.pump(const Duration(milliseconds: 600));
      await tester.tap(next);
      await tester.pumpAndSettle();
      expect(under, findsOneWidget);
    });
  });

  group('celebration sequence', () {
    final next = find.byKey(const Key('celebration-continue'));
    String headline(WidgetTester tester) => tester
        .widget<Text>(find.byKey(const Key('celebration-headline')))
        .data!;

    /// The first session of the day: the streak 12 → 13, the goal of one
    /// card reached and level 12 → 13.
    final several = (
      before: progressJson(current: 12, goalKind: 'cards', goal: 1, xp: 3890),
      after: progressJson(
        current: 13,
        cards: 1,
        goalKind: 'cards',
        goal: 1,
        xp: 3901,
      ),
    );

    /// Answers the session's only card with [before] and [after] as the
    /// progress, up to the first celebration page.
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
      await tester.pump();
      await tester.pump();
      return app;
    }

    testWidgets('several events in order, then the summary', (tester) async {
      // Several events (progress spec) and Finish with streak and goal
      // (study-session spec).
      await finish(
        tester,
        withGoal('cards', 1),
        before: several.before,
        after: several.after,
      );
      expect(find.text('Session complete'), findsNothing);
      expect(await continueToSummary(tester), [
        '13 day streak',
        'Daily goal reached',
        'Level 13',
      ]);
      expect(find.text('Session complete'), findsOneWidget);
      expect(find.text('Streak extended to 13 days'), findsOneWidget);
      expect(
        find.descendant(
          of: find.byKey(const Key('reward-card')),
          matching: find.text('Daily goal reached'),
        ),
        findsOneWidget,
      );
    });

    testWidgets('nothing earned: the summary at once', (tester) async {
      // Nothing earned (progress spec).
      await finish(
        tester,
        FakeGo(),
        before: progressJson(current: 12, cards: 5, xp: 3700),
        after: progressJson(current: 12, cards: 6, xp: 3701),
      );
      expect(next, findsNothing);
      await tester.pumpAndSettle();
      expect(next, findsNothing);
      expect(find.text('Session complete'), findsOneWidget);
    });

    testWidgets('Escape skips to the summary', (tester) async {
      // Skip (progress spec): Escape on the first of three pages.
      await finish(
        tester,
        withGoal('cards', 1),
        before: several.before,
        after: several.after,
      );
      await tester.pumpAndSettle();
      expect(headline(tester), '13 day streak');
      await tester.sendKeyEvent(LogicalKeyboardKey.escape);
      await tester.pumpAndSettle();
      expect(next, findsNothing);
      expect(find.text('Session complete'), findsOneWidget);
      // The screen's own keys work again on the summary.
      expect(
        FocusManager.instance.primaryFocus?.debugLabel,
        isNot('celebration continue'),
      );
    });

    testWidgets('system Back shows the summary instead of leaving', (
      tester,
    ) async {
      await finish(
        tester,
        withGoal('cards', 1),
        before: several.before,
        after: several.after,
      );
      await tester.pumpAndSettle();
      expect(headline(tester), '13 day streak');
      await tester.binding.handlePopRoute();
      await tester.pumpAndSettle();
      expect(next, findsNothing);
      expect(find.text('Session complete'), findsOneWidget);
    });

    testWidgets('a rating key does nothing', (tester) async {
      // Rating key (progress spec).
      final go = withGoal('cards', 1);
      await finish(tester, go, before: several.before, after: several.after);
      await tester.pumpAndSettle();
      await tester.sendKeyEvent(LogicalKeyboardKey.digit3);
      await tester.pumpAndSettle();
      expect(headline(tester), '13 day streak');
      expect(go.callsTo('study.Answer'), hasLength(1));
    });

    testWidgets('the level page after a banner named the level', (
      tester,
    ) async {
      // Banner already shown (progress spec), which names a level while
      // Show level-ups while studying is off.
      final go = FakeGo();
      go.settings = {...go.settings, 'showLevelUpsWhileStudying': false};
      FakeStudy(go, [
        studyCard(1, 'hola', 'hello'),
        studyCard(2, 'uno', 'one'),
      ]);
      go.progress = progressJson(current: 3, cards: 5, xp: 3890);
      final app = await pumpStudy(tester, loadSettings: true);
      go.progress = progressJson(current: 3, cards: 6, xp: 3901);
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('rate-3')));
      await tester.pump();
      await app.progress.load();
      await tester.pump();
      expect(
        find.descendant(
          of: find.byKey(const Key('reward-banner')),
          matching: find.text('Level 13'),
        ),
        findsOneWidget,
      );
      await tester.pumpAndSettle();
      go.progress = progressJson(current: 3, cards: 7, xp: 3902);
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('rate-3')));
      await tester.pumpAndSettle();
      expect(await continueToSummary(tester), ['Level 13']);
    });

    testWidgets('no new pages on a second summary', (tester) async {
      // Continued after the summary (progress spec).
      final go = withGoal('cards', 1);
      noLearnAhead(go);
      final cards = [studyCard(1, 'hola', 'hello')];
      final study = FakeStudy(go, cards);
      go.progress = progressJson(
        current: 3,
        cards: 5,
        goalKind: 'cards',
        goal: 9,
      );
      await pumpStudy(tester, loadSettings: true);
      go.progress = progressJson(
        current: 3,
        cards: 9,
        goalKind: 'cards',
        goal: 9,
      );
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      study.learningDue = inAMinute();
      await tester.tap(find.byKey(const Key('rate-1')));
      await tester.pumpAndSettle();
      expect(await continueToSummary(tester), ['Daily goal reached']);
      cards.add(studyCard(2, 'perro', 'dog', kind: 'learning'));
      study.learningDue = null;
      await waitForLearningCard(tester);
      expect(find.text('Learning cards are due now'), findsOneWidget);
      await tester.tap(find.text('Continue'));
      await tester.pumpAndSettle();
      go.progress = progressJson(
        current: 3,
        cards: 10,
        goalKind: 'cards',
        goal: 9,
      );
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('rate-3')));
      await tester.pumpAndSettle();
      expect(next, findsNothing);
      expect(find.text('Session complete'), findsOneWidget);
    });

    testWidgets('none while streaks and goals are hidden', (tester) async {
      // Streaks and goals off (progress spec): the level raised, no pages.
      final go = FakeGo();
      go.settings = {...go.settings, 'showProgress': false};
      await finish(
        tester,
        go,
        before: progressJson(current: 3, cards: 5, xp: 3890),
        after: progressJson(current: 3, cards: 6, xp: 3901),
      );
      await tester.pumpAndSettle();
      expect(next, findsNothing);
      expect(find.text('Session complete'), findsOneWidget);
    });

    testWidgets('the XP the session earned', (tester) async {
      // XP earned (progress spec): 3,712 → 3,760.
      await finish(
        tester,
        FakeGo(),
        before: progressJson(current: 12, cards: 5, xp: 3712),
        after: progressJson(current: 12, cards: 6, xp: 3760),
      );
      await tester.pumpAndSettle();
      await continueToSummary(tester);
      expect(find.text('+48 XP this session'), findsOneWidget);
    });

    testWidgets('no XP line without XP earned', (tester) async {
      // All answers undone (progress spec): the XP of the start at the end.
      await finish(
        tester,
        FakeGo(),
        before: progressJson(current: 12, cards: 5, xp: 3712),
        after: progressJson(current: 12, cards: 6, xp: 3712),
      );
      await tester.pumpAndSettle();
      expect(find.text('Session complete'), findsOneWidget);
      expect(find.textContaining('XP this session'), findsNothing);
    });
  });
}
