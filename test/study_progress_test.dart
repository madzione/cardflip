import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter/semantics.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/data/scope.dart';
import 'package:cardflip/features/progress/achievement_badge.dart';
import 'package:cardflip/features/progress/achievements_screen.dart';

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

  group('reward card', () {
    final card = find.byKey(const Key('reward-card'));
    final celebration = find.byKey(const Key('celebration'));

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

    testWidgets('it settles within 3 seconds, buttons usable', (tester) async {
      // Goal reached (Celebrations, progress spec): the ring fills once
      // while the summary's buttons stay usable.
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
      // The level ring starts at the start's level.
      expect(ringLevel(tester), '12');
      await tester.pump(const Duration(milliseconds: 1300));
      expect(ringLevel(tester), '13');
      expect(celebration, findsOneWidget);
      // The burst lies over the buttons, which still work.
      expect(
        tester.getRect(celebration).overlaps(tester.getRect(card)),
        isTrue,
      );
      await tester.tap(find.text('Back to decks'));
      await tester.pump();
      expect(find.text('Session complete'), findsOneWidget);
      // Settled: nothing more moves after 3 seconds.
      await tester.pump(const Duration(milliseconds: 1700));
      await tester.pump();
      expect(celebration, findsNothing);
      expect(tester.binding.hasScheduledFrame, isFalse);
      // Once: not again when the progress reloads.
      await app.progress.load();
      await tester.pumpAndSettle();
      expect(celebration, findsNothing);
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
      await tester.pump(const Duration(milliseconds: 1300));
      expect(celebration, findsOneWidget);
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
      // Level and tier from one answer (progress spec): one banner, the
      // Level 10 tier left out, the Answers badge.
      final go = FakeGo();
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
      // Rewards while studying off (progress spec): no banner; the summary
      // still says the goal was reached.
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
      await tester.pump(const Duration(seconds: 2));
      expect(banner, findsOneWidget);
      await tester.pump(const Duration(milliseconds: 600));
      expect(banner, findsNothing);
    });
  });
}
