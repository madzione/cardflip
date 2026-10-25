import 'package:flutter/material.dart';
import 'package:cardflip/l10n/l10n.dart';

/// Colours of the new, learning and review counts.
@immutable
class CountColors extends ThemeExtension<CountColors> {
  const CountColors({
    required this.newCards,
    required this.learning,
    required this.review,
  });

  final Color newCards;
  final Color learning;
  final Color review;

  static const light = CountColors(
    newCards: Color(0xFF1E6FD9),
    learning: Color(0xFFC62828),
    review: Color(0xFF2E7D32),
  );

  static const dark = CountColors(
    newCards: Color(0xFF64B5F6),
    learning: Color(0xFFEF9A9A),
    review: Color(0xFF81C784),
  );

  static CountColors of(BuildContext context) =>
      Theme.of(context).extension<CountColors>() ?? light;

  @override
  CountColors copyWith({Color? newCards, Color? learning, Color? review}) =>
      CountColors(
        newCards: newCards ?? this.newCards,
        learning: learning ?? this.learning,
        review: review ?? this.review,
      );

  @override
  CountColors lerp(CountColors? other, double t) {
    if (other == null) return this;
    return CountColors(
      newCards: Color.lerp(newCards, other.newCards, t)!,
      learning: Color.lerp(learning, other.learning, t)!,
      review: Color.lerp(review, other.review, t)!,
    );
  }
}

/// The number of card flags.
const flagCount = 7;

/// The colour name of flag 1 to 7, its name until the user renames it.
String flagColorName(int flag, [AppLocalizations? l10n]) {
  final l = l10n ?? appL10n;
  return switch (flag) {
    1 => l.flagRed,
    2 => l.flagOrange,
    3 => l.flagGreen,
    4 => l.flagBlue,
    5 => l.flagPink,
    6 => l.flagTurquoise,
    _ => l.flagPurple,
  };
}

/// Colours of the seven card flags and of marked notes, and the row tints
/// of flagged, marked and suspended cards.
@immutable
class FlagColors extends ThemeExtension<FlagColors> {
  const FlagColors({
    required this.flags,
    required this.marked,
    required this.suspended,
    required this.tintAlpha,
  });

  /// The colours of flags 1 to 7.
  final List<Color> flags;

  /// The colour of marked notes: their star in study and their row tint.
  final Color marked;

  /// The tint of suspended cards without a flag or mark.
  final Color suspended;

  /// The opacity of a row tint over the row background; low enough to keep
  /// the text readable.
  final double tintAlpha;

  static const light = FlagColors(
    flags: [
      Color(0xFFD32F2F),
      Color(0xFFEF6C00),
      Color(0xFF2E7D32),
      Color(0xFF1565C0),
      Color(0xFFD81B60),
      Color(0xFF00897B),
      Color(0xFF7B1FA2),
    ],
    marked: Color(0xFF9575CD),
    suspended: Color(0xFFFBC02D),
    tintAlpha: 0.16,
  );

  static const dark = FlagColors(
    flags: [
      Color(0xFFEF9A9A),
      Color(0xFFFFB74D),
      Color(0xFF81C784),
      Color(0xFF64B5F6),
      Color(0xFFF48FB1),
      Color(0xFF4DB6AC),
      Color(0xFFCE93D8),
    ],
    marked: Color(0xFFB39DDB),
    suspended: Color(0xFFFFEE58),
    tintAlpha: 0.2,
  );

  static FlagColors of(BuildContext context) =>
      Theme.of(context).extension<FlagColors>() ?? light;

  /// The colour of flag 1 to 7; null for 0 (no flag).
  Color? flag(int n) => n >= 1 && n <= flags.length ? flags[n - 1] : null;

  /// The row tint of a card: its flag's colour, else purple when its note
  /// is marked, else yellow when it is suspended, else none.
  Color? rowTint({
    required int flag,
    bool marked = false,
    required bool suspended,
  }) {
    final base =
        this.flag(flag) ??
        (marked ? this.marked : null) ??
        (suspended ? this.suspended : null);
    return base?.withValues(alpha: tintAlpha);
  }

  @override
  FlagColors copyWith({
    List<Color>? flags,
    Color? marked,
    Color? suspended,
    double? tintAlpha,
  }) => FlagColors(
    flags: flags ?? this.flags,
    marked: marked ?? this.marked,
    suspended: suspended ?? this.suspended,
    tintAlpha: tintAlpha ?? this.tintAlpha,
  );

