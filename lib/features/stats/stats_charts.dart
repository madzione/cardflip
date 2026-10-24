import 'dart:math' as math;

import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';

import 'package:cardflip/app/theme.dart';
import 'package:cardflip/l10n/l10n.dart';
import 'package:cardflip/widgets/time_format.dart';

/// Colours of the statistics graphs: what a card was when it was answered
/// (learning, young, mature, relearning, filtered) and card states.
@immutable
class StatsColors {
  const StatsColors({
    required this.learning,
    required this.young,
    required this.mature,
    required this.relearning,
    required this.filtered,
    required this.newCards,
    required this.suspended,
    required this.buried,
    required this.bar,
    required this.line,
  });

  final Color learning;
  final Color young;
  final Color mature;
  final Color relearning;
  final Color filtered;
  final Color newCards;
  final Color suspended;
  final Color buried;

  /// The colour of single-series bars and of lines.
  final Color bar;
  final Color line;

  static const light = StatsColors(
    learning: Color(0xFFEF6C00),
    young: Color(0xFF81C784),
    mature: Color(0xFF2E7D32),
    relearning: Color(0xFFC62828),
    filtered: Color(0xFF1E88E5),
    newCards: Color(0xFF1E6FD9),
    suspended: Color(0xFFF9A825),
    buried: Color(0xFF9E9E9E),
    bar: Color(0xFF3F6FB5),
    line: Color(0xFF6A3FB5),
  );

  static const dark = StatsColors(
    learning: Color(0xFFFFB74D),
    young: Color(0xFFA5D6A7),
    mature: Color(0xFF66BB6A),
    relearning: Color(0xFFEF9A9A),
    filtered: Color(0xFF64B5F6),
    newCards: Color(0xFF64B5F6),
    suspended: Color(0xFFFFEE58),
    buried: Color(0xFFBDBDBD),
    bar: Color(0xFF8AB4F8),
    line: Color(0xFFCE93D8),
  );

  static StatsColors of(BuildContext context) {
    final dark = Theme.of(context).brightness == Brightness.dark;
    final base = dark ? StatsColors.dark : StatsColors.light;
    final counts = CountColors.of(context);
    return StatsColors(
      learning: base.learning,
      young: base.young,
      mature: base.mature,
      relearning: base.relearning,
      filtered: base.filtered,
      newCards: counts.newCards,
      suspended: base.suspended,
      buried: base.buried,
      bar: base.bar,
      line: base.line,
    );
  }

  /// The colours of a review history's series, in [reviewKinds] order.
  List<Color> get kinds => [learning, young, mature, relearning, filtered];
}

/// The series of a review history, in the order they are stacked.
List<String> get reviewKinds => [
  appL10n.kindLearning,
  appL10n.kindYoung,
  appL10n.kindMature,
  appL10n.kindRelearning,
  appL10n.kindFiltered,
];

/// A coloured square and a label, for chart legends.
class LegendItem extends StatelessWidget {
  const LegendItem({super.key, required this.color, required this.label});

  final Color color;
  final String label;

  @override
  Widget build(BuildContext context) => Row(
    mainAxisSize: MainAxisSize.min,
    children: [
      Container(
        width: 12,
        height: 12,
        decoration: BoxDecoration(
          color: color,
          borderRadius: BorderRadius.circular(2),
        ),
      ),
      const SizedBox(width: 6),
      Text(label, style: Theme.of(context).textTheme.bodySmall),
    ],
  );
}

/// A legend of [labels] in [colors].
class Legend extends StatelessWidget {
  const Legend({super.key, required this.labels, required this.colors});

  final List<String> labels;
  final List<Color> colors;

  @override
  Widget build(BuildContext context) => Wrap(
    spacing: 16,
    runSpacing: 4,
    children: [
      for (final (i, l) in labels.indexed)
        LegendItem(color: colors[i], label: l),
    ],
  );
}

/// A compact number for an axis: 950, 1.2k, 35k, 1.5M.
String compactNumber(num v) {
  final a = v.abs();
  String trim(double x) {
    final s = x.toStringAsFixed(x < 10 ? 1 : 0);
    return s.endsWith('.0') ? s.substring(0, s.length - 2) : s;
  }

  if (a >= 1e6) return '${trim(v / 1e6)}M';
  if (a >= 1e3) return '${trim(v / 1e3)}k';
  if (v == v.roundToDouble()) return v.round().toString();
  return v.toStringAsFixed(1);
}

/// A step for [max] that gives about four grid lines at a round value.
double niceStep(double max) {
  if (max <= 0) return 1;
  final raw = max / 4;
  final mag = math.pow(10, (math.log(raw) / math.ln10).floor()).toDouble();
  for (final m in [1, 2, 2.5, 5, 10]) {
    if (raw <= m * mag) return m * mag;
  }
  return 10 * mag;
}

