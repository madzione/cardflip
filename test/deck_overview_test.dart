import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/features/decks/deck_options_screen.dart';
import 'package:cardflip/features/decks/deck_overview_screen.dart';
import 'package:cardflip/features/decks/deck_progress.dart';
import 'package:cardflip/features/stats/stats_charts.dart';
import 'package:cardflip/features/study/study_screen.dart';

import 'support/fake_go.dart';

/// The style of the text span reading [text] in the description.
TextStyle? spanStyle(WidgetTester tester, String text) {
  TextStyle? found;
  for (final rich in tester.widgetList<RichText>(
    find.descendant(
      of: find.byKey(const Key('deck-description')),
      matching: find.byType(RichText),
    ),
  )) {
    rich.text.visitChildren((span) {
      if (span is TextSpan && span.text == text) found = span.style;
      return found == null;
    });
  }
  return found;
}

Future<void> openOverview(WidgetTester tester, String deck) async {
  await tester.tap(find.text(deck));
  await tester.pumpAndSettle();
  expect(find.byType(DeckOverviewScreen), findsOneWidget);
}

String countText(WidgetTester tester, String key) =>
    tester.widget<Text>(find.byKey(Key(key))).data!;

/// Answers `study.Start` with a session that has nothing due.
void handleStudy(FakeGo go) => go.handle(
  'study.Start',
  (args) => {
    'deckId': args[0],
    'deckName': 'Spanish',
    'card': null,
    'counts': {'new': 0, 'learning': 0, 'review': 0},
    'canUndo': false,
    'studied': 0,
    'timeSpent': 0,
    'nextDue': null,
    'learningDue': null,
    'leech': false,
    'leechSuspended': false,
  },
);

/// The flex and colour of each part of the deck progress bar.
List<(int, Color)> barParts(WidgetTester tester) => [
  for (final e in tester.widgetList<Expanded>(
    find.descendant(
      of: find.byType(DeckProgressBar),
      matching: find.byType(Expanded),
    ),
  ))
    (e.flex, (e.child as ColoredBox).color),
];

