import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import 'package:cardflip/data/scope.dart';
import 'package:cardflip/features/progress/achievements_screen.dart';
import 'package:cardflip/features/study/study_screen.dart' show ratingNames;
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/time_format.dart';
import 'package:cardflip/l10n/l10n.dart';

import 'stats_charts.dart';

/// The sections of the statistics screen, in screen order.
enum StatsSection {
  today,
  streaks,
  futureDue,
  calendar,
  reviews,
  reviewTime,
  counts,
  intervals,
  stability,
  difficulty,
  retrievability,
  hourly,
  buttons,
  trueRetention;

  /// The section's heading, in the interface language.
  String get title {
    final l = appL10n;
    return switch (this) {
      today => l.secToday,
      streaks => l.secStreaks,
      futureDue => l.secFutureDue,
      calendar => l.secCalendar,
      reviews => l.secReviews,
      reviewTime => l.secReviewTime,
      counts => l.secCounts,
      intervals => l.secIntervals,
      stability => l.secStability,
      difficulty => l.secDifficulty,
      retrievability => l.secRetrievability,
      hourly => l.secHourly,
      buttons => l.secButtons,
      trueRetention => l.secTrueRetention,
    };
  }

  /// Whether the section depends on the period.
  bool get timeBased => switch (this) {
    streaks || futureDue || reviews || reviewTime || hourly || buttons => true,
    _ => false,
  };

  /// Asks Go for the section's data.
  Future<Object> load(golang.Scope scope, String period) => switch (this) {
    today => golang.stats.today(scope, period),
    streaks => golang.stats.streaks(scope, period),
    futureDue => golang.stats.futureDue(scope, period),
    calendar => golang.stats.calendar(scope, period),
    reviews => golang.stats.reviews(scope, period),
    reviewTime => golang.stats.reviewTime(scope, period),
    counts => golang.stats.counts(scope, period),
    intervals => golang.stats.intervals(scope, period),
    stability => golang.stats.stability(scope, period),
    difficulty => golang.stats.difficulty(scope, period),
    retrievability => golang.stats.retrievability(scope, period),
    hourly => golang.stats.hourly(scope, period),
    buttons => golang.stats.buttons(scope, period),
    trueRetention => golang.stats.trueRetention(scope, period),
  };

  /// The section's content for [data], as [load] returned it.
  Widget content(Object data) => switch (this) {
    today => TodayView(data as golang.TodayStats),
    streaks => StreaksView(data as golang.StreakSection),
    futureDue => FutureDueView(data as golang.Forecast),
    calendar => CalendarView(data as golang.Calendar),
    reviews => ReviewHistoryView(data as golang.ReviewHistory),
    reviewTime => ReviewHistoryView(data as golang.ReviewHistory, time: true),
    counts => CardCountsView(data as golang.StateCounts),
    intervals => HistogramView(
      data as golang.Histogram,
      kind: HistogramKind.intervals,
    ),
    stability => HistogramView(
      data as golang.Histogram,
      kind: HistogramKind.stability,
    ),
    difficulty => HistogramView(
      data as golang.Histogram,
      kind: HistogramKind.difficulty,
    ),
    retrievability => HistogramView(
      data as golang.Histogram,
      kind: HistogramKind.retrievability,
    ),
    hourly => HourlyView(data as List<golang.HourBucket>),
    buttons => AnswerButtonsView(data as golang.ButtonStats),
    trueRetention => TrueRetentionView(data as golang.RetentionTable),
  };
}

/// The periods of the time-based sections, with their labels.
Map<String, String> get statsPeriods => {
  'month': appL10n.period1Month,
  '3months': appL10n.period3Months,
  'year': appL10n.period1Year,
  'all': appL10n.periodAll,
};

/// "90%", from a share between 0 and 1.
String percent(double share) =>
    NumberFormat.percentPattern(appL10n.localeName).format(share);

/// "90.0%", from a share between 0 and 1.
String percent1(double share) => NumberFormat.decimalPercentPattern(
  locale: appL10n.localeName,
  decimalDigits: 1,
).format(share);

/// A count, "1,234".
String count(num n) =>
    NumberFormat.decimalPattern(appL10n.localeName).format(n);

/// A number with one decimal, "12.5".
String oneDecimal(num n) => NumberFormat('0.0', appL10n.localeName).format(n);

/// Parses a "2006-01-02" date from Go.
DateTime parseDay(String date) {
  final [y, m, d] = date.split('-').map(int.parse).toList();
  return DateTime(y, m, d);
}

