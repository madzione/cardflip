import 'package:intl/intl.dart';

import 'package:cardflip/l10n/l10n.dart';

// The formats below use the interface language, [appL10n], unless they are
// given one.

/// The time of day, "14:05".
String _time(DateTime t, AppLocalizations l) {
  ensureDateFormats();
  return DateFormat.Hm(l.localeName).format(t);
}

/// Formats times for people: "at 14:05", "tomorrow", "in 3 days", "on 2026-11-02".
String formatDue(DateTime due, {DateTime? now, AppLocalizations? l10n}) {
  final l = l10n ?? appL10n;
  final n = now ?? DateTime.now();
  final local = due.toLocal();
  final today = DateTime(n.year, n.month, n.day);
  final day = DateTime(local.year, local.month, local.day);
  final days = day.difference(today).inDays;
  if (days <= 0) return l.dueAt(_time(local, l));
  if (days == 1) return l.dueTomorrow;
  if (days < 7) return l.dueInDays(days);
  return l.dueOn(formatDate(local, l10n: l));
}

/// "45 seconds", "1 minute", "12 minutes", "1 h 5 min".
String formatSpent(Duration d, {AppLocalizations? l10n}) {
  final l = l10n ?? appL10n;
  if (d.inSeconds < 60) return l.spentSeconds(d.inSeconds);
  if (d.inMinutes < 60) return l.spentMinutes(d.inMinutes);
  return l.spentHours(d.inHours, d.inMinutes % 60);
}

/// "2026-10-13 14:05", in local time.
String formatDateTime(DateTime t, {AppLocalizations? l10n}) {
  final l = l10n ?? appL10n;
  final local = t.toLocal();
  return '${formatDate(local, l10n: l)} ${_time(local, l)}';
}

/// "14:05" today, otherwise "2026-10-13 14:05", in local time.
String formatRecent(DateTime t, {DateTime? now, AppLocalizations? l10n}) {
  final l = l10n ?? appL10n;
  final n = now ?? DateTime.now();
  final local = t.toLocal();
  return local.year == n.year && local.month == n.month && local.day == n.day
      ? _time(local, l)
      : formatDateTime(local, l10n: l);
}

/// "2026-10-13", in local time.
String formatDate(DateTime t, {AppLocalizations? l10n}) {
  final l = l10n ?? appL10n;
  ensureDateFormats();
  return DateFormat(l.datePattern, l.localeName).format(t.toLocal());
}

/// A scheduling interval: "10 minutes", "5 hours", "1 day", "12 days",
/// "2.5 months", "1.3 years".
String formatInterval(Duration d, {AppLocalizations? l10n}) {
  final l = l10n ?? appL10n;
  final days = d.inMinutes / (24 * 60);
  if (d.inMinutes < 60) {
    return l.intervalMinutes(d.inMinutes < 1 ? 1 : d.inMinutes);
  }
  if (d.inHours < 24) return l.intervalHours(d.inHours);
  if (days < 30) return l.intervalDays(days.round());
  final oneDecimal = NumberFormat('0.0', l.localeName);
  if (days < 365) return l.intervalMonths(oneDecimal.format(days / 30));
  return l.intervalYears(oneDecimal.format(days / 365));
}

/// A rating button's interval, short: "<10m", "5h", "1.5h", "3d", "1.2mo",
/// "2y" (as Go's core.FormatInterval).
String formatIntervalLabel(Duration d, {AppLocalizations? l10n}) {
  final l = l10n ?? appL10n;
  // One decimal at most, without a trailing zero.
  String num(double v) => NumberFormat('0.#', l.localeName).format(v);
  const day = Duration(days: 1);
  if (d < const Duration(hours: 1)) {
    final m = (d.inSeconds / 60).ceil();
    return l.intervalLabelMinutes(m < 1 ? 1 : m);
  }
  if (d < day) return l.intervalLabelHours(num(d.inSeconds / 3600));
  final days = (d.inSeconds / day.inSeconds).round();
  if (days < 30) return l.intervalLabelDays(days);
  if (days < 365) return l.intervalLabelMonths(num(days / 30));
  return l.intervalLabelYears(num(days / 365));
}

/// "512 bytes", "1.5 KB", "12.3 MB", "1.1 GB".
String formatBytes(int bytes, {AppLocalizations? l10n}) {
  final l = l10n ?? appL10n;
  if (bytes < 1024) return l.sizeBytes(bytes);
  const units = ['KB', 'MB', 'GB'];
  var value = bytes / 1024;
  var unit = 0;
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024;
    unit++;
  }
  return '${NumberFormat('0.0', l.localeName).format(value)} ${units[unit]}';
}
