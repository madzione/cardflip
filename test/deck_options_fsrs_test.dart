import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/features/decks/deck_options_screen.dart';

import 'deck_list_test.dart' show openDeckMenu;
import 'support/fake_go.dart';

/// Fitted parameters, as Optimize finds them.
const fitted = [
  0.5428, 3.8365, 4.1718, 4.1718, 7.2786, 0.5261, 1.3833, 0.001, 1.4182, //
  0.1563, 0.9142, 1.903, 0.1412, 0.2488, 2.2624, 0.227, 2.87, 0.4423, 0.5417,
];

Future<void> tapVisible(WidgetTester tester, Finder finder) async {
  await tester.ensureVisible(finder);
  await tester.pumpAndSettle();
  await tester.tap(finder);
  await tester.pumpAndSettle();
}

/// Taps [finder] and lets the optimization start; its progress bar moves
/// until the first progress event, so the test cannot wait to settle.
Future<void> startOptimize(WidgetTester tester, Finder finder) async {
  await tester.ensureVisible(finder);
  await tester.pumpAndSettle();
  await tester.tap(finder);
  await tester.pump();
}

String fieldText(WidgetTester tester, String key) =>
    tester.widget<TextField>(find.byKey(Key(key))).controller!.text;

void main() {
  testWidgets('Optimize shows progress and fills in better parameters', (
    tester,
  ) async {
    final go = FakeGo();
    final events = StreamController<Object?>();
    go.optimizeEvents = (_, _) => events.stream;
    await pumpApp(tester);
    await openDeckMenu(tester, 'Default', 'Options');
    expect(fieldText(tester, 'fsrsWeights'), '');
    await startOptimize(tester, find.byKey(const Key('optimize')));
    expect(find.text('Starting…'), findsOneWidget);

    final call = go.callsTo('decks.Optimize').single;
    expect(call[0], 1);
    expect((call[1] as Map)['historicalRetention'], 0.9);
    events.add(optimizeProgress('training', 30, 120));
    await tester.pumpAndSettle();
    expect(find.text('Optimizing… 25%'), findsOneWidget);
    expect(find.byKey(const Key('optimize')), findsNothing);

    events.add(optimizeDone(optimizeResultJson(weights: fitted)));
    await events.close();
    await tester.pumpAndSettle();
    expect(fieldText(tester, 'fsrsWeights'), formatWeights(fitted));
    expect(
      find.text('New parameters fitted to 5,000 reviews. Save to use them.'),
      findsOneWidget,
    );
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    final saved = go.callsTo('decks.SaveSettings').single[2] as Map;
    expect(saved['weights'], fitted);
    expect(find.byType(DeckOptionsScreen), findsNothing);
  });

  testWidgets('with too few reviews Optimize explains and changes nothing', (
    tester,
  ) async {
    final go = FakeGo();
    go.optimizeEvents = (_, _) => Stream.value(
      optimizeDone(optimizeResultJson(notEnough: true, reviews: 150)),
    );
    await pumpApp(tester);
    await openDeckMenu(tester, 'Default', 'Options');
    await tapVisible(tester, find.byKey(const Key('optimize')));
    expect(
      find.text(
        'Optimizing needs at least 400 reviews, but there are only 150 '
        'reviews to learn from. The parameters are unchanged.',
      ),
      findsOneWidget,
    );
    expect(fieldText(tester, 'fsrsWeights'), '');
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    expect((go.callsTo('decks.SaveSettings').single[2] as Map)['weights'], []);
  });

  testWidgets('a cancelled Optimize stops and changes nothing', (tester) async {
    final go = FakeGo();
    var cancelled = false;
    final events = StreamController<Object?>(onCancel: () => cancelled = true);
    go.optimizeEvents = (_, _) => events.stream;
    await pumpApp(tester);
    await openDeckMenu(tester, 'Default', 'Options');
    await startOptimize(tester, find.byKey(const Key('optimize')));
    events.add(optimizeProgress('reading', 0, 1));
    await tester.pumpAndSettle();
    expect(find.text('Reading the review history…'), findsOneWidget);
    // Saving waits until optimizing ends.
    expect(
      tester
          .widget<FilledButton>(find.widgetWithText(FilledButton, 'Save'))
          .onPressed,
      isNull,
    );

    await tapVisible(tester, find.byKey(const Key('cancelOptimize')));
    expect(cancelled, isTrue);
    expect(
      find.text('Optimizing was cancelled. Nothing changed.'),
      findsOneWidget,
    );
    expect(find.byKey(const Key('optimize')), findsOneWidget);
    expect(fieldText(tester, 'fsrsWeights'), '');
    // A late event of the cancelled run is ignored.
    events.add(optimizeDone(optimizeResultJson(weights: fitted)));
    await tester.pumpAndSettle();
    expect(fieldText(tester, 'fsrsWeights'), '');
    // Its listener is gone, so closing never completes.
    unawaited(events.close());
  });

  testWidgets('saving a change that reschedules asks first', (tester) async {
    final go = FakeGo();
    go.rescheduleCount = 3;
    await pumpApp(tester);
    await openDeckMenu(tester, 'Default', 'Options');
    await tapVisible(tester, find.byKey(const Key('rescheduleOnChange')));
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    expect(find.text('Reschedule cards?'), findsOneWidget);
    expect(find.textContaining('3 review cards will get new due'), findsOne);
    await tester.tap(find.text('Cancel'));
    await tester.pumpAndSettle();
    expect(go.callsTo('decks.SaveSettings'), isEmpty);
    expect(find.byType(DeckOptionsScreen), findsOneWidget);

    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Reschedule'));
    await tester.pumpAndSettle();
    final saved = go.callsTo('decks.SaveSettings').single[2] as Map;
    expect(saved['rescheduleOnChange'], isTrue);
    expect(find.byType(DeckOptionsScreen), findsNothing);
  });

  testWidgets('Optimize all asks first and reports the presets', (
    tester,
  ) async {
    final go = FakeGo();
    go.rescheduleCountAll = 12;
    await pumpApp(tester);
    await openDeckMenu(tester, 'Default', 'Options');
    await tapVisible(tester, find.byKey(const Key('optimizeAll')));
    expect(find.textContaining('Up to 12 review cards'), findsOneWidget);
    await tester.tap(find.text('Optimize').last);
    await tester.pumpAndSettle();
    expect(go.callsTo('decks.OptimizeAll'), hasLength(1));
    expect(
      find.text(
        'No preset got new parameters. 1 preset has too few reviews to '
        'optimize.',
      ),
      findsOneWidget,
    );
  });

  testWidgets('invalid parameters are explained and not saved', (tester) async {
    final go = FakeGo();
    await pumpApp(tester);
    await openDeckMenu(tester, 'Default', 'Options');
    final field = find.byKey(const Key('fsrsWeights'));
    await tester.ensureVisible(field);
    await tester.enterText(field, '0.4, 1.2, x');
    await tester.pump();
    expect(find.text('"x" is not a number.'), findsOneWidget);
    await tester.enterText(field, '0.4, 1.2, 3.1');
    await tester.pump();
    expect(
      find.text('Enter 19 numbers, or leave empty for the default parameters.'),
      findsOneWidget,
    );
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    expect(go.callsTo('decks.SaveSettings'), isEmpty);
    expect(find.text('Correct the FSRS parameters first.'), findsOneWidget);
  });

  test('parameters are parsed and formatted', () {
    expect(parseWeights(''), isEmpty);
    expect(parseWeights(formatWeights(fitted)), fitted);
    expect(parseWeights(fitted.join(' ')), fitted);
    expect(formatDate(DateTime(2026, 1, 5)), '2026-01-05');
  });

  testWidgets('Easy Days are chosen per weekday and saved', (tester) async {
    final go = FakeGo();
    // Seven three-way rows fit a phone.
    tester.view.physicalSize = const Size(360, 800);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.reset);
    await pumpApp(tester);
    await openDeckMenu(tester, 'Default', 'Options');
    expect(find.text('Easy Days'), findsOneWidget);
    for (var i = 0; i < 7; i++) {
      final selector = tester.widget<SegmentedButton<int>>(
        find.byKey(Key('easyDay$i')),
      );
      expect(selector.selected, {0});
    }
    Finder level(int day, String label) => find.descendant(
      of: find.byKey(Key('easyDay$day')),
      matching: find.text(label),
    );
    await tapVisible(tester, level(5, 'Reduced'));
    await tapVisible(tester, level(6, 'Minimum'));
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    final saved = go.callsTo('decks.SaveSettings').single[2] as Map;
    expect(saved['easyDays'], [0, 0, 0, 0, 0, 1, 2]);
  });
}