/// "Oct 13", from a date.
String shortDate(DateTime d) {
  ensureDateFormats();
  return DateFormat.MMMd(appL10n.localeName).format(d);
}

/// "Oct 2026", from a date.
String monthYear(DateTime d) {
  ensureDateFormats();
  return DateFormat.yMMM(appL10n.localeName).format(d);
}

/// Label–value rows under a graph.
class StatLines extends StatelessWidget {
  const StatLines(this.lines, {super.key});

  final List<(String, String)> lines;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        for (final (label, value) in lines)
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 2),
            child: Row(
              children: [
                Expanded(child: Text(label, style: theme.textTheme.bodyMedium)),
                Flexible(
                  child: Text(
                    value,
                    textAlign: TextAlign.end,
                    style: theme.textTheme.bodyMedium?.copyWith(
                      fontWeight: FontWeight.w600,
                      fontFeatures: const [FontFeature.tabularFigures()],
                    ),
                  ),
                ),
              ],
            ),
          ),
      ],
    );
  }
}

/// The text of the today summary, shared with the PDF.
List<String> todayLines(golang.TodayStats t) {
  final l = appL10n;
  if (t.cards == 0) return [l.noCardsStudiedToday];
  return [
    l.studiedToday(t.cards, formatSpent(t.time)),
    l.againCount(count(t.again), percent(t.correct)),
    l.todayKinds(
      count(t.learning),
      count(t.review),
      count(t.relearning),
      count(t.filtered),
    ),
  ];
}

class TodayView extends StatelessWidget {
  const TodayView(this.today, {super.key});

  final golang.TodayStats today;

  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      for (final line in todayLines(today))
        Padding(
          padding: const EdgeInsets.symmetric(vertical: 2),
          child: Text(line),
        ),
    ],
  );
}

/// The label of day [i] of a forecast: "Today", "Tomorrow", "+5".
String forecastDay(int i) => switch (i) {
  0 => appL10n.today,
  1 => appL10n.tomorrow,
  _ => '+$i',
};

List<(String, String)> futureDueLines(golang.Forecast f) => [
  (appL10n.overdue, count(f.overdue)),
  (appL10n.dueTomorrowLabel, count(f.dueTomorrow)),
  (appL10n.total, count(f.total)),
  (appL10n.average, appL10n.cardsPerDay(oneDecimal(f.average))),
];

class FutureDueView extends StatelessWidget {
  const FutureDueView(this.forecast, {super.key});

  final golang.Forecast forecast;

  @override
  Widget build(BuildContext context) {
    final colors = StatsColors.of(context);
    final f = forecast;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        StackedBarChart(
          bars: [
            for (final d in f.days) [d.toDouble()],
          ],
          colors: [colors.bar],
          xLabel: (i) => i == 0 ? context.l10n.today : '+$i',
          tooltip: (i) => context.l10n.forecastTip(
            forecastDay(i),
            f.days[i],
            count(f.cumulative[i]),
          ),
        ),
        const SizedBox(height: 8),
        Text(
          context.l10n.runningTotal,
          style: Theme.of(context).textTheme.labelMedium,
        ),
        ValueLineChart(
          values: [for (final c in f.cumulative) c.toDouble()],
          xLabel: forecastDay,
          color: colors.line,
        ),
        const SizedBox(height: 8),
        StatLines(futureDueLines(f)),
      ],
    );
  }
}

class CalendarView extends StatelessWidget {
  const CalendarView(this.calendar, {super.key});

  final golang.Calendar calendar;

  @override
  Widget build(BuildContext context) {
    final c = calendar;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        CalendarHeatmap(
          today: parseDay(c.today),
          counts: {for (final d in c.days) d.day: d.count},
          max: c.max,
        ),
        const SizedBox(height: 8),
        StatLines(calendarLines(c)),
      ],
    );
  }
}

/// The Streaks section: its lines and, while streaks and goals are shown,
/// a link to the achievements.
class StreaksView extends StatelessWidget {
  const StreaksView(this.streaks, {super.key});

  final golang.StreakSection streaks;

  @override
  Widget build(BuildContext context) {
    final lines = StatLines(streakLines(streaks));
    final settings = AppScope.maybeRead(context)?.settings;
    if (settings == null) return lines;
    return ListenableBuilder(
      listenable: settings,
      builder: (context, _) => Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          lines,
          if (settings.showProgress)
            Align(
              alignment: AlignmentDirectional.centerStart,
              child: TextButton.icon(
                key: const Key('streaks-achievements'),
                icon: const Icon(Icons.emoji_events_outlined),
                label: Text(context.l10n.achievements),
                onPressed: () => showAchievements(context),
              ),
            ),
        ],
      ),
    );
  }
}

