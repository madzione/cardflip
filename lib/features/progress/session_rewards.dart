import 'dart:math' as math;

import 'package:cardflip/golang/golang.dart' as golang;

/// What one progress update found new in a session: the daily goal
/// reached, a level reached and achievement tiers, for one reward banner.
class RewardEvents {
  const RewardEvents({this.goal = false, this.level, this.tiers = const []});

  final bool goal;
  final int? level;
  final List<golang.NewAchievement> tiers;

  bool get isEmpty => !goal && level == null && tiers.isEmpty;

  /// The tiers to name: without Level tiers when [level] names the level
  /// reached (progress spec, "Reward banner during study").
  List<golang.NewAchievement> get named => [
    for (final a in tiers)
      if (level == null || a.id != 'level') a,
  ];

  /// The keys that record these events as shown in [SessionRewards.shown];
  /// a Level tier's key is that of its level.
  Set<String> get keys => {
    if (goal) SessionRewards.goalKey,
    if (level case final level?) SessionRewards.levelKey(level),
    for (final a in tiers) a.key,
  };
}

/// One page of the celebration sequence after a session (progress spec,
/// "Celebration sequence").
sealed class CelebrationEvent {
  const CelebrationEvent();
}

/// The streak extended from [from] to [to] days, with the [week] of
/// `Progress.week`; [best] when it is a new personal best.
class StreakCelebration extends CelebrationEvent {
  const StreakCelebration({
    required this.from,
    required this.to,
    required this.week,
    required this.best,
  });

  final int from;
  final int to;
  final List<String> week;
  final bool best;
}

/// The daily goal reached: [progress] of [goal] [kind] ("cards" or
/// "minutes"), earning [xp] XP.
class GoalCelebration extends CelebrationEvent {
  const GoalCelebration({
    required this.progress,
    required this.goal,
    required this.kind,
    required this.xp,
  });

  final int progress;
  final int goal;
  final String kind;
  final int xp;
}

/// The level raised from [fromLevel], [fromFraction] (0 to 1) into it, to
/// [toLevel], with [xpLeft] XP to the next level.
class LevelCelebration extends CelebrationEvent {
  const LevelCelebration({
    required this.fromLevel,
    required this.fromFraction,
    required this.toLevel,
    required this.xpLeft,
  });

  final int fromLevel;
  final double fromFraction;
  final int toLevel;
  final int xpLeft;
}

/// One achievement tier earned.
class AchievementCelebration extends CelebrationEvent {
  const AchievementCelebration(this.tier);

  final golang.NewAchievement tier;
}

/// The tiers earned beyond the first three, on one page.
class MoreAchievementsCelebration extends CelebrationEvent {
  const MoreAchievementsCelebration(this.tiers);

  final List<golang.NewAchievement> tiers;
}

/// The rewards of one study session (add-study-rewards design decision 1):
/// the goal and level at its start, what it earned, and what a reward
/// banner already showed. The banner, the summary's reward card and the
/// deck list's message after leaving all read from it, so they agree.
class SessionRewards {
  /// Starts from [before], the progress when the session started, or from
  /// a later [start] while that is unknown.
  SessionRewards([golang.Progress? before]) {
    if (before != null) start(before);
  }

  static const goalKey = 'goal';
  static const streakKey = 'streak';
  static String levelKey(int level) => 'level:$level';

  /// Whether the goal was reached and the level when the session started,
  /// and how far into that level its XP were (0 to 1); null until known.
  bool? goalReachedBefore;
  int? levelBefore;
  double levelFractionBefore = 0;

  /// The current streak and the XP when the session started; null until
  /// known.
  int? streakBefore;
  int? xpBefore;

  /// Whether the session reached the goal, the highest level it reached
  /// (null for none) and its tiers, in the order they were found.
  bool goalEarned = false;
  int? levelEarned;
  final tiers = <golang.NewAchievement>[];

  /// The keys of the events a banner showed: [goalKey], [levelKey] and
  /// tier keys.
  final shown = <String>{};

  /// The keys of the events a celebration page showed: those of [shown]
  /// and [streakKey]. A banner does not replace a page.
  final celebrated = <String>{};

  /// The level a later celebration fills from, and how far into it.
  int? _celebratedLevel;
  double _celebratedFraction = 0;

  /// The XP the session earned by [now], 0 while unknown.
  int xpEarned(golang.Progress now) => now.xp - (xpBefore ?? now.xp);

  bool get started => levelBefore != null;

  /// Whether the session earned anything.
  bool get isEmpty => !goalEarned && levelEarned == null && tiers.isEmpty;

  /// Sets the start, unless it is known already.
  void start(golang.Progress before) {
    if (started) return;
    goalReachedBefore = before.goalReached;
    levelBefore = before.level;
    levelFractionBefore = levelFraction(before);
    streakBefore = before.current;
    xpBefore = before.xp;
  }

