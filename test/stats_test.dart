import 'dart:async';
import 'dart:convert';

import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/features/stats/stats_charts.dart';
import 'package:cardflip/features/stats/stats_pdf.dart';
import 'package:cardflip/features/stats/stats_screen.dart';
import 'package:cardflip/features/stats/stats_sections.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/time_format.dart';

import 'deck_list_test.dart' show openDeckMenu;
import 'support/fake_go.dart';

/// Starts the app and opens collection statistics from the deck list.
Future<FakeGo> openStats(
  WidgetTester tester, {
  FakeGo? go,
  FakePicker? picker,
  Size size = desktop,
  bool settle = true,
}) async {
  go ??= FakeGo();
  await pumpApp(tester, size: size, picker: picker);
  await tester.tap(find.byTooltip('Statistics'));
  if (settle) {
    await tester.pumpAndSettle();
  } else {
    for (var i = 0; i < 5; i++) {
      await tester.pump(const Duration(milliseconds: 100));
    }
  }
  return go;
}

/// The scopes and periods the last calls to every section asked for.
Set<String> lastRequests(FakeGo go) => {
  for (final s in statsSections) jsonEncode(go.callsTo('stats.$s').last),
};

Finder section(StatsSection s) => find.byKey(Key('stats-${s.name}'));

Finder inSection(StatsSection s, Finder f) =>
    find.descendant(of: section(s), matching: f);

/// The value of the stat line [label] in section [s].
String statValue(WidgetTester tester, StatsSection s, String label) {
  final lines = inSection(s, find.byType(StatLines));
  final row = find
      .ancestor(
        of: find.descendant(of: lines, matching: find.text(label)),
        matching: find.byType(Row),
      )
      .first;
  final texts = tester
      .widgetList<Text>(find.descendant(of: row, matching: find.byType(Text)))
      .map((t) => t.data)
      .toList();
  return texts.last!;
}

BarChartData barChart(WidgetTester tester, StatsSection s) =>
    tester.widget<BarChart>(inSection(s, find.byType(BarChart)).first).data;

/// A section's data decoded as the generated API decodes it.
Object decodeSection(StatsSection s) {
  final json = jsonDecode(jsonEncode(statsJson()[statsSections[s.index]]));
  return switch (s) {
    StatsSection.today => golang.TodayStats.fromJson(json),
    StatsSection.streaks => golang.StreakSection.fromJson(json),
    StatsSection.futureDue => golang.Forecast.fromJson(json),
    StatsSection.calendar => golang.Calendar.fromJson(json),
    StatsSection.reviews ||
    StatsSection.reviewTime => golang.ReviewHistory.fromJson(json),
    StatsSection.counts => golang.StateCounts.fromJson(json),
    StatsSection.intervals ||
    StatsSection.stability ||
    StatsSection.difficulty ||
    StatsSection.retrievability => golang.Histogram.fromJson(json),
    StatsSection.hourly => [
      for (final h in json as List) golang.HourBucket.fromJson(h),
    ],
    StatsSection.buttons => golang.ButtonStats.fromJson(json),
    StatsSection.trueRetention => golang.RetentionTable.fromJson(json),
  };
}

/// The text a PDF draws, its strings joined by spaces.
String pdfText(List<int> bytes) =>
    RegExp(r'\[\((.*?)\)\]TJ')
        .allMatches(latin1.decode(bytes))
        .map((m) => m[1]!.replaceAllMapped(RegExp(r'\\(.)'), (e) => e[1]!))
        .join(' ');