/// The lines of the Streaks section, shared with the PDF. Days the goal was
/// met are left out without a daily goal.
List<(String, String)> streakLines(golang.StreakSection s) {
  final l = appL10n;
  final longest = l.countDays(s.longest);
  return [
    (l.currentStreak, l.countDays(s.current)),
    (
      l.longestStreak,
      switch ((s.longestFirst, s.longestLast)) {
        (final first?, final last?) => l.longestStreakValue(
          longest,
          shortDate(parseDay(first)),
          shortDate(parseDay(last)),
        ),
        _ => longest,
      },
    ),
    (l.studyDaysInPeriod, l.countOf(count(s.studyDays), count(s.periodDays))),
    if (s.goalDays case final days?) (l.goalDaysMet, count(days)),
  ];
}

List<(String, String)> calendarLines(golang.Calendar c) => [
  (appL10n.daysStudiedYear, appL10n.countOf(count(c.days.length), count(365))),
  (appL10n.mostReviewsDay, count(c.max)),
];

/// The label of a review history bucket: a date, or a month.
String historyLabel(golang.ReviewHistory h, int i) {
  final start = parseDay(h.today).add(Duration(days: h.buckets[i].start));
  return h.unit == 'month' ? monthYear(start) : shortDate(start);
}

List<double> historyValues(golang.ReviewBucket b, {bool time = false}) => [
  for (final v in [b.learning, b.young, b.mature, b.relearning, b.filtered])
    time ? v / 60000 : v.toDouble(),
];

/// Minutes for review time, counts for reviews.
String historyAmount(num v, {required bool time}) =>
    time ? appL10n.minutesAmount(compactNumber(v / 60000)) : compactNumber(v);

List<(String, String)> historyLines(
  golang.ReviewHistory h, {
  bool time = false,
}) {
  final perDay = h.daysStudied == 0 ? 0 : h.total / h.daysStudied;
  final overPeriod = h.periodDays == 0 ? 0 : h.total / h.periodDays;
  return [
    (
      appL10n.daysStudied,
      appL10n.countOf(count(h.daysStudied), count(h.periodDays)),
    ),
    (appL10n.total, historyAmount(h.total, time: time)),
    (appL10n.averageDaysStudied, historyAmount(perDay, time: time)),
    (appL10n.averageOverPeriod, historyAmount(overPeriod, time: time)),
  ];
}

class ReviewHistoryView extends StatelessWidget {
  const ReviewHistoryView(this.history, {super.key, this.time = false});

  final golang.ReviewHistory history;

  /// Whether this is review time, in milliseconds, rather than counts.
  final bool time;

  @override
  Widget build(BuildContext context) {
    final colors = StatsColors.of(context);
    final h = history;
    String tip(int i) {
      final b = h.buckets[i];
      final v = [b.learning, b.young, b.mature, b.relearning, b.filtered];
      return [
        historyLabel(h, i),
        for (final (k, name) in reviewKinds.indexed)
          if (v[k] > 0) '$name: ${historyAmount(v[k], time: time)}',
      ].join('\n');
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        StackedBarChart(
          bars: [for (final b in h.buckets) historyValues(b, time: time)],
          colors: colors.kinds,
          xLabel: (i) => historyLabel(h, i),
          yLabel: time
              ? (v) => context.l10n.minutesAxis(compactNumber(v))
              : compactNumber,
          tooltip: tip,
        ),
        const SizedBox(height: 8),
        Legend(labels: reviewKinds, colors: colors.kinds),
        const SizedBox(height: 8),
        StatLines(historyLines(h, time: time)),
      ],
    );
  }
}

/// The rows of the card counts: label and count.
List<(String, int)> countRows(golang.StateCounts c) => [
  (appL10n.countNew, c.new_),
  (appL10n.kindLearning, c.learning),
  (appL10n.kindRelearning, c.relearning),
  (appL10n.kindYoung, c.young),
  (appL10n.kindMature, c.mature),
  (appL10n.kindSuspended, c.suspended),
  (appL10n.kindBuried, c.buried),
];

class CardCountsView extends StatelessWidget {
  const CardCountsView(this.counts, {super.key});

  final golang.StateCounts counts;