void main() {
  testWidgets('overview before studying', (tester) async {
    final go = FakeGo();
    go.tree = [
      deckNode(1, 'Default'),
      deckNode(2, 'Spanish', newCount: 5, review: 30),
    ];
    handleStudy(go);
    await pumpApp(tester);
    await openOverview(tester, 'Spanish');
    // Overview before studying (app-shell spec): the counts and Study now,
    // and no session has started.
    expect(countText(tester, 'overview-new'), '5');
    expect(countText(tester, 'overview-learning'), '0');
    expect(countText(tester, 'overview-review'), '30');
    expect(find.text('Study now'), findsOneWidget);
    expect(find.byKey(const Key('buried')), findsNothing);
    expect(go.callsTo('study.Start'), isEmpty);
    expect(go.callsTo('decks.Overview'), [
      [2],
    ]);

    // Study from the overview (study-session spec).
    await tester.tap(find.text('Study now'));
    await tester.pumpAndSettle();
    expect(find.byType(StudyScreen), findsOneWidget);
    expect(go.callsTo('study.Start'), [
      [2],
    ]);
  });

  testWidgets('nothing due says when the next card is due', (tester) async {
    final go = FakeGo();
    go.tree = [deckNode(1, 'Default', cards: 3)];
    final now = DateTime.now();
    go.nextDue[1] = DateTime(now.year, now.month, now.day + 1, 4);
    await pumpApp(tester);
    await openOverview(tester, 'Default');
    expect(find.text('Nothing is due right now.'), findsOneWidget);
    expect(find.text('The next card is due tomorrow.'), findsOneWidget);
    final study = tester.widget<ButtonStyleButton>(
      find.ancestor(
        of: find.text('Study now'),
        matching: find.bySubtype<ButtonStyleButton>(),
      ),
    );
    expect(study.onPressed, isNull);
    expect(go.callsTo('study.Start'), isEmpty);
  });

  testWidgets('a deck without cards to study says so', (tester) async {
    FakeGo();
    await pumpApp(tester);
    await openOverview(tester, 'Default');
    expect(
      find.text(
        'This deck has no cards to study. Add cards or import some to get started.',
      ),
      findsOneWidget,
    );
  });

  testWidgets('unbury makes buried cards due again', (tester) async {
    final go = FakeGo();
    go.tree = [deckNode(1, 'Default', review: 2)];
    go.buried[1] = 3;
    await pumpApp(tester);
    await openOverview(tester, 'Default');
    expect(find.text('3 buried cards'), findsOneWidget);

    await tester.tap(find.text('Unbury'));
    await tester.pumpAndSettle();
    expect(go.callsTo('decks.Unbury'), [
      [1],
    ]);
    // Unbury manually buried cards (collection spec): the cards are due
    // again and the buried count is gone.
    expect(find.byKey(const Key('buried')), findsNothing);
    expect(countText(tester, 'overview-review'), '5');
    expect(find.text('Unburied “Default”'), findsOneWidget);
  });

  testWidgets('description is shown as HTML and can be edited', (tester) async {
    final go = FakeGo();
    go.tree = [deckNode(1, 'Default', newCount: 1)];
    await pumpApp(tester);
    await openOverview(tester, 'Default');
    expect(find.byKey(const Key('deck-description')), findsNothing);

    // Edit description (collection spec).
    await tester.tap(find.byTooltip('Edit description'));
    await tester.pumpAndSettle();
    await tester.enterText(
      find.byType(TextField),
      'Vocabulary from <b>Book 1</b>',
    );
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    expect(go.callsTo('decks.SetDescription'), [
      [1, 'Vocabulary from <b>Book 1</b>'],
    ]);
    expect(spanStyle(tester, 'Book 1')?.fontWeight, FontWeight.bold);
    expect(spanStyle(tester, 'Vocabulary from '), isNotNull);

    // The editor starts with the current description; Cancel keeps it.
    await tester.tap(find.byTooltip('Edit description'));
    await tester.pumpAndSettle();
    expect(find.text('Vocabulary from <b>Book 1</b>'), findsOneWidget);
    await tester.tap(find.text('Cancel'));
    await tester.pumpAndSettle();
    expect(go.callsTo('decks.SetDescription'), hasLength(1));
  });

  testWidgets('links in the description open in the system browser', (
    tester,
  ) async {
    final opened = <Uri>[];
    final open = DeckOverviewScreen.openUrl;
    DeckOverviewScreen.openUrl = (url) async {
      opened.add(url);
      return true;
    };
    addTearDown(() => DeckOverviewScreen.openUrl = open);
    final go = FakeGo();
    go.descriptions[1] = '<a href="https://example.com/deck">Deck home</a>';
    await pumpApp(tester);
    await openOverview(tester, 'Default');
    await tester.tapOnText(find.textRange.ofSubstring('Deck home'));
    await tester.pumpAndSettle();
    expect(opened, [Uri.parse('https://example.com/deck')]);
  });

  testWidgets('options and the overview follow changes', (tester) async {
    final go = FakeGo();
    go.tree = [deckNode(1, 'Default', newCount: 2)];
    await pumpApp(tester);
    await openOverview(tester, 'Default');
    await tester.tap(find.text('Options'));
    await tester.pumpAndSettle();
    expect(find.byType(DeckOptionsScreen), findsOneWidget);
    await tester.pageBack();
    await tester.pumpAndSettle();

    // A change elsewhere (e.g. a rename) reloads the overview.
    go.tree = [deckNode(1, 'Main', newCount: 4)];
    go.emitChange();
    await tester.pumpAndSettle();
    expect(find.widgetWithText(AppBar, 'Main'), findsOneWidget);
    expect(countText(tester, 'overview-new'), '4');

    // Deleting the deck closes its overview.
    go.tree = [deckNode(3, 'Other')];
    go.emitChange();
    await tester.pumpAndSettle();
    expect(find.byType(DeckOverviewScreen), findsNothing);
    expect(find.text('Other'), findsOneWidget);
  });

  group('deck progress', () {
    const spanish = {'new': 18, 'learning': 4, 'young': 12, 'mature': 56};

    testWidgets('shows the parts and the mature share', (tester) async {
      final go = FakeGo();
      go.tree = [
        deckNode(1, 'Default'),
        deckNode(2, 'Spanish', cards: 100, review: 3, progress: spanish),
      ];
      await pumpApp(tester);
      await openOverview(tester, 'Spanish');
      // Deck progress (app-shell spec): of 90 cards that are not
      // suspended, 18 new, 4 learning, 12 young and 56 mature.
      expect(find.text('62% mature'), findsOneWidget);
      for (final label in [
        'New: 18',
        'Learning: 4',
        'Young: 12',
        'Mature: 56',
      ]) {
        expect(find.text(label), findsOneWidget);
      }
      const colors = StatsColors.light;
      expect(barParts(tester), [
        (18, colors.newCards),
        (4, colors.learning),
        (12, colors.young),
        (56, colors.mature),
      ]);
    });

    testWidgets('an empty deck has an empty bar', (tester) async {
      FakeGo().tree = [deckNode(1, 'Default')];
      await pumpApp(tester);
      await openOverview(tester, 'Default');
      expect(find.text('0% mature'), findsOneWidget);
      expect(find.byType(DeckProgressBar), findsOneWidget);
      expect(barParts(tester), isEmpty);
    });

    testWidgets('uses the dark statistics colours', (tester) async {
      final go = FakeGo();
      go.settings = {...go.settings, 'theme': 'dark'};
      go.tree = [deckNode(1, 'Default', cards: 90, progress: spanish)];
      await pumpApp(tester);
      await openOverview(tester, 'Default');
      const colors = StatsColors.dark;
      expect(barParts(tester).map((p) => p.$2), [
        colors.newCards,
        colors.learning,
        colors.young,
        colors.mature,
      ]);
    });

    testWidgets('follows the deck list', (tester) async {
      final go = FakeGo();
      go.tree = [
        deckNode(1, 'Default', cards: 2, progress: {'new': 2}),
      ];
      await pumpApp(tester);
      await openOverview(tester, 'Default');
      expect(find.text('0% mature'), findsOneWidget);
      go.tree = [
        deckNode(1, 'Default', cards: 2, progress: {'young': 1, 'mature': 1}),
      ];
      go.emitChange();
      await tester.pumpAndSettle();
      expect(find.text('50% mature'), findsOneWidget);
    });
  });
}