FlTitlesData _titles(
  BuildContext context, {
  required int count,
  required String Function(int i) xLabel,
  required int labelEvery,
  required String Function(double v) yLabel,
  required double step,
}) {
  final style = Theme.of(context).textTheme.labelSmall;
  return FlTitlesData(
    topTitles: const AxisTitles(),
    rightTitles: const AxisTitles(),
    leftTitles: AxisTitles(
      sideTitles: SideTitles(
        showTitles: true,
        reservedSize: 40,
        interval: step,
        getTitlesWidget: (v, meta) => SideTitleWidget(
          meta: meta,
          child: Text(yLabel(v), style: style),
        ),
      ),
    ),
    bottomTitles: AxisTitles(
      sideTitles: SideTitles(
        showTitles: true,
        reservedSize: 22,
        interval: 1,
        getTitlesWidget: (v, meta) {
          final i = v.round();
          if (v != i || i < 0 || i >= count) return const SizedBox.shrink();
          final last = i == count - 1;
          if (i % labelEvery != 0 && !last) return const SizedBox.shrink();
          if (!last && count - 1 - i < labelEvery / 2) {
            return const SizedBox.shrink();
          }
          return SideTitleWidget(
            meta: meta,
            child: Text(xLabel(i), style: style),
          );
        },
      ),
    ),
  );
}

FlGridData _grid(BuildContext context, double step) => FlGridData(
  drawVerticalLine: false,
  horizontalInterval: step,
  getDrawingHorizontalLine: (_) => FlLine(
    color: Theme.of(context).colorScheme.outlineVariant,
    strokeWidth: 0.5,
  ),
);

/// Bars of one or more stacked series: [bars] holds each bar's value per
/// series, drawn bottom up in [colors]. Hovering or tapping a bar shows
/// [tooltip] for it.
class StackedBarChart extends StatelessWidget {
  const StackedBarChart({
    super.key,
    required this.bars,
    required this.colors,
    required this.xLabel,
    required this.tooltip,
    this.yLabel = compactNumber,
    this.height = 180,
  });

  final List<List<double>> bars;
  final List<Color> colors;
  final String Function(int i) xLabel;
  final String Function(int i) tooltip;
  final String Function(double v) yLabel;
  final double height;

  @override
  Widget build(BuildContext context) {
    final totals = [for (final b in bars) b.fold<double>(0, (a, v) => a + v)];
    final top = totals.fold<double>(0, math.max);
    final step = niceStep(top);
    final maxY = top <= 0 ? 1.0 : (top / step).ceil() * step;
    return SizedBox(
      height: height,
      child: LayoutBuilder(
        builder: (context, constraints) {
          final n = math.max(bars.length, 1);
          final plot = math.max(constraints.maxWidth - 48, 1.0);
          final width = (plot / n * 0.7).clamp(1.0, 28.0);
          final labelEvery = math.max(1, (n * 48 / plot).ceil());
          return BarChart(
            duration: Duration.zero,
            BarChartData(
              maxY: maxY,
              alignment: BarChartAlignment.spaceAround,
              barGroups: [
                for (final (i, b) in bars.indexed)
                  BarChartGroupData(
                    x: i,
                    barRods: [
                      BarChartRodData(
                        toY: totals[i],
                        width: width,
                        color: colors.first,
                        borderRadius: BorderRadius.zero,
                        rodStackItems: b.length < 2
                            ? null
                            : [
                                for (var s = 0, from = 0.0; s < b.length; s++)
                                  BarChartRodStackItem(
                                    from,
                                    from += b[s],
                                    colors[s % colors.length],
                                  ),
                              ],
                      ),
                    ],
                  ),
              ],
              titlesData: _titles(
                context,
                count: bars.length,
                xLabel: xLabel,
                labelEvery: labelEvery,
                yLabel: yLabel,
                step: step,
              ),
              gridData: _grid(context, step),
              borderData: FlBorderData(show: false),
              barTouchData: BarTouchData(
                touchTooltipData: BarTouchTooltipData(
                  fitInsideHorizontally: true,
                  fitInsideVertically: true,
                  getTooltipColor: (_) =>
                      Theme.of(context).colorScheme.inverseSurface,
                  getTooltipItem: (group, _, _, _) => BarTooltipItem(
                    tooltip(group.x),
                    TextStyle(
                      color: Theme.of(context).colorScheme.onInverseSurface,
                      fontSize: 12,
                    ),
                  ),
                ),
              ),
            ),
          );
        },
      ),
    );
  }
}

/// A line through [values], one point per x position.
class ValueLineChart extends StatelessWidget {
  const ValueLineChart({
    super.key,
    required this.values,
    required this.xLabel,
    required this.color,
    this.yLabel = compactNumber,
    this.maxY,
    this.height = 120,
  });

  final List<double> values;
  final String Function(int i) xLabel;
  final String Function(double v) yLabel;
  final Color color;

  /// The top of the scale; by default a round value above the largest.
  final double? maxY;
  final double height;

  @override
  Widget build(BuildContext context) => MultiLineChart(
    lines: [values],
    colors: [color],
    xLabel: xLabel,
    yLabel: yLabel,
    maxY: maxY,
    height: height,
  );
}