  @override
  Widget build(BuildContext context) {
    final colors = StatsColors.of(context);
    final palette = [
      colors.newCards,
      colors.learning,
      colors.relearning,
      colors.young,
      colors.mature,
      colors.suspended,
      colors.buried,
    ];
    final rows = countRows(counts);
    final total = counts.total;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (total > 0)
          ClipRRect(
            borderRadius: BorderRadius.circular(4),
            child: SizedBox(
              height: 16,
              child: Row(
                children: [
                  for (final (i, (_, n)) in rows.indexed)
                    if (n > 0)
                      Expanded(
                        flex: n,
                        child: ColoredBox(color: palette[i]),
                      ),
                ],
              ),
            ),
          ),
        const SizedBox(height: 8),
        for (final (i, (label, n)) in rows.indexed)
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 2),
            child: Row(
              children: [
                LegendItem(color: palette[i], label: label),
                const Spacer(),
                Text(
                  total == 0 ? count(n) : '${count(n)} (${percent(n / total)})',
                  style: const TextStyle(
                    fontFeatures: [FontFeature.tabularFigures()],
                  ),
                ),
              ],
            ),
          ),
        const Divider(),
        StatLines([(context.l10n.total, count(total))]),
      ],
    );
  }
}

/// What a histogram shows.
enum HistogramKind { intervals, stability, difficulty, retrievability }

/// The label of a histogram bucket: "1d", "1-7d", "20-25%". The last
/// bucket of intervals includes its upper end.
String bucketLabel(
  golang.HistogramBucket b,
  HistogramKind kind, {
  bool last = false,
}) {
  switch (kind) {
    case HistogramKind.difficulty || HistogramKind.retrievability:
      return '${(b.lo * 100).round()}-${(b.hi * 100).round()}%';
    case HistogramKind.intervals:
      final hi = last ? b.hi : b.hi - 1;
      if (hi <= b.lo) return appL10n.bucketDays(compactNumber(b.lo));
      return appL10n.bucketDays('${compactNumber(b.lo)}-${compactNumber(hi)}');
    case HistogramKind.stability:
      return appL10n.bucketDays(
        '${compactNumber(b.lo)}-${compactNumber(b.hi)}',
      );
  }
}

List<(String, String)> histogramLines(golang.Histogram h, HistogramKind kind) =>
    switch (kind) {
      HistogramKind.intervals => [
        (appL10n.reviewCards, count(h.count)),
        (appL10n.averageInterval, appL10n.daysDecimal(oneDecimal(h.average))),
      ],
      HistogramKind.stability => [
        (appL10n.cards, count(h.count)),
        (appL10n.averageStability, appL10n.daysDecimal(oneDecimal(h.average))),
      ],
      HistogramKind.difficulty => [
        (appL10n.cards, count(h.count)),
        (appL10n.averageDifficulty, percent(h.average)),
      ],
      HistogramKind.retrievability => [
        (appL10n.cards, count(h.count)),
        (appL10n.averageRetrievability, percent(h.average)),
        (appL10n.estimatedKnowledge, appL10n.countCards(h.sum.round())),
      ],
    };

class HistogramView extends StatelessWidget {
  const HistogramView(this.histogram, {super.key, required this.kind});

  final golang.Histogram histogram;
  final HistogramKind kind;

  @override
  Widget build(BuildContext context) {
    final h = histogram;
    if (h.count == 0) {
      return Text(context.l10n.noCardsInfo);
    }
    final colors = StatsColors.of(context);
    final last = h.buckets.length - 1;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        StackedBarChart(
          bars: [
            for (final b in h.buckets) [b.count.toDouble()],
          ],
          colors: [
            kind == HistogramKind.intervals ? colors.mature : colors.bar,
          ],
          xLabel: (i) => bucketLabel(h.buckets[i], kind, last: i == last),
          tooltip: (i) =>
              '${bucketLabel(h.buckets[i], kind, last: i == last)}: '
              '${context.l10n.countCards(h.buckets[i].count)}',
        ),
        const SizedBox(height: 8),
        StatLines(histogramLines(h, kind)),
      ],
    );
  }
}

class HourlyView extends StatelessWidget {
  const HourlyView(this.hours, {super.key});

  final List<golang.HourBucket> hours;

