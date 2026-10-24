import 'dart:math' as math;
import 'dart:typed_data';

import 'package:pdf/pdf.dart';
import 'package:pdf/widgets.dart' as pw;

import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/time_format.dart';
import 'package:cardflip/l10n/l10n.dart';

import 'stats_charts.dart' show reviewKinds;
import 'stats_sections.dart';

/// The file name of a statistics PDF: "Statistics - Spanish - 2026-10-13.pdf".
String statsPdfName(String scope, DateTime date) {
  final safe = scope
      .replaceAll('::', ' - ')
      .replaceAll(RegExp(r'[\\/:*?"<>|\x00-\x1f]'), '_')
      .trim();
  return appL10n.statsPdfName(safe, formatDate(date));
}

/// The colours of the PDF's graphs, in the order of [reviewKinds].
const _kindColors = [
  PdfColor.fromInt(0xFFEF6C00),
  PdfColor.fromInt(0xFF81C784),
  PdfColor.fromInt(0xFF2E7D32),
  PdfColor.fromInt(0xFFC62828),
  PdfColor.fromInt(0xFF1E88E5),
];
const _barColor = PdfColor.fromInt(0xFF3F6FB5);
const _grey = PdfColor.fromInt(0xFF757575);
const _empty = PdfColor.fromInt(0xFFEEEEEE);

/// Builds a PDF of the statistics screen: a heading naming [scope],
/// [period] and [date], then every section in [data] (sections that failed
/// to load are missing from it and left out). [compress] is off in tests,
/// so the text can be read from the bytes.
Future<Uint8List> buildStatsPdf({
  required String scope,
  required String period,
  required DateTime date,
  required Map<StatsSection, Object> data,
  bool compress = true,
}) {
  final doc = pw.Document(
    compress: compress,
    title: appL10n.statsTitleScope(scope),
    creator: 'Cardflip',
  );
  doc.addPage(
    pw.MultiPage(
      pageFormat: PdfPageFormat.a4,
      margin: const pw.EdgeInsets.all(36),
      header: (context) => context.pageNumber == 1
          ? pw.SizedBox()
          : pw.Text(
              appL10n.statsTitleScope(scope),
              style: const pw.TextStyle(color: _grey, fontSize: 9),
            ),
      footer: (context) => pw.Align(
        alignment: pw.Alignment.centerRight,
        child: pw.Text(
          '${context.pageNumber} / ${context.pagesCount}',
          style: const pw.TextStyle(color: _grey, fontSize: 9),
        ),
      ),
      build: (_) => [
        pw.Text(
          appL10n.statistics,
          style: pw.TextStyle(fontSize: 22, fontWeight: pw.FontWeight.bold),
        ),
        pw.SizedBox(height: 4),
        pw.Text(appL10n.statsTimeBased(scope, period)),
        pw.Text(
          appL10n.createdAt(formatDateTime(date)),
          style: const pw.TextStyle(color: _grey, fontSize: 10),
        ),
        pw.SizedBox(height: 12),
        for (final s in StatsSection.values)
          if (data[s] case final d?) _section(s, d),
      ],
    ),
  );
  return doc.save();
}

pw.Widget _section(StatsSection s, Object data) => pw.Padding(
  padding: const pw.EdgeInsets.only(top: 14),
  child: pw.Column(
    crossAxisAlignment: pw.CrossAxisAlignment.stretch,
    children: [
      pw.Text(
        s.title,
        style: pw.TextStyle(fontSize: 14, fontWeight: pw.FontWeight.bold),
      ),
      pw.Divider(thickness: 0.5, color: _grey),
      ..._content(s, data),
    ],
  ),
);

