// The statistics check shared by Linux and the browser. It avoids dart:io
// so that it also runs in the browser.
import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/app/bootstrap.dart';
import 'package:cardflip/features/stats/stats_screen.dart';
import 'package:cardflip/features/stats/stats_sections.dart';
import 'package:cardflip/golang/golang.dart' as golang;

/// The widget each section shows once it has loaded.
const _views = {
  StatsSection.today: TodayView,
  StatsSection.futureDue: FutureDueView,
  StatsSection.calendar: CalendarView,
  StatsSection.reviews: ReviewHistoryView,
  StatsSection.reviewTime: ReviewHistoryView,
  StatsSection.counts: CardCountsView,
  StatsSection.intervals: HistogramView,
  StatsSection.stability: HistogramView,
  StatsSection.difficulty: HistogramView,
  StatsSection.retrievability: HistogramView,
  StatsSection.hourly: HourlyView,
  StatsSection.buttons: AnswerButtonsView,
  StatsSection.trueRetention: TrueRetentionView,
};

/// Imports 50,000 Basic notes, answers [answers] of them, opens collection
/// statistics from the deck list and checks that every section appears.
Future<void> runStatsFlow(
  WidgetTester tester, {
  required String dataDir,
  required bool isWeb,
  int answers = 60,
}) async {
  // Wide enough for the two-column layout, which builds every section.
  tester.view.physicalSize =
      const Size(1400, 900) * tester.view.devicePixelRatio;
  addTearDown(tester.view.resetPhysicalSize);
  await tester.pumpWidget(
    CardflipBootstrap(dataDir: () async => dataDir, isWeb: isWeb),
  );
  await tester.pumpAndSettle();
  // A browser that refuses persistent storage gets a one-time notice.
  if (!(await golang.app.status()).persistent) {
    await tester.tap(find.text('OK'));
    await tester.pumpAndSettle();
  }

  final csv = StringBuffer();
  for (var i = 0; i < 50000; i++) {
    csv.writeln('Question $i about topic ${i % 97}\tAnswer $i');
  }
  final stage = await golang.transfer.stageBytes(
    'synthetic.tsv',
    Uint8List.fromList(utf8.encode(csv.toString())),
  );
  final preview = (await golang.transfer.inspect(stage.id)).csv!;
  final events = await golang.transfer
      .import_(
        stage.id,
        golang.ImportOptions(
          typeID: preview.typeID,
          deckID: preview.deckID,
          deckName: '',
          mapping: preview.mapping,
          hasHeader: false,
          duplicates: 'duplicate',
          duplicateScope: 'noteType',
          separator: '',
          allowHTML: preview.allowHTML,
          tags: preview.tags,
          keepScheduling: false,
          updateNotes: 'ifNewer',
          updateNoteTypes: 'ifNewer',
          mergeNoteTypes: false,
        ),
      )
      .toList();
  expect(events.last.summary!.notesAdded, 50000);

  // Some review history: new cards answered with every button.
  // The session ends early once only learning cards that are not due yet
  // are left.
  var state = await golang.study.start(preview.deckID);
  var answered = 0;
  for (; answered < answers; answered++) {
    final card = state.card;
    if (card == null) break;
    state = await golang.study.answer(card.cardID, 1 + answered % 4, 2000);
  }
  expect(answered, greaterThan(0));
  await tester.pumpAndSettle();

  // ignore: avoid_print
  print('imported and answered $answered cards');
  final opened = Stopwatch()..start();
  await tester.tap(find.byTooltip('Statistics'));
  await tester.pump();
  final screen = find.byType(StatsScreen);
  bool loading() => find
      .descendant(of: screen, matching: find.byType(CircularProgressIndicator))
      .evaluate()
      .isNotEmpty;
  while (screen.evaluate().isEmpty || loading()) {
    await tester.pump(const Duration(milliseconds: 16));
    if (opened.elapsed > const Duration(seconds: 30)) {
      final waiting = [
        for (final s in StatsSection.values)
          if (find
              .descendant(
                of: find.byKey(Key('stats-${s.name}')),
                matching: find.byType(CircularProgressIndicator),
              )
              .evaluate()
              .isNotEmpty)
            s.title,
      ];
      fail(
        'statistics did not finish loading: ${screen.evaluate().isEmpty ? 'screen not open' : waiting.join(', ')}',
      );
    }
  }
  opened.stop();
  // ignore: avoid_print
  print('all statistics sections after ${opened.elapsedMilliseconds} ms');
  await tester.pumpAndSettle();

  for (final s in StatsSection.values) {
    final card = find.byKey(Key('stats-${s.name}'));
    expect(card, findsOneWidget, reason: s.title);
    expect(
      find.descendant(of: card, matching: find.byType(_views[s]!)),
      findsOneWidget,
      reason: '${s.title} shows its content',
    );
  }
  expect(find.textContaining('Studied $answered cards in'), findsOneWidget);
  final counts = find.byKey(Key('stats-${StatsSection.counts.name}'));
  expect(
    find.descendant(of: counts, matching: find.text('50000')),
    findsOneWidget,
  );
}
