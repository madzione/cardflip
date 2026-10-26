import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/features/progress/session_rewards.dart';
import 'package:cardflip/golang/golang.dart' as golang;

import 'support/fake_go.dart';

void main() {
  golang.Progress progress({
    int cards = 0,
    int goal = 0,
    int xp = 0,
    List<Map<String, Object?>> newAchievements = const [],
  }) => golang.Progress.fromJson(
    progressJson(
      current: 3,
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
}
