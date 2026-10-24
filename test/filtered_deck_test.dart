import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/features/decks/deck_options_screen.dart';
import 'package:cardflip/features/decks/filtered_deck_screen.dart';
import 'package:cardflip/features/study/study_screen.dart';

import 'browse_test.dart' show openBrowse;
import 'deck_list_test.dart' show openDeckMenu;
import 'deck_overview_test.dart' show handleStudy;
import 'support/fake_go.dart';

FakeGo withFiltered() {
  final go = FakeGo();
  go.tree = [
    deckNode(1, 'Default'),
    deckNode(2, 'Spanish', newCount: 3)..['cards'] = 30,
    deckNode(5, 'Cram', newCount: 12, cards: 12, filtered: true),
  ];
  go.filteredOptions[5] = filteredOptionsJson('Cram', search: 'tag:exam');
  return go;
}

void main() {
  testWidgets('filtered decks are marked, with Rebuild and Empty (spec)', (
    tester,
  ) async {
    final go = withFiltered();
    go.filterMatches = (_) => 8;
    await pumpApp(tester);
    expect(find.byKey(const Key('filtered-5')), findsOneWidget);
    expect(find.byKey(const Key('filtered-2')), findsNothing);

    final row = find
        .ancestor(of: find.text('Cram'), matching: find.byType(InkWell))
        .first;
    await tester.tap(
      find.descendant(of: row, matching: find.byTooltip('Deck actions')),
    );
    await tester.pumpAndSettle();
    expect(find.text('Rebuild'), findsOneWidget);
    expect(find.text('Empty'), findsOneWidget);
    expect(find.text('Add card'), findsNothing);
    expect(find.text('Custom study'), findsNothing);
    await tester.tap(find.text('Rebuild'));
    await tester.pumpAndSettle();
    expect(go.callsTo('decks.Rebuild'), [
      [5],
    ]);
    expect(find.text('Rebuilt “Cram”: 8 cards'), findsOneWidget);
    final counts = find.byKey(const Key('counts-5'));
    expect(
      find.descendant(of: counts, matching: find.text('8')),
      findsOneWidget,
    );

    await openDeckMenu(tester, 'Cram', 'Empty');
    expect(go.callsTo('decks.Empty'), [
      [5],
    ]);
    expect(find.descendant(of: counts, matching: find.text('8')), findsNothing);
  });

  testWidgets('create a filtered deck from the deck list', (tester) async {
    final go = withFiltered();
    go.settings = {
      ...go.settings,
      'recentDecks': [5, 2],
    };
    await pumpApp(tester);
    await tester.tap(find.byTooltip('Create filtered deck'));
    await tester.pumpAndSettle();
    expect(find.byType(FilteredDeckScreen), findsOneWidget);
    expect(find.text('New filtered deck'), findsOneWidget);
    // The most recently studied regular deck.
    expect(go.callsTo('decks.NewFilteredOptions'), [
      ['deck:Spanish'],
    ]);
    expect(find.text('Filtered Deck 1'), findsOneWidget);

    await tester.enterText(find.byKey(const Key('first-limit')), '50');
    await tester.tap(find.byKey(const Key('second-filter')));
    await tester.pumpAndSettle();
    await tester.enterText(find.byKey(const Key('second-search')), 'is:new');
    await tester.tap(find.byKey(const Key('build-filtered')));
    await tester.pumpAndSettle();

    final options = go.callsTo('decks.CreateFiltered').single[0] as Map;
    expect(options['name'], 'Filtered Deck 1');
    expect(options['terms'], [
      {'search': 'deck:Spanish', 'limit': 50, 'order': 'random'},
      {'search': 'is:new', 'limit': 20, 'order': 'due'},
    ]);
    expect(options['reschedule'], true);
    expect(find.byType(FilteredDeckScreen), findsNothing);
    expect(find.text('“Filtered Deck 1” holds 10 cards'), findsOneWidget);
    expect(find.text('Filtered Deck 1'), findsOneWidget);
  });

  testWidgets('no matching card: the error is shown and the screen stays', (
    tester,
  ) async {
    final go = withFiltered();
    go.filterMatches = (_) => 0;
    await pumpApp(tester);
    await tester.tap(find.byTooltip('Create filtered deck'));
    await tester.pumpAndSettle();
    await tester.enterText(find.byKey(const Key('first-limit')), '0');
    await tester.tap(find.byKey(const Key('build-filtered')));
    await tester.pumpAndSettle();
    expect(find.textContaining('number of cards from 1'), findsOneWidget);
    expect(go.callsTo('decks.CreateFiltered'), isEmpty);

    await tester.enterText(find.byKey(const Key('first-limit')), '100');
    await tester.tap(find.byKey(const Key('build-filtered')));
    await tester.pumpAndSettle();
    expect(find.text(FakeGo.noMatchError), findsOneWidget);
    expect(find.byType(FilteredDeckScreen), findsOneWidget);
    expect(go.tree, hasLength(3));
  });

  testWidgets('a filtered deck\'s options: preview mode with delays', (
    tester,
  ) async {
    final go = withFiltered();
    await pumpApp(tester);
    await openDeckMenu(tester, 'Cram', 'Options');
    expect(find.byType(DeckOptionsScreen), findsNothing);
    expect(find.text('Filtered deck options'), findsOneWidget);
    expect(find.text('tag:exam'), findsOneWidget);
    expect(find.byKey(const Key('delay-again')), findsNothing);

    await tester.tap(find.byKey(const Key('reschedule')));
    await tester.pumpAndSettle();
    await tester.enterText(find.byKey(const Key('delay-good')), '30');
    await tester.enterText(find.byKey(const Key('filtered-name')), 'Preview');
    await tester.tap(find.widgetWithText(FilledButton, 'Rebuild'));
    await tester.pumpAndSettle();

    final [id, options as Map] = go.callsTo('decks.UpdateFiltered').single;
    expect(id, 5);
    expect(options['name'], 'Preview');
    expect(options['reschedule'], false);
    expect(options['previewDelays'], [60, 600, 30]);
    expect(options['terms'], hasLength(1));
    expect(find.byType(FilteredDeckScreen), findsNothing);
    expect(find.text('Preview'), findsOneWidget);
  });

  testWidgets('add screen: filtered decks are not offered (spec)', (
    tester,
  ) async {
    withFiltered();
    await pumpApp(tester);
    await tester.tap(find.text('Add').last);
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('deck-picker')));
    await tester.pumpAndSettle();
    expect(find.text('Spanish').hitTestable(), findsOneWidget);
    expect(find.text('Cram').hitTestable(), findsNothing);
  });

  testWidgets(
    'custom study: review forgotten cards starts the session (spec)',
    (tester) async {
      final go = withFiltered();
      handleStudy(go);
      await pumpApp(tester);
      await openDeckMenu(tester, 'Spanish', 'Custom study');
      await tester.tap(find.byKey(const Key('custom-forgotten')));
      await tester.pumpAndSettle();
      await tester.enterText(find.byKey(const Key('customDays')), '2');
      await tester.tap(find.text('OK'));
      await tester.pumpAndSettle();
      expect(go.callsTo('decks.CustomStudy').single, [
        2,
        {
          'kind': 'forgotten',
          'days': 2,
          'state': '',
          'tags': <String>[],
          'cards': 0,
        },
      ]);
      expect(find.byType(StudyScreen), findsOneWidget);
      final deckId = go.callsTo('study.Start').single[0];
      expect(go.filteredOptions[deckId]!['name'], 'Custom Study Session');
      expect(go.callsTo('decks.IncreaseToday'), isEmpty);
    },
  );

  testWidgets('custom study by card state and tags; no match is shown', (
    tester,
  ) async {
    final go = withFiltered();
    handleStudy(go);
    go.filterMatches = (_) => 0;
    await pumpApp(tester);
    await openDeckMenu(tester, 'Spanish', 'Custom study');
    await tester.tap(find.byKey(const Key('custom-byState')));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('cardState')));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Due cards only').last);
    await tester.pumpAndSettle();
    await tester.enterText(find.byKey(const Key('customTags')), 'verb, noun');
    await tester.enterText(find.byKey(const Key('customCards')), '25');
    await tester.tap(find.text('OK'));
    await tester.pumpAndSettle();
    final [deck, choice as Map] = go.callsTo('decks.CustomStudy').single;
    expect(deck, 2);
    expect(choice['kind'], 'byState');
    expect(choice['state'], 'due');
    expect(choice['tags'], ['verb', 'noun']);
    expect(choice['cards'], 25);
    expect(find.text(FakeGo.noMatchError), findsOneWidget);
    expect(find.byType(StudyScreen), findsNothing);

    // Invalid numbers are explained and not sent.
    await tester.tap(find.byKey(const Key('custom-ahead')));
    await tester.pumpAndSettle();
    await tester.enterText(find.byKey(const Key('customDays')), '0');
    await tester.tap(find.text('OK'));
    await tester.pumpAndSettle();
    expect(find.text('Enter a whole number from 1 to 365.'), findsOneWidget);
    expect(go.callsTo('decks.CustomStudy'), hasLength(1));
  });

  testWidgets('filtered deck from search (spec): the browser search is used', (
    tester,
  ) async {
    final go = withFiltered();
    await openBrowse(tester);
    await tester.enterText(find.byType(TextField).first, 'tag:exam is:due');
    await tester.testTextInput.receiveAction(TextInputAction.search);
    await tester.pumpAndSettle();
    await tester.tap(find.byTooltip('More'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Create filtered deck'));
    await tester.pumpAndSettle();
    expect(go.callsTo('decks.NewFilteredOptions'), [
      ['tag:exam is:due'],
    ]);
    expect(find.byType(FilteredDeckScreen), findsOneWidget);
    expect(find.text('tag:exam is:due'), findsOneWidget);
  });
}