  @override
  Widget build(BuildContext context) {
    final colors = StatsColors.of(context);
    final total = hours.fold<int>(0, (a, h) => a + h.reviews);
    if (total == 0) return Text(context.l10n.noReviewsPeriod);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        StackedBarChart(
          bars: [
            for (final h in hours) [h.reviews.toDouble()],
          ],
          colors: [colors.bar],
          xLabel: (i) => '$i',
          tooltip: (i) =>
              context.l10n.hourTip('$i:00-${i + 1}:00', hours[i].reviews) +
              (hours[i].reviews > 0
                  ? context.l10n.hourTipCorrect(percent(hours[i].correct))
                  : ''),
        ),
        const SizedBox(height: 8),
        Text(
          context.l10n.successRate,
          style: Theme.of(context).textTheme.labelMedium,
        ),
        ValueLineChart(
          values: [for (final h in hours) h.correct * 100],
          xLabel: (i) => '$i:00',
          yLabel: (v) => percent(v / 100),
          maxY: 100,
          color: colors.line,
        ),
      ],
    );
  }
}

/// The groups of the answer buttons: label and counts.
List<(String, golang.ButtonGroup)> buttonGroups(golang.ButtonStats b) => [
  (appL10n.kindLearning, b.learning),
  (appL10n.kindYoung, b.young),
  (appL10n.kindMature, b.mature),
];

/// The names of the answer buttons, Again to Easy.
List<String> get buttonNames => ratingNames(appL10n);

String buttonLine(golang.ButtonGroup g) => g.total == 0
    ? appL10n.noAnswers
    : appL10n.percentCorrect(percent(g.correct), count(g.total));

class AnswerButtonsView extends StatelessWidget {
  const AnswerButtonsView(this.buttons, {super.key});

  final golang.ButtonStats buttons;

  @override
  Widget build(BuildContext context) {
    final colors = StatsColors.of(context);
    final groups = buttonGroups(buttons);
    final bars = <List<double>>[];
    final labels = <String>[];
    final tips = <String>[];
    for (final (name, g) in groups) {
      for (final (i, n) in g.counts.indexed) {
        bars.add([n.toDouble()]);
        labels.add(i == 0 ? name : '');
        tips.add('$name, ${buttonNames[i]}: ${count(n)}');
      }
    }
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        StackedBarChart(
          bars: bars,
          colors: [colors.bar],
          xLabel: (i) => labels[i],
          tooltip: (i) => tips[i],
        ),
        const SizedBox(height: 8),
        Text(
          context.l10n.eachGroupShows,
          style: Theme.of(context).textTheme.bodySmall,
        ),
        const SizedBox(height: 4),
        StatLines([for (final (name, g) in groups) (name, buttonLine(g))]),
      ],
    );
  }
}

/// The label of a true retention period.
String retentionPeriod(String period) => switch (period) {
  'today' => appL10n.today,
  'yesterday' => appL10n.yesterday,
  'week' => appL10n.lastWeek,
  'month' => appL10n.lastMonth,
  'year' => appL10n.lastYear,
  _ => appL10n.allTime,
};

/// "90.0% (180/200)", or "-" without reviews.
String retentionCell(golang.RetentionCount c) {
  final r = c.retention;
  if (r == null) return '-';
  return '${percent1(r)} (${count(c.passed)}/${count(c.passed + c.failed)})';
}

class TrueRetentionView extends StatelessWidget {
  const TrueRetentionView(this.table, {super.key});

  final golang.RetentionTable table;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final bold = theme.textTheme.bodyMedium?.copyWith(
      fontWeight: FontWeight.w600,
    );
    Widget cell(String text, {TextStyle? style, bool end = true}) => Padding(
      padding: const EdgeInsets.symmetric(vertical: 4, horizontal: 4),
      child: Text(
        text,
        textAlign: end ? TextAlign.end : TextAlign.start,
        style: (style ?? theme.textTheme.bodyMedium)?.copyWith(
          fontFeatures: const [FontFeature.tabularFigures()],
        ),
      ),
    );
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        SingleChildScrollView(
          scrollDirection: Axis.horizontal,
          child: Table(
            defaultColumnWidth: const IntrinsicColumnWidth(),
            children: [
              TableRow(
                children: [
                  cell('', end: false),
                  cell(context.l10n.kindYoung, style: bold),
                  cell(context.l10n.kindMature, style: bold),
                  cell(context.l10n.total, style: bold),
                ],
              ),
              for (final r in table.rows)
                TableRow(
                  children: [
                    cell(retentionPeriod(r.period), end: false),
                    cell(retentionCell(r.young)),
                    cell(retentionCell(r.mature)),
                    cell(retentionCell(r.total)),
                  ],
                ),
            ],
          ),
        ),
        const SizedBox(height: 8),
        StatLines([(context.l10n.desiredRetention, percent(table.desired))]),
      ],
    );
  }
}