  @override
  FlagColors lerp(FlagColors? other, double t) {
    if (other == null) return this;
    return FlagColors(
      flags: [
        for (var i = 0; i < flags.length; i++)
          Color.lerp(flags[i], other.flags[i], t)!,
      ],
      marked: Color.lerp(marked, other.marked, t)!,
      suspended: Color.lerp(suspended, other.suspended, t)!,
      tintAlpha: tintAlpha + (other.tintAlpha - tintAlpha) * t,
    );
  }
}

/// The colours of a badge's rim: the top and bottom of the ring's gradient
/// and the scalloped seal around it, which also outlines the disc and the
/// stars.
@immutable
class RimColors {
  const RimColors(this.top, this.bottom, this.seal);

  final Color top;
  final Color bottom;
  final Color seal;

  @override
  bool operator ==(Object other) =>
      other is RimColors &&
      other.top == top &&
      other.bottom == bottom &&
      other.seal == seal;

  @override
  int get hashCode => Object.hash(top, bottom, seal);

  static RimColors lerp(RimColors a, RimColors b, double t) => RimColors(
    Color.lerp(a.top, b.top, t)!,
    Color.lerp(a.bottom, b.bottom, t)!,
    Color.lerp(a.seal, b.seal, t)!,
  );
}

/// The colours of achievement badges: a rim per tier, and the grey rim,
/// disc and padlock of a locked badge. They follow the draft designs in
/// assets/badges/; the metals stay the same in the dark theme, the locked
/// greys are darker there.
@immutable
class BadgeColors extends ThemeExtension<BadgeColors> {
  const BadgeColors({
    required this.bronze,
    required this.silver,
    required this.gold,
    required this.platinum,
    required this.locked,
    required this.lockedDiscTop,
    required this.lockedDiscBottom,
    required this.padlock,
  });

  final RimColors bronze;
  final RimColors silver;
  final RimColors gold;
  final RimColors platinum;
  final RimColors locked;
  final Color lockedDiscTop;
  final Color lockedDiscBottom;

  /// The disc behind the padlock of a locked badge.
  final Color padlock;

  static const _bronze = RimColors(
    Color(0xFFF0B484),
    Color(0xFFA9612F),
    Color(0xFF7A4320),
  );
  static const _silver = RimColors(
    Color(0xFFF4F6F9),
    Color(0xFF9AA3AF),
    Color(0xFF6B7480),
  );
  static const _gold = RimColors(
    Color(0xFFFFE38F),
    Color(0xFFD49A1E),
    Color(0xFF9E6C0C),
  );
  static const _platinum = RimColors(
    Color(0xFFF2F7FC),
    Color(0xFF8FA6BF),
    Color(0xFF4F6478),
  );

  static const light = BadgeColors(
    bronze: _bronze,
    silver: _silver,
    gold: _gold,
    platinum: _platinum,
    locked: RimColors(Color(0xFFDADADA), Color(0xFFA6A6A6), Color(0xFF8C8C8C)),
    lockedDiscTop: Color(0xFFCFCFCF),
    lockedDiscBottom: Color(0xFFB5B5B5),
    padlock: Color(0xFF6F6F6F),
  );

  static const dark = BadgeColors(
    bronze: _bronze,
    silver: _silver,
    gold: _gold,
    platinum: _platinum,
    locked: RimColors(Color(0xFF8A8A8A), Color(0xFF5C5C5C), Color(0xFF474747)),
    lockedDiscTop: Color(0xFF7A7A7A),
    lockedDiscBottom: Color(0xFF606060),
    padlock: Color(0xFF3A3A3A),
  );

  static BadgeColors of(BuildContext context) =>
      Theme.of(context).extension<BadgeColors>() ?? light;

  /// The rim of [tier] ("bronze" to "platinum").
  RimColors rim(String tier) => switch (tier) {
    'bronze' => bronze,
    'silver' => silver,
    'platinum' => platinum,
    _ => gold,
  };

  @override
  BadgeColors copyWith({
    RimColors? bronze,
    RimColors? silver,
    RimColors? gold,
    RimColors? platinum,
    RimColors? locked,
    Color? lockedDiscTop,
    Color? lockedDiscBottom,
    Color? padlock,
  }) => BadgeColors(
    bronze: bronze ?? this.bronze,
    silver: silver ?? this.silver,
    gold: gold ?? this.gold,
    platinum: platinum ?? this.platinum,
    locked: locked ?? this.locked,
    lockedDiscTop: lockedDiscTop ?? this.lockedDiscTop,
    lockedDiscBottom: lockedDiscBottom ?? this.lockedDiscBottom,
    padlock: padlock ?? this.padlock,
  );

