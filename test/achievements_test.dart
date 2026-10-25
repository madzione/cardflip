import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/data/scope.dart';
import 'package:cardflip/features/progress/achievement_badge.dart';
import 'package:cardflip/features/progress/achievements_screen.dart';

import 'support/fake_go.dart';

void main() {
  /// Opens the achievements screen from the deck list menu.
  Future<void> openAchievements(WidgetTester tester) async {
    await pumpApp(tester);
    await tester.tap(find.byTooltip('More'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Achievements'));
    await tester.pumpAndSettle();
    expect(find.byType(AchievementsScreen), findsOneWidget);
  }

  /// The texts of achievement [id]'s tile.
  List<String?> tile(WidgetTester tester, String id) => tester
      .widgetList<Text>(
        find.descendant(
          of: find.byKey(Key('achievement-$id')),
          matching: find.byType(Text),
        ),
      )
      .map((t) => t.data)
      .toList();

  AchievementBadge badge(WidgetTester tester, String id) =>
      tester.widget<AchievementBadge>(
        find.descendant(
          of: find.byKey(Key('achievement-$id')),
          matching: find.byType(AchievementBadge),
        ),
      );

  testWidgets('nothing earned', (tester) async {
    // Nothing earned (spec): every badge locked with zero progress; the
    // Daily goal needs a goal.
    final go = FakeGo();
    go.achievements = [
      for (final id in achievementTiers.keys)
        achievementJson(id, needsGoal: id == 'daily-goal'),
    ];
    await openAchievements(tester);
    for (final id in achievementTiers.keys) {
      await tester.scrollUntilVisible(
        find.byKey(Key('achievement-$id')),
        100,
        scrollable: find.byType(Scrollable).last,
      );
      expect(badge(tester, id).tier, isNull, reason: id);
    }
    expect(tile(tester, 'answers'), [
      'Answers',
      'Bronze · 100 answers',
      '0 / 100',
    ]);
    expect(tile(tester, 'daily-goal'), [
      'Daily goal',
      'Bronze · met on 7 days',
      'Set a daily goal to start',
    ]);
    expect(tile(tester, 'comeback'), [
      'Comeback',
      'Gold · One-time',
      'Study again after 14 days or more without study',
    ]);
  });

  testWidgets('partly earned', (tester) async {
    // Locked with progress (spec): a longest streak of 212 days has the
    // 100-day tier with its date and 212 / 365 toward the next.
    final go = FakeGo();
    go.achievements = [
      achievementJson(
        'answers',
        progress: 1020,
        earned: ['2024-01-08', '2024-03-12'],
      ),
      achievementJson(
        'streak',
        progress: 212,
        earned: ['2026-01-07', '2026-01-30', '2026-04-10'],
      ),
      achievementJson('comeback', progress: 1, earned: ['2026-09-01']),
      achievementJson('long-term', progress: 40),
    ];
    await openAchievements(tester);
    expect(tile(tester, 'answers'), [
      'Answers',
      'Silver · 1,000 answers',
      'Earned 2024-03-12',
      '1,020 / 10,000',
    ]);
    expect(tile(tester, 'streak'), [
      'Streak',
      'Gold · 100 days',
      'Earned 2026-04-10',
      '212 / 365',
    ]);
    final streak = badge(tester, 'streak');
    expect((streak.tier, streak.stars), ('gold', 3));
    final bar = tester.widget<LinearProgressIndicator>(
      find.descendant(
        of: find.byKey(const Key('achievement-streak')),
        matching: find.byType(LinearProgressIndicator),
      ),
    );
    expect(bar.value, closeTo(212 / 365, 1e-9));
    await tester.scrollUntilVisible(
      find.byKey(const Key('achievement-long-term')),
      100,
      scrollable: find.byType(Scrollable).last,
    );
    expect(tile(tester, 'comeback'), [
      'Comeback',
      'Gold · One-time',
      'Earned 2026-09-01',
      'All tiers earned',
    ]);
    expect(badge(tester, 'comeback').stars, 0);
    expect(badge(tester, 'long-term').tier, isNull);
    expect(tile(tester, 'long-term').last, '40 / 100');
  });

  testWidgets('the Streaks statistics open it', (tester) async {
    await pumpApp(tester, size: desktop);
    await tester.tap(find.byTooltip('Statistics'));
    await tester.pumpAndSettle();
    await tester.scrollUntilVisible(
      find.byKey(const Key('streaks-achievements')),
      100,
      scrollable: find.byType(Scrollable).first,
    );
    await tester.tap(find.byKey(const Key('streaks-achievements')));
    await tester.pumpAndSettle();
    expect(find.byType(AchievementsScreen), findsOneWidget);
  });

  testWidgets('hidden with streaks and goals', (tester) async {
    // Hidden (spec): no entry in the deck list menu or the statistics.
    final go = FakeGo();
    go.settings = {...go.settings, 'showProgress': false};
    await pumpApp(tester, size: desktop);
    await tester.tap(find.byTooltip('Statistics'));
    await tester.pumpAndSettle();
    await tester.scrollUntilVisible(
      find.text('Current streak'),
      100,
      scrollable: find.byType(Scrollable).first,
    );
    expect(find.byKey(const Key('streaks-achievements')), findsNothing);
    await tester.pageBack();
    await tester.pumpAndSettle();
    await tester.tap(find.byTooltip('More'));
    await tester.pumpAndSettle();
    expect(find.text('Undo'), findsOneWidget);
    expect(find.text('Achievements'), findsNothing);
  });

  group('new achievement message', () {
    final message = find.byKey(const Key('new-achievements'));

    String messageText(WidgetTester tester) => tester
        .widget<Text>(
          find.descendant(of: message, matching: find.byType(Text)).first,
        )
        .data!;

    testWidgets('earned on another device', (tester) async {
      // Earned on another device (spec): the deck list tells once.
      final go = FakeGo()
        ..progress = progressJson(
          current: 3,
          newAchievements: [newAchievementJson('study-days', 100)],
        );
      await pumpApp(tester);
      expect(messageText(tester), 'New achievement: 100 study days');
      expect(go.callsTo('progress.MarkSeen'), [
        [
          ['study-days:100'],
        ],
      ]);
      // Shown once: a reload does not tell again.
      await tester.pump(const Duration(seconds: 10));
      await tester.pumpAndSettle();
      go.emitChange(kind: 'sync');
      await tester.pumpAndSettle();
      expect(message, findsNothing);
      expect(go.callsTo('progress.MarkSeen'), hasLength(1));
    });

    testWidgets('up to three are named', (tester) async {
      FakeGo().progress = progressJson(
        newAchievements: [
          newAchievementJson('answers', 1000),
          newAchievementJson('streak', 7),
          newAchievementJson('comeback', 1),
        ],
      );
      await pumpApp(tester);
      expect(
        messageText(tester),
        'New achievements: 1,000 answers, 7-day streak, Comeback',
      );
    });

    testWidgets('an import earns many', (tester) async {
      // Import earns many (spec): one message counts 14 tiers.
      final go = FakeGo()..progress = progressJson();
      await pumpApp(tester);
      expect(message, findsNothing);
      go.progress = progressJson(
        newAchievements: [
          for (final MapEntry(key: id, value: tiers)
              in achievementTiers.entries)
            for (final t in tiers.take(3)) newAchievementJson(id, t),
        ].take(14).toList(),
      );
      go.emitChange(kind: 'import');
      await tester.pumpAndSettle();
      expect(messageText(tester), '14 achievements from your history');
      expect((go.callsTo('progress.MarkSeen').single[0] as List).length, 14);
      // Show opens the achievements.
      await tester.tap(find.text('Show'));
      await tester.pumpAndSettle();
      expect(find.byType(AchievementsScreen), findsOneWidget);
    });

    testWidgets('waits while a session is open', (tester) async {
      // No pop-ups during study: the summary tells instead.
      final go = FakeGo()..progress = progressJson();
      await pumpApp(tester);
      final app = AppScope.read(tester.element(find.byType(Scaffold).first));
      app.openSessions.value++;
      go.progress = progressJson(
        cards: 1,
        newAchievements: [newAchievementJson('answers', 100)],
      );
      go.emitChange(kind: 'answer');
      await tester.pumpAndSettle();
      expect(message, findsNothing);
      app.openSessions.value--;
      await tester.pumpAndSettle();
      expect(messageText(tester), 'New achievement: 100 answers');
    });

    testWidgets('hidden', (tester) async {
      // Hidden (spec): no notice.
      final go = FakeGo()
        ..progress = progressJson(
          newAchievements: [newAchievementJson('answers', 1000)],
        );
      go.settings = {...go.settings, 'showProgress': false};
      await pumpApp(tester);
      expect(message, findsNothing);
      expect(go.callsTo('progress.MarkSeen'), isEmpty);
    });
  });
}