List<pw.Widget> _content(StatsSection s, Object data) {
  switch (s) {
    case StatsSection.today:
      return [
        for (final l in todayLines(data as golang.TodayStats)) pw.Text(l),
      ];
    case StatsSection.futureDue:
      final f = data as golang.Forecast;
      return [
        _bars(
          [
            for (final d in f.days) [d.toDouble()],
          ],
          [_barColor],
          first: appL10n.today,
          last: '+${f.days.length - 1}',
        ),
        _lines(futureDueLines(f)),
      ];
    case StatsSection.calendar:
      final c = data as golang.Calendar;
      return [_calendar(c), _lines(calendarLines(c))];
    case StatsSection.reviews || StatsSection.reviewTime:
      final h = data as golang.ReviewHistory;
      final time = s == StatsSection.reviewTime;
      return [
        _bars(
          [for (final b in h.buckets) historyValues(b, time: time)],
          _kindColors,
          first: h.buckets.isEmpty ? '' : historyLabel(h, 0),
          last: h.buckets.isEmpty ? '' : historyLabel(h, h.buckets.length - 1),
        ),
        _legend(),
        _lines(historyLines(h, time: time)),
      ];
    case StatsSection.counts:
      final c = data as golang.StateCounts;
      return [
        _lines([
          for (final (label, n) in countRows(c)) (label, count(n)),
          (appL10n.total, count(c.total)),
        ]),
      ];
    case StatsSection.intervals ||
        StatsSection.stability ||
        StatsSection.difficulty ||
        StatsSection.retrievability:
      final h = data as golang.Histogram;
      final kind = switch (s) {
        StatsSection.intervals => HistogramKind.intervals,
        StatsSection.stability => HistogramKind.stability,
        StatsSection.difficulty => HistogramKind.difficulty,
        _ => HistogramKind.retrievability,
      };
      if (h.count == 0) return [pw.Text(appL10n.noCardsInfo)];
      final last = h.buckets.length - 1;
      return [
        _bars(
          [
            for (final b in h.buckets) [b.count.toDouble()],
          ],
          [_barColor],
          first: bucketLabel(h.buckets.first, kind),
          last: bucketLabel(h.buckets[last], kind, last: true),
        ),
        _lines(histogramLines(h, kind)),
      ];
    case StatsSection.hourly:
      final hours = data as List<golang.HourBucket>;
      return [
        _bars(
          [
            for (final h in hours) [h.reviews.toDouble()],
          ],
          [_barColor],
          first: '0:00',
          last: '23:00',
        ),
        _table(
          [appL10n.hour, appL10n.secReviews, appL10n.correct],
          [
            for (final (i, h) in hours.indexed)
              if (h.reviews > 0)
                ['$i:00', count(h.reviews), percent(h.correct)],
          ],
        ),
      ];
    case StatsSection.buttons:
      final b = data as golang.ButtonStats;
      return [
        _table(
          ['', ...buttonNames, appL10n.correct],
          [
            for (final (name, g) in buttonGroups(b))
              [
                name,
                for (final n in g.counts) count(n),
                g.total == 0 ? '-' : percent(g.correct),
              ],
          ],
        ),
      ];
    case StatsSection.trueRetention:
      final t = data as golang.RetentionTable;
      return [
        _table(
          ['', appL10n.kindYoung, appL10n.kindMature, appL10n.total],
          [
            for (final r in t.rows)
              [
                retentionPeriod(r.period),
                retentionCell(r.young),
                retentionCell(r.mature),
                retentionCell(r.total),
              ],
          ],
        ),
        _lines([(appL10n.desiredRetention, percent(t.desired))]),
      ];
  }
}