  @override
  BadgeColors lerp(BadgeColors? other, double t) {
    if (other == null) return this;
    return BadgeColors(
      bronze: RimColors.lerp(bronze, other.bronze, t),
      silver: RimColors.lerp(silver, other.silver, t),
      gold: RimColors.lerp(gold, other.gold, t),
      platinum: RimColors.lerp(platinum, other.platinum, t),
      locked: RimColors.lerp(locked, other.locked, t),
      lockedDiscTop: Color.lerp(lockedDiscTop, other.lockedDiscTop, t)!,
      lockedDiscBottom: Color.lerp(
        lockedDiscBottom,
        other.lockedDiscBottom,
        t,
      )!,
      padlock: Color.lerp(padlock, other.padlock, t)!,
    );
  }
}

const _seed = Color(0xFF3B82F6);

/// The light theme; [minimalist] flattens it (see [minimalistTheme]).
ThemeData lightTheme({bool minimalist = false}) {
  final theme = ThemeData(
    colorScheme: ColorScheme.fromSeed(seedColor: _seed),
    extensions: const [CountColors.light, FlagColors.light, BadgeColors.light],
  );
  return minimalist ? minimalistTheme(theme) : theme;
}

/// The dark theme; [minimalist] flattens it (see [minimalistTheme]).
ThemeData darkTheme({bool minimalist = false}) {
  final theme = ThemeData(
    colorScheme: ColorScheme.fromSeed(
      seedColor: _seed,
      brightness: Brightness.dark,
    ),
    extensions: const [CountColors.dark, FlagColors.dark, BadgeColors.dark],
  );
  return minimalist ? minimalistTheme(theme) : theme;
}

/// [theme] for minimalist mode: surfaces are flat, without shadows,
/// elevation tints or tinted fills; an outline separates cards, dialogs and
/// menus instead, and dividers are thinner.
ThemeData minimalistTheme(ThemeData theme) {
  final colors = theme.colorScheme;
  const none = Colors.transparent;
  final outline = BorderSide(color: colors.outlineVariant);
  return theme.copyWith(
    colorScheme: colors.copyWith(surfaceTint: none, shadow: none),
    shadowColor: none,
    cardTheme: CardThemeData(
      elevation: 0,
      color: colors.surface,
      shadowColor: none,
      surfaceTintColor: none,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: outline,
      ),
    ),
    dialogTheme: DialogThemeData(
      elevation: 0,
      backgroundColor: colors.surface,
      shadowColor: none,
      surfaceTintColor: none,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(28),
        side: outline,
      ),
    ),
    popupMenuTheme: PopupMenuThemeData(
      elevation: 0,
      color: colors.surface,
      shadowColor: none,
      surfaceTintColor: none,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(4),
        side: outline,
      ),
    ),
    appBarTheme: const AppBarTheme(
      elevation: 0,
      scrolledUnderElevation: 0,
      shadowColor: none,
      surfaceTintColor: none,
    ),
    navigationBarTheme: NavigationBarThemeData(
      elevation: 0,
      backgroundColor: colors.surface,
      shadowColor: none,
      surfaceTintColor: none,
    ),
    navigationRailTheme: NavigationRailThemeData(
      elevation: 0,
      backgroundColor: colors.surface,
    ),
    snackBarTheme: const SnackBarThemeData(elevation: 0),
    chipTheme: const ChipThemeData(
      elevation: 0,
      pressElevation: 0,
      shadowColor: none,
      surfaceTintColor: none,
    ),
    elevatedButtonTheme: ElevatedButtonThemeData(
      style: ElevatedButton.styleFrom(
        elevation: 0,
        shadowColor: none,
        surfaceTintColor: none,
      ),
    ),
    dividerTheme: DividerThemeData(
      thickness: 0.5,
      color: colors.outlineVariant,
    ),
  );
}

/// Maps the stored theme setting to a [ThemeMode].
ThemeMode themeModeOf(String setting) => switch (setting) {
  'light' => ThemeMode.light,
  'dark' => ThemeMode.dark,
  _ => ThemeMode.system,
};

String themeSettingOf(ThemeMode mode) => switch (mode) {
  ThemeMode.light => 'light',
  ThemeMode.dark => 'dark',
  ThemeMode.system => 'system',
};
