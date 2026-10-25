import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/data/changes.dart';
import 'package:cardflip/data/deck_list_model.dart';
import 'package:cardflip/features/decks/deck_options_screen.dart';
import 'package:cardflip/features/decks/deck_overview_screen.dart';
import 'package:cardflip/features/decks/deck_picker.dart';
import 'package:cardflip/features/import/import_screen.dart';
import 'package:cardflip/features/study/study_screen.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/time_format.dart';

import 'deck_overview_test.dart' show handleStudy;
import 'support/fake_go.dart';

Color? colorOf(WidgetTester tester, Finder scope, String text) {
  final t = tester.widget<Text>(
    find.descendant(of: scope, matching: find.text(text)),
  );
  return t.style?.color;
}

Future<void> openDeckMenu(
  WidgetTester tester,
  String deck,
  String action,
) async {
  final row = find
      .ancestor(of: find.text(deck), matching: find.byType(InkWell))
      .first;
  await tester.tap(
    find.descendant(of: row, matching: find.byTooltip('Deck actions')),
  );
  await tester.pumpAndSettle();
  await tester.tap(find.text(action).last);
  await tester.pumpAndSettle();
}

void main() {
  testWidgets('due counts in three distinguishable colours', (tester) async {
    final go = FakeGo();
    go.tree = [
      deckNode(1, 'Default', newCount: 5, learning: 2, review: 30)
        ..['cards'] = 40,
    ];
    await pumpApp(tester);
    final counts = find.byKey(const Key('counts-1'));
    expect(
      find.descendant(of: counts, matching: find.text('5')),
      findsOneWidget,
    );
    expect(
      find.descendant(of: counts, matching: find.text('2')),
      findsOneWidget,
    );
    expect(
      find.descendant(of: counts, matching: find.text('30')),
      findsOneWidget,
    );
    final colors = {
      colorOf(tester, counts, '5'),
      colorOf(tester, counts, '2'),
      colorOf(tester, counts, '30'),
    };
    expect(colors, hasLength(3));
    expect(find.text('Your collection is empty'), findsNothing);
  });

  testWidgets('first run: Default deck and a hint to add or import', (
    tester,
  ) async {
    FakeGo();
    await pumpApp(tester);
    expect(find.text('Default'), findsOneWidget);
    expect(find.text('Your collection is empty'), findsOneWidget);
    expect(find.text('Add a card'), findsOneWidget);
    await tester.tap(find.text('Import a file'));
    await tester.pumpAndSettle();
    expect(find.byType(ImportScreen), findsOneWidget);
  });

  testWidgets('deleting a deck asks for confirmation naming the card count', (
    tester,
  ) async {
    final go = FakeGo();
    go.tree = [
      deckNode(1, 'Default')..['cards'] = 3,
      deckNode(2, 'Spanish', hasChildren: true)..['cards'] = 12,
      deckNode(3, 'Spanish::Verbs', parentId: 2)..['cards'] = 7,
    ];
    go.handle(
      'decks.DeletionInfo',
      (_) => deletionInfoJson(decks: 2, cards: 12),
    );
    go.handle('decks.Delete', (_) => null);
    await pumpApp(tester);

    await openDeckMenu(tester, 'Spanish', 'Delete');
    expect(find.text('Delete “Spanish”?'), findsOneWidget);
    expect(find.textContaining('12 cards'), findsOneWidget);
    expect(find.textContaining('1 subdeck'), findsOneWidget);
    await tester.tap(find.text('Cancel'));
    await tester.pumpAndSettle();
    expect(go.callsTo('decks.Delete'), isEmpty);

    await openDeckMenu(tester, 'Spanish', 'Delete');
    await tester.tap(find.widgetWithText(FilledButton, 'Delete'));
    await tester.pumpAndSettle();
    expect(go.callsTo('decks.Delete'), [
      [2],
    ]);
  });

  testWidgets(
    'deleting an empty deck needs no confirmation; errors are shown',
    (tester) async {
      final go = FakeGo();
      go.handle('decks.DeletionInfo', (_) => deletionInfoJson());
      go.handle(
        'decks.Delete',
        (_) => throw 'The last deck cannot be deleted.',
      );
      await pumpApp(tester);
      await openDeckMenu(tester, 'Default', 'Delete');
      expect(find.text('The last deck cannot be deleted.'), findsOneWidget);
    },
  );

  testWidgets('parents collapse, and counts refresh after a change', (
    tester,
  ) async {
    final go = FakeGo();
    go.tree = [
      deckNode(2, 'Spanish', hasChildren: true)..['cards'] = 1,
      deckNode(3, 'Spanish::Verbs', parentId: 2)..['cards'] = 1,
    ];
    go.handle('decks.SetCollapsed', (args) {
      go.tree[0] = deckNode(
        2,
        'Spanish',
        hasChildren: true,
        collapsed: args[1] as bool,
      )..['cards'] = 1;
      go.emitChange();
      return null;
    });
    await pumpApp(tester);
    expect(find.text('Verbs'), findsOneWidget);
    await tester.tap(find.byTooltip('Collapse'));
    await tester.pumpAndSettle();
    expect(go.callsTo('decks.SetCollapsed'), [
      [2, true],
    ]);
    expect(find.text('Verbs'), findsNothing);

    // An import finishing elsewhere: the list shows the new deck and counts.
    go.tree = [...go.tree, deckNode(4, 'Imported', newCount: 9)..['cards'] = 9];
    go.emitChange(all: true);
    await tester.pumpAndSettle();
    expect(find.text('Imported'), findsOneWidget);
    expect(
      find.descendant(
        of: find.byKey(const Key('counts-4')),
        matching: find.text('9'),
      ),
      findsOneWidget,
    );
  });

  testWidgets('custom study: ten more new cards for today', (tester) async {
    final go = FakeGo();
    go.tree = [
      deckNode(1, 'Default'),
      deckNode(2, 'Spanish', newCount: 0, review: 3)..['cards'] = 60,
    ];
    go.extraCards[2] = (newCards: 30, reviews: 0);
    await pumpApp(tester);
    await openDeckMenu(tester, 'Spanish', 'Custom study');
    expect(find.text('30 more available'), findsOneWidget);
    expect(find.text('0 more available'), findsOneWidget);
    await tester.enterText(find.byKey(const Key('increaseBy')), '10');
    await tester.tap(find.text('OK'));
    await tester.pumpAndSettle();
    expect(find.byType(DeckOptionsScreen), findsNothing);
    expect(go.callsTo('decks.IncreaseToday').single, [2, 10, 0]);
    expect(go.callsTo('decks.SaveSettings'), isEmpty);
    final counts = find.byKey(const Key('counts-2'));
    expect(
      find.descendant(of: counts, matching: find.text('10')),
      findsOneWidget,
    );
  });

  testWidgets('custom study: review limit, and invalid numbers', (
    tester,
  ) async {
    final go = FakeGo();
    go.extraCards[1] = (newCards: 0, reviews: 40);
    await pumpApp(tester);
    await openDeckMenu(tester, 'Default', 'Custom study');
    await tester.tap(find.text("Increase today's review card limit"));
    await tester.enterText(find.byKey(const Key('increaseBy')), '0');
    await tester.tap(find.text('OK'));
    await tester.pumpAndSettle();
    expect(find.text('Enter a whole number from 1 to 9999.'), findsOneWidget);
    expect(go.callsTo('decks.IncreaseToday'), isEmpty);
    await tester.enterText(find.byKey(const Key('increaseBy')), '25');
    await tester.tap(find.text('OK'));
    await tester.pumpAndSettle();
    expect(go.callsTo('decks.IncreaseToday').single, [1, 0, 25]);
  });

  testWidgets('deck options are saved; invalid values show the error', (
    tester,
  ) async {
    final go = FakeGo();
    await pumpApp(tester);
    await openDeckMenu(tester, 'Default', 'Options');
    await tester.enterText(find.byKey(const Key('newPerDay')), '50000');
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    expect(
      find.text('New cards per day must be between 0 and 9999.'),
      findsOneWidget,
    );
    await tester.enterText(find.byKey(const Key('newPerDay')), '15');
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    final saved = go.callsTo('decks.SaveSettings').last;
    expect(saved[0], 1);
    expect((saved[2] as Map)['newPerDay'], 15);
    expect((saved[2] as Map)['desiredRetention'], 0.9);
    expect((saved[2] as Map)['disableAutoplay'], false);
    expect(find.byType(DeckOptionsScreen), findsNothing);
  });

  testWidgets('the audio deck options are saved', (tester) async {
    final go = FakeGo();
    await pumpApp(tester);
    await openDeckMenu(tester, 'Default', 'Options');
    for (final label in [
      "Don't play audio automatically",
      'Skip question when replaying answer',
    ]) {
      final tile = find.widgetWithText(SwitchListTile, label);
      await tester.ensureVisible(tile);
      expect(tester.widget<SwitchListTile>(tile).value, isFalse);
      await tester.tap(tile);
      await tester.pumpAndSettle();
    }
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    final saved = go.callsTo('decks.SaveSettings').single[2] as Map;
    expect(saved['disableAutoplay'], true);
    expect(saved['skipQuestionOnReplay'], true);
  });

  testWidgets('limit behaviour options are saved', (tester) async {
    final go = FakeGo();
    await pumpApp(tester);
    await openDeckMenu(tester, 'Default', 'Options');
    // Defaults (spec): new cards ignore the review limit, limits do not
    // start from the top.
    final ignore = find.byKey(const Key('newIgnoresReviewLimit'));
    final fromTop = find.byKey(const Key('limitsStartFromTop'));
    expect(tester.widget<SwitchListTile>(ignore).value, isTrue);
    expect(tester.widget<SwitchListTile>(fromTop).value, isFalse);
    for (final tile in [ignore, fromTop]) {
      await tester.ensureVisible(tile);
      await tester.tap(tile);
      await tester.pumpAndSettle();
    }
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    final saved = go.callsTo('decks.SaveSettings').single[2] as Map;
    expect(saved['newIgnoresReviewLimit'], false);
    expect(saved['limitsStartFromTop'], true);
    expect(saved['newPerDay'], 20);
  });

  testWidgets('interval limit options are saved', (tester) async {
    final go = FakeGo();
    await pumpApp(tester);
    await openDeckMenu(tester, 'Default', 'Options');
    final max = find.byKey(const Key('maxInterval'));
    final min = find.byKey(const Key('minRelearnInterval'));
    expect(tester.widget<TextField>(max).controller!.text, '36500');
    expect(tester.widget<TextField>(min).controller!.text, '1');
    await tester.ensureVisible(min);
    await tester.enterText(max, '180');
    await tester.enterText(min, '3');
    await tester.pumpAndSettle();
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    final saved = go.callsTo('decks.SaveSettings').single[2] as Map;
    expect(saved['maxInterval'], 180);
    expect(saved['minRelearnInterval'], 3);
    expect(find.byType(DeckOptionsScreen), findsNothing);
  });

  testWidgets('invalid interval limits are not saved', (tester) async {
    final go = FakeGo();
    await pumpApp(tester);
    await openDeckMenu(tester, 'Default', 'Options');
    final max = find.byKey(const Key('maxInterval'));
    final min = find.byKey(const Key('minRelearnInterval'));
    await tester.ensureVisible(min);
    // Minimum above maximum (spec): rejected with an explanation.
    await tester.enterText(max, '30');
    await tester.enterText(min, '60');
    await tester.pumpAndSettle();
    expect(
      find.text('Must not be above the maximum interval.'),
      findsOneWidget,
    );
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    expect(find.text('Correct the intervals first.'), findsOneWidget);
    for (final bad in ['0', '36501', 'abc', '']) {
      await tester.enterText(max, bad);
      await tester.pumpAndSettle();
      expect(
        find.text('Enter a whole number of days from 1 to 36500.'),
        findsOneWidget,
        reason: bad,
      );
    }
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    expect(go.callsTo('decks.SaveSettings'), isEmpty);
    expect(find.byType(DeckOptionsScreen), findsOneWidget);
  });

  test('steps parse and format', () {
    // Valid steps (spec): shown the same way when the options are reopened.
    expect(parseSteps('30s 5m 1h 2d'), [30, 300, 3600, 172800]);
    expect(formatSteps(parseSteps('30s 5m 1h 2d')), '30s 5m 1h 2d');
    expect(formatSteps([90, 5400, 600, 86400 * 3]), '90s 90m 10m 3d');
    expect(parseSteps('  1.5m   10m '), [90, 600]);
    expect(parseSteps(''), isEmpty);
    for (final bad in ['5x', '0m', '10', 'm', '-5m', '366d']) {
      expect(() => parseSteps(bad), throwsFormatException, reason: bad);
    }
    expect(
      () => parseSteps(List.filled(21, '1m').join(' ')),
      throwsFormatException,
    );
  });

  testWidgets('learning step options are saved', (tester) async {
    final go = FakeGo();
    await pumpApp(tester);
    await openDeckMenu(tester, 'Default', 'Options');
    expect(find.text('1m 10m'), findsOneWidget);
    expect(find.text('10m'), findsOneWidget);
    await tester.enterText(
      find.byKey(const Key('learningSteps')),
      '30s 5m 1h 2d',
    );
    await tester.enterText(find.byKey(const Key('relearningSteps')), '');
    await tester.ensureVisible(find.byKey(const Key('interdayOrder')));
    await tester.tap(find.byKey(const Key('interdayOrder')));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Show before reviews').last);
    await tester.pumpAndSettle();
    final bury = find.widgetWithText(
      SwitchListTile,
      'Bury interday learning siblings',
    );
    await tester.ensureVisible(bury);
    await tester.tap(bury);
    await tester.pumpAndSettle();
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    final saved = go.callsTo('decks.SaveSettings').single[2] as Map;
    expect(saved['learningSteps'], [30, 300, 3600, 172800]);
    expect(saved['relearningSteps'], isEmpty);
    expect(saved['interdayLearningOrder'], 'before');
    expect(saved['buryInterdayLearning'], true);
    expect(find.byType(DeckOptionsScreen), findsNothing);
  });

  testWidgets('burying options are saved', (tester) async {
    final go = FakeGo();
    await pumpApp(tester);
    await openDeckMenu(tester, 'Default', 'Options');
    SwitchListTile switchOf(String key) =>
        tester.widget<SwitchListTile>(find.byKey(Key(key)));
    // Defaults (spec): both burying options are off.
    expect(find.text('Burying'), findsOneWidget);
    expect(switchOf('buryNew').value, false);
    expect(switchOf('buryReview').value, false);
    expect(switchOf('buryInterdayLearning').value, false);
    await tester.ensureVisible(find.byKey(const Key('buryReview')));
    await tester.tap(find.byKey(const Key('buryReview')));
    await tester.pumpAndSettle();
    expect(switchOf('buryReview').value, true);
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    final saved = go.callsTo('decks.SaveSettings').single[2] as Map;
    expect(saved['buryNew'], false);
    expect(saved['buryReview'], true);
    expect(saved['buryInterdayLearning'], false);
  });

  testWidgets('display order options are saved', (tester) async {
    final go = FakeGo();
    await pumpApp(tester);
    await openDeckMenu(tester, 'Default', 'Options');
    Future<void> choose(String key, String option) async {
      await tester.ensureVisible(find.byKey(Key(key)));
      await tester.tap(find.byKey(Key(key)));
      await tester.pumpAndSettle();
      await tester.tap(find.text(option).last);
      await tester.pumpAndSettle();
    }

    await choose('newGatherOrder', 'Random notes');
    await choose('newSortOrder', 'Card type, then random');
    await choose('newReviewOrder', 'Show after reviews');
    await choose('reviewSortOrder', 'Ascending retrievability');
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    final saved = go.callsTo('decks.SaveSettings').single[2] as Map;
    expect(saved['newGatherOrder'], 'randomNotes');
    expect(saved['newSortOrder'], 'typeThenRandom');
    expect(saved['newReviewOrder'], 'after');
    expect(saved['reviewSortOrder'], 'retrievabilityAscending');
    expect(saved['interdayLearningOrder'], 'mix');
    expect(find.byType(DeckOptionsScreen), findsNothing);
  });

  testWidgets('insertion order defaults to sequential and is saved', (
    tester,
  ) async {
    final go = FakeGo();
    await pumpApp(tester);
    await openDeckMenu(tester, 'Default', 'Options');
    await tester.ensureVisible(find.byKey(const Key('insertionOrder')));
    expect(find.text('Sequential (oldest cards first)'), findsOneWidget);
    await tester.tap(find.byKey(const Key('insertionOrder')));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Random').last);
    await tester.pumpAndSettle();
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    final saved = go.callsTo('decks.SaveSettings').single[2] as Map;
    expect(saved['insertionOrder'], 'random');
  });

  testWidgets('leech options are saved', (tester) async {
    final go = FakeGo();
    await pumpApp(tester);
    await openDeckMenu(tester, 'Default', 'Options');
    final threshold = find.byKey(const Key('leechThreshold'));
    await tester.ensureVisible(threshold);
    expect(find.text('8'), findsOneWidget);
    expect(find.text('Tag only'), findsOneWidget);
    await tester.enterText(threshold, 'x');
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    expect(
      find.text('Enter a whole number for the leech threshold.'),
      findsOneWidget,
    );
    expect(go.callsTo('decks.SaveSettings'), isEmpty);
    await tester.enterText(threshold, '0');
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    expect(
      find.text('Leech threshold must be between 1 and 99 lapses.'),
      findsOneWidget,
    );
    // Options stored (spec): threshold 5, suspend.
    await tester.enterText(threshold, '5');
    await tester.ensureVisible(find.byKey(const Key('leechAction')));
    await tester.tap(find.byKey(const Key('leechAction')));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Suspend card and tag').last);
    await tester.pumpAndSettle();
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    final saved = go.callsTo('decks.SaveSettings').last[2] as Map;
    expect(saved['leechThreshold'], 5);
    expect(saved['leechAction'], 'suspend');
    expect(find.byType(DeckOptionsScreen), findsNothing);
  });

  testWidgets('timer options are saved', (tester) async {
    final go = FakeGo();
    await pumpApp(tester);
    await openDeckMenu(tester, 'Default', 'Options');
    final seconds = find.byKey(const Key('maxAnswerSeconds'));
    await tester.ensureVisible(seconds);
    expect(find.text('60'), findsOneWidget);
    expect(
      tester.widget<SwitchListTile>(find.byKey(const Key('showTimer'))).value,
      isFalse,
    );
    final stop = find.byKey(const Key('stopTimerOnAnswer'));
    expect(tester.widget<SwitchListTile>(stop).value, isTrue);
    // Out of range (spec): 0 is rejected and nothing is saved.
    await tester.enterText(seconds, '0');
    await tester.pump();
    expect(
      find.text('Enter a whole number of seconds from 1 to 9999.'),
      findsOneWidget,
    );
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    expect(go.callsTo('decks.SaveSettings'), isEmpty);
    expect(find.byType(DeckOptionsScreen), findsOneWidget);
    await tester.enterText(seconds, '90');
    await tester.ensureVisible(find.byKey(const Key('showTimer')));
    await tester.tap(find.byKey(const Key('showTimer')));
    await tester.pumpAndSettle();
    await tester.ensureVisible(stop);
    await tester.tap(stop);
    await tester.pumpAndSettle();
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    final saved = go.callsTo('decks.SaveSettings').single[2] as Map;
    expect(saved['maxAnswerSeconds'], 90);
    expect(saved['showTimer'], isTrue);
    expect(saved['stopTimerOnAnswer'], isFalse);
    expect(find.byType(DeckOptionsScreen), findsNothing);
  });

  testWidgets('auto advance options are saved', (tester) async {
    final go = FakeGo();
    await pumpApp(tester);
    await openDeckMenu(tester, 'Default', 'Options');
    final question = find.byKey(const Key('secondsToShowQuestion'));
    final answer = find.byKey(const Key('secondsToShowAnswer'));
    await tester.ensureVisible(answer);
    expect(tester.widget<TextField>(question).controller!.text, '0');
    expect(tester.widget<TextField>(answer).controller!.text, '0');
    // Out of range: rejected while typing and not saved.
    await tester.enterText(answer, '10000');
    await tester.pump();
    expect(
      find.text('Enter a whole number of seconds from 0 to 9999.'),
      findsOneWidget,
    );
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    expect(go.callsTo('decks.SaveSettings'), isEmpty);
    // Options stored (spec): 5 seconds for the question, 3 for the answer,
    // answer Good.
    await tester.enterText(question, '5');
    await tester.enterText(answer, '3');
    final action = find.byKey(const Key('answerAction'));
    await tester.ensureVisible(action);
    await tester.tap(action);
    await tester.pumpAndSettle();
    await tester.tap(find.text('Answer Good').last);
    await tester.pumpAndSettle();
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    final saved = go.callsTo('decks.SaveSettings').single[2] as Map;
    expect(saved['secondsToShowQuestion'], 5);
    expect(saved['secondsToShowAnswer'], 3);
    expect(saved['questionAction'], 'showAnswer');
    expect(saved['answerAction'], 'answerGood');
    expect(find.byType(DeckOptionsScreen), findsNothing);
  });

  testWidgets('invalid steps are explained and not saved', (tester) async {
    final go = FakeGo();
    await pumpApp(tester);
    await openDeckMenu(tester, 'Default', 'Options');
    // Invalid step (spec): rejected with an explanation while typing.
    await tester.enterText(find.byKey(const Key('learningSteps')), '1m 5x');
    await tester.pump();
    expect(find.textContaining('"5x" is not a delay'), findsOneWidget);
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    expect(go.callsTo('decks.SaveSettings'), isEmpty);
    expect(find.byType(DeckOptionsScreen), findsOneWidget);
    await tester.enterText(find.byKey(const Key('learningSteps')), '1m 5m');
    await tester.pumpAndSettle();
    expect(find.textContaining('is not a delay'), findsNothing);
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    final saved = go.callsTo('decks.SaveSettings').single[2] as Map;
    expect(saved['learningSteps'], [60, 300]);
    expect(saved['relearningSteps'], [600]);
  });

  /// A fake with the decks Default (1), Languages (2) with the subdecks
  /// Spanish (3) and French (4); Spanish and French use the preset
  /// Languages (50), which has 10 new cards per day.
  FakeGo presetsFake() {
    final go = FakeGo();
    go.tree = [
      deckNode(1, 'Default'),
      deckNode(2, 'Languages', hasChildren: true),
      deckNode(3, 'Languages::Spanish', parentId: 2),
      deckNode(4, 'Languages::French', parentId: 2),
    ];
    go.presets[50] = {
      'id': 50,
      'name': 'Languages',
      'options': {...deckOptionsJson(), 'newPerDay': 10},
    };
    go.deckPresets.addAll({3: 50, 4: 50});
    return go;
  }

  Future<void> choosePreset(WidgetTester tester, String name) async {
    await tester.tap(find.byKey(const Key('preset')));
    await tester.pumpAndSettle();
    await tester.tap(find.text(name).last);
    await tester.pumpAndSettle();
  }

  Future<void> presetMenu(WidgetTester tester, String action) async {
    await tester.tap(find.byTooltip('Preset actions'));
    await tester.pumpAndSettle();
    await tester.tap(find.text(action).last);
    await tester.pumpAndSettle();
  }

  String fieldText(WidgetTester tester, String key) =>
      tester.widget<TextField>(find.byKey(Key(key))).controller!.text;

  testWidgets('switching presets shows and saves the chosen preset', (
    tester,
  ) async {
    final go = presetsFake();
    await pumpApp(tester);
    await openDeckMenu(tester, 'Default', 'Options');
    expect(find.text('Default'), findsWidgets);
    expect(
      find.text('Used by 2 decks. Changes apply to all of them.'),
      findsOneWidget,
    );
    expect(fieldText(tester, 'newPerDay'), '20');
    // Unsaved changes are not dropped without asking.
    await tester.enterText(find.byKey(const Key('newPerDay')), '7');
    await choosePreset(tester, 'Languages');
    expect(find.text('Discard changes?'), findsOneWidget);
    await tester.tap(find.text('Cancel'));
    await tester.pumpAndSettle();
    expect(fieldText(tester, 'newPerDay'), '7');
    await choosePreset(tester, 'Languages');
    await tester.tap(find.text('Discard'));
    await tester.pumpAndSettle();
    // The chosen preset's options are shown; once saved, Default deck is
    // its third deck.
    expect(fieldText(tester, 'newPerDay'), '10');
    expect(
      find.text('Used by 3 decks. Changes apply to all of them.'),
      findsOneWidget,
    );
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    final saved = go.callsTo('decks.SaveSettings').single;
    expect(saved[0], 1);
    expect(saved[1], 50);
    expect((saved[2] as Map)['newPerDay'], 10);
    expect(go.deckPresets[1], 50);
    expect(find.byType(DeckOptionsScreen), findsNothing);
  });

  testWidgets('deleting a preset moves its decks to Default', (tester) async {
    final go = presetsFake();
    await pumpApp(tester);
    await openDeckMenu(tester, 'Spanish', 'Options');
    expect(fieldText(tester, 'newPerDay'), '10');
    await presetMenu(tester, 'Delete preset');
    expect(
      find.text(
        'The preset "Languages" is used by 2 decks. They use Default instead.',
      ),
      findsOneWidget,
    );
    await tester.tap(find.text('Delete'));
    await tester.pumpAndSettle();
    expect(go.callsTo('decks.DeletePreset').single, [50]);
    expect(go.deckPresets, isEmpty);
    // The deck now shows Default, which cannot be deleted.
    expect(fieldText(tester, 'newPerDay'), '20');
    expect(
      find.text('Used by 4 decks. Changes apply to all of them.'),
      findsOneWidget,
    );
    await tester.tap(find.byTooltip('Preset actions'));
    await tester.pumpAndSettle();
    final delete = tester.widget<PopupMenuItem<Object?>>(
      find.ancestor(
        of: find.text('Delete preset'),
        matching: find.byWidgetPredicate((w) => w is PopupMenuItem),
      ),
    );
    expect(delete.enabled, isFalse);
  });

  testWidgets('this-deck limits override the preset for one deck', (
    tester,
  ) async {
    final go = presetsFake();
    await pumpApp(tester);
    await openDeckMenu(tester, 'French', 'Options');
    expect(fieldText(tester, 'deckNewLimit'), '');
    await tester.enterText(find.byKey(const Key('deckNewLimit')), 'x');
    await tester.pump();
    expect(find.text('Enter a whole number'), findsOneWidget);
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    expect(go.callsTo('decks.SaveSettings'), isEmpty);
    expect(
      find.text('Enter whole numbers for the limits of this deck.'),
      findsOneWidget,
    );
    // One deck lower (spec): French gets 5, the preset keeps 10.
    await tester.enterText(find.byKey(const Key('deckNewLimit')), '5');
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    final saved = go.callsTo('decks.SaveSettings').single;
    expect(saved[1], 50);
    expect((saved[2] as Map)['newPerDay'], 10);
    expect(saved[3], {'newPerDay': 5, 'reviewsPerDay': null});
    // Reopened, the screen shows the deck's own limit.
    await openDeckMenu(tester, 'French', 'Options');
    expect(fieldText(tester, 'deckNewLimit'), '5');
    expect(fieldText(tester, 'deckReviewLimit'), '');
  });

  testWidgets('presets are added, renamed and saved to subdecks', (
    tester,
  ) async {
    final go = presetsFake();
    await pumpApp(tester);
    await openDeckMenu(tester, 'Languages', 'Options');
    // Add: the deck uses the new preset, which has the default options.
    await presetMenu(tester, 'Add preset');
    await tester.enterText(find.byType(TextField).last, 'languages');
    await tester.tap(find.widgetWithText(FilledButton, 'Add'));
    await tester.pumpAndSettle();
    expect(
      find.text('A preset named "Languages" already exists.'),
      findsOneWidget,
    );
    await presetMenu(tester, 'Add preset');
    await tester.enterText(find.byType(TextField).last, 'Exam');
    await tester.tap(find.widgetWithText(FilledButton, 'Add'));
    await tester.pumpAndSettle();
    expect(go.callsTo('decks.AddPreset').last, [2, 'Exam']);
    expect(go.deckPresets[2], 100);
    expect(find.text('Exam'), findsOneWidget);
    expect(find.text('Used by 1 deck.'), findsOneWidget);
    // Rename keeps unsaved changes.
    await tester.enterText(find.byKey(const Key('newPerDay')), '30');
    await presetMenu(tester, 'Rename preset');
    await tester.enterText(find.byType(TextField).last, 'Finals');
    await tester.tap(find.widgetWithText(FilledButton, 'Rename'));
    await tester.pumpAndSettle();
    expect(go.callsTo('decks.RenamePreset').single, [100, 'Finals']);
    expect(find.text('Finals'), findsOneWidget);
    expect(fieldText(tester, 'newPerDay'), '30');
    // Save to all subdecks saves the shown options first.
    await presetMenu(tester, 'Save to all subdecks');
    expect(
      (go.callsTo('decks.SaveSettings').single[2] as Map)['newPerDay'],
      30,
    );
    expect(go.callsTo('decks.SavePresetToSubdecks').single, [2]);
    expect(go.deckPresets, {2: 100, 3: 100, 4: 100});
    expect(
      find.text('Used by 3 decks. Changes apply to all of them.'),
      findsOneWidget,
    );
  });

  testWidgets('create, rename and study', (tester) async {
    final go = FakeGo();
    go.handle('decks.Create', (_) => 5);
    go.handle('decks.Rename', (_) => null);
    await pumpApp(tester);
    await tester.tap(find.byTooltip('Create deck'));
    await tester.pumpAndSettle();
    await tester.enterText(find.byType(TextField), 'Languages::Spanish');
    await tester.tap(find.text('Create'));
    await tester.pumpAndSettle();
    expect(go.callsTo('decks.Create'), [
      ['Languages::Spanish'],
    ]);

    await openDeckMenu(tester, 'Default', 'Rename');
    await tester.enterText(find.byType(TextField), 'Main');
    await tester.tap(find.widgetWithText(FilledButton, 'Rename'));
    await tester.pumpAndSettle();
    expect(go.callsTo('decks.Rename'), [
      [1, 'Main'],
    ]);

    // Tapping a deck opens its overview; the menu's Study starts studying
    // directly.
    await tester.tap(find.text('Default'));
    await tester.pumpAndSettle();
    expect(find.byType(DeckOverviewScreen), findsOneWidget);
    expect(find.byType(StudyScreen), findsNothing);
    await tester.pageBack();
    await tester.pumpAndSettle();
    await openDeckMenu(tester, 'Default', 'Study');
    expect(find.byType(StudyScreen), findsOneWidget);
    expect(find.byType(DeckOverviewScreen), findsNothing);
  });

  testWidgets('learning count updates when a learning card becomes due', (
    tester,
  ) async {
    final go = FakeGo();
    go.nextChange = () => DateTime.now().add(const Duration(seconds: 60));
    var calls = 0;
    go.handle('decks.Tree', (_) {
      calls++;
      return [deckNode(1, 'Default', learning: calls == 1 ? 0 : 1)];
    });
    await pumpApp(tester);
    final counts = find.byKey(const Key('counts-1'));
    expect(find.descendant(of: counts, matching: find.text('1')), findsNothing);
    await tester.pump(const Duration(seconds: 61));
    await tester.pump();
    expect(
      find.descendant(of: counts, matching: find.text('1')),
      findsOneWidget,
    );
  });

  testWidgets('reloads when the app returns to the foreground', (tester) async {
    final go = FakeGo();
    await pumpApp(tester);
    final before = go.callsTo('decks.Tree').length;
    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.inactive);
    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.resumed);
    await tester.pump();
    expect(go.callsTo('decks.Tree').length, before + 1);
  });

  group('newly due learning cards', () {
    /// Loads a model against [go] and records its reminders. Call the
    /// returned close function before the test ends, to cancel its timer.
    Future<(DeckListModel, List<LearningDue>, void Function())> loaded(
      WidgetTester tester,
      FakeGo go,
    ) async {
      final hub = ChangeHub();
      final model = DeckListModel(hub);
      final events = <LearningDue>[];
      final sub = model.learningDue.listen(events.add);
      await model.load();
      await tester.pump();
      return (
        model,
        events,
        () {
          sub.cancel();
          model.dispose();
          hub.dispose();
        },
      );
    }

    List<Map<String, Object?>> tree({int spanish = 0, int biology = 0}) => [
      deckNode(1, 'Spanish', learning: spanish, hasChildren: true),
      deckNode(2, 'Spanish::Verbs', learning: spanish, parentId: 1),
      deckNode(3, 'Biology', learning: biology),
    ];

    testWidgets('a time-driven increase reports the deck once', (tester) async {
      final go = FakeGo()..tree = tree();
      // The counts change right now: the next reload is after that moment.
      go.nextChange = DateTime.now;
      final (model, events, close) = await loaded(tester, go);
      go.tree = tree(spanish: 1);
      await tester.pump(const Duration(seconds: 2));
      await tester.pump(const Duration(seconds: 2));
      expect(events, hasLength(1));
      expect(events.single.decks.map((d) => d.fullName), ['Spanish']);
      expect(model.decks.first.counts.learning, 1);
      close();
    });

    testWidgets('a reload before the due moment reports nothing', (
      tester,
    ) async {
      final go = FakeGo()..tree = tree();
      go.nextChange = () => DateTime.now().add(const Duration(minutes: 1));
      final (model, events, close) = await loaded(tester, go);
      // Answering a card reloads the list before the due moment.
      go.tree = tree(spanish: 1);
      await model.load();
      await tester.pump();
      expect(model.decks.first.counts.learning, 1);
      expect(events, isEmpty);
      close();
    });

    testWidgets('two decks increasing are reported in one event', (
      tester,
    ) async {
      final go = FakeGo()..tree = tree();
      go.nextChange = DateTime.now;
      final (_, events, close) = await loaded(tester, go);
      go.tree = tree(spanish: 2, biology: 1);
      await tester.pump(const Duration(seconds: 2));
      await tester.pump(const Duration(seconds: 2));
      expect(events, hasLength(1));
      expect(events.single.decks.map((d) => d.fullName), [
        'Spanish',
        'Biology',
      ]);
      close();
    });
  });

  group('export', () {
    FakeGo spanishGo() => FakeGo()
      ..tree = [
        deckNode(1, 'Default'),
        deckNode(2, 'Languages', hasChildren: true),
        deckNode(3, 'Languages::Spanish', parentId: 2),
      ];

    bool checked(WidgetTester tester, String key) =>
        tester.widget<CheckboxListTile>(find.byKey(Key(key))).value!;

    testWidgets('the deck menu exports the deck with the chosen options', (
      tester,
    ) async {
      final go = spanishGo();
      final picker = FakePicker();
      await pumpApp(tester, picker: picker);
      await openDeckMenu(tester, 'Spanish', 'Export');
      expect(find.text('Export “Spanish”'), findsOneWidget);
      // Defaults: no scheduling, no presets, media, the current format.
      expect(checked(tester, 'export-scheduling'), isFalse);
      expect(checked(tester, 'export-presets'), isFalse);
      expect(checked(tester, 'export-media'), isTrue);
      expect(checked(tester, 'export-legacy'), isFalse);
      await tester.tap(find.text('Include scheduling information'));
      await tester.ensureVisible(find.text('Support older Anki versions'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Support older Anki versions'));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(FilledButton, 'Export'));
      await tester.pumpAndSettle();
      expect(go.callsTo('transfer.ExportDeck'), [
        [
          3,
          {'scheduling': true, 'presets': false, 'media': true, 'legacy': true},
        ],
      ]);
      expect(picker.saved.keys, ['Spanish.apkg']);
      expect(find.byType(AlertDialog), findsNothing);
      expect(find.text('Exported Spanish.apkg'), findsOneWidget);
    });

    testWidgets('progress is shown while the package is written', (
      tester,
    ) async {
      final go = spanishGo();
      final events = StreamController<Object?>();
      go.exportEvents = (_, _) => events.stream;
      final picker = FakePicker();
      await pumpApp(tester, picker: picker);
      await openDeckMenu(tester, 'Spanish', 'Export');
      await tester.tap(find.widgetWithText(FilledButton, 'Export'));
      await tester.pump();
      events.add(exportProgress('collection', 0, 1));
      await tester.pump();
      expect(find.text('Writing the cards…'), findsOneWidget);
      events.add(exportProgress('media', 2, 4));
      await tester.pump();
      expect(find.text('Adding media (2 of 4)…'), findsOneWidget);
      expect(
        tester
            .widget<LinearProgressIndicator>(
              find.byKey(const Key('export-progress')),
            )
            .value,
        0.5,
      );
      // The options and Export are locked while exporting.
      expect(
        tester
            .widget<FilledButton>(find.widgetWithText(FilledButton, 'Export'))
            .onPressed,
        isNull,
      );
      expect(
        tester
            .widget<CheckboxListTile>(find.byKey(const Key('export-media')))
            .onChanged,
        isNull,
      );
      events.add(exportDone('Spanish.apkg'));
      await tester.pumpAndSettle();
      expect(picker.saved.keys, ['Spanish.apkg']);
      expect(find.byType(AlertDialog), findsNothing);
      unawaited(events.close());
    });

    testWidgets('a failed export shows its error and can be retried', (
      tester,
    ) async {
      final go = spanishGo();
      var fail = true;
      go.exportEvents = (channel, args) => Stream.value(
        fail
            ? exportFailed('that deck no longer exists.')
            : exportDone(go.exportName(channel, args)),
      );
      final picker = FakePicker();
      await pumpApp(tester, picker: picker);
      await openDeckMenu(tester, 'Spanish', 'Export');
      await tester.tap(find.widgetWithText(FilledButton, 'Export'));
      await tester.pumpAndSettle();
      expect(find.text('That deck no longer exists.'), findsOneWidget);
      expect(picker.saved, isEmpty);
      fail = false;
      await tester.tap(find.widgetWithText(FilledButton, 'Export'));
      await tester.pumpAndSettle();
      expect(picker.saved.keys, ['Spanish.apkg']);
    });

    testWidgets('web: the package is downloaded as Spanish.apkg', (
      tester,
    ) async {
      final go = spanishGo();
      final picker = FakePicker();
      await pumpApp(tester, picker: picker, isWeb: true);
      await openDeckMenu(tester, 'Spanish', 'Export');
      await tester.tap(find.widgetWithText(FilledButton, 'Export'));
      await tester.pumpAndSettle();
      expect(go.callsTo('transfer.ExportDeck').single[0], 3);
      expect(picker.saved.keys, ['Spanish.apkg']);
    });

    Future<void> chooseFormat(WidgetTester tester, String format) async {
      await tester.tap(find.byKey(const Key('export-format')));
      await tester.pumpAndSettle();
      await tester.tap(find.text(format).last);
      await tester.pumpAndSettle();
    }

    testWidgets('the deck menu exports the notes as text with tags', (
      tester,
    ) async {
      final go = spanishGo();
      final picker = FakePicker();
      await pumpApp(tester, picker: picker);
      await openDeckMenu(tester, 'Spanish', 'Export');
      await chooseFormat(tester, 'Notes in plain text (.txt)');
      // Only the column options: tags on, the others off.
      expect(find.byKey(const Key('export-media')), findsNothing);
      expect(find.byKey(const Key('export-scheduling')), findsNothing);
      expect(checked(tester, 'export-tags'), isTrue);
      expect(checked(tester, 'export-deck'), isFalse);
      expect(checked(tester, 'export-note-type'), isFalse);
      expect(checked(tester, 'export-guid'), isFalse);
      await tester.tap(find.text('Include deck name'));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(FilledButton, 'Export'));
      await tester.pumpAndSettle();
      expect(go.callsTo('transfer.ExportNotesText'), [
        [
          {'deckId': 3, 'cardIds': <int>[], 'all': false},
          {'tags': true, 'deck': true, 'noteType': false, 'guid': false},
        ],
      ]);
      expect(go.callsTo('transfer.ExportDeck'), isEmpty);
      expect(picker.saved.keys, ['Spanish.txt']);
      expect(find.text('Exported Spanish.txt'), findsOneWidget);
    });

    testWidgets('a cards export shows rendering progress', (tester) async {
      final go = spanishGo();
      final events = StreamController<Object?>();
      go.exportEvents = (_, _) => events.stream;
      final picker = FakePicker();
      await pumpApp(tester, picker: picker);
      await openDeckMenu(tester, 'Spanish', 'Export');
      await chooseFormat(tester, 'Cards in plain text (.txt)');
      expect(find.byType(CheckboxListTile), findsNothing);
      await tester.tap(find.widgetWithText(FilledButton, 'Export'));
      await tester.pump();
      expect(go.callsTo('transfer.ExportCardsText'), [
        [
          {'deckId': 3, 'cardIds': <int>[], 'all': false},
        ],
      ]);
      events.add(exportProgress('cards', 1, 4));
      await tester.pump();
      expect(find.text('Rendering the cards (1 of 4)…'), findsOneWidget);
      expect(
        tester
            .widget<LinearProgressIndicator>(
              find.byKey(const Key('export-progress')),
            )
            .value,
        0.25,
      );
      events.add(exportProgress('saving', 0, 1));
      await tester.pump();
      expect(find.text('Writing the file…'), findsOneWidget);
      events.add(exportDone('Spanish.txt'));
      await tester.pumpAndSettle();
      expect(picker.saved.keys, ['Spanish.txt']);
      unawaited(events.close());
    });

    testWidgets('Cancel closes the dialog without exporting', (tester) async {
      final go = spanishGo();
      await pumpApp(tester);
      await openDeckMenu(tester, 'Spanish', 'Export');
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(find.byType(AlertDialog), findsNothing);
      expect(go.callsTo('transfer.ExportDeck'), isEmpty);
    });
  });

  pickerTests();
}