/// Stacked bars, with the first and last labels below.
pw.Widget _bars(
  List<List<double>> bars,
  List<PdfColor> colors, {
  required String first,
  required String last,
}) {
  final top = bars
      .map((b) => b.fold<double>(0, (a, v) => a + v))
      .fold<double>(0, math.max);
  const height = 70.0;
  return pw.Column(
    crossAxisAlignment: pw.CrossAxisAlignment.stretch,
    children: [
      pw.Row(
        crossAxisAlignment: pw.CrossAxisAlignment.end,
        children: [
          pw.SizedBox(
            width: 32,
            height: height,
            child: pw.Column(
              mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
              crossAxisAlignment: pw.CrossAxisAlignment.end,
              children: [
                pw.Text(compact(top), style: const pw.TextStyle(fontSize: 7)),
                pw.Text('0', style: const pw.TextStyle(fontSize: 7)),
              ],
            ),
          ),
          pw.SizedBox(width: 4),
          pw.Expanded(
            child: pw.SizedBox(
              height: height,
              child: pw.Row(
                crossAxisAlignment: pw.CrossAxisAlignment.end,
                children: [
                  for (final b in bars)
                    pw.Expanded(
                      child: pw.Padding(
                        padding: const pw.EdgeInsets.symmetric(horizontal: 0.4),
                        child: pw.Column(
                          mainAxisAlignment: pw.MainAxisAlignment.end,
                          children: [
                            for (var i = b.length - 1; i >= 0; i--)
                              if (b[i] > 0 && top > 0)
                                pw.Container(
                                  height: height * b[i] / top,
                                  color: colors[i % colors.length],
                                ),
                          ],
                        ),
                      ),
                    ),
                ],
              ),
            ),
          ),
        ],
      ),
      pw.Padding(
        padding: const pw.EdgeInsets.only(left: 36, top: 2, bottom: 4),
        child: pw.Row(
          mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
          children: [
            pw.Text(first, style: const pw.TextStyle(fontSize: 7)),
            pw.Text(last, style: const pw.TextStyle(fontSize: 7)),
          ],
        ),
      ),
    ],
  );
}

String compact(double v) =>
    v == v.roundToDouble() ? count(v.round()) : oneDecimal(v);

pw.Widget _legend() => pw.Padding(
  padding: const pw.EdgeInsets.only(bottom: 4),
  child: pw.Wrap(
    spacing: 10,
    children: [
      for (final (i, k) in reviewKinds.indexed)
        pw.Row(
          mainAxisSize: pw.MainAxisSize.min,
          children: [
            pw.Container(width: 7, height: 7, color: _kindColors[i]),
            pw.SizedBox(width: 3),
            pw.Text(k, style: const pw.TextStyle(fontSize: 8)),
          ],
        ),
    ],
  ),
);

/// The last year as weeks of days, shaded by count.
pw.Widget _calendar(golang.Calendar c) {
  final counts = {for (final d in c.days) d.day: d.count};
  const days = 365;
  final todayRow = parseDay(c.today).weekday - 1;
  final firstRow = (todayRow - (days - 1)) % 7;
  final weeks = ((days + firstRow + (6 - todayRow)) / 7).ceil();
  return pw.Row(
    children: [
      for (var w = 0; w < weeks; w++)
        pw.Column(
          children: [
            for (var r = 0; r < 7; r++)
              () {
                final rel = -(days - 1) - firstRow + w * 7 + r;
                final n = rel < -(days - 1) || rel > 0
                    ? null
                    : counts[rel] ?? 0;
                final l = n == null || n == 0 || c.max == 0
                    ? 0
                    : (4 * n / c.max).ceil().clamp(1, 4);
                return pw.Container(
                  width: 7,
                  height: 7,
                  margin: const pw.EdgeInsets.all(0.6),
                  color: n == null
                      ? PdfColors.white
                      : l == 0
                      ? _empty
                      : PdfColor.fromInt(0xFF2E7D32).shade(0.75 - 0.15 * l),
                );
              }(),
          ],
        ),
    ],
  );
}

pw.Widget _lines(List<(String, String)> lines) => pw.Padding(
  padding: const pw.EdgeInsets.only(top: 4),
  child: pw.Column(
    children: [
      for (final (label, value) in lines)
        pw.Row(
          children: [
            pw.Expanded(child: pw.Text(label)),
            pw.Text(value, style: pw.TextStyle(fontWeight: pw.FontWeight.bold)),
          ],
        ),
    ],
  ),
);

pw.Widget _table(List<String> header, List<List<String>> rows) =>
    pw.TableHelper.fromTextArray(
      headers: header,
      data: rows,
      border: null,
      headerStyle: pw.TextStyle(fontWeight: pw.FontWeight.bold, fontSize: 9),
      cellStyle: const pw.TextStyle(fontSize: 9),
      cellAlignment: pw.Alignment.centerRight,
      headerAlignment: pw.Alignment.centerRight,
      cellAlignments: {0: pw.Alignment.centerLeft},
      headerCellDecoration: const pw.BoxDecoration(
        border: pw.Border(bottom: pw.BorderSide(color: _grey, width: 0.5)),
      ),
    );