/// Lines through each of [lines] in [colors], one point per x position.
/// With [lineLabels], the tooltip names each line's value; otherwise it
/// shows the x label and the value.
class MultiLineChart extends StatelessWidget {
  const MultiLineChart({
    super.key,
    required this.lines,
    required this.colors,
    required this.xLabel,
    this.lineLabels,
    this.yLabel = compactNumber,
    this.maxY,
    this.height = 120,
  });

  final List<List<double>> lines;
  final List<Color> colors;
  final List<String>? lineLabels;
  final String Function(int i) xLabel;
  final String Function(double v) yLabel;

  /// The top of the scale; by default a round value above the largest.
  final double? maxY;
  final double height;

  @override
  Widget build(BuildContext context) {
    final count = lines.fold<int>(0, (a, l) => math.max(a, l.length));
    final top =
        maxY ?? lines.fold<double>(0, (a, l) => l.fold<double>(a, math.max));
    final step = niceStep(top);
    final scaleTop = maxY ?? (top <= 0 ? 1.0 : (top / step).ceil() * step);
    final labels = lineLabels;
    return SizedBox(
      height: height,
      child: LayoutBuilder(
        builder: (context, constraints) {
          final n = math.max(count, 1);
          final plot = math.max(constraints.maxWidth - 48, 1.0);
          final labelEvery = math.max(1, (n * 48 / plot).ceil());
          final tooltipStyle = TextStyle(
            color: Theme.of(context).colorScheme.onInverseSurface,
            fontSize: 12,
          );
          return LineChart(
            duration: Duration.zero,
            LineChartData(
              minY: 0,
              maxY: scaleTop,
              minX: 0,
              maxX: math.max(count - 1, 1).toDouble(),
              lineBarsData: [
                for (final (k, values) in lines.indexed)
                  LineChartBarData(
                    spots: [
                      for (final (i, v) in values.indexed)
                        FlSpot(i.toDouble(), v),
                    ],
                    color: colors[k % colors.length],
                    barWidth: 2,
                    dotData: const FlDotData(show: false),
                  ),
              ],
              titlesData: _titles(
                context,
                count: count,
                xLabel: xLabel,
                labelEvery: labelEvery,
                yLabel: yLabel,
                step: step,
              ),
              gridData: _grid(context, step),
              borderData: FlBorderData(show: false),
              lineTouchData: LineTouchData(
                touchTooltipData: LineTouchTooltipData(
                  fitInsideHorizontally: true,
                  getTooltipColor: (_) =>
                      Theme.of(context).colorScheme.inverseSurface,
                  getTooltipItems: (spots) => [
                    for (final (j, s) in spots.indexed)
                      LineTooltipItem(
                        labels == null
                            ? '${xLabel(s.x.round())}: ${yLabel(s.y)}'
                            : '${j == 0 ? '${xLabel(s.x.round())}\n' : ''}'
                                  '${labels[s.barIndex]}: ${yLabel(s.y)}',
                        tooltipStyle,
                      ),
                  ],
                ),
              ),
            ),
          );
        },
      ),
    );
  }
}

/// A year of days as a grid of weeks (columns) by weekday (rows), each day
/// shaded by its count: darker for more. The newest week is on the right.
class CalendarHeatmap extends StatelessWidget {
  const CalendarHeatmap({
    super.key,
    required this.today,
    required this.counts,
    required this.max,
    this.days = 365,
  });

  /// Today's date, and the count per day relative to it (0 today, -1
  /// yesterday).
  final DateTime today;
  final Map<int, int> counts;
  final int max;
  final int days;

  /// The shade of a count: 0 for none, then 1 to 4.
  static int level(int count, int max) {
    if (count <= 0 || max <= 0) return 0;
    return (4 * count / max).ceil().clamp(1, 4);
  }

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final base = StatsColors.of(context).mature;
    final empty = scheme.surfaceContainerHighest;
    // Columns start on Monday; the first column holds the oldest day.
    final todayRow = today.weekday - 1;
    final first = -(days - 1);
    final firstRow = (todayRow + first) % 7;
    final weeks = ((days + firstRow + (6 - todayRow)) / 7).ceil();
    const cell = 11.0, gap = 2.0;
    Widget day(int rel) {
      final n = counts[rel] ?? 0;
      final l = level(n, max);
      final date = today.add(Duration(days: rel));
      return Tooltip(
        message: context.l10n.calendarTip(formatDate(date), n),
        child: Container(
          key: l > 0 ? Key('calendar-day-$rel') : null,
          width: cell,
          height: cell,
          margin: const EdgeInsets.all(gap / 2),
          decoration: BoxDecoration(
            color: l == 0
                ? empty
                : base.withValues(alpha: 0.25 + 0.25 * (l - 1)),
            borderRadius: BorderRadius.circular(2),
          ),
        ),
      );
    }

    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      reverse: true,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          for (var w = 0; w < weeks; w++)
            Column(
              children: [
                for (var r = 0; r < 7; r++)
                  () {
                    final rel = first - firstRow + w * 7 + r;
                    if (rel < first || rel > 0) {
                      return const SizedBox(
                        width: cell + gap,
                        height: cell + gap,
                      );
                    }
                    return day(rel);
                  }(),
              ],
            ),
        ],
      ),
    );
  }
}