/// The full names the open deck picker lists, in order.
List<String> pickedNames(WidgetTester tester) => [
  for (final tile in tester.widgetList<ListTile>(
    find.descendant(
      of: find.byType(DeckPicker),
      matching: find.byType(ListTile),
    ),
  ))
    (tile.title! as Text).data!,
];

/// The full name of the deck selected in the open deck picker.
String selectedName(WidgetTester tester) =>
    (tester
                .widgetList<ListTile>(find.byType(ListTile))
                .singleWhere((t) => t.selected)
                .title!
            as Text)
        .data!;

/// Decks for the picker; Spanish verbs (5) has cards to study.
List<Map<String, Object?>> languages() => [
  deckNode(1, 'Default'),
  deckNode(2, 'Biology'),
  deckNode(3, 'Languages', hasChildren: true),
  deckNode(4, 'Languages::French', parentId: 3, hasChildren: true),
  deckNode(5, 'Languages::French::Verbs', parentId: 4),
  deckNode(6, 'Languages::Spanish', parentId: 3, hasChildren: true),
  deckNode(7, 'Languages::Spanish::Verbs', parentId: 6, newCount: 4),
];

Future<void> openPicker(WidgetTester tester) async {
  await tester.tap(find.byTooltip('Study deck'));
  await tester.pumpAndSettle();
  expect(find.byType(DeckPicker), findsOneWidget);
}