  /// Records what [p] shows earned since the start and returns what is new
  /// since the last update and not shown: the goal now reached, a level
  /// above every level found so far, and tiers of [unseen] (by default
  /// those of [p]) not found yet. An undo lowers [p] but takes nothing
  /// back, so answering again finds nothing new.
  RewardEvents update(
    golang.Progress p, [
    List<golang.NewAchievement>? unseen,
  ]) {
    if (!started) start(p);
    final goal = p.goalReached && goalReachedBefore == false && !goalEarned;
    if (goal) goalEarned = true;
    final level = p.level > levelBefore! && p.level > (levelEarned ?? 0)
        ? p.level
        : null;
    if (level != null) levelEarned = level;
    final known = {for (final a in tiers) a.key};
    final fresh = [
      for (final a in unseen ?? p.newAchievements)
        if (known.add(a.key)) a,
    ];
    tiers.addAll(fresh);
    return RewardEvents(
      goal: goal && !shown.contains(goalKey),
      level: level != null && !shown.contains(levelKey(level)) ? level : null,
      tiers: [
        for (final a in fresh)
          if (!shown.contains(a.key)) a,
      ],
    );
  }

  /// Records everything earned so far as shown, as the summary does.
  void showAll() {
    if (goalEarned) shown.add(goalKey);
    if (levelEarned case final level?) shown.add(levelKey(level));
    shown.addAll(tiers.map((a) => a.key));
  }

  /// What the session earned that nothing showed, judged against [now]
  /// (the progress after the session's last answer): the goal when it is
  /// reached now, the level when it is above the start, and the tiers.
  RewardEvents unshown(golang.Progress now) {
    final before = levelBefore;
    return RewardEvents(
      goal:
          now.goalReached &&
          goalReachedBefore == false &&
          !shown.contains(goalKey),
      level:
          before != null &&
              now.level > before &&
              !shown.contains(levelKey(now.level))
          ? now.level
          : null,
      tiers: [
        for (final a in tiers)
          if (!shown.contains(a.key)) a,
      ],
    );
  }

  /// The celebration pages for the session ending with [now], in the order
  /// streak, goal, level, achievements (add-celebration-screens design
  /// decision 2): what it earned that no page celebrated yet, whether or
  /// not a banner showed it. The streak counts when the session made today
  /// a study day ([studiedBefore] false). A Level tier is left out beside a
  /// level page, and tiers beyond the first three share one page. Records
  /// the pages as celebrated and shown, so a second summary in the session
  /// only gets new events.
  List<CelebrationEvent> celebrations(
    golang.Progress now, {
    required bool? studiedBefore,
  }) {
    final out = <CelebrationEvent>[];
    final keys = <String>{};
    if (studiedBefore == false &&
        now.todayStudied &&
        !celebrated.contains(streakKey)) {
      keys.add(streakKey);
      out.add(
        StreakCelebration(
          from: math.max(0, math.min(streakBefore ?? 0, now.current - 1)),
          to: now.current,
          week: now.week,
          best: now.current > 1 && now.current == now.longest,
        ),
      );
    }
    if (now.goalReached &&
        goalReachedBefore == false &&
        !celebrated.contains(goalKey)) {
      keys.add(goalKey);
      out.add(
        GoalCelebration(
          progress: now.goalKind == 'minutes'
              ? now.todayMinutes
              : now.todayCards,
          goal: now.goal,
          kind: now.goalKind,
          xp: now.goalXP,
        ),
      );
    }
    final level = _level(now);
    if (level != null) {
      keys.add(levelKey(level.toLevel));
      out.add(level);
    }
    final fresh = [
      for (final a in tiers)
        if (!celebrated.contains(a.key)) a,
    ];
    keys.addAll(fresh.map((a) => a.key));
    final paged = [
      for (final a in fresh)
        if (level == null || a.id != 'level') a,
    ];
    out.addAll(paged.take(3).map(AchievementCelebration.new));
    if (paged.length > 3) {
      out.add(MoreAchievementsCelebration(paged.sublist(3)));
    }
    celebrated.addAll(keys);
    shown.addAll(keys.difference({streakKey}));
    return out;
  }

  /// The level page during study for [now] (add-celebration-screens design
  /// decision 2): the level when it is above the start and above every
  /// level celebrated so far, or null. Records it and the Level tiers found
  /// so far as celebrated and shown, so [celebrations] and a banner leave
  /// them out, and a further level-up fills from where this page ended.
  LevelCelebration? levelCelebration(golang.Progress now) {
    final level = _level(now);
    if (level != null) {
      final keys = {
        levelKey(level.toLevel),
        for (final a in tiers)
          if (a.id == 'level') a.key,
      };
      celebrated.addAll(keys);
      shown.addAll(keys);
    }
    return level;
  }

  /// The level page for [now], filling from the start or from where the
  /// last level page ended, or null when [now] reached no level beyond
  /// those. Records where the page ends.
  LevelCelebration? _level(golang.Progress now) {
    final before = levelBefore;
    if (before == null ||
        now.level <= (_celebratedLevel ?? before) ||
        celebrated.contains(levelKey(now.level))) {
      return null;
    }
    final page = LevelCelebration(
      fromLevel: _celebratedLevel ?? before,
      fromFraction: _celebratedLevel == null
          ? levelFractionBefore
          : _celebratedFraction,
      toLevel: now.level,
      xpLeft: now.nextLevelXP - now.xp,
    );
    _celebratedLevel = now.level;
    _celebratedFraction = levelFraction(now);
    return page;
  }
}

/// How far into its level [p]'s XP are, from 0 to 1.
double levelFraction(golang.Progress p) {
  final span = p.nextLevelXP - p.levelStart;
  return span <= 0 ? 0 : ((p.xp - p.levelStart) / span).clamp(0, 1);
}
