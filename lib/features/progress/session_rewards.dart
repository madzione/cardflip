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
  static String levelKey(int level) => 'level:$level';

  /// Whether the goal was reached and the level when the session started,
  /// and how far into that level its XP were (0 to 1); null until known.
  bool? goalReachedBefore;
  int? levelBefore;
  double levelFractionBefore = 0;

  /// Whether the session reached the goal, the highest level it reached
  /// (null for none) and its tiers, in the order they were found.
  bool goalEarned = false;
  int? levelEarned;
  final tiers = <golang.NewAchievement>[];

  /// The keys of the events a banner showed: [goalKey], [levelKey] and
  /// tier keys.
  final shown = <String>{};

  bool get started => levelBefore != null;

  /// Whether the session earned anything.
  bool get isEmpty => !goalEarned && levelEarned == null && tiers.isEmpty;

  /// Sets the start, unless it is known already.
  void start(golang.Progress before) {
    if (started) return;
    goalReachedBefore = before.goalReached;
    levelBefore = before.level;
    levelFractionBefore = levelFraction(before);
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
}

/// How far into its level [p]'s XP are, from 0 to 1.
double levelFraction(golang.Progress p) {
  final span = p.nextLevelXP - p.levelStart;
  return span <= 0 ? 0 : ((p.xp - p.levelStart) / span).clamp(0, 1);
}