void pickerTests() {
  group('study deck picker', () {
    test('matches all words in any order, ignoring case', () {
      final decks = [for (final d in languages()) golang.DeckNode.fromJson(d)];
      List<int> ids(String query, [List<int> recent = const []]) => [
        for (final d in filterDecks(decks, recent, query)) d.id,
      ];
      expect(ids('span verb'), [7]);
      expect(ids('VERB  span'), [7]);
      expect(ids('verbs'), [5, 7]);
      expect(ids('lang span'), [6, 7]);
      expect(ids('span german'), isEmpty);
      // Recent decks first, in their order; deleted ones are skipped.
      expect(ids('verbs', [7, 99]), [7, 5]);
      expect(ids('', [2, 6]), [2, 6, 1, 3, 4, 5, 7]);
    });

    testWidgets('find a subdeck: span verb, Enter, Enter studies it (spec)', (
      tester,
    ) async {
      final go = FakeGo()..tree = languages();
      handleStudy(go);
      await pumpApp(tester, size: desktop);
      await openPicker(tester);
      // The search field has the focus.
      await tester.enterText(find.byType(TextField), 'span verb');
      await tester.pumpAndSettle();
      expect(pickedNames(tester), ['Languages::Spanish::Verbs']);

      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pumpAndSettle();
      expect(find.byType(DeckPicker), findsNothing);
      expect(find.byType(DeckOverviewScreen), findsOneWidget);
      expect(find.text('Languages::Spanish::Verbs'), findsOneWidget);

      // Study now has the focus, so Enter starts studying.
      await tester.sendKeyEvent(LogicalKeyboardKey.enter);
      await tester.pumpAndSettle();
      expect(find.byType(StudyScreen), findsOneWidget);
      expect(go.callsTo('study.Start'), [
        [7],
      ]);
    });

    testWidgets('recently studied decks are listed first (spec)', (
      tester,
    ) async {
      final go = FakeGo()..tree = languages();
      go.settings = {
        ...go.settings,
        'recentDecks': [2, 6],
      };
      await pumpApp(tester);
      await openPicker(tester);
      expect(pickedNames(tester), [
        'Biology',
        'Languages::Spanish',
        'Default',
        'Languages',
        'Languages::French',
        'Languages::French::Verbs',
        'Languages::Spanish::Verbs',
      ]);
      expect(selectedName(tester), 'Biology');
    });

    testWidgets('the arrow keys move the selection, Enter chooses it', (
      tester,
    ) async {
      final go = FakeGo()..tree = languages();
      await pumpApp(tester, size: desktop);
      await openPicker(tester);
      await tester.enterText(find.byType(TextField), 'verbs');
      await tester.pumpAndSettle();
      expect(selectedName(tester), 'Languages::French::Verbs');
      await tester.sendKeyEvent(LogicalKeyboardKey.arrowDown);
      await tester.pump();
      expect(selectedName(tester), 'Languages::Spanish::Verbs');
      // The selection stops at the ends.
      await tester.sendKeyEvent(LogicalKeyboardKey.arrowDown);
      await tester.pump();
      expect(selectedName(tester), 'Languages::Spanish::Verbs');
      await tester.sendKeyEvent(LogicalKeyboardKey.arrowUp);
      await tester.sendKeyEvent(LogicalKeyboardKey.arrowUp);
      await tester.pump();
      expect(selectedName(tester), 'Languages::French::Verbs');
      await tester.sendKeyEvent(LogicalKeyboardKey.arrowDown);
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pumpAndSettle();
      expect(find.byType(DeckOverviewScreen), findsOneWidget);
      expect(go.callsTo('decks.Overview'), [
        [7],
      ]);
    });

    testWidgets('a tap chooses a deck; Cancel and no match choose none', (
      tester,
    ) async {
      final go = FakeGo()..tree = languages();
      await pumpApp(tester);
      await openPicker(tester);
      await tester.enterText(find.byType(TextField), 'german');
      await tester.pumpAndSettle();
      expect(find.text('No matching decks'), findsOneWidget);
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pumpAndSettle();
      expect(find.byType(DeckPicker), findsOneWidget);
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(find.byType(DeckPicker), findsNothing);
      expect(find.byType(DeckOverviewScreen), findsNothing);

      await openPicker(tester);
      await tester.tap(find.byKey(const ValueKey('pick-2')));
      await tester.pumpAndSettle();
      expect(find.byType(DeckOverviewScreen), findsOneWidget);
      expect(go.callsTo('decks.Overview'), [
        [2],
      ]);
    });
  });

  group('sync indicator', () {
    const home = {'id': 'wifi:Home', 'wifi': 'Home'};
    Map<String, Object?> linked({
      bool trusted = true,
      bool needsChoice = false,
      DateTime? lastSynced,
      bool syncing = false,
      Map<String, Object?>? progress,
      String? error,
    }) => syncStatusJson(
      trusted: trusted ? [home] : const [],
      paired: [pairedDeviceJson('PHONEDEVICEID001', 'Phone')],
      links: [
        linkInfoJson(
          'link-1',
          'PHONEDEVICEID001',
          'Phone',
          needsChoice: needsChoice,
          lastSynced: lastSynced,
        ),
      ],
      syncing: syncing,
      progress: progress,
      lastSynced: lastSynced,
      error: error,
    );
    final indicator = find.byKey(const Key('sync-indicator'));

    testWidgets('appears only with a link and shows the state', (tester) async {
      final go = FakeGo()..syncStatus = syncStatusJson(trusted: [home]);
      await pumpApp(tester);
      expect(indicator, findsNothing);

      go.setSyncStatus(linked());
      await tester.pumpAndSettle();
      expect(indicator, findsOneWidget);
      expect(find.byTooltip('Not synced yet'), findsOneWidget);

      final synced = DateTime.now().subtract(const Duration(minutes: 5));
      go.setSyncStatus(linked(lastSynced: synced));
      await tester.pumpAndSettle();
      expect(find.byTooltip('Synced ${formatRecent(synced)}'), findsOneWidget);

      go.setSyncStatus(
        linked(
          syncing: true,
          progress: {'phase': 'receiving', 'received': 3, 'sent': 0},
        ),
      );
      await tester.pumpAndSettle();
      expect(find.byTooltip('Syncing: 3 received, 0 sent'), findsOneWidget);

      go.setSyncStatus(linked(needsChoice: true));
      await tester.pumpAndSettle();
      expect(find.byTooltip('Sync needs your choice'), findsOneWidget);

      go.setSyncStatus(linked(trusted: false));
      await tester.pumpAndSettle();
      expect(find.byTooltip('Not on a trusted network'), findsOneWidget);
    });

    testWidgets('a failed automatic sync shows no dialog, only the reason '
        'when tapped', (tester) async {
      final go = FakeGo()..syncStatus = syncStatusJson(trusted: [home]);
      await pumpApp(tester);
      go.setSyncStatus(linked(error: 'The connection to Phone was lost.'));
      await tester.pumpAndSettle();
      expect(find.byType(AlertDialog), findsNothing);
      expect(find.byTooltip('Sync failed'), findsOneWidget);

      await tester.tap(indicator);
      await tester.pumpAndSettle();
      expect(find.text('The connection to Phone was lost.'), findsOneWidget);
      go.syncNowResult = {
        'received': 1,
        'sent': 0,
        'media': 0,
        'reachable': true,
      };
      await tester.tap(find.text('Sync again'));
      await tester.pumpAndSettle();
      expect(go.syncCalls, contains('SyncNow'));
      expect(find.text('Synced: 1 received, 0 sent'), findsOneWidget);
    });

    testWidgets('Sync now reports what was received and sent', (tester) async {
      final go = FakeGo()..syncStatus = linked();
      await pumpApp(tester);
      go.syncNowResult = {
        'received': 3,
        'sent': 2,
        'media': 0,
        'reachable': true,
      };
      await tester.tap(indicator);
      await tester.pumpAndSettle();
      expect(go.syncCalls.where((c) => c == 'SyncNow'), hasLength(1));
      expect(find.text('Synced: 3 received, 2 sent'), findsOneWidget);

      go.syncNowResult = {
        'received': 0,
        'sent': 0,
        'media': 0,
        'reachable': false,
      };
      await tester.tap(indicator);
      await tester.pumpAndSettle();
      expect(
        find.text('No linked device is reachable on this network.'),
        findsOneWidget,
      );
    });

    testWidgets('the label is shown beside the icon on wide screens', (
      tester,
    ) async {
      FakeGo().syncStatus = linked();
      await pumpApp(tester, size: desktop);
      expect(
        find.descendant(of: indicator, matching: find.text('Not synced yet')),
        findsOneWidget,
      );
    });

    testWidgets('the first sync asks to merge or replace', (tester) async {
      final go = FakeGo()..syncStatus = linked(needsChoice: true);
      await pumpApp(tester);
      final link = linkInfoJson(
        'link-1',
        'PHONEDEVICEID001',
        'Phone',
        remoteName: 'Anna',
        needsChoice: true,
      );
      var choice = go.chooseFirstSync(link);
      await tester.pumpAndSettle();
      expect(find.text('Sync with Phone'), findsOneWidget);
      expect(find.text('Replace the profile on Phone'), findsOneWidget);
      expect(
        find.text(
          'This device gets exactly the cards and history of “Anna” on Phone.',
        ),
        findsOneWidget,
      );
      await tester.tap(find.byKey(const Key('first-sync-replaceOther')));
      await tester.pumpAndSettle();
      expect(await choice, 'replaceOther');

      choice = go.chooseFirstSync(link);
      await tester.pumpAndSettle();
      await tester.tap(find.text('Later'));
      await tester.pumpAndSettle();
      expect(await choice, '');

      choice = go.chooseFirstSync(link);
      await tester.pumpAndSettle();
      await tester.tap(find.text('Merge'));
      await tester.pumpAndSettle();
      expect(await choice, 'merge');
    });
  });

  group('progress header', () {
    final header = find.byKey(const Key('progress-header'));

    Future<FakeGo> pumpHeader(
      WidgetTester tester,
      Map<String, Object?>? progress, {
      bool showProgress = true,
      bool minimalist = false,
    }) async {
      final go = FakeGo()
        ..tree = [deckNode(1, 'Spanish', newCount: 3)]
        ..progress = progress;
      go.settings = {
        ...go.settings,
        'showProgress': showProgress,
        'minimalist': minimalist,
      };
      await pumpApp(tester);
      return go;
    }

    String headerText(WidgetTester tester) => tester
        .widgetList<Text>(
          find.descendant(of: header, matching: find.byType(Text)),
        )
        .map((t) => t.data)
        .join(' | ');

    testWidgets('no goal: only the streak', (tester) async {
      // No goal (spec).
      await pumpHeader(tester, progressJson(current: 12, cards: 4));
      expect(headerText(tester), '12 days');
      expect(find.byKey(const Key('goal-ring')), findsNothing);
      // Outside minimalist mode the row has a quiet fill.
      expect(tester.widget<Material>(header).color, isNot(Colors.transparent));
    });

    testWidgets('goal in progress', (tester) async {
      // Goal in progress (spec): 12 days and 18 / 30 cards.
      await pumpHeader(
        tester,
        progressJson(current: 12, cards: 18, goalKind: 'cards', goal: 30),
      );
      expect(headerText(tester), '12 days | 18 / 30 cards');
      final ring = tester.widget<CircularProgressIndicator>(
        find.byKey(const Key('goal-ring')),
      );
      expect(ring.value, closeTo(0.6, 1e-9));
    });

    testWidgets('goal reached', (tester) async {
      await pumpHeader(
        tester,
        progressJson(
          current: 3,
          minutes: 25,
          cards: 40,
          goalKind: 'minutes',
          goal: 20,
        ),
      );
      expect(headerText(tester), '3 days | 25 / 20 minutes');
      expect(
        tester
            .widget<CircularProgressIndicator>(
              find.byKey(const Key('goal-ring')),
            )
            .value,
        1,
      );
    });

    testWidgets('at risk', (tester) async {
      // At risk (spec), without warning colours.
      await pumpHeader(tester, progressJson(current: 12));
      expect(headerText(tester), 'Study today to keep your 12-day streak');
      final theme = Theme.of(tester.element(header));
      final text = tester.widget<Text>(
        find.descendant(of: header, matching: find.byType(Text)),
      );
      expect(text.style?.color, isNot(theme.colorScheme.error));
      final icon = tester.widget<Icon>(
        find.descendant(of: header, matching: find.byType(Icon)),
      );
      expect(icon.color, isNot(theme.colorScheme.error));
    });

    testWidgets('no streak', (tester) async {
      // No streak (spec): 0 days, nothing about losing it.
      await pumpHeader(tester, progressJson());
      expect(headerText(tester), '0 days');
    });

    testWidgets('hidden', (tester) async {
      // Hidden (spec): no streak row, and the history is not read.
      final go = await pumpHeader(
        tester,
        progressJson(current: 12),
        showProgress: false,
      );
      expect(header, findsNothing);
      expect(go.callsTo('progress.Today'), isEmpty);
      expect(find.text('Spanish'), findsOneWidget);
    });

    testWidgets('minimalist: no shadow or fill', (tester) async {
      // Minimalist (spec).
      await pumpHeader(tester, progressJson(current: 12), minimalist: true);
      final m = tester.widget<Material>(header);
      expect(m.elevation, 0);
      expect(m.color, Colors.transparent);
    });

    testWidgets('updates after answers', (tester) async {
      // After a session (spec): the header follows the collection.
      final go = await pumpHeader(
        tester,
        progressJson(current: 12, goalKind: 'cards', goal: 30),
      );
      go.progress = progressJson(
        current: 13,
        cards: 12,
        goalKind: 'cards',
        goal: 30,
      );
      go.emitChange(kind: 'answer');
      await tester.pumpAndSettle();
      expect(headerText(tester), '13 days | 12 / 30 cards');
    });
  });
}
