import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/features/progress/session_rewards.dart';
import 'package:cardflip/golang/golang.dart' as golang;

import 'support/fake_go.dart';

/// The only element of [pages], which must be a [T].
T single<T>(List<CelebrationEvent> pages) {
  expect(pages, [isA<T>()]);
  return pages.single as T;
}

void main() {
  golang.Progress progress({
    int current = 3,
    int? longest,
    int cards = 0,
    int goal = 0,
    int xp = 0,
    List<Map<String, Object?>> newAchievements = const [],
  }) => golang.Progress.fromJson(
    progressJson(
      current: current,
      longest: longest,
      cards: cards,
      goalKind: goal > 0 ? 'cards' : '',
      goal: goal,
      xp: xp,
      newAchievements: newAchievements,
    ),
  );

  test('the goal reached', () {
    final rewards = SessionRewards(progress(cards: 29, goal: 30));
    expect(rewards.update(progress(cards: 29, goal: 30)).isEmpty, isTrue);
    final events = rewards.update(progress(cards: 30, goal: 30));
    expect(events.goal, isTrue);
    expect(events.keys, {'goal'});
    expect(rewards.goalEarned, isTrue);
    // Once.
    expect(rewards.update(progress(cards: 31, goal: 30)).isEmpty, isTrue);
  });

  test('not when the goal was reached before the session', () {
    final rewards = SessionRewards(progress(cards: 30, goal: 30));
    expect(rewards.update(progress(cards: 31, goal: 30)).isEmpty, isTrue);
    expect(rewards.isEmpty, isTrue);
  });

  test('the level raised twice in one session', () {
    final rewards = SessionRewards(progress(xp: 3890)); // level 12
    expect(rewards.levelBefore, 12);
    expect(rewards.levelFractionBefore, closeTo(590 / 600, 1e-9));
    expect(rewards.update(progress(xp: 3901)).level, 13);
    expect(rewards.update(progress(xp: 4000)).isEmpty, isTrue);
    expect(rewards.update(progress(xp: 4551)).level, 14);
    expect(rewards.levelEarned, 14);
  });

  test('tiers from one update', () {
    final rewards = SessionRewards(progress(cards: 5));
    final events = rewards.update(
      progress(
        cards: 6,
        newAchievements: [
          newAchievementJson('answers', 1000),
          newAchievementJson('study-days', 100),
        ],
      ),
    );
    expect(events.tiers.map((a) => a.key), ['answers:1000', 'study-days:100']);
    expect(rewards.tiers, hasLength(2));
    // Go may report them again until they are marked seen.
    final again = rewards.update(
      progress(
        cards: 7,
        newAchievements: [newAchievementJson('answers', 1000)],
      ),
    );
    expect(again.isEmpty, isTrue);
    expect(rewards.tiers, hasLength(2));
  });

  test('the unseen tiers given instead of those of the progress', () {
    final rewards = SessionRewards(progress());
    final events = rewards.update(
      progress(
        cards: 1,
        newAchievements: [newAchievementJson('answers', 1000)],
      ),
      const [],
    );
    expect(events.isEmpty, isTrue);
  });

  test('no event after an undo reload', () {
    final rewards = SessionRewards(progress(cards: 29, goal: 30, xp: 3890));
    rewards.update(progress(cards: 30, goal: 30, xp: 3901));
    // Undo lowers the progress again.
    expect(
      rewards.update(progress(cards: 29, goal: 30, xp: 3890)).isEmpty,
      isTrue,
    );
    expect(rewards.goalEarned, isTrue);
    expect(rewards.levelEarned, 13);
  });

  test('the same event again after undo', () {
    // Undone and answered again (spec): nothing new.
    final rewards = SessionRewards(progress(cards: 29, goal: 30, xp: 3890));
    final first = rewards.update(progress(cards: 30, goal: 30, xp: 3901));
    rewards.shown.addAll(first.keys);
    rewards.update(progress(cards: 29, goal: 30, xp: 3890));
    expect(
      rewards.update(progress(cards: 30, goal: 30, xp: 3901)).isEmpty,
      isTrue,
    );
  });

  test('a Level tier left out when the level is reported', () {
    // Level and tier from one answer (spec).
    final rewards = SessionRewards(progress(xp: 2240)); // level 9
    final events = rewards.update(
      progress(
        xp: 2276,
        newAchievements: [
          newAchievementJson('level', 10),
          newAchievementJson('answers', 1000),
        ],
      ),
    );
    expect(events.level, 10);
    expect(events.named.map((a) => a.key), ['answers:1000']);
    // Still earned and recorded as shown with the banner.
    expect(events.keys, {'level:10', 'answers:1000'});
    expect(rewards.tiers, hasLength(2));
    // Named when the level is not.
    final alone = RewardEvents(tiers: events.tiers);
    expect(alone.named, hasLength(2));
  });

  test('the start comes later while unknown', () {
    final rewards = SessionRewards();
    expect(rewards.started, isFalse);
    rewards.start(progress(cards: 29, goal: 30));
    rewards.start(progress(cards: 30, goal: 30));
    expect(rewards.goalReachedBefore, isFalse);
  });

  test('what nothing showed', () {
    final rewards = SessionRewards(progress(cards: 29, goal: 30, xp: 3890));
    final events = rewards.update(
      progress(
        cards: 30,
        goal: 30,
        xp: 3901,
        newAchievements: [newAchievementJson('answers', 1000)],
      ),
    );
    final now = progress(cards: 30, goal: 30, xp: 3901);
    final all = rewards.unshown(now);
    expect(all.goal, isTrue);
    expect(all.level, 13);
    expect(all.tiers, hasLength(1));
    rewards.shown.addAll(events.keys);
    expect(rewards.unshown(now).isEmpty, isTrue);
  });

  test('everything shown on the summary', () {
    final rewards = SessionRewards(progress(cards: 29, goal: 30, xp: 3890));
    rewards.update(
      progress(
        cards: 30,
        goal: 30,
        xp: 3901,
        newAchievements: [newAchievementJson('answers', 1000)],
      ),
    );
    rewards.showAll();
    expect(rewards.shown, {'goal', 'level:13', 'answers:1000'});
    // Only later events after Check again.
    expect(rewards.update(progress(cards: 31, goal: 30, xp: 4551)).level, 14);
  });

  group('celebrations', () {
    test('in the order streak, goal, level, achievements', () {
      // Several events (spec): the first session of the day extends the
      // streak to 13, reaches the goal and raises the level to 13.
      final rewards = SessionRewards(
        progress(current: 12, longest: 40, goal: 30, xp: 3890),
      );
      final now = progress(
        current: 13,
        longest: 40,
        cards: 34,
        goal: 30,
        xp: 3910,
        newAchievements: [newAchievementJson('answers', 1000)],
      );
      rewards.update(now);
      final pages = rewards.celebrations(now, studiedBefore: false);
      expect(pages, hasLength(4));
      final streak = pages[0] as StreakCelebration;
      final goal = pages[1] as GoalCelebration;
      final level = pages[2] as LevelCelebration;
      final tier = pages[3] as AchievementCelebration;
      expect((streak.from, streak.to, streak.best), (12, 13, false));
      expect(streak.week, now.week);
      expect(
        (goal.progress, goal.goal, goal.kind, goal.xp),
        (34, 30, 'cards', 20),
      );
      expect(
        (level.fromLevel, level.toLevel, level.xpLeft),
        (12, 13, 640), // 3,910 XP: 640 XP to level 14
      );
      expect(level.fromFraction, closeTo(590 / 600, 1e-9));
      expect(tier.tier.key, 'answers:1000');
      expect(tier.tier.xp, 50);
      // Every celebration counts as shown.
      expect(rewards.shown, {'goal', 'level:13', 'answers:1000'});
    });

    test('the streak only when the session made today a study day', () {
      for (final before in [true, null]) {
        final rewards = SessionRewards(progress(current: 12));
        final now = progress(current: 13, cards: 1);
        rewards.update(now);
        expect(rewards.celebrations(now, studiedBefore: before), isEmpty);
      }
      // All answers undone: today is no study day.
      final rewards = SessionRewards(progress(current: 12));
      final now = progress(current: 12);
      expect(rewards.celebrations(now, studiedBefore: false), isEmpty);
      // Personal best (spec), and the first study day without one.
      final best = SessionRewards(progress(current: 40, longest: 40));
      final b = single<StreakCelebration>(
        best.celebrations(
          progress(current: 41, cards: 1),
          studiedBefore: false,
        ),
      );
      expect((b.from, b.to, b.best), (40, 41, true));
      final first = SessionRewards(progress(current: 0));
      final f = single<StreakCelebration>(
        first.celebrations(
          progress(current: 1, cards: 1),
          studiedBefore: false,
        ),
      );
      expect((f.from, f.to, f.best), (0, 1, false));
    });

    test('a Level tier left out beside a level page', () {
      // Level tier with level page (spec).
      final rewards = SessionRewards(progress(xp: 2240)); // level 9
      final now = progress(
        xp: 2276,
        newAchievements: [newAchievementJson('level', 10)],
      );
      rewards.update(now);
      final pages = rewards.celebrations(now, studiedBefore: true);
      expect(pages, [isA<LevelCelebration>()]);
      expect(rewards.celebrated, {'level:10'});
    });

    test('five tiers: three pages and one for the rest', () {
      // Five tiers (spec).
      final rewards = SessionRewards(progress());
      final now = progress(
        cards: 1,
        newAchievements: [
          newAchievementJson('answers', 100),
          newAchievementJson('streak', 7),
          newAchievementJson('study-days', 30),
          newAchievementJson('study-time', 10),
          newAchievementJson('long-term', 100),
        ],
      );
      rewards.update(now);
      final pages = rewards.celebrations(now, studiedBefore: true);
      expect(pages, hasLength(4));
      expect(
        [for (final p in pages.take(3)) (p as AchievementCelebration).tier.key],
        ['answers:100', 'streak:7', 'study-days:30'],
      );
      expect(
        (pages[3] as MoreAchievementsCelebration).tiers.map((a) => a.key),
        ['study-time:10', 'long-term:100'],
      );
    });

    test('events a banner showed are still celebrated', () {
      // Banner already shown (spec): the level page all the same.
      final rewards = SessionRewards(progress(xp: 3890));
      final now = progress(xp: 3901);
      rewards.shown.addAll(rewards.update(now).keys);
      expect(rewards.celebrations(now, studiedBefore: true), [
        isA<LevelCelebration>(),
      ]);
    });

    test('a second summary only gets new events', () {
      // Continued after the summary (spec).
      final rewards = SessionRewards(
        progress(current: 12, goal: 30, cards: 0, xp: 3890),
      );
      final first = progress(current: 13, cards: 30, goal: 30, xp: 3901);
      rewards.update(first);
      expect(rewards.celebrations(first, studiedBefore: false), hasLength(3));
      rewards.showAll();
      final same = progress(current: 13, cards: 31, goal: 30, xp: 3902);
      rewards.update(same);
      expect(rewards.celebrations(same, studiedBefore: false), isEmpty);
      // A new level fills from where the first page ended.
      final higher = progress(current: 13, cards: 40, goal: 30, xp: 4551);
      rewards.update(higher);
      final level = single<LevelCelebration>(
        rewards.celebrations(higher, studiedBefore: false),
      );
      expect((level.fromLevel, level.toLevel), (13, 14));
      expect(level.fromFraction, closeTo(1 / 650, 1e-9));
    });

    test('the level page during study', () {
      // Level-up mid-session (spec): the page once, then not again at the
      // end, and a further level fills from where it ended.
      final rewards = SessionRewards(progress(xp: 2240)); // level 9
      final first = progress(
        xp: 2276,
        newAchievements: [newAchievementJson('level', 10)],
      );
      rewards.update(first);
      final page = rewards.levelCelebration(first)!;
      expect((page.fromLevel, page.toLevel, page.xpLeft), (9, 10, 2750 - 2276));
      expect(page.fromFraction, closeTo(440 / 450, 1e-9));
      expect(rewards.celebrated, {'level:10'});
      expect(rewards.shown, {'level:10'});
      expect(rewards.levelCelebration(first), isNull);
      // Not again at the end (spec), nor its Level tier.
      expect(rewards.celebrations(first, studiedBefore: true), isEmpty);
      final higher = progress(xp: 2760); // level 11
      rewards.update(higher);
      final level = single<LevelCelebration>(
        rewards.celebrations(higher, studiedBefore: true),
      );
      expect((level.fromLevel, level.toLevel), (10, 11));
      expect(level.fromFraction, closeTo(26 / 500, 1e-9));
      // No level reached: nothing.
      // Two levels at once: the Level tier between them is left out too.
      final jump = SessionRewards(progress(xp: 2240)); // level 9
      final eleven = progress(
        xp: 2760,
        newAchievements: [newAchievementJson('level', 10)],
      );
      jump.update(eleven);
      expect(jump.levelCelebration(eleven)?.toLevel, 11);
      expect(jump.shown, {'level:11', 'level:10'});
      expect(jump.celebrations(eleven, studiedBefore: true), isEmpty);
      final same = SessionRewards(progress(xp: 3890));
      expect(same.levelCelebration(progress(xp: 3895)), isNull);
      expect(same.celebrated, isEmpty);
    });

    test('the XP the session earned', () {
      final rewards = SessionRewards(progress(xp: 3712));
      expect(rewards.xpEarned(progress(xp: 3760)), 48);
      expect(rewards.xpEarned(progress(xp: 3712)), 0);
      expect(SessionRewards().xpEarned(progress(xp: 10)), 0);
    });
  });
}
