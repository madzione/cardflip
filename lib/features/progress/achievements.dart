import 'package:flutter/material.dart';

import 'package:cardflip/features/stats/stats_sections.dart' show parseDay;
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';
import 'package:cardflip/widgets/time_format.dart';

/// How the app shows an achievement that Go defines: the icon on its
/// badge's disc, the disc's colours (top and bottom of its gradient) and
/// its name. A new achievement in Go needs an entry here.
class AchievementLook {
  const AchievementLook(this.icon, this.discTop, this.discBottom);

  final IconData icon;
  final Color discTop;
  final Color discBottom;

  /// The look of achievement [id]; an unknown one gets a neutral star.
  static AchievementLook of(String id) => _looks[id] ?? _unknown;

  static const _unknown = AchievementLook(
    Icons.star,
    Color(0xFF8A94A6),
    Color(0xFF5B6475),
  );

  // The colours of the draft designs in assets/badges/.
  static const _looks = {
    'answers': AchievementLook(
      Icons.check_box,
      Color(0xFF5598F0),
      Color(0xFF1E6FD9),
    ),
    'streak': AchievementLook(
      Icons.local_fire_department,
      Color(0xFFFF9147),
      Color(0xFFE8590C),
    ),
    'study-days': AchievementLook(
      Icons.event_available,
      Color(0xFF55B05A),
      Color(0xFF2E7D32),
    ),
    'daily-goal': AchievementLook(
      Icons.track_changes,
      Color(0xFFEA6464),
      Color(0xFFC62828),
    ),
    'long-term': AchievementLook(
      Icons.eco,
      Color(0xFF33BBAA),
      Color(0xFF00897B),
    ),
    'comeback': AchievementLook(
      Icons.restore,
      Color(0xFF9470DA),
      Color(0xFF6A3FB5),
    ),
  };
}

/// The name of achievement [id], "Long-term memory".
String achievementName(AppLocalizations l, String id) => switch (id) {
  'answers' => l.achievementAnswers,
  'streak' => l.achievementStreak,
  'study-days' => l.achievementStudyDays,
  'daily-goal' => l.achievementDailyGoal,
  'long-term' => l.achievementLongTerm,
  'comeback' => l.achievementComeback,
  _ => id,
};

/// The name of a tier, "Gold".
String tierName(AppLocalizations l, String tier) => switch (tier) {
  'bronze' => l.tierBronze,
  'silver' => l.tierSilver,
  'gold' => l.tierGold,
  'platinum' => l.tierPlatinum,
  _ => tier,
};

/// A tier's threshold as the badge's caption shows it: "100 days",
/// "met on 30 days", "One-time".
String tierThreshold(AppLocalizations l, String id, int threshold) =>
    switch (id) {
      'answers' => l.tierAnswers(threshold),
      'streak' || 'study-days' => l.tierDays(threshold),
      'daily-goal' => l.tierGoalDays(threshold),
      'long-term' => l.tierCards(threshold),
      'comeback' => l.tierOneTime,
      _ => '$threshold',
    };

/// A tier as a notice names it: "1,000 answers", "100-day streak".
String newTierLabel(AppLocalizations l, String id, int threshold) =>
    switch (id) {
      'answers' => l.newAnswers(threshold),
      'streak' => l.newStreak(threshold),
      'study-days' => l.newStudyDays(threshold),
      'daily-goal' => l.newDailyGoal(threshold),
      'long-term' => l.newLongTerm(threshold),
      'comeback' => l.achievementComeback,
      _ => '${achievementName(l, id)} $threshold',
    };

/// The tiers named in a notice: "1,000 answers, 100 study days".
String newTierLabels(AppLocalizations l, List<golang.NewAchievement> tiers) =>
    tiers.map((a) => newTierLabel(l, a.id, a.threshold)).join(', ');

/// The deck list's message for [tiers] earned through sync or import: one
/// to three are named, more are counted (progress spec, "New achievement
/// notice").
String newAchievementsMessage(
  AppLocalizations l,
  List<golang.NewAchievement> tiers,
) => switch (tiers.length) {
  1 => l.newAchievement(newTierLabel(l, tiers[0].id, tiers[0].threshold)),
  <= 3 => l.newAchievements(newTierLabels(l, tiers)),
  _ => l.achievementsFromHistory(tiers.length),
};

/// The summary's line for [tiers] earned in the session.
String newAchievementsLine(
  AppLocalizations l,
  List<golang.NewAchievement> tiers,
) => tiers.length <= 3
    ? l.newAchievementsLine(newTierLabels(l, tiers))
    : l.newAchievementsCountLine(tiers.length);

/// The date a tier was earned, from Go's "2006-01-02".
String earnedDate(String date) => formatDate(parseDay(date));