void main() {
  test('sections and their channels line up', () {
    expect(StatsSection.values, hasLength(statsSections.length));
    for (final s in StatsSection.values) {
      expect(
        statsSections[s.index].toLowerCase(),
        s.name.toLowerCase(),
        reason: 'fake channel order',
      );
    }
  });

  group('opening statistics', () {
    testWidgets('collection statistics from the deck list', (tester) async {
      // Collection statistics (app-shell spec).
      final go = await openStats(tester, size: phone);
      expect(find.byType(StatsScreen), findsOneWidget);
      expect(find.text('Whole collection'), findsOneWidget);
      expect(lastRequests(go), {
        jsonEncode([
          {'kind': 'collection'},
          'month',
        ]),
      });
    });

    testWidgets('deck statistics from the deck menu', (tester) async {
      // Deck scope (statistics spec): Spanish with its subdecks.
      final go = FakeGo();
      go.tree = [
        deckNode(1, 'Default'),
        deckNode(2, 'Spanish', hasChildren: true),
        deckNode(3, 'Spanish::Verbs', parentId: 2),
      ];
      await pumpApp(tester, size: phone);
      await openDeckMenu(tester, 'Spanish', 'Statistics');
      expect(find.byType(StatsScreen), findsOneWidget);
      expect(
        tester.widget<Text>(find.byKey(const Key('stats-scope-label'))).data,
        'Spanish',
      );
      expect(lastRequests(go), {
        jsonEncode([
          {'kind': 'deck', 'deckId': 2},
          'month',
        ]),
      });
    });

    testWidgets('search scope', (tester) async {
      final go = await openStats(tester, size: phone);
      await tester.tap(find.text('Search'));
      await tester.pumpAndSettle();
      await tester.enterText(
        find.byKey(const Key('stats-search')),
        ' tag:verbs ',
      );
      await tester.testTextInput.receiveAction(TextInputAction.search);
      await tester.pumpAndSettle();
      expect(find.text('Search: tag:verbs'), findsOneWidget);
      expect(lastRequests(go), {
        jsonEncode([
          {'kind': 'search', 'query': 'tag:verbs'},
          'month',
        ]),
      });
    });

    testWidgets('deck scope picks a deck', (tester) async {
      final go = FakeGo();
      go.tree = [deckNode(1, 'Default'), deckNode(2, 'French')];
      await openStats(tester, go: go, size: phone);
      await tester.tap(find.text('Deck'));
      await tester.pumpAndSettle();
      expect(jsonDecode(lastRequests(go).single), [
        {'kind': 'deck', 'deckId': 1},
        'month',
      ]);
      await tester.tap(find.byKey(const Key('stats-deck')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('French').last);
      await tester.pumpAndSettle();
      expect(jsonDecode(lastRequests(go).single), [
        {'kind': 'deck', 'deckId': 2},
        'month',
      ]);
      expect(find.byKey(const Key('stats-scope-label')), findsOneWidget);
      expect(
        tester.widget<Text>(find.byKey(const Key('stats-scope-label'))).data,
        'French',
      );
    });

    testWidgets('a new period reloads only the time-based sections', (
      tester,
    ) async {
      final go = await openStats(tester);
      await tester.tap(find.text('1 year'));
      await tester.pumpAndSettle();
      for (final s in StatsSection.values) {
        final calls = go.callsTo('stats.${statsSections[s.index]}');
        expect(calls, hasLength(s.timeBased ? 2 : 1), reason: s.name);
        if (s.timeBased) expect(calls.last[1], 'year');
      }
    });

    testWidgets('sections appear as they are ready', (tester) async {
      final go = FakeGo();
      final gate = Completer<void>();
      go.statsGates['Reviews'] = gate;
      await openStats(tester, go: go, settle: false);
      expect(inSection(StatsSection.today, find.byType(Text)), findsWidgets);
      expect(
        find.text('Studied 40 cards in 10 minutes today.'),
        findsOneWidget,
      );
      expect(
        inSection(StatsSection.reviews, find.byType(CircularProgressIndicator)),
        findsOneWidget,
      );
      gate.complete();
      await tester.pumpAndSettle();
      expect(
        inSection(StatsSection.reviews, find.byType(CircularProgressIndicator)),
        findsNothing,
      );
      expect(inSection(StatsSection.reviews, find.byType(BarChart)), findsOne);
    });

    testWidgets('a failed section shows why; the others still show', (
      tester,
    ) async {
      final go = FakeGo();
      go.statsErrors['Calendar'] = 'deck: no such deck';
      await openStats(tester, go: go);
      expect(
        inSection(StatsSection.calendar, find.textContaining('no such deck')),
        findsOneWidget,
      );
      expect(
        inSection(StatsSection.futureDue, find.byType(BarChart)),
        findsOne,
      );
    });
  });

  group('sections', () {
    testWidgets('today', (tester) async {
      // After a session (statistics spec).
      await openStats(tester);
      expect(
        inSection(
          StatsSection.today,
          find.text('Studied 40 cards in 10 minutes today.'),
        ),
        findsOneWidget,
      );
      expect(find.text('Again count: 4 (90% correct)'), findsOneWidget);
      expect(
        find.text('Learning: 6, Review: 30, Relearning: 4, Filtered: 0'),
        findsOneWidget,
      );
    });

    testWidgets('today without answers', (tester) async {
      final go = FakeGo();
      go.stats['Today'] = {
        ...go.stats['Today'] as Map<String, Object?>,
        'cards': 0,
        'time': 0,
        'again': 0,
      };
      await openStats(tester, go: go);
      expect(find.text('No cards have been studied today.'), findsOneWidget);
    });

    testWidgets('future due', (tester) async {
      // Forecast (statistics spec): 30 due tomorrow, 12 the day after.
      await openStats(tester);
      const s = StatsSection.futureDue;
      final bars = barChart(tester, s).barGroups;
      expect(bars, hasLength(30));
      expect(bars[1].barRods.single.toY, 30);
      expect(bars[2].barRods.single.toY, 12);
      final line = tester
          .widget<LineChart>(inSection(s, find.byType(LineChart)))
          .data;
      expect(line.lineBarsData.single.spots[2].y, 47);
      expect(statValue(tester, s, 'Due tomorrow'), '30');
      expect(statValue(tester, s, 'Overdue'), '3');
      expect(statValue(tester, s, 'Total'), '47');
      expect(statValue(tester, s, 'Average'), '1.6 cards/day');
    });

    testWidgets('calendar', (tester) async {
      // Calendar (statistics spec): 20 of the last 30 days, darker for
      // more reviews.
      final go = FakeGo();
      go.stats['Calendar'] = {
        'today': '2026-10-13',
        'days': [
          for (var d = -29; d <= 0; d++)
            if (d % 3 != 0) {'day': d, 'count': 1 + (-d % 4)},
        ],
        'max': 4,
      };
      await openStats(tester, go: go);
      final days = find.byWidgetPredicate(
        (w) =>
            w.key is ValueKey<String> &&
            (w.key! as ValueKey<String>).value.startsWith('calendar-day-'),
      );
      expect(days, findsNWidgets(20));
      Color colorOf(int day) =>
          ((tester.widget<Container>(find.byKey(Key('calendar-day-$day'))))
                      .decoration!
                  as BoxDecoration)
              .color!;
      // -1 has 2 reviews, -2 has 3.
      expect(colorOf(-2).a, greaterThan(colorOf(-1).a));
      expect(find.byKey(const Key('calendar-day-0')), findsNothing);
      expect(
        statValue(
          tester,
          StatsSection.calendar,
          'Days studied in the last year',
        ),
        '20 of 365',
      );
      expect(CalendarHeatmap.level(0, 4), 0);
      expect(CalendarHeatmap.level(1, 40), 1);
      expect(CalendarHeatmap.level(40, 40), 4);
    });

    testWidgets('reviews', (tester) async {
      await openStats(tester);
      const s = StatsSection.reviews;
      final bars = barChart(tester, s).barGroups;
      expect(bars, hasLength(30));
      final today = bars.last.barRods.single;
      expect(today.toY, 40);
      expect(today.rodStackItems.map((i) => i.toY - i.fromY), [
        6,
        20,
        10,
        4,
        0,
      ]);
      for (final kind in reviewKinds) {
        expect(inSection(s, find.text(kind)), findsOneWidget);
      }
      expect(statValue(tester, s, 'Days studied'), '1 of 30');
      expect(statValue(tester, s, 'Total'), '40');
    });

    testWidgets('review time', (tester) async {
      await openStats(tester);
      const s = StatsSection.reviewTime;
      expect(barChart(tester, s).barGroups.last.barRods.single.toY, 10);
      expect(statValue(tester, s, 'Total'), '10 min');
      expect(statValue(tester, s, 'Average for days studied'), '10 min');
    });

    testWidgets('card counts', (tester) async {
      await openStats(tester);
      const s = StatsSection.counts;
      expect(statValue(tester, s, 'Total'), '211');
      final row = find
          .ancestor(
            of: inSection(s, find.text('New')),
            matching: find.byType(Row),
          )
          .at(1);
      expect(
        find.descendant(of: row, matching: find.text('100 (47%)')),
        findsOneWidget,
      );
    });

    testWidgets('intervals, stability and difficulty', (tester) async {
      await openStats(tester);
      expect(
        statValue(tester, StatsSection.intervals, 'Average interval'),
        '9.0 days',
      );
      expect(barChart(tester, StatsSection.intervals).barGroups, hasLength(3));
      expect(
        statValue(tester, StatsSection.stability, 'Average stability'),
        '14.0 days',
      );
      expect(
        statValue(tester, StatsSection.difficulty, 'Average difficulty'),
        '22%',
      );
      expect(
        barChart(tester, StatsSection.difficulty).barGroups,
        hasLength(20),
      );
    });

    testWidgets('retrievability', (tester) async {
      // Retrievability (statistics spec): 100 cards averaging 0.92.
      final go = FakeGo();
      go.stats['Retrievability'] = histogramJson(
        [...List.filled(18, 0), 80, 20],
        width: 0.05,
        average: 0.92,
        sum: 92,
      );
      await openStats(tester, go: go);
      const s = StatsSection.retrievability;
      expect(statValue(tester, s, 'Average retrievability'), '92%');
      expect(statValue(tester, s, 'Estimated total knowledge'), '92 cards');
      expect(statValue(tester, s, 'Cards'), '100');
    });

    testWidgets('a histogram without cards', (tester) async {
      final go = FakeGo();
      go.stats['Stability'] = histogramJson([]);
      await openStats(tester, go: go);
      expect(
        inSection(
          StatsSection.stability,
          find.text('No cards with this information yet.'),
        ),
        findsOneWidget,
      );
    });

    testWidgets('hourly breakdown', (tester) async {
      await openStats(tester);
      const s = StatsSection.hourly;
      final bars = barChart(tester, s).barGroups;
      expect(bars, hasLength(24));
      expect(bars[9].barRods.single.toY, 40);
      final line = tester
          .widget<LineChart>(inSection(s, find.byType(LineChart)))
          .data;
      expect(line.lineBarsData.single.spots[9].y, 90);
      expect(inSection(s, find.text('Success rate')), findsOneWidget);
    });

    testWidgets('answer buttons', (tester) async {
      await openStats(tester);
      const s = StatsSection.buttons;
      expect(barChart(tester, s).barGroups, hasLength(12));
      expect(statValue(tester, s, 'Learning'), '67% correct (6)');
      expect(statValue(tester, s, 'Young'), '95% correct (20)');
      expect(statValue(tester, s, 'Mature'), '90% correct (10)');
    });

    testWidgets('true retention', (tester) async {
      // Compare with target (statistics spec).
      await openStats(tester);
      const s = StatsSection.trueRetention;
      final row = tester
          .widget<Table>(inSection(s, find.byType(Table)))
          .children
          .firstWhere(
            (r) =>
                (r.children.first as Padding).child is Text &&
                ((r.children.first as Padding).child as Text).data ==
                    'Last month',
          );
      String cell(int i) => ((row.children[i] as Padding).child as Text).data!;
      expect(cell(2), '90.0% (180/200)');
      expect(cell(1), '90.9% (100/110)');
      expect(inSection(s, find.text('-')), findsNWidgets(3)); // yesterday
      expect(statValue(tester, s, 'Desired retention'), '90%');
    });

    testWidgets('streaks', (tester) async {
      // Collection streaks (statistics spec): a 9-day streak, the longest
      // 41 days with its dates, 24 of 30 study days. No goal (spec): no
      // goal days.
      await openStats(tester);
      const s = StatsSection.streaks;
      expect(statValue(tester, s, 'Current streak'), '9 days');
      expect(
        statValue(tester, s, 'Longest streak'),
        '41 days, from Mar 2 to Apr 11',
      );
      expect(statValue(tester, s, 'Study days'), '24 of 30');
      expect(inSection(s, find.text('Days the goal was met')), findsNothing);
    });

    testWidgets('streaks with a goal', (tester) async {
      // Goal days (statistics spec): 18 days met the goal.
      final go = FakeGo();
      go.stats = {...go.stats, 'Streaks': streakSectionJson(goalDays: 18)};
      await openStats(tester, go: go);
      expect(
        statValue(tester, StatsSection.streaks, 'Days the goal was met'),
        '18',
      );
    });

    testWidgets('the deck list header opens them at Streaks', (tester) async {
      // Open statistics (progress spec): the collection statistics, scrolled
      // to the Streaks section.
      final go = FakeGo()..progress = progressJson(current: 12, cards: 3);
      await pumpApp(tester, size: phone);
      await tester.tap(find.byKey(const Key('progress-header')));
      await tester.pumpAndSettle();
      expect(find.byType(StatsScreen), findsOneWidget);
      expect(lastRequests(go), {
        jsonEncode([
          {'kind': 'collection'},
          'month',
        ]),
      });
      final top = tester.getTopLeft(section(StatsSection.streaks)).dy;
      final appBar = tester.getBottomLeft(find.byType(AppBar)).dy;
      expect(top, lessThan(appBar + 100));
      expect(top, greaterThanOrEqualTo(appBar));
    });
  });

  group('save as PDF', () {
    test('the PDF has every section', () async {
      final data = {for (final s in StatsSection.values) s: decodeSection(s)};
      final bytes = await buildStatsPdf(
        scope: 'Spanish',
        period: '1 month',
        date: DateTime(2026, 10, 13, 12),
        data: data,
        compress: false,
      );
      final text = pdfText(bytes);
      expect(latin1.decode(bytes.sublist(0, 4)), '%PDF');
      for (final s in StatsSection.values) {
        expect(text, contains(s.title), reason: s.title);
      }
      expect(text, contains('Spanish - time-based sections: 1 month'));
      expect(text, contains('Studied 40 cards in 10 minutes today.'));
      expect(text, contains('Current streak 9 days'));
      expect(text, contains('41 days, from Mar 2 to Apr 11'));
      expect(text, contains('Study days 24 of 30'));
      expect(text, contains('Last month 90.9% (100/110) 90.0% (180/200)'));

      // A section that failed to load is left out.
      data.remove(StatsSection.calendar);
      final without = pdfText(
        await buildStatsPdf(
          scope: 'Spanish',
          period: '1 month',
          date: DateTime(2026, 10, 13, 12),
          data: data,
          compress: false,
        ),
      );
      expect(without, isNot(contains('Calendar')));
      expect(without, contains('Future due'));
    });

    test('file name', () {
      final date = DateTime(2026, 10, 13);
      expect(
        statsPdfName('Languages::Spanish', date),
        'Statistics - Languages - Spanish - 2026-10-13.pdf',
      );
      expect(
        statsPdfName('Search: deck:"a/b"', date),
        'Statistics - Search_ deck__a_b_ - 2026-10-13.pdf',
      );
    });

    testWidgets('saves the screen as a PDF', (tester) async {
      // PDF (statistics spec).
      final picker = FakePicker();
      await openStats(tester, picker: picker);
      await tester.tap(find.byTooltip('Save as PDF'));
      // Building the PDF takes real time.
      for (var i = 0; i < 200 && picker.saved.isEmpty; i++) {
        await tester.runAsync(
          () => Future<void>.delayed(const Duration(milliseconds: 10)),
        );
        await tester.pump();
      }
      await tester.pumpAndSettle();
      final name =
          'Statistics - Whole collection - ${formatDate(DateTime.now())}.pdf';
      expect(picker.saved.keys, [name]);
      expect(latin1.decode(picker.saved[name]!.sublist(0, 4)), '%PDF');
      expect(find.text('Saved $name'), findsOneWidget);
    });
  });
}
