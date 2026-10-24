import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/features/decks/deck_options_screen.dart';
import 'package:cardflip/features/decks/simulator_screen.dart';
import 'package:cardflip/features/stats/stats_charts.dart';

import 'deck_list_test.dart' show openDeckMenu;
import 'deck_options_fsrs_test.dart' show tapVisible;
import 'support/fake_go.dart';

/// Opens the simulator from the Default deck's options, in a view tall
/// enough to show the results without scrolling.
Future<void> openSimulator(WidgetTester tester) async {
  await pumpApp(tester, size: const Size(600, 2400));
  await openDeckMenu(tester, 'Default', 'Options');
  await tapVisible(tester, find.byKey(const Key('simulateWorkload')));
  expect(find.byType(SimulatorScreen), findsOneWidget);
}

void main() {
  testWidgets('Simulate compares two retentions and changes nothing', (
    tester,
  ) async {
    final go = FakeGo();
    await openSimulator(tester);
    // The inputs start from the deck options.
    expect(find.text('Desired retention: 90%'), findsOneWidget);
    expect(
      tester
          .widget<TextField>(find.byKey(const Key('simNewPerDay')))
          .controller!
          .text,
      '20',
    );
    await tester.tap(find.byKey(const Key('simCompare')));
    await tester.pumpAndSettle();
    expect(find.text('Compared retention: 95%'), findsOneWidget);
    await tapVisible(tester, find.byKey(const Key('simulate')));

    final call = go.callsTo('decks.Simulate').single;
    expect(call[0], 1);
    expect((call[1] as Map)['desiredRetention'], 0.9);
    expect(call[2], {
      'days': 365,
      'newPerDay': 20,
      'reviewsPerDay': 200,
      'retention': 0.9,
      'compareRetention': 0.95,
    });
    // Both curves, with a legend, and the totals of each.
    expect(find.byKey(const Key('simReviewsChart')), findsOneWidget);
    expect(find.byKey(const Key('simMemorizedChart')), findsOneWidget);
    final chart = tester.widget<MultiLineChart>(
      find.byKey(const Key('simReviewsChart')),
    );
    expect(chart.lines, hasLength(2));
    expect(chart.lines[1].first, greaterThan(chart.lines[0].first));
    expect(chart.lineLabels, ['90%', '95%']);
    expect(find.text('At 90%'), findsOneWidget);
    expect(find.text('At 95%'), findsOneWidget);
    expect(find.textContaining('10,950 reviews (30.0 a day)'), findsOneWidget);
    expect(find.text('Memorised at the end: 365 cards'), findsNWidgets(2));

    // Leaving changes nothing: the options keep their retention.
    await tester.pageBack();
    await tester.pumpAndSettle();
    expect(find.byType(DeckOptionsScreen), findsOneWidget);
    expect(go.callsTo('decks.SaveSettings'), isEmpty);
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    expect(
      (go.callsTo('decks.SaveSettings').single[2] as Map)['desiredRetention'],
      0.9,
    );
  });

  testWidgets('Help me decide suggests a retention, used after confirming', (
    tester,
  ) async {
    final go = FakeGo();
    await openSimulator(tester);
    await tapVisible(tester, find.byKey(const Key('helpMeDecide')));
    expect(go.callsTo('decks.SuggestRetention'), hasLength(1));
    expect(find.text('Suggested desired retention: 85%'), findsOneWidget);
    final curve = tester.widget<MultiLineChart>(
      find.byKey(const Key('simCostChart')),
    );
    expect(curve.lines.single, hasLength(30));

    // Cancelling the confirmation keeps the simulator open.
    await tapVisible(tester, find.byKey(const Key('simApply')));
    expect(find.text('Use 85% desired retention?'), findsOneWidget);
    await tester.tap(find.text('Cancel'));
    await tester.pumpAndSettle();
    expect(find.byType(SimulatorScreen), findsOneWidget);

    await tapVisible(tester, find.byKey(const Key('simApply')));
    await tester.tap(find.widgetWithText(FilledButton, 'Use'));
    await tester.pumpAndSettle();
    // The options show it, unsaved, until Save.
    expect(find.byType(SimulatorScreen), findsNothing);
    expect(find.text('Desired retention: 85%'), findsOneWidget);
    expect(go.callsTo('decks.SaveSettings'), isEmpty);
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    expect(
      (go.callsTo('decks.SaveSettings').single[2] as Map)['desiredRetention'],
      0.85,
    );
  });

  testWidgets('progress, cancel and errors', (tester) async {
    final go = FakeGo();
    var cancelled = false;
    final events = StreamController<Object?>(onCancel: () => cancelled = true);
    go.simulationEvents = (_, _) => events.stream;
    await openSimulator(tester);
    await tester.tap(find.byKey(const Key('helpMeDecide')));
    await tester.pump();
    events.add(simulationProgress('simulating', 15, 30));
    await tester.pump();
    expect(find.text('Simulating… 50%'), findsOneWidget);
    expect(find.byKey(const Key('simulate')), findsNothing);
    await tester.tap(find.byKey(const Key('simCancel')));
    await tester.pumpAndSettle();
    expect(cancelled, isTrue);
    expect(find.byKey(const Key('simulate')), findsOneWidget);

    // An error from the simulation is shown.
    go.simulationEvents = (_, _) => Stream.value({
      ...simulationProgress('', 0, 0),
      'error': 'The deck has no cards to simulate.',
    });
    await tapVisible(tester, find.byKey(const Key('helpMeDecide')));
    expect(find.text('The deck has no cards to simulate.'), findsOneWidget);

    // Invalid inputs are not sent.
    await tester.enterText(find.byKey(const Key('simDays')), '0');
    await tapVisible(tester, find.byKey(const Key('simulate')));
    expect(
      find.text('Enter between 1 and 3650 days to simulate.'),
      findsOneWidget,
    );
    expect(go.callsTo('decks.Simulate'), isEmpty);
  });
}
