import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/features/browse/card_info_screen.dart';
import 'package:cardflip/widgets/time_format.dart';

import 'support/fake_go.dart';

Future<void> pumpCardInfo(
  WidgetTester tester,
  Map<String, Object?> info, {
  Size size = phone,
}) async {
  final go = FakeGo();
  go.handle('browse.CardInfo', (args) {
    expect(args, [info['cardId']]);
    return info;
  });
  setSize(tester, size);
  await tester.pumpWidget(
    MaterialApp(home: CardInfoScreen(cardId: info['cardId'] as int)),
  );
  await tester.pumpAndSettle();
}

/// The text of the summary row [label].
String summary(WidgetTester tester, String label) {
  // Labels are Texts and values SelectableTexts, both in row order.
  final table = find.byKey(const Key('card-info-summary'));
  final texts = tester
      .widgetList<SelectableText>(
        find.descendant(of: table, matching: find.byType(SelectableText)),
      )
      .map((t) => t.data!)
      .toList();
  final labels = tester
      .widgetList<Text>(find.descendant(of: table, matching: find.byType(Text)))
      .map((t) => t.data)
      .whereType<String>()
      .toList();
  return texts[labels.indexOf(label)];
}

/// A Review card reviewed 5 times with 1 lapse, as Go reports it.
Map<String, Object?> reviewCardInfo() {
  final last = DateTime.utc(2026, 10, 10, 9);
  return cardInfoJson(
    7,
    state: 'review',
    due: DateTime.utc(2026, 10, 30, 9),
    interval: const Duration(days: 20),
    reviews: 5,
    lapses: 1,
    history: [
      historyEntry(
        last,
        3,
        kind: 'relearning',
        interval: const Duration(days: 20),
      ),
      historyEntry(
        last.subtract(const Duration(minutes: 10)),
        1,
        interval: const Duration(minutes: 10),
      ),
      historyEntry(
        DateTime.utc(2026, 9, 20, 9),
        3,
        interval: const Duration(days: 20),
      ),
      historyEntry(
        DateTime.utc(2026, 9, 2, 9, 10),
        3,
        kind: 'learning',
        interval: const Duration(days: 18),
      ),
      historyEntry(
        DateTime.utc(2026, 9, 2, 9),
        3,
        kind: 'learning',
        interval: const Duration(minutes: 10),
      ),
    ],
  );
}

void main() {
  testWidgets('a review card shows its counts, memory state and due date '
      '(spec)', (tester) async {
    final info = reviewCardInfo();
    await pumpCardInfo(tester, info);
    expect(summary(tester, 'Reviews'), '5');
    expect(summary(tester, 'Lapses'), '1');
    expect(summary(tester, 'Stability'), '13 days');
    expect(summary(tester, 'Difficulty'), '5.2');
    expect(summary(tester, 'Retrievability'), '93%');
    expect(
      summary(tester, 'Due'),
      formatDateTime(DateTime.utc(2026, 10, 30, 9)),
    );
    expect(summary(tester, 'Interval'), '20 days');
    expect(summary(tester, 'Average time'), '4.0 s');
    expect(summary(tester, 'Total time'), '20.0 s');
    expect(
      summary(tester, 'First review'),
      formatDate(DateTime.utc(2026, 9, 2, 9)),
    );
    expect(summary(tester, 'Card ID'), '7');
    expect(summary(tester, 'Note ID'), '107');

    // The history, newest first, below the summary.
    await tester.scrollUntilVisible(
      find.byKey(const Key('history-4')),
      200,
      scrollable: find.byType(Scrollable).first,
    );
    final newest = find.byKey(const Key('history-0'));
    expect(
      find.descendant(
        of: newest,
        matching: find.text('Good · Relearning', findRichText: true),
      ),
      findsOneWidget,
    );
    expect(
      find.descendant(of: newest, matching: find.text('20 days')),
      findsOneWidget,
    );
    final lapse = find.byKey(const Key('history-1'));
    expect(
      find.descendant(
        of: lapse,
        matching: find.text('Again · Review', findRichText: true),
      ),
      findsOneWidget,
    );
    expect(
      find.descendant(of: lapse, matching: find.text('10 minutes')),
      findsOneWidget,
    );
    expect(
      find.descendant(
        of: find.byKey(const Key('history-4')),
        matching: find.text(formatDateTime(DateTime.utc(2026, 9, 2, 9))),
      ),
      findsOneWidget,
    );
  });

  testWidgets('a new card shows its position and no memory state (spec)', (
    tester,
  ) async {
    await pumpCardInfo(tester, cardInfoJson(3, position: 4));
    expect(summary(tester, 'Due'), 'New #4');
    for (final label in [
      'Stability',
      'Difficulty',
      'Retrievability',
      'Interval',
      'First review',
      'Average time',
    ]) {
      expect(summary(tester, label), '—', reason: label);
    }
    expect(summary(tester, 'Reviews'), '0');
    expect(find.text('No reviews yet.'), findsOneWidget);
  });

  testWidgets('wide layouts show the history as a table beside the summary', (
    tester,
  ) async {
    await pumpCardInfo(tester, reviewCardInfo(), size: desktop);
    for (final header in ['Date', 'Kind', 'Rating', 'Time']) {
      expect(find.text(header), findsOneWidget);
    }
    // Interval is also a summary row.
    expect(find.text('Interval'), findsNWidgets(2));
    expect(find.text('Relearning'), findsOneWidget);
    expect(find.text('Again'), findsOneWidget);
    expect(find.text('Good'), findsNWidgets(4));
    final summaryBox = tester.getRect(
      find.byKey(const Key('card-info-summary')),
    );
    final historyBox = tester.getRect(find.text('Review history'));
    expect(historyBox.left, greaterThan(summaryBox.right));
  });

  testWidgets('a card that no longer exists says so', (tester) async {
    final go = FakeGo();
    go.handle('browse.CardInfo', (_) => throw 'That card no longer exists.');
    setSize(tester, phone);
    await tester.pumpWidget(const MaterialApp(home: CardInfoScreen(cardId: 1)));
    await tester.pumpAndSettle();
    expect(find.textContaining('no longer exists'), findsOneWidget);
  });

  test('intervals read naturally', () {
    expect(formatInterval(const Duration(seconds: 20)), '1 minute');
    expect(formatInterval(const Duration(minutes: 10)), '10 minutes');
    expect(formatInterval(const Duration(hours: 5)), '5 hours');
    expect(formatInterval(const Duration(days: 1)), '1 day');
    expect(formatInterval(const Duration(days: 75)), '2.5 months');
    expect(formatInterval(const Duration(days: 730)), '2.0 years');
  });
}
