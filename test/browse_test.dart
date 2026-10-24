import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/app/theme.dart';
import 'package:cardflip/features/browse/browse_sidebar.dart';
import 'package:cardflip/features/editor/field_drop_target.dart';
import 'package:cardflip/features/editor/occlusion_editor.dart';
import 'package:cardflip/features/editor/rich_field.dart';
import 'package:cardflip/widgets/time_format.dart';

import 'support/fake_go.dart';
import 'support/fake_media.dart';
import 'support/occlusion.dart';
import '../integration_test/rich_fields.dart';

Map<String, Object?> row(
  int id,
  String question, {
  String deck = 'Default',
  String state = 'new',
  bool suspended = false,
  int flag = 0,
  bool marked = false,
  int? noteId,
  bool cloze = false,
  String? noteType,
  int? position,
  String answer = '',
  int? interval,
  int lapses = 0,
  String tags = '',
  String? sortField,
}) => {
  'cardId': id,
  'noteId': noteId ?? id + 1000,
  'question': question,
  'answer': answer,
  'sortField': sortField ?? question,
  'deck': deck,
  'noteType': noteType ?? (cloze ? 'Cloze' : 'Basic'),
  'cardType': 'Card 1',
  'state': state,
  'due': null,
  'position': position,
  'interval': interval,
  'difficulty': null,
  'stability': null,
  'retrievability': null,
  'reviews': lapses,
  'lapses': lapses,
  'tags': tags,
  'created': '2026-10-01T10:00:00Z',
  'cardModified': '2026-10-02T10:00:00Z',
  'noteModified': '2026-10-03T10:00:00Z',
  'suspended': suspended,
  'flag': flag,
  'marked': marked,
};

/// The browser's Go side over a list of rows. Suspending and deleting change
/// the rows and notify the app.
class FakeBrowse {
  FakeBrowse(this.go, this.rows) {
    go.handle('browse.Search', (args) {
      final q = args[0] as String;
      if (q.contains('is:bogus')) {
        throw 'is:bogus: unknown state; use is:new, is:due, is:review, is:learn or is:suspended';
      }
      if ('('.allMatches(q).length > ')'.allMatches(q).length) {
        throw 'a parenthesis is not closed (character ${q.indexOf('(') + 1})';
      }
      final found = [
        for (final r in rows)
          if (_matches(r, q)) r,
      ];
      if (args[1] == 'lapses') {
        found.sort(
          (a, b) => (a['lapses'] as int).compareTo(b['lapses'] as int),
        );
      }
      if (args[1] == 'sortField') {
        found.sort(
          (a, b) => (a['sortField'] as String).toLowerCase().compareTo(
            (b['sortField'] as String).toLowerCase(),
          ),
        );
      }
      return [
        for (final r in args[2] as bool ? found.reversed : found) r['cardId'],
      ];
    });
    go.handle('browse.SearchNotes', (args) {
      final q = args[0] as String;
      final notes = <Object?>[];
      for (final r in rows) {
        final match = q == 'is:suspended'
            ? r['suspended'] == true
            : q.isEmpty || (r['question'] as String).contains(q);
        if (match && !notes.contains(r['noteId'])) notes.add(r['noteId']);
      }
      return args[2] as bool ? notes.reversed.toList() : notes;
    });
    go.handle('browse.NoteRows', (args) {
      final ids = (args[0] as List).cast<int>();
      pages.add(ids.length);
      return [
        for (final id in ids)
          if (_cardsOfNote(id) case [final first, ...] && final cards)
            {
              'noteId': id,
              'cardId': first['cardId'],
              'sortField': first['sortField'],
              'noteType': first['noteType'],
              'cards': cards.length,
              'deck': first['deck'],
              'multipleDecks': cards.any((r) => r['deck'] != first['deck']),
              'tags': first['tags'],
              'state': first['state'],
              'due': null,
              'position': first['position'],
              'suspended': cards.every((r) => r['suspended'] == true),
              'marked': first['marked'],
              'created': first['created'],
            },
      ];
    });
    go.handle('browse.NoteCards', (args) {
      final notes = (args[0] as List).toSet();
      return [
        for (final r in rows)
          if (notes.contains(r['noteId'])) r['cardId'],
      ];
    });
    go.handle('browse.CardsOfNotes', (args) {
      final cards = (args[0] as List).toSet();
      final notes = {
        for (final r in rows)
          if (cards.contains(r['cardId'])) r['noteId'],
      };
      return [
        for (final r in rows)
          if (notes.contains(r['noteId'])) r['cardId'],
      ];
    });
    go.handle('browse.Rows', (args) {
      final ids = (args[0] as List).cast<int>();
      pages.add(ids.length);
      return [for (final id in ids) rows.firstWhere((r) => r['cardId'] == id)];
    });
    go.handle('browse.Card', (args) {
      final r = rows.firstWhere((r) => r['cardId'] == args[0]);
      final cloze = r['noteType'] == 'Cloze';
      final cardIds = [
        for (final o in rows)
          if (o['noteId'] == r['noteId']) o['cardId'],
      ];
      return {
        'cardId': r['cardId'],
        'deckId': 1,
        'deck': r['deck'],
        'cardName': 'Card 1',
        'question': '<div>${r['question']}</div>',
        'answer': '<div>${r['question']}<hr>answer</div>',
        'questionMedia': r['questionMedia'] ?? const [],
        'questionTts': r['questionTts'] ?? const [],
        'state': r['state'],
        'due': null,
        'position': r['position'],
        'intervalDays': 0,
        'stability': null,
        'difficulty': null,
        'reviews': 0,
        'lapses': 0,
        'suspended': r['suspended'],
        'flag': r['flag'],
        'note': {
          'id': r['noteId'],
          'typeId': cloze ? 12 : 10,
          'fields': [r['question'], cloze ? '' : 'answer'],
          'tags': <String>[],
          'cardIds': cardIds,
        },
        'noteType': cloze
            ? noteTypeJson(12, 'Cloze', ['Text', 'Back Extra'], cloze: true)
            : noteTypeJson(10, 'Basic', ['Front', 'Back']),
        'noteCards': cardIds.length,
      };
    });
    go.handle('browse.SetSuspended', (args) {
      final ids = (args[0] as List).toSet();
      for (var i = 0; i < rows.length; i++) {
        if (ids.contains(rows[i]['cardId'])) {
          rows[i] = {...rows[i], 'suspended': args[1]};
        }
      }
      go.emitChange();
      return null;
    });
    go.handle('browse.SetFlag', (args) {
      final ids = (args[0] as List).toSet();
      var flag = args[1] as int;
      final picked = rows.where((r) => ids.contains(r['cardId']));
      if (picked.every((r) => r['flag'] == flag)) flag = 0;
      for (var i = 0; i < rows.length; i++) {
        if (ids.contains(rows[i]['cardId'])) {
          rows[i] = {...rows[i], 'flag': flag};
        }
      }
      go.emitChange();
      return null;
    });
    go.handle('browse.ToggleMark', (args) {
      final notes = _notesOf(args[0] as List);
      final marked = rows
          .where((r) => notes.contains(r['noteId']))
          .any((r) => r['marked'] != true);
      for (var i = 0; i < rows.length; i++) {
        if (notes.contains(rows[i]['noteId'])) {
          rows[i] = {...rows[i], 'marked': marked};
        }
      }
      go.emitChange();
      return null;
    });
    go.handle('browse.DeleteImpact', (args) {
      final notes = _notesOf(args[0] as List);
      return {
        'notes': notes.length,
        'cards': rows.where((r) => notes.contains(r['noteId'])).length,
      };
    });
    go.handle('browse.DeleteNotes', (args) {
      final notes = _notesOf(args[0] as List);
      rows.removeWhere((r) => notes.contains(r['noteId']));
      go.emitChange();
      return null;
    });
    _findReplace();
  }

  /// Find and Replace over the questions of the rows' notes: the fake
  /// changes nothing but counts the notes whose question matches.
  void _findReplace() {
    int count(List<dynamic> args) {
      final opts = args[2] as Map;
      final find = opts['find'] as String;
      final RegExp re;
      try {
        re = RegExp(
          opts['regex'] as bool ? find : RegExp.escape(find),
          caseSensitive: !(opts['ignoreCase'] as bool),
        );
      } on FormatException {
        throw 'Invalid regular expression: missing closing ): `$find`';
      }
      final notes = args[1] as bool
          ? {for (final r in rows) r['noteId']}
          : _notesOf(args[0] as List);
      return {
        for (final r in rows)
          if (notes.contains(r['noteId']) &&
              re.hasMatch(r['question'] as String))
            r['noteId'],
      }.length;
    }

    go.handle('browse.FieldNames', (_) => ['Front', 'Back']);
    go.handle('browse.FindReplacePreview', count);
    go.handle('browse.FindReplace', (args) {
      final n = count(args);
      go.emitChange();
      return n;
    });
  }

  /// Whether row [r] matches [q]: all of its `deck:`, `tag:` and `is:due`
  /// terms when it has only those, else its text in the question.
  static bool _matches(Map<String, Object?> r, String q) {
    final terms = q.split(' ').where((t) => t.isNotEmpty).toList();
    bool term(String t) => switch (t.split(':')) {
      ['deck', final deck] => r['deck'] == deck,
      ['tag', final tag] => (r['tags'] as String).split(' ').contains(tag),
      ['is', 'due'] => r['state'] == 'review',
      _ => false,
    };
    final qualified = RegExp(r'^(deck|tag):\S+$|^is:due$');
    if (terms.isNotEmpty && terms.every(qualified.hasMatch)) {
      return terms.every(term);
    }
    return q.isEmpty || (r['question'] as String).contains(q);
  }

  List<Map<String, Object?>> _cardsOfNote(int note) => [
    for (final r in rows)
      if (r['noteId'] == note) r,
  ];

  Set<Object?> _notesOf(List<dynamic> cardIds) => {
    for (final r in rows)
      if (cardIds.contains(r['cardId'])) r['noteId'],
  };

  final FakeGo go;
  final List<Map<String, Object?>> rows;
  final pages = <int>[];
}

/// The text of card [id]'s [column] in the wide table.
String cell(WidgetTester tester, int id, String column) =>
    tester.widget<Text>(find.byKey(Key('cell-$id-$column'))).data!;

Future<void> openBrowse(WidgetTester tester, {Size size = phone}) async {
  await pumpApp(tester, size: size);
  await tester.tap(find.text('Browse').last);
  await tester.pumpAndSettle();
}

void main() {
  group('shortcut hints', () {
    Future<void> select(WidgetTester tester) async {
      FakeBrowse(FakeGo(), [row(1, 'uno'), row(2, 'dos')]);
      await openBrowse(tester);
      await tester.longPress(find.text('uno'));
      await tester.pumpAndSettle();
    }

    Future<void> openMenu(WidgetTester tester, String tooltip) async {
      await tester.tap(find.byTooltip(tooltip));
      await tester.pumpAndSettle();
    }

    testWidgets('mobile browser (spec)', (tester) async {
      await select(tester);
      await openMenu(tester, 'Bulk actions');
      expect(find.text('Set due date'), findsOneWidget);
      expect(find.textContaining('⇧⌘D'), findsNothing);
      expect(find.textContaining('Ctrl+Shift+D'), findsNothing);
      await tester.tapAt(Offset.zero);
      await tester.pumpAndSettle();
      await openMenu(tester, 'Selection');
      expect(find.text('Invert selection'), findsOneWidget);
      expect(find.text('⌘⌥S'), findsNothing);
      expect(find.text('Ctrl+Alt+S'), findsNothing);
    }, variant: TargetPlatformVariant.only(TargetPlatform.iOS));

    testWidgets('the browser menus show the shortcuts on Linux', (
      tester,
    ) async {
      await select(tester);
      await openMenu(tester, 'Bulk actions');
      expect(find.text('Set due date (Ctrl+Shift+D)'), findsOneWidget);
      await tester.tapAt(Offset.zero);
      await tester.pumpAndSettle();
      await openMenu(tester, 'Selection');
      expect(find.text('Invert selection'), findsOneWidget);
      expect(find.text('Ctrl+Alt+S'), findsOneWidget);
    }, variant: TargetPlatformVariant.only(TargetPlatform.linux));
  });

  testWidgets('search lists matching cards; an invalid term shows an error', (
    tester,
  ) async {
    final go = FakeGo();
    FakeBrowse(go, [row(1, 'comer'), row(2, 'perro'), row(3, 'comida')]);
    await openBrowse(tester);
    expect(find.text('3 cards'), findsOneWidget);
    await tester.enterText(find.byKey(const Key('search')), 'com');
    await tester.testTextInput.receiveAction(TextInputAction.done);
    await tester.pumpAndSettle();
    expect(find.text('comer'), findsOneWidget);
    expect(find.text('comida'), findsOneWidget);
    expect(find.text('perro'), findsNothing);
    expect(go.callsTo('browse.Search').last, ['com', 'created', false]);

    await tester.enterText(find.byKey(const Key('search')), 'verb is:bogus');
    await tester.testTextInput.receiveAction(TextInputAction.done);
    await tester.pumpAndSettle();
    expect(find.textContaining('is:bogus: unknown state'), findsOneWidget);
  });

  testWidgets('an unclosed group shows an error instead of a list', (
    tester,
  ) async {
    final go = FakeGo();
    FakeBrowse(go, [row(1, 'comer'), row(2, 'manger')]);
    await openBrowse(tester);
    expect(find.text('2 cards'), findsOneWidget);

    await tester.enterText(
      find.byKey(const Key('search')),
      '(deck:Spanish or deck:French',
    );
    await tester.testTextInput.receiveAction(TextInputAction.done);
    await tester.pumpAndSettle();
    expect(
      find.text('a parenthesis is not closed (character 1)'),
      findsOneWidget,
    );
    expect(find.text('2 cards'), findsNothing);
  });

  testWidgets('rows load lazily in pages of 100', (tester) async {
    final go = FakeGo();
    final browse = FakeBrowse(go, [
      for (var i = 1; i <= 250; i++) row(i, 'card $i'),
    ]);
    await openBrowse(tester);
    expect(find.text('250 cards'), findsOneWidget);
    expect(browse.pages, [100]);
    await tester.fling(
      find.byKey(const Key('browse-list')),
      const Offset(0, -20000),
      5000,
    );
    await tester.pumpAndSettle();
    expect(browse.pages.every((n) => n <= 100), isTrue);
    expect(browse.pages.length, greaterThan(1));
    expect(find.text('card 250'), findsOneWidget);
  });

  testWidgets(
    'edit from the browser: the row and preview show the new content',
    (tester) async {
      final go = FakeGo();
      final rows = [row(1, 'comer'), row(2, 'perro')];
      FakeBrowse(go, rows);
      go.handle('notes.Update', (args) {
        rows[0] = row(1, (args[1] as List).first as String);
        go.emitChange();
        return null;
      });
      await openBrowse(tester, size: desktop);
      await tester.tap(find.text('comer'));
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('card-details')), findsOneWidget);
      await enterField(tester, 0, 'comer (to eat)', prefix: 'details-field');
      await tester.dragUntilVisible(
        find.text('Save'),
        find.byKey(const Key('card-details')),
        const Offset(0, -200),
      );
      await tester.ensureVisible(find.text('Save'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Save'));
      await tester.pumpAndSettle();
      expect(find.text('Card saved'), findsOneWidget);
      expect(go.callsTo('notes.Update').single[1], [
        'comer (to eat)',
        'answer',
      ]);
      expect(find.text('comer (to eat)'), findsWidgets);
      expect(
        find.textContaining('comer (to eat)', findRichText: true),
        findsWidgets,
      );
    },
  );

  testWidgets('bulk move: select cards and move them to a deck', (
    tester,
  ) async {
    final go = FakeGo();
    go.tree = [deckNode(1, 'Default'), deckNode(2, 'Archive')];
    FakeBrowse(go, [row(1, 'uno'), row(2, 'dos'), row(3, 'tres')]);
    go.handle('browse.Move', (_) => null);
    await openBrowse(tester);
    await tester.longPress(find.text('uno'));
    await tester.pumpAndSettle();
    expect(find.text('1 selected'), findsOneWidget);
    await tester.tap(find.text('tres'));
    await tester.pumpAndSettle();
    expect(find.text('2 selected'), findsOneWidget);
    await tester.tap(find.byTooltip('Move to deck'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Archive'));
    await tester.pumpAndSettle();
    final call = go.callsTo('browse.Move').single;
    expect((call[0] as List).toSet(), {1, 3});
    expect(call[1], 2);
  });

  testWidgets('Undo from the snackbar: moved cards go back', (tester) async {
    final go = FakeGo();
    go.tree = [deckNode(1, 'Default'), deckNode(2, 'Archive')];
    final rows = [for (var i = 1; i <= 30; i++) row(i, 'card $i')];
    FakeBrowse(go, rows);
    go.handle('browse.Move', (args) {
      final ids = (args[0] as List).toSet();
      for (var i = 0; i < rows.length; i++) {
        if (ids.contains(rows[i]['cardId'])) {
          rows[i] = {...rows[i], 'deck': 'Archive'};
        }
      }
      go.undoStep = 'Move ${ids.length} cards';
      go.emitChange();
      return null;
    });
    go.onUndo = () {
      for (var i = 0; i < rows.length; i++) {
        rows[i] = {...rows[i], 'deck': 'Default'};
      }
      go.emitChange();
    };
    await openBrowse(tester, size: desktop);
    await tester.longPress(find.text('card 1'));
    await tester.pumpAndSettle();
    await tester.tap(find.byTooltip('Select all'));
    await tester.pumpAndSettle();
    await tester.tap(find.byTooltip('Move to deck'));
    await tester.pumpAndSettle();
    // The sidebar lists the deck too.
    await tester.tap(
      find.descendant(
        of: find.byType(SimpleDialog),
        matching: find.text('Archive'),
      ),
    );
    await tester.pumpAndSettle();
    expect(rows.every((r) => r['deck'] == 'Archive'), isTrue);
    expect(find.text('Moved 30 cards'), findsOneWidget);
    await tester.tap(find.widgetWithText(SnackBarAction, 'Undo'));
    await tester.pumpAndSettle();
    expect(go.callsTo('app.Undo'), hasLength(1));
    expect(rows.every((r) => r['deck'] == 'Default'), isTrue);
    expect(find.text('Undone: Move 30 cards'), findsOneWidget);
  });

  testWidgets('deleting cards offers Undo', (tester) async {
    final go = FakeGo();
    final rows = [row(1, 'uno'), row(2, 'dos')];
    FakeBrowse(go, rows);
    go.onUndo = () {
      rows.add(row(1, 'uno'));
      go.emitChange();
    };
    await openBrowse(tester);
    await tester.longPress(find.text('uno'));
    await tester.pumpAndSettle();
    await tester.tap(find.byTooltip('Bulk actions'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Delete cards'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Delete'));
    go.undoStep = 'Delete 1 note';
    await tester.pumpAndSettle();
    expect(find.text('uno'), findsNothing);
    expect(find.text('Deleted 1 card'), findsOneWidget);
    await tester.tap(find.widgetWithText(SnackBarAction, 'Undo'));
    await tester.pumpAndSettle();
    expect(go.callsTo('app.Undo'), hasLength(1));
    expect(find.text('uno'), findsOneWidget);
  });

  testWidgets('deleting notes needs confirmation', (tester) async {
    final go = FakeGo();
    FakeBrowse(go, [row(1, 'uno'), row(2, 'dos')]);
    await openBrowse(tester);
    await tester.longPress(find.text('uno'));
    await tester.pumpAndSettle();
    await tester.tap(find.byTooltip('Select all'));
    await tester.pumpAndSettle();
    await tester.tap(find.byTooltip('Bulk actions'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Delete cards'));
    await tester.pumpAndSettle();
    expect(find.text('Delete 2 cards?'), findsOneWidget);
    expect(
      find.text('You can undo this until you close Cardflip.'),
      findsOneWidget,
    );
    await tester.tap(find.text('Cancel'));
    await tester.pumpAndSettle();
    expect(go.callsTo('browse.DeleteNotes'), isEmpty);
  });

  testWidgets('bulk delete counts the other cards of the picked notes', (
    tester,
  ) async {
    final go = FakeGo();
    FakeBrowse(go, [
      row(1, 'gato', noteId: 7),
      row(2, 'cat', noteId: 7),
      row(3, 'perro'),
    ]);
    await openBrowse(tester);
    await tester.longPress(find.text('gato'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('perro'));
    await tester.pumpAndSettle();
    await tester.tap(find.byTooltip('Bulk actions'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Delete cards'));
    await tester.pumpAndSettle();
    expect(find.text('Delete 3 cards?'), findsOneWidget);
    expect(
      find.text(
        'Including 1 other card from the same notes. You can undo this until you close Cardflip.',
      ),
      findsOneWidget,
    );
    await tester.tap(find.text('Delete'));
    await tester.pumpAndSettle();
    expect((go.callsTo('browse.DeleteNotes').single[0] as List).toSet(), {
      1,
      3,
    });
    expect(find.text('No cards yet.'), findsOneWidget);
    expect(find.byTooltip('Select'), findsOneWidget);
  });

  testWidgets('the Select button picks cards without long-press', (
    tester,
  ) async {
    final go = FakeGo();
    FakeBrowse(go, [row(1, 'uno'), row(2, 'dos'), row(3, 'tres')]);
    await openBrowse(tester);
    expect(find.byTooltip('Add card'), findsOneWidget);
    await tester.tap(find.byTooltip('Select'));
    await tester.pumpAndSettle();
    expect(find.text('Select cards'), findsOneWidget);
    expect(
      tester
          .widget<IconButton>(
            find.widgetWithIcon(IconButton, Icons.drive_file_move_outline),
          )
          .onPressed,
      isNull,
    );
    await tester.tap(find.text('uno'));
    await tester.tap(find.text('tres'));
    await tester.pumpAndSettle();
    expect(find.text('2 selected'), findsOneWidget);
    expect(find.byKey(const Key('card-details')), findsNothing);
    await tester.tap(find.byTooltip('Bulk actions'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Suspend'));
    await tester.pumpAndSettle();
    final call = go.callsTo('browse.SetSuspended').single;
    expect((call[0] as List).toSet(), {1, 3});
    expect(call[1], true);

    // Deselecting the last card keeps the mode; Clear selection leaves it.
    await tester.tap(find.text('uno'));
    await tester.tap(find.text('tres'));
    await tester.pumpAndSettle();
    expect(find.text('Select cards'), findsOneWidget);
    await tester.tap(find.byTooltip('Clear selection'));
    await tester.pumpAndSettle();
    expect(find.text('Browse'), findsWidgets);
    expect(find.byType(Checkbox), findsNothing);
  });

  group('marked notes', () {
    Color? tint(WidgetTester tester, int id) =>
        tester.widget<ListTile>(find.byKey(Key('row-$id'))).tileColor;
    const colors = FlagColors.light;

    testWidgets('Toggle mark marks the selection, purple rows (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      FakeBrowse(go, [for (var i = 1; i <= 4; i++) row(i, 'card $i')]);
      await openBrowse(tester);
      await tester.tap(find.byTooltip('Select'));
      await tester.pumpAndSettle();
      for (var i = 1; i <= 3; i++) {
        await tester.tap(find.text('card $i'));
      }
      await tester.pumpAndSettle();

      // Bulk mark (spec).
      await tester.tap(find.byTooltip('Bulk actions'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Toggle mark'));
      await tester.pumpAndSettle();
      final call = go.callsTo('browse.ToggleMark').single;
      expect((call[0] as List).toSet(), {1, 2, 3});
      for (var i = 1; i <= 3; i++) {
        expect(
          tint(tester, i),
          colors.marked.withValues(alpha: colors.tintAlpha),
        );
      }
      expect(tint(tester, 4), isNull);

      // Toggling again unmarks them.
      await tester.tap(find.byTooltip('Bulk actions'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Toggle mark'));
      await tester.pumpAndSettle();
      for (var i = 1; i <= 3; i++) {
        expect(tint(tester, i), isNull);
      }
    });

    testWidgets('a flag wins over marked, marked over suspended', (
      tester,
    ) async {
      final go = FakeGo();
      FakeBrowse(go, [
        row(1, 'comer', marked: true, flag: 2),
        row(2, 'perro', marked: true, suspended: true),
        row(3, 'gato', suspended: true),
      ]);
      await openBrowse(tester);
      expect(tint(tester, 1), colors.rowTint(flag: 2, suspended: false));
      expect(
        tint(tester, 2),
        colors.rowTint(flag: 0, marked: true, suspended: false),
      );
      expect(tint(tester, 3), colors.rowTint(flag: 0, suspended: true));
      expect(tint(tester, 2), isNot(tint(tester, 3)));
    });

    testWidgets('Notes mode tints marked notes', (tester) async {
      final go = FakeGo();
      go.settings['browserMode'] = 'notes';
      FakeBrowse(go, [
        row(1, 'comer', noteId: 50, marked: true),
        row(2, 'perro', noteId: 51),
      ]);
      await openBrowse(tester);
      expect(
        tint(tester, 50),
        colors.rowTint(flag: 0, marked: true, suspended: false),
      );
      expect(tint(tester, 51), isNull);
    });
  });

  group('flags', () {
    Color? tint(WidgetTester tester, int id) =>
        tester.widget<ListTile>(find.byKey(Key('row-$id'))).tileColor;
    Color? marker(WidgetTester tester, int id) =>
        tester.widget<Icon>(find.byKey(Key('row-flag-$id'))).color;
    const colors = FlagColors.light;

    testWidgets('bulk Flag sets a flag on the selection, No flag clears it', (
      tester,
    ) async {
      final go = FakeGo();
      FakeBrowse(go, [for (var i = 1; i <= 6; i++) row(i, 'card $i')]);
      await openBrowse(tester);
      await tester.tap(find.byTooltip('Select'));
      await tester.pumpAndSettle();
      for (var i = 1; i <= 5; i++) {
        await tester.tap(find.text('card $i'));
      }
      await tester.pumpAndSettle();

      // Flag selection (spec).
      await tester.tap(find.byTooltip('Flag'));
      await tester.pumpAndSettle();
      expect(find.text('Blue'), findsOneWidget);
      await tester.tap(find.byKey(const Key('flag-item-4')));
      await tester.pumpAndSettle();
      final call = go.callsTo('browse.SetFlag').single;
      expect((call[0] as List).toSet(), {1, 2, 3, 4, 5});
      expect(call[1], 4);
      for (var i = 1; i <= 5; i++) {
        expect(marker(tester, i), colors.flag(4));
        expect(tint(tester, i), colors.rowTint(flag: 4, suspended: false));
      }
      expect(find.byKey(const Key('row-flag-6')), findsNothing);
      expect(tint(tester, 6), isNull);

      // Clear flags (spec).
      await tester.tap(find.byTooltip('Flag'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('No flag'));
      await tester.pumpAndSettle();
      expect(go.callsTo('browse.SetFlag').last[1], 0);
      for (var i = 1; i <= 5; i++) {
        expect(find.byKey(Key('row-flag-$i')), findsNothing);
        expect(tint(tester, i), isNull);
      }
    });

    testWidgets('suspended rows are yellow; a flag wins over suspended', (
      tester,
    ) async {
      final go = FakeGo();
      FakeBrowse(go, [
        row(1, 'comer', suspended: true),
        row(2, 'perro', suspended: true, flag: 2),
        row(3, 'gato'),
      ]);
      await openBrowse(tester);
      // Suspended row (spec): yellow, and it still says "suspended".
      expect(tint(tester, 1), colors.rowTint(flag: 0, suspended: true));
      expect(tint(tester, 1)!.toARGB32() & 0xFFFFFF, 0xFBC02D);
      expect(
        find.descendant(
          of: find.byKey(const Key('row-1')),
          matching: find.text('Default · suspended'),
        ),
        findsOneWidget,
      );
      // Flag wins over suspended (spec): orange.
      expect(tint(tester, 2), colors.rowTint(flag: 2, suspended: false));
      expect(marker(tester, 2), colors.flag(2));
      expect(tint(tester, 3), isNull);
    });

    testWidgets('the details flag a card, by the flag names', (tester) async {
      final go = FakeGo();
      go.settings['flagNames'] = ['Fix later', '', '', '', '', '', ''];
      FakeBrowse(go, [row(1, 'comer'), row(2, 'perro')]);
      await openBrowse(tester, size: desktop);
      await tester.tap(find.text('comer'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Flag'));
      await tester.pumpAndSettle();
      // Rename a flag (spec): the browser's flag action shows the name.
      expect(find.text('Fix later'), findsOneWidget);
      expect(find.text('Red'), findsNothing);
      await tester.tap(find.byKey(const Key('flag-item-1')));
      await tester.pumpAndSettle();
      expect(go.callsTo('browse.SetFlag').single, [
        [1],
        1,
      ]);
      expect(marker(tester, 1), colors.flag(1));
      expect(find.text('Fix later'), findsOneWidget); // in the details
      await tester.tap(find.byTooltip('Flag'));
      await tester.pumpAndSettle();
      expect(
        tester
            .widget<CheckedPopupMenuItem<int>>(
              find.byKey(const Key('flag-item-1')),
            )
            .checked,
        isTrue,
      );
    });
  });

  testWidgets('a shared note is disclosed in the editor', (tester) async {
    final go = FakeGo();
    FakeBrowse(go, [row(1, 'gato', noteId: 7), row(2, 'cat', noteId: 7)]);
    await openBrowse(tester, size: desktop);
    await tester.tap(find.text('gato'));
    await tester.pumpAndSettle();
    expect(find.text('Edit card'), findsOneWidget);
    expect(
      find.text('Changes also apply to 1 other card from this note.'),
      findsOneWidget,
    );
  });

  testWidgets('unsuspend from the details', (tester) async {
    final go = FakeGo();
    FakeBrowse(go, [row(1, 'comer', suspended: true), row(2, 'perro')]);
    await openBrowse(tester, size: desktop);
    expect(cell(tester, 1, 'due'), 'suspended');
    await tester.tap(find.text('comer'));
    await tester.pumpAndSettle();
    expect(find.text('New (suspended)'), findsOneWidget);
    await tester.tap(find.byTooltip('Unsuspend'));
    await tester.pumpAndSettle();
    expect(go.callsTo('browse.SetSuspended').single, [
      [1],
      false,
    ]);
    expect(cell(tester, 1, 'due'), 'new');
    expect(cell(tester, 2, 'due'), 'new');
    // The details show the state without the suspension.
    expect(find.text('New (suspended)'), findsNothing);
    expect(find.text('New'), findsWidgets);
    expect(find.byTooltip('Unsuspend'), findsNothing);
    expect(find.byTooltip('Suspend'), findsOneWidget);
  });

  for (final size in [phone, desktop]) {
    testWidgets('Card info in the details opens it with its history (spec, '
        '${size.width.toInt()} wide)', (tester) async {
      final go = FakeGo();
      FakeBrowse(go, [row(1, 'comer'), row(2, 'perro', state: 'review')]);
      go.handle(
        'browse.CardInfo',
        (args) => cardInfoJson(
          args[0] as int,
          state: 'review',
          due: DateTime.utc(2026, 10, 20),
          reviews: 1,
          history: [
            historyEntry(DateTime.utc(2026, 10, 10, 9), 3, kind: 'learning'),
          ],
        ),
      );
      await openBrowse(tester, size: size);
      await tester.tap(find.text('perro'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Card info'));
      await tester.pumpAndSettle();
      expect(go.callsTo('browse.CardInfo'), [
        [2],
      ]);
      expect(find.text('Review history'), findsOneWidget);
      expect(
        find.text(formatDateTime(DateTime.utc(2026, 10, 10, 9))),
        findsOneWidget,
      );
      await tester.pageBack();
      await tester.pumpAndSettle();
      expect(find.byTooltip('Card info'), findsOneWidget);
    });
  }

  testWidgets('a new card shows its position as New #42 (spec)', (
    tester,
  ) async {
    final go = FakeGo();
    FakeBrowse(go, [row(1, 'comer', position: 42), row(2, 'perro')]);
    await openBrowse(tester, size: desktop);
    expect(cell(tester, 1, 'due'), 'New #42');
    expect(cell(tester, 2, 'due'), 'new');
    await tester.tap(find.text('comer'));
    await tester.pumpAndSettle();
    expect(find.text('New #42'), findsNWidgets(2)); // and in the details

    // On a phone, the due column is in the subtitle.
    await openBrowse(tester);
    expect(find.text('Default · New #42'), findsOneWidget);
  });

  group('Reposition', () {
    Future<void> openDialog(WidgetTester tester) async {
      await tester.longPress(find.text('uno'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Select all'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Bulk actions'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Reposition'));
      await tester.pumpAndSettle();
    }

    testWidgets('sends start, step and options; says which were not new', (
      tester,
    ) async {
      final go = FakeGo();
      FakeBrowse(go, [
        row(1, 'uno', position: 1),
        row(2, 'dos', position: 2),
        row(3, 'tres', state: 'review'),
      ]);
      go.handle('browse.Reposition', (_) => 1);
      await openBrowse(tester);
      await openDialog(tester);
      expect(find.text('Reposition new cards'), findsOneWidget);
      await tester.enterText(find.byKey(const Key('reposition-start')), '5');
      await tester.enterText(find.byKey(const Key('reposition-step')), '2');
      await tester.tap(find.text('Randomize order'));
      await tester.tap(find.text('Shift position of existing cards'));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(FilledButton, 'Reposition'));
      await tester.pumpAndSettle();
      final call = go.callsTo('browse.Reposition').single;
      expect((call[0] as List).toSet(), {1, 2, 3});
      expect(call.sublist(1), [5, 2, true, true]);
      expect(
        find.text(
          'Repositioned 2 cards. 1 card was not new and left unchanged.',
        ),
        findsOneWidget,
      );
    });

    testWidgets('defaults to start 0, step 1; rejects a step of 0', (
      tester,
    ) async {
      final go = FakeGo();
      FakeBrowse(go, [row(1, 'uno'), row(2, 'dos')]);
      go.handle('browse.Reposition', (_) => 0);
      await openBrowse(tester);
      await openDialog(tester);
      await tester.enterText(find.byKey(const Key('reposition-step')), '0');
      await tester.tap(find.widgetWithText(FilledButton, 'Reposition'));
      await tester.pumpAndSettle();
      expect(find.text('A whole number from 1 to 1000000000'), findsOneWidget);
      expect(go.callsTo('browse.Reposition'), isEmpty);
      await tester.enterText(find.byKey(const Key('reposition-step')), '1');
      await tester.tap(find.widgetWithText(FilledButton, 'Reposition'));
      await tester.pumpAndSettle();
      expect(go.callsTo('browse.Reposition').single.sublist(1), [
        0,
        1,
        false,
        false,
      ]);
      expect(find.text('Repositioned 2 cards.'), findsOneWidget);
    });
  });

  group('Set due date', () {
    testWidgets('bulk postpone starts with the last input; Go\'s error shows '
        'inline (spec)', (tester) async {
      final go = FakeGo();
      go.settings = {...go.settings, 'lastDueDate': '7'};
      FakeBrowse(go, [row(1, 'uno'), row(2, 'dos'), row(3, 'tres')]);
      go.handle('browse.SetDueDate', (args) {
        if (args[1] == '10-5') {
          throw 'The range 10-5 ends before it starts; put the smaller number first.';
        }
        return null;
      });
      await openBrowse(tester);
      await tester.longPress(find.text('uno'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Select all'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Bulk actions'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Set due date'));
      await tester.pumpAndSettle();
      expect(find.text('Set due date of 3 cards'), findsOneWidget);
      expect(find.textContaining('60-90 = a random day'), findsOneWidget);
      final input = find.byKey(const Key('due-date-input'));
      expect(tester.widget<TextField>(input).controller!.text, '7');

      await tester.enterText(input, '10-5');
      await tester.tap(find.widgetWithText(FilledButton, 'Set due date'));
      await tester.pumpAndSettle();
      expect(
        find.text(
          'The range 10-5 ends before it starts; put the smaller number first.',
        ),
        findsOneWidget,
      );
      expect(find.text('Set due date of 3 cards'), findsOneWidget);

      await tester.enterText(input, '3');
      await tester.tap(find.widgetWithText(FilledButton, 'Set due date'));
      await tester.pumpAndSettle();
      final calls = go.callsTo('browse.SetDueDate');
      expect(calls, hasLength(2));
      expect((calls.last[0] as List).toSet(), {1, 2, 3});
      expect(calls.last[1], '3');
      // The dialog closed; the undo offer names the change.
      expect(input, findsNothing);
      expect(find.text('Set due date of 3 cards'), findsOneWidget);
    });

    testWidgets('Ctrl+Shift+D sets the current card\'s; the details have a '
        'button', (tester) async {
      final go = FakeGo();
      FakeBrowse(go, [row(1, 'uno'), row(2, 'dos')]);
      go.handle('browse.SetDueDate', (_) => null);
      await openBrowse(tester, size: desktop);
      await tester.tap(find.text('dos'));
      await tester.pumpAndSettle();
      await tester.sendKeyDownEvent(LogicalKeyboardKey.controlLeft);
      await tester.sendKeyDownEvent(LogicalKeyboardKey.shiftLeft);
      await tester.sendKeyEvent(LogicalKeyboardKey.keyD);
      await tester.sendKeyUpEvent(LogicalKeyboardKey.shiftLeft);
      await tester.sendKeyUpEvent(LogicalKeyboardKey.controlLeft);
      await tester.pumpAndSettle();
      expect(find.text('Set due date'), findsWidgets);
      await tester.enterText(find.byKey(const Key('due-date-input')), '0');
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pumpAndSettle();
      expect(go.callsTo('browse.SetDueDate').single, [
        [2],
        '0',
      ]);
      expect(find.text('Set due date of 1 card'), findsOneWidget); // undo

      await tester.tap(find.byTooltip('Set due date'));
      await tester.pumpAndSettle();
      await tester.enterText(find.byKey(const Key('due-date-input')), '30!');
      await tester.tap(find.widgetWithText(FilledButton, 'Set due date'));
      await tester.pumpAndSettle();
      expect(go.callsTo('browse.SetDueDate').last, [
        [2],
        '30!',
      ]);
    });
  });

  group('Reset', () {
    bool checked(WidgetTester tester, String key) =>
        tester.widget<CheckboxListTile>(find.byKey(Key(key))).value!;

    testWidgets('the options start with their defaults; nothing changes '
        'until confirmed (spec)', (tester) async {
      final go = FakeGo();
      FakeBrowse(go, [
        for (final (i, w) in ['uno', 'dos', 'tres', 'cuatro', 'cinco'].indexed)
          row(i + 1, w),
      ]);
      go.handle('browse.ResetToNew', (_) => null);
      await openBrowse(tester);
      await tester.longPress(find.text('uno'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Select all'));
      await tester.pumpAndSettle();
      Future<void> openReset() async {
        await tester.tap(find.byTooltip('Bulk actions'));
        await tester.pumpAndSettle();
        await tester.tap(find.text('Reset to new'));
        await tester.pumpAndSettle();
      }

      await openReset();
      expect(find.text('Reset 5 cards to new?'), findsOneWidget);
      expect(checked(tester, 'reset-restore-position'), isTrue);
      expect(checked(tester, 'reset-counts'), isFalse);
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(go.callsTo('browse.ResetToNew'), isEmpty);

      await openReset();
      await tester.tap(find.text('Reset repetition and lapse counts'));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(FilledButton, 'Reset'));
      await tester.pumpAndSettle();
      final call = go.callsTo('browse.ResetToNew').single;
      expect((call[0] as List).toSet(), {1, 2, 3, 4, 5});
      expect(call[1], {'restorePosition': true, 'resetCounts': true});
      expect(find.text('Reset 5 cards to new'), findsOneWidget); // undo
    });

    testWidgets('the details have a button', (tester) async {
      final go = FakeGo();
      FakeBrowse(go, [row(1, 'uno'), row(2, 'dos')]);
      go.handle('browse.ResetToNew', (_) => null);
      await openBrowse(tester, size: desktop);
      await tester.tap(find.text('dos'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Reset to new'));
      await tester.pumpAndSettle();
      expect(find.text('Reset 1 card to new?'), findsOneWidget);
      await tester.tap(find.widgetWithText(FilledButton, 'Reset'));
      await tester.pumpAndSettle();
      expect(go.callsTo('browse.ResetToNew').single, [
        [2],
        {'restorePosition': true, 'resetCounts': false},
      ]);
    });
  });

  testWidgets('delete from the details on desktop closes them', (tester) async {
    final go = FakeGo();
    FakeBrowse(go, [row(1, 'comer'), row(2, 'perro')]);
    await openBrowse(tester, size: desktop);
    await tester.tap(find.text('comer'));
    await tester.pumpAndSettle();
    await tester.tap(find.byTooltip('Delete'));
    await tester.pumpAndSettle();
    expect(find.text('Delete 1 card?'), findsOneWidget);
    expect(
      find.text('You can undo this until you close Cardflip.'),
      findsOneWidget,
    );
    await tester.tap(find.text('Delete'));
    await tester.pumpAndSettle();
    expect(go.callsTo('browse.DeleteNotes').single, [
      [1],
    ]);
    expect(find.text('comer'), findsNothing);
    expect(find.byKey(const Key('card-details')), findsNothing);
    expect(find.text('Select a card to see its details.'), findsOneWidget);
  });

  testWidgets('back from the details during a video stops it before the '
      'player closes (spec)', (tester) async {
    final backend = FakeBackend();
    useFakeMedia(backend, FakeMediaFiles(['clip.mp4']));
    final go = FakeGo();
    FakeBrowse(go, [
      row(1, 'gato')
        ..['question'] = '<cf-video idx="0"></cf-video> gato'
        ..['questionMedia'] = [
          {'file': 'clip.mp4', 'kind': 'video'},
        ],
    ]);
    await openBrowse(tester);
    await tester.tap(find.byKey(const Key('row-1')));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('card-details')), findsOneWidget);
    await tester.tap(find.byTooltip('Play clip.mp4'));
    await tester.pumpAndSettle();
    expect(backend.playing, 'clip.mp4');

    await tester.pageBack();
    await tester.pump();
    expect(backend.log, ['stop']);
    expect(find.byKey(const Key('fake-video')), findsNothing);
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('card-details')), findsNothing);
    expect(find.byKey(const Key('row-1')), findsOneWidget);
    expect(backend.log, ['stop', 'close']);
  });

  testWidgets('delete from the details on a phone closes them', (tester) async {
    final go = FakeGo();
    FakeBrowse(go, [
      row(1, 'gato', noteId: 7),
      row(2, 'cat', noteId: 7),
      row(3, 'perro'),
    ]);
    await openBrowse(tester);
    await tester.tap(find.text('gato'));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('card-details')), findsOneWidget);
    await tester.tap(find.byTooltip('Delete'));
    await tester.pumpAndSettle();
    expect(find.text('Delete 2 cards?'), findsOneWidget);
    expect(
      find.text(
        'Including 1 other card from the same note. You can undo this until you close Cardflip.',
      ),
      findsOneWidget,
    );
    await tester.tap(find.text('Delete'));
    await tester.pumpAndSettle();
    expect(go.callsTo('browse.DeleteImpact'), isEmpty);
    expect(go.callsTo('browse.DeleteNotes').single, [
      [1],
    ]);
    expect(find.byKey(const Key('card-details')), findsNothing);
    expect(find.text('1 card'), findsOneWidget);
    expect(find.text('perro'), findsOneWidget);
  });

  testWidgets('cancelling a delete from the details keeps the card', (
    tester,
  ) async {
    final go = FakeGo();
    FakeBrowse(go, [row(1, 'comer')]);
    await openBrowse(tester, size: desktop);
    await tester.tap(find.text('comer'));
    await tester.pumpAndSettle();
    await tester.tap(find.byTooltip('Delete'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Cancel'));
    await tester.pumpAndSettle();
    expect(go.callsTo('browse.DeleteNotes'), isEmpty);
    expect(find.byKey(const Key('card-details')), findsOneWidget);
  });

  testWidgets('Edit in the details focuses the first field', (tester) async {
    final go = FakeGo();
    FakeBrowse(go, [row(1, 'comer')]);
    await openBrowse(tester);
    await tester.tap(find.text('comer'));
    await tester.pumpAndSettle();
    await tester.tap(find.byTooltip('Edit'));
    await tester.pumpAndSettle();
    final field = tester.widget<RichField>(
      find.byKey(const Key('details-field-0')),
    );
    expect(field.focusNode.hasFocus, isTrue);
  });

  testWidgets('the details editor has the math menu and shortcuts', (
    tester,
  ) async {
    final go = FakeGo();
    FakeBrowse(go, [row(1, 'comer')]);
    await openBrowse(tester);
    await tester.tap(find.text('comer'));
    await tester.pumpAndSettle();
    await enterField(tester, 1, 'x^2', prefix: 'details-field');
    selectInField(tester, 1, 0, 3, prefix: 'details-field');
    final c = fieldController(tester, 1, prefix: 'details-field');
    await tester.ensureVisible(find.byTooltip('Math'));
    await tester.tap(find.byTooltip('Math'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('LaTeX equation'));
    await tester.pumpAndSettle();
    expect(c.html, r'[$]x^2[/$]');

    await tester.sendKeyDownEvent(LogicalKeyboardKey.controlLeft);
    await tester.sendKeyEvent(LogicalKeyboardKey.keyM);
    await tester.sendKeyUpEvent(LogicalKeyboardKey.controlLeft);
    await tester.sendKeyEvent(LogicalKeyboardKey.keyM);
    await tester.pump();
    expect(c.html, r'[$]x^2\(\)[/$]');
  });

  group('mask editor in the details editor', () {
    const shapes =
        '{{c1::image-occlusion:rect:left=.05:top=.05:width=.2:height=.2}}<br>'
        '{{c2::image-occlusion:rect:left=.6:top=.6:width=.3:height=.25}}';

    /// Opens Browse over the two cards of an Image Occlusion note with
    /// shapes 1 and 2, and selects the first. Saving stores the fields in
    /// the returned list.
    Future<(FakeGo, List<Object?>)> openOcclusion(
      WidgetTester tester, {
      Size size = desktop,
      FakePicker? picker,
    }) async {
      final go = FakeGo();
      go.media['heart.png'] = await occlusionImage(tester);
      FakeBrowse(go, [
        row(1, 'heart 1', noteId: 1001, noteType: 'Image Occlusion'),
        row(2, 'heart 2', noteId: 1001, noteType: 'Image Occlusion'),
      ]);
      final fields = <Object?>[
        shapes,
        '<img src="heart.png">',
        'Heart',
        '',
        '',
      ];
      go.handle(
        'browse.Card',
        (args) => {
          'cardId': args[0],
          'deckId': 1,
          'deck': 'Default',
          'cardName': 'Image Occlusion',
          'question': '<div>${fields[2]}</div>',
          'answer': '<div>${fields[2]}<hr>answer</div>',
          'questionMedia': const [],
          'questionTts': const [],
          'state': 'new',
          'due': null,
          'position': null,
          'intervalDays': 0,
          'stability': null,
          'difficulty': null,
          'reviews': 0,
          'lapses': 0,
          'suspended': false,
          'flag': 0,
          'note': {
            'id': 1001,
            'typeId': 13,
            'fields': [...fields],
            'tags': <String>[],
            'cardIds': [1, 2],
            'cardOrds': [0, 1],
          },
          'noteType': occlusionTypeJson(),
          'noteCards': 2,
        },
      );
      go.handle('notes.Update', (args) {
        fields.setAll(0, args[1] as List);
        go.emitChange();
        return null;
      });
      await pumpApp(tester, size: size, picker: picker);
      await tester.tap(find.text('Browse').last);
      await tester.pumpAndSettle();
      await tester.tap(find.text('heart 1'));
      await tester.pumpAndSettle();
      await waitForCanvas(tester);
      return (go, fields);
    }

    Future<void> save(WidgetTester tester) async {
      final save = find.descendant(
        of: find.byKey(const Key('details-save-row')),
        matching: find.byType(FilledButton),
      );
      await tester.ensureVisible(save);
      await tester.pumpAndSettle();
      await tester.tap(save);
      await tester.pumpAndSettle();
    }

    testWidgets('an Image Occlusion card shows the mask editor (spec)', (
      tester,
    ) async {
      await openOcclusion(tester, size: phone);
      expect(find.byKey(const Key('occlusion-section')), findsOneWidget);
      for (final tool in ['select', 'rect', 'ellipse', 'polygon', 'text']) {
        expect(find.byKey(Key('occlusion-tool-$tool')), findsOneWidget);
      }
      expect(find.byKey(const Key('details-field-0')), findsNothing);
      expect(find.byKey(const Key('details-field-1')), findsNothing);
      for (final i in [2, 3, 4]) {
        expect(find.byKey(Key('details-field-$i')), findsOneWidget);
      }
      // The narrow layout's screen fits the canvas.
      final canvas = tester.getRect(find.byKey(const Key('occlusion-canvas')));
      expect(canvas.left, greaterThanOrEqualTo(0));
      expect(canvas.right, lessThanOrEqualTo(phone.width));
    });

    testWidgets('a new shape gets a new number (spec)', (tester) async {
      final (go, _) = await openOcclusion(tester);
      await tester.ensureVisible(find.byKey(const Key('occlusion-tool-rect')));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('occlusion-tool-rect')));
      await showCanvas(tester);
      await tester.pump();
      await tester.dragFrom(onCanvas(tester, .3, .1), const Offset(40, 20));
      await tester.pump();
      await save(tester);
      expect(find.text('Card saved'), findsOneWidget);
      final [occlusion as String, img, header, _, _] =
          go.callsTo('notes.Update').single[1] as List<dynamic>;
      expect(parseOcclusionField(occlusion).map((s) => s.ord), [1, 2, 3]);
      expect(occlusion, startsWith(shapes));
      expect(img, '<img src="heart.png">');
      expect(header, 'Heart');
      // The pane reloads, and the editor keeps its shapes.
      expect(find.byKey(const Key('occlusion-canvas')), findsOneWidget);
    });

    testWidgets('a note without shapes is not saved (spec)', (tester) async {
      final (go, _) = await openOcclusion(tester);
      await showCanvas(tester);
      for (final (x, y) in [(.1, .1), (.7, .7)]) {
        await tester.tapAt(onCanvas(tester, x, y));
        await tester.pump();
        await tester.sendKeyEvent(LogicalKeyboardKey.delete);
        await tester.pump();
      }
      await save(tester);
      expect(
        tester.widget<Text>(find.byKey(const Key('details-error'))).data,
        'Draw a shape over each part to ask.',
      );
      expect(go.callsTo('notes.Update'), isEmpty);
    });

    testWidgets('Change image stores the new picture (spec)', (tester) async {
      final lungs = await occlusionImage(
        tester,
        color: const Color(0xFFCC4444),
      );
      final picker = FakePicker()..media = [(name: 'lungs.png', bytes: lungs)];
      final (go, _) = await openOcclusion(tester, picker: picker);
      await tester.ensureVisible(find.byKey(const Key('occlusion-pick-image')));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('occlusion-pick-image')));
      await tester.pumpAndSettle();
      expect(picker.mediaPicks, 1);
      expect(
        tester.widget<OcclusionEditor>(find.byType(OcclusionEditor)).image,
        lungs,
      );
      expect(go.callsTo('media.Add').single[0], 'lungs.png');
      expect(go.media['lungs.png'], lungs);

      await save(tester);
      final [occlusion, img, _, _, _] =
          go.callsTo('notes.Update').single[1] as List<dynamic>;
      expect(occlusion, shapes);
      expect(img, '<img src="lungs.png">');
    });

    testWidgets('Edit focuses the Header field (spec)', (tester) async {
      await openOcclusion(tester, size: phone);
      await tester.tap(find.byTooltip('Edit'));
      await tester.pumpAndSettle();
      final field = tester.widget<RichField>(
        find.byKey(const Key('details-field-2')),
      );
      expect(field.focusNode.hasFocus, isTrue);
    });
  });

  group('editing buttons in the details editor', () {
    const text = '{{c1::Paris}} is in France';

    Future<void> openCloze(WidgetTester tester) async {
      final go = FakeGo();
      FakeBrowse(go, [row(1, text, cloze: true)]);
      await openBrowse(tester, size: desktop);
      await tester.tap(find.text(text));
      await tester.pumpAndSettle();
    }

    testWidgets('a cloze note has the same group; Save stays alone (spec)', (
      tester,
    ) async {
      await openCloze(tester);
      expect(toolbarTools(tester), [
        'Cloze deletion',
        'Math',
        'Attach picture, audio or video',
        'Record audio',
      ]);
      final saveRow = find.byKey(const Key('details-save-row'));
      expect(
        find.descendant(of: saveRow, matching: find.byType(IconButton)),
        findsNothing,
      );
      expect(
        find.descendant(of: saveRow, matching: find.byType(FilledButton)),
        findsOneWidget,
      );
      expect(
        find.descendant(of: saveRow, matching: find.text('Save')),
        findsOneWidget,
      );
    });

    testWidgets('Cloze deletion wraps the selection in c2 (spec)', (
      tester,
    ) async {
      await openCloze(tester);
      expect(fieldHtml(tester, 0, prefix: 'details-field'), text);
      selectInField(tester, 0, 20, 26, prefix: 'details-field');
      final cloze = find.byKey(const Key('insert-cloze'));
      await tester.ensureVisible(cloze);
      await tester.pumpAndSettle();
      await tester.tap(cloze);
      await tester.pumpAndSettle();
      await tester.tap(find.text('Next cloze number'));
      await tester.pumpAndSettle();
      expect(
        fieldHtml(tester, 0, prefix: 'details-field'),
        '{{c1::Paris}} is in {{c2::France}}',
      );
    });

    testWidgets('Ctrl+Alt+Shift+C reuses the highest number (spec)', (
      tester,
    ) async {
      await openCloze(tester);
      await enterField(tester, 0, text, prefix: 'details-field');
      selectInField(tester, 0, 20, 26, prefix: 'details-field');
      await pressKeys(tester, [
        LogicalKeyboardKey.controlLeft,
        LogicalKeyboardKey.altLeft,
        LogicalKeyboardKey.shiftLeft,
        LogicalKeyboardKey.keyC,
      ]);
      expect(
        fieldHtml(tester, 0, prefix: 'details-field'),
        '{{c1::Paris}} is in {{c1::France}}',
      );
    });

    testWidgets('a Basic note has no Cloze deletion', (tester) async {
      final go = FakeGo();
      FakeBrowse(go, [row(1, 'comer')]);
      await openBrowse(tester, size: desktop);
      await tester.tap(find.text('comer'));
      await tester.pumpAndSettle();
      expect(toolbarTools(tester), [
        'Math',
        'Attach picture, audio or video',
        'Record audio',
      ]);
    });
  });

  group('media in the details editor', () {
    Future<void> openDetails(WidgetTester tester, {FakePicker? picker}) async {
      await pumpApp(tester, size: desktop, picker: picker);
      await tester.tap(find.text('Browse').last);
      await tester.pumpAndSettle();
      await tester.tap(find.text('comer'));
      await tester.pumpAndSettle();
    }

    Future<void> tapInDetails(WidgetTester tester, Finder button) async {
      await tester.ensureVisible(button);
      await tester.pumpAndSettle();
      await tester.tap(button);
      await tester.pumpAndSettle();
    }

    testWidgets('attach an image to an existing note and save it (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      FakeBrowse(go, [row(1, 'comer')]);
      go.handle('notes.Update', (_) => null);
      final picker = FakePicker()..media = [(name: 'eat.png', bytes: pngPixel)];
      await openDetails(tester, picker: picker);
      await enterField(tester, 1, 'to eat ', prefix: 'details-field');
      await tapInDetails(tester, find.byKey(const Key('attach-media')));
      expect(go.media.keys, ['eat.png']);
      expect(
        fieldHtml(tester, 1, prefix: 'details-field'),
        'to eat <img src="eat.png">',
      );
      await tapInDetails(tester, find.text('Save'));
      expect(go.callsTo('notes.Update').single[1], [
        'comer',
        'to eat <img src="eat.png">',
      ]);
    });

    testWidgets('record, paste and drop work there too', (tester) async {
      final recorder = FakeRecorder();
      useFakeRecorder(recorder);
      FakeClipboard(imageBytes: pngPixel).install();
      final go = FakeGo();
      FakeBrowse(go, [row(1, 'comer')]);
      await openDetails(tester);

      await enterField(tester, 0, '', prefix: 'details-field');
      await tapInDetails(tester, find.byKey(const Key('record-audio')));
      await tester.tap(find.byKey(const Key('record-insert')));
      await tester.pumpAndSettle();
      expect(
        fieldHtml(tester, 0, prefix: 'details-field'),
        '[sound:media-1.m4a]',
      );

      await enterField(tester, 1, '', prefix: 'details-field');
      await pressKeys(tester, [
        LogicalKeyboardKey.controlLeft,
        LogicalKeyboardKey.keyV,
      ]);
      expect(
        fieldHtml(tester, 1, prefix: 'details-field'),
        '<img src="paste-2.jpg">',
      );

      tester
          .widget<FieldDropTarget>(find.byKey(const Key('details-drop-1')))
          .onDrop([
            (name: 'hola.mp3', bytes: Uint8List.fromList([1])),
          ]);
      await tester.pumpAndSettle();
      expect(
        fieldHtml(tester, 1, prefix: 'details-field'),
        '<img src="paste-2.jpg">[sound:hola.mp3]',
      );
    });
  });

  testWidgets('exports the notes of the selected cards', (tester) async {
    final go = FakeGo();
    final picker = FakePicker();
    // Ten cards of eight notes: notes 1 and 2 have two cards each.
    FakeBrowse(go, [
      for (var i = 1; i <= 10; i++)
        row(i, 'card $i', noteId: i <= 4 ? 100 + (i + 1) ~/ 2 : 100 + i - 2),
    ]);
    await pumpApp(tester, picker: picker, size: desktop);
    await tester.tap(find.text('Browse').last);
    await tester.pumpAndSettle();
    await tester.tap(find.byTooltip('Select'));
    await tester.pumpAndSettle();
    await tester.tap(find.byTooltip('Select all'));
    await tester.pumpAndSettle();
    expect(find.text('10 selected'), findsOneWidget);
    await tester.tap(find.byTooltip('Bulk actions'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Export').last);
    await tester.pumpAndSettle();
    expect(find.text('Export the notes of 10 cards'), findsOneWidget);
    // The same options as a deck export.
    for (final key in ['scheduling', 'presets', 'media', 'legacy']) {
      expect(find.byKey(Key('export-$key')), findsOneWidget);
    }
    await tester.tap(find.widgetWithText(FilledButton, 'Export'));
    await tester.pumpAndSettle();
    final call = go.callsTo('transfer.ExportNotes').single;
    expect((call[0] as List).toSet(), {for (var i = 1; i <= 10; i++) i});
    expect(call[1], {
      'scheduling': false,
      'presets': false,
      'media': true,
      'legacy': false,
    });
    expect(picker.saved.keys, ['Selected notes.apkg']);
  });

  testWidgets('exports the selected notes or cards as text', (tester) async {
    final go = FakeGo();
    final picker = FakePicker();
    FakeBrowse(go, [for (var i = 1; i <= 5; i++) row(i, 'card $i')]);
    await pumpApp(tester, picker: picker, size: desktop);
    await tester.tap(find.text('Browse').last);
    await tester.pumpAndSettle();
    await tester.tap(find.byTooltip('Select'));
    await tester.pumpAndSettle();
    await tester.tap(find.byTooltip('Select all'));
    await tester.pumpAndSettle();

    Future<void> export(String format) async {
      await tester.tap(find.byTooltip('Bulk actions'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Export').last);
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('export-format')));
      await tester.pumpAndSettle();
      await tester.tap(find.text(format).last);
      await tester.pumpAndSettle();
    }

    // Export selected notes (spec).
    await export('Notes in plain text (.txt)');
    expect(find.text('Export the notes of 5 cards'), findsOneWidget);
    await tester.tap(find.widgetWithText(FilledButton, 'Export'));
    await tester.pumpAndSettle();
    final notes = go.callsTo('transfer.ExportNotesText').single;
    expect(((notes[0] as Map)['cardIds'] as List).toSet(), {1, 2, 3, 4, 5});
    expect(picker.saved.keys, ['Selected notes.txt']);

    // The selected cards themselves.
    await export('Cards in plain text (.txt)');
    expect(find.text('Export 5 cards'), findsOneWidget);
    await tester.tap(find.widgetWithText(FilledButton, 'Export'));
    await tester.pumpAndSettle();
    final cards = go.callsTo('transfer.ExportCardsText').single;
    expect((cards[0] as Map)['all'], isFalse);
    expect(picker.saved.keys, ['Selected notes.txt', 'Selected cards.txt']);
  });

  group('Notes mode', () {
    /// Three reversed notes (two cards each) and a Basic note.
    List<Map<String, Object?>> notes() => [
      for (final (i, word) in ['gato', 'perro', 'pájaro'].indexed) ...[
        row(i * 2 + 1, word, noteId: 100 + i, noteType: 'Reversed'),
        row(i * 2 + 2, 'cat $word', noteId: 100 + i, noteType: 'Reversed'),
      ],
      row(7, 'uno', noteId: 103),
    ];

    Future<void> showNotes(WidgetTester tester) async {
      await tester.tap(
        find.descendant(
          of: find.byKey(const Key('browse-mode')),
          matching: find.text('Notes'),
        ),
      );
      await tester.pumpAndSettle();
    }

    for (final size in [phone, desktop]) {
      final layout = size == phone ? 'list' : 'table';
      testWidgets('a reversed note is listed once, with 2 cards '
          '(spec, $layout)', (tester) async {
        final go = FakeGo();
        FakeBrowse(go, notes());
        await openBrowse(tester, size: size);
        expect(find.text('7 cards'), findsOneWidget);
        await showNotes(tester);
        expect(find.text('4 notes'), findsOneWidget);
        expect(go.callsTo('browse.SearchNotes').last, ['', 'created', false]);
        expect(find.text('gato'), findsOneWidget);
        expect(find.text('cat gato'), findsNothing);
        if (size == desktop) {
          expect(cell(tester, 100, 'cards'), '2 cards');
          expect(cell(tester, 100, 'noteType'), 'Reversed');
          expect(cell(tester, 103, 'cards'), '1 card');
          // Selecting a note shows its first card's details.
          await tester.tap(find.text('perro'));
          await tester.pumpAndSettle();
          expect(go.callsTo('browse.Card').last, [3]);
        } else {
          expect(find.text('2 cards · Reversed'), findsNWidgets(3));
          expect(find.text('1 card · Basic'), findsOneWidget);
        }
      });
    }

    testWidgets('a note matches through one of its cards (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      final rows = notes();
      rows[1] = {...rows[1], 'suspended': true};
      FakeBrowse(go, rows);
      await openBrowse(tester);
      await showNotes(tester);
      await tester.enterText(find.byKey(const Key('search')), 'is:suspended');
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pumpAndSettle();
      expect(find.text('1 note'), findsOneWidget);
      expect(find.text('gato'), findsOneWidget);
    });

    testWidgets('suspending 3 reversed notes suspends all 6 cards (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      FakeBrowse(go, notes());
      await openBrowse(tester);
      await showNotes(tester);
      await tester.tap(find.byTooltip('Select'));
      await tester.pumpAndSettle();
      expect(find.text('Select notes'), findsOneWidget);
      for (final word in ['gato', 'perro', 'pájaro']) {
        await tester.tap(find.text(word));
      }
      await tester.pumpAndSettle();
      expect(find.text('3 selected'), findsOneWidget);
      await tester.tap(find.byTooltip('Bulk actions'));
      await tester.pumpAndSettle();
      expect(find.text('Delete notes'), findsOneWidget);
      await tester.tap(find.text('Suspend'));
      await tester.pumpAndSettle();
      final call = go.callsTo('browse.SetSuspended').single;
      expect((call[0] as List).toSet(), {1, 2, 3, 4, 5, 6});
      expect(call[1], true);
    });

    testWidgets('the mode is remembered and sorts by note columns', (
      tester,
    ) async {
      final go = FakeGo();
      FakeBrowse(go, notes());
      await openBrowse(tester, size: desktop);
      // Sorting cards by a column notes don't have falls back to created.
      await tester.tap(find.byKey(const Key('column-question')));
      await tester.pumpAndSettle();
      await showNotes(tester);
      expect(go.callsTo('app.SetBrowserMode').single, ['notes']);
      expect(go.settings['browserMode'], 'notes');
      expect(go.callsTo('browse.SearchNotes').last, ['', 'created', false]);
      expect(find.byTooltip('Columns'), findsNothing);
      // Beside the sidebar the table scrolls sideways.
      await tester.ensureVisible(find.byKey(const Key('column-due')));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('column-due')));
      await tester.pumpAndSettle();
      expect(go.callsTo('browse.SearchNotes').last, ['', 'due', false]);
      // Only some note columns sort.
      final calls = go.callsTo('browse.SearchNotes').length;
      await tester.tap(find.byKey(const Key('column-deck')));
      await tester.pumpAndSettle();
      expect(go.callsTo('browse.SearchNotes').length, calls);
      await tester.tap(find.byTooltip('Sort'));
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('sort-sortField')), findsOneWidget);
      expect(find.byKey(const Key('sort-question')), findsNothing);
      await tester.tap(find.byKey(const Key('sort-sortField')));
      await tester.pumpAndSettle();
      expect(go.callsTo('browse.SearchNotes').last, ['', 'sortField', false]);

      // A restart opens the browser in Notes mode.
      await tester.pumpWidget(const SizedBox());
      final again = FakeGo()..settings['browserMode'] = 'notes';
      FakeBrowse(again, notes());
      await openBrowse(tester, size: desktop);
      expect(find.text('4 notes'), findsOneWidget);
      expect(again.callsTo('browse.Search'), isEmpty);
    });
  });

  group('columns', () {
    /// The table's headers, left to right.
    List<String> headers(WidgetTester tester) => [
      for (final e
          in find
              .byWidgetPredicate(
                (w) =>
                    w.key is ValueKey<String> &&
                    (w.key! as ValueKey<String>).value.startsWith('column-'),
              )
              .evaluate())
        (e.widget.key! as ValueKey<String>).value.substring('column-'.length),
    ];

    /// Taps [finder] in a dialog or menu, scrolling to it first.
    Future<void> tapItem(WidgetTester tester, Finder finder) async {
      await tester.scrollUntilVisible(
        finder,
        100,
        scrollable: find.byType(Scrollable).last,
      );
      await tester.tap(finder);
      await tester.pumpAndSettle();
    }

    Future<void> columnMenu(WidgetTester tester, String column) async {
      await tester.tap(
        find.byKey(Key('column-$column')),
        buttons: kSecondaryButton,
      );
      await tester.pumpAndSettle();
    }

    for (final size in [phone, desktop]) {
      final wide = size == desktop;
      final layout = wide ? 'table' : 'list';

      testWidgets('a large collection loads one page of rows ($layout)', (
        tester,
      ) async {
        final go = FakeGo();
        final browse = FakeBrowse(go, [
          for (var i = 1; i <= 50000; i++) row(i, 'card $i'),
        ]);
        await openBrowse(tester, size: size);
        expect(find.text('50,000 cards'), findsOneWidget);
        expect(find.text('card 1'), findsOneWidget);
        expect(browse.pages, [100]);
        expect(go.callsTo('browse.Rows').single[1], [
          'question',
          'deck',
          'due',
          'noteType',
        ]);
        if (wide) {
          expect(headers(tester), ['question', 'deck', 'due', 'noteType']);
          expect(cell(tester, 1, 'noteType'), 'Basic');
        } else {
          expect(find.text('Default · new'), findsWidgets);
        }
      });

      testWidgets('sort by the sort field (spec, $layout)', (tester) async {
        final go = FakeGo();
        // The note type sorts by its second field, the meaning.
        FakeBrowse(go, [
          row(1, 'perro', sortField: 'Dog'),
          row(2, 'agua', sortField: 'water'),
          row(3, 'gato', sortField: 'cat'),
        ]);
        await openBrowse(tester, size: size);
        await tester.tap(find.byTooltip('Sort'));
        await tester.pumpAndSettle();
        await tapItem(tester, find.byKey(const Key('sort-sortField')));
        expect(go.callsTo('browse.Search').last, ['', 'sortField', false]);
        double top(String text) => tester.getTopLeft(find.text(text)).dy;
        expect(top('gato'), lessThan(top('perro')));
        expect(top('perro'), lessThan(top('agua')));
      });

      testWidgets('sort by lapses, descending (spec, $layout)', (tester) async {
        final go = FakeGo();
        FakeBrowse(go, [
          row(1, 'comer', lapses: 1),
          row(2, 'perro', lapses: 7),
          row(3, 'gato', lapses: 3),
        ]);
        await openBrowse(tester, size: size);
        if (wide) {
          // Show Lapses from a header's menu, then click its header twice.
          await columnMenu(tester, 'due');
          await tapItem(tester, find.byKey(const Key('show-column-lapses')));
          expect(headers(tester), [
            'question',
            'deck',
            'due',
            'lapses',
            'noteType',
          ]);
          expect(cell(tester, 2, 'lapses'), '7');
          await tester.tap(find.byKey(const Key('column-lapses')));
          await tester.pumpAndSettle();
          await tester.tap(find.byKey(const Key('column-lapses')));
          await tester.pumpAndSettle();
          expect(
            tester.widget<Icon>(find.byKey(const Key('sort-indicator'))).icon,
            Icons.arrow_downward,
          );
        } else {
          // Show Lapses in the Columns dialog, then sort by it twice.
          await tester.tap(find.byTooltip('Columns'));
          await tester.pumpAndSettle();
          await tapItem(tester, find.byKey(const Key('columns-lapses')));
          await tester.tap(find.text('Save'));
          await tester.pumpAndSettle();
          for (var i = 0; i < 2; i++) {
            await tester.tap(find.byTooltip('Sort'));
            await tester.pumpAndSettle();
            await tapItem(tester, find.byKey(const Key('sort-lapses')));
          }
        }
        expect(go.callsTo('browse.Search').last, ['', 'lapses', true]);
        expect(go.settings['browserColumns'], contains('lapses'));
        // The card with the most lapses first.
        final first = tester.getTopLeft(find.text('perro')).dy;
        expect(first, lessThan(tester.getTopLeft(find.text('gato')).dy));
        expect(
          tester.getTopLeft(find.text('gato')).dy,
          lessThan(tester.getTopLeft(find.text('comer')).dy),
        );
      });

      testWidgets('the layout is remembered (spec, $layout)', (tester) async {
        final go = FakeGo();
        FakeBrowse(go, [row(1, 'comer', state: 'review', interval: 12)]);
        await openBrowse(tester, size: size);
        if (wide) {
          // Hide Note type, add Interval, from the header menu.
          await columnMenu(tester, 'noteType');
          await tester.tap(find.text('Hide column'));
          await tester.pumpAndSettle();
          await columnMenu(tester, 'due');
          await tapItem(tester, find.byKey(const Key('show-column-interval')));
        } else {
          await tester.tap(find.byTooltip('Columns'));
          await tester.pumpAndSettle();
          await tapItem(tester, find.byKey(const Key('columns-noteType')));
          await tapItem(tester, find.byKey(const Key('columns-interval')));
          await tester.tap(find.text('Save'));
          await tester.pumpAndSettle();
        }
        const want = ['question', 'deck', 'due', 'interval'];
        expect(go.settings['browserColumns'], want);

        // Restart the app.
        await tester.pumpWidget(const SizedBox());
        await openBrowse(tester, size: size);
        expect(go.callsTo('browse.Rows').last[1], want);
        if (wide) {
          expect(headers(tester), want);
          expect(cell(tester, 1, 'interval'), '12 days');
        } else {
          await tester.tap(find.byTooltip('Columns'));
          await tester.pumpAndSettle();
          final checked = [
            for (final key in [
              'question',
              'deck',
              'due',
              'noteType',
              'interval',
            ])
              if (tester
                  .widget<CheckboxListTile>(find.byKey(Key('columns-$key')))
                  .value!)
                key,
          ];
          expect(checked, ['question', 'deck', 'due', 'interval']);
        }
      });
    }

    testWidgets('a narrow row shows the first column as title, the next two '
        'as subtitle', (tester) async {
      final go = FakeGo();
      go.settings['browserColumns'] = ['answer', 'tags', 'lapses', 'deck'];
      FakeBrowse(go, [
        row(1, 'comer', answer: 'to eat', tags: 'verb', lapses: 2),
      ]);
      await openBrowse(tester);
      expect(find.text('to eat'), findsOneWidget);
      expect(find.text('verb · 2'), findsOneWidget);
      expect(find.text('comer'), findsNothing);
    });

    testWidgets('the Columns dialog reorders by dragging and keeps one '
        'column', (tester) async {
      final go = FakeGo();
      go.settings['browserColumns'] = ['question', 'deck'];
      FakeBrowse(go, [row(1, 'comer')]);
      await openBrowse(tester, size: desktop);
      await tester.tap(find.byTooltip('Columns'));
      await tester.pumpAndSettle();
      // Drag Deck above Question.
      final handle = find.descendant(
        of: find.byKey(const Key('columns-deck')),
        matching: find.byIcon(Icons.drag_handle),
      );
      final gesture = await tester.startGesture(tester.getCenter(handle));
      for (var i = 0; i < 5; i++) {
        await gesture.moveBy(const Offset(0, -20));
        await tester.pump();
      }
      await gesture.up();
      await tester.pumpAndSettle();
      expect(
        tester.getTopLeft(find.byKey(const Key('columns-deck'))).dy,
        lessThan(
          tester.getTopLeft(find.byKey(const Key('columns-question'))).dy,
        ),
      );
      await tester.tap(find.text('Save'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.SetBrowserColumns').single, [
        ['deck', 'question'],
      ]);
      expect(headers(tester), ['deck', 'question']);

      // Hiding Question leaves Deck, which cannot be hidden.
      await tester.tap(find.byTooltip('Columns'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('columns-question')));
      await tester.pumpAndSettle();
      expect(
        tester
            .widget<CheckboxListTile>(find.byKey(const Key('columns-deck')))
            .onChanged,
        isNull,
      );
    });

    testWidgets('columns move left and right from the header menu', (
      tester,
    ) async {
      final go = FakeGo();
      FakeBrowse(go, [row(1, 'comer')]);
      await openBrowse(tester, size: desktop);
      await columnMenu(tester, 'deck');
      await tester.tap(find.text('Move left'));
      await tester.pumpAndSettle();
      expect(headers(tester), ['deck', 'question', 'due', 'noteType']);
      await columnMenu(tester, 'question');
      await tester.tap(find.text('Move right'));
      await tester.pumpAndSettle();
      expect(headers(tester), ['deck', 'due', 'question', 'noteType']);
    });
  });

  group('Find and replace', () {
    Future<void> openForSelection(WidgetTester tester) async {
      await tester.longPress(find.text('recieve 0'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Select all'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Bulk actions'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Find and replace'));
      await tester.pumpAndSettle();
    }

    String countText(WidgetTester tester) =>
        tester.widget<Text>(find.byKey(const Key('find-replace-count'))).data!;

    testWidgets('shows how many notes will change before replacing (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      FakeBrowse(go, [
        for (var i = 0; i < 7; i++) row(i + 1, 'recieve $i'),
        for (var i = 7; i < 10; i++) row(i + 1, 'other $i'),
      ]);
      await openBrowse(tester);
      await openForSelection(tester);
      expect(countText(tester), 'Enter the text to find.');
      await tester.enterText(
        find.byKey(const Key('find-replace-find')),
        'RECIEVE',
      );
      await tester.enterText(
        find.byKey(const Key('find-replace-replace')),
        'receive',
      );
      await tester.pumpAndSettle(const Duration(milliseconds: 400));
      expect(countText(tester), '7 notes will change.');
      // Ignore case is on by default; turning it off changes the count.
      await tester.tap(find.text('Ignore case'));
      await tester.pumpAndSettle();
      expect(countText(tester), 'No notes will change.');
      expect(
        tester
            .widget<FilledButton>(find.byKey(const Key('find-replace-apply')))
            .onPressed,
        isNull,
      );
      await tester.tap(find.text('Ignore case'));
      await tester.pumpAndSettle();
      expect(countText(tester), '7 notes will change.');

      await tester.tap(find.byKey(const Key('find-replace-apply')));
      await tester.pumpAndSettle();
      final call = go.callsTo('browse.FindReplace').single;
      expect((call[0] as List).toSet(), {for (var i = 1; i <= 10; i++) i});
      expect(call[1], isFalse);
      expect(call[2], {
        'find': 'RECIEVE',
        'replace': 'receive',
        'field': '',
        'tags': false,
        'regex': false,
        'ignoreCase': true,
      });
      expect(find.text('Find and replace'), findsNothing);
      expect(find.text('Changed 7 notes'), findsOneWidget);
    });

    testWidgets('an invalid regular expression is refused with a message', (
      tester,
    ) async {
      final go = FakeGo();
      FakeBrowse(go, [row(1, 'recieve 0'), row(2, 'other')]);
      await openBrowse(tester);
      await openForSelection(tester);
      await tester.tap(find.text('Treat input as regular expression'));
      await tester.enterText(find.byKey(const Key('find-replace-find')), '(');
      await tester.pumpAndSettle(const Duration(milliseconds: 400));
      expect(find.textContaining('Invalid regular expression'), findsOneWidget);
      expect(find.byKey(const Key('find-replace-count')), findsNothing);
      await tester.tap(find.byKey(const Key('find-replace-apply')));
      await tester.pumpAndSettle();
      expect(go.callsTo('browse.FindReplace'), isEmpty);
      expect(find.text('Find and replace'), findsOneWidget);
    });

    testWidgets('without a selection it covers all notes, in a chosen field', (
      tester,
    ) async {
      final go = FakeGo();
      FakeBrowse(go, [row(1, 'recieve 0'), row(2, 'recieve 1')]);
      await openBrowse(tester);
      await tester.tap(find.byTooltip('More'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Find and replace'));
      await tester.pumpAndSettle();
      final selected = tester.widget<CheckboxListTile>(
        find.byKey(const Key('find-replace-selected')),
      );
      expect(selected.value, isFalse);
      expect(selected.onChanged, isNull);
      await tester.tap(find.byKey(const Key('find-replace-in')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Back').last);
      await tester.pumpAndSettle();
      await tester.enterText(
        find.byKey(const Key('find-replace-find')),
        'recieve',
      );
      await tester.pumpAndSettle(const Duration(milliseconds: 400));
      final call = go.callsTo('browse.FindReplacePreview').last;
      expect(call[0], isEmpty);
      expect(call[1], isTrue);
      expect((call[2] as Map)['field'], 'Back');
      expect(go.callsTo('browse.FieldNames').single, [<int>[], true]);
    });
  });

  group('Change note type', () {
    /// The note types of the tests: the reversed type has two card types,
    /// and Vocab has Basic's fields in another order plus one more.
    void types(FakeGo go) => go.noteTypes = [
      noteTypeJson(10, 'Basic', ['Front', 'Back']),
      noteTypeJson(
        11,
        'Basic (and reversed card)',
        ['Front', 'Back'],
        templates: ['Card 1', 'Card 2'],
      ),
      noteTypeJson(12, 'Cloze', ['Text', 'Back Extra'], cloze: true),
      noteTypeJson(
        13,
        'Vocab',
        ['Back', 'Front', 'Notes'],
        templates: ['Card 1', 'Card 2'],
      ),
    ];

    Future<void> openChange(WidgetTester tester, String first) async {
      await tester.longPress(find.text(first));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Select all'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Bulk actions'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Change note type'));
      await tester.pumpAndSettle();
    }

    Future<void> pickTarget(WidgetTester tester, String name) async {
      final target = find.byKey(const Key('change-note-type-target'));
      await tester.ensureVisible(target);
      await tester.pumpAndSettle();
      await tester.tap(target);
      await tester.pumpAndSettle();
      await tester.tap(find.text(name).last);
      await tester.pumpAndSettle();
    }

    testWidgets('fields and card types map by name, then by position', (
      tester,
    ) async {
      final go = FakeGo();
      types(go);
      FakeBrowse(go, [row(1, 'hola'), row(2, 'gato')]);
      await openBrowse(tester);
      await openChange(tester, 'hola');
      expect(find.text('Current note type: Basic'), findsOneWidget);
      // The first other note type: Front and Back by name, Card 1 by name,
      // Card 2 has no card type to take.
      await tester.tap(find.byKey(const Key('change-note-type-apply')));
      await tester.pumpAndSettle();
      var call = go.callsTo('browse.ChangeNoteTypePreview').last;
      expect((call[0] as List).toSet(), {1, 2});
      expect(call.sublist(1), [
        11,
        [0, 1],
        [0, -1],
      ]);
      await tester.tap(find.text('Cancel').last);
      await tester.pumpAndSettle();
      expect(go.callsTo('browse.ChangeNoteType'), isEmpty);

      // Vocab: Back and Front by name in their new places, Notes takes
      // nothing (its position's field is taken by name).
      await pickTarget(tester, 'Vocab');
      await tester.tap(find.byKey(const Key('change-note-type-apply')));
      await tester.pumpAndSettle();
      call = go.callsTo('browse.ChangeNoteTypePreview').last;
      expect(call.sublist(1), [
        13,
        [1, 0, -1],
        [0, -1],
      ]);
      await tester.tap(find.text('Cancel').last);
      await tester.pumpAndSettle();

      // A card type's cards go to one card type only: choosing Card 1 for
      // Card 2 clears it from Card 1.
      await tester.ensureVisible(find.byKey(const Key('change-template-1')));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('change-template-1')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Card 1').last);
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('change-note-type-apply')));
      await tester.pumpAndSettle();
      expect(go.callsTo('browse.ChangeNoteTypePreview').last[3], [-1, 0]);
      await tester.tap(find.text('Cancel').last);
      await tester.pumpAndSettle();

      // A cloze target maps no card types.
      await pickTarget(tester, 'Cloze');
      expect(find.byKey(const Key('change-template-0')), findsNothing);
      expect(find.textContaining('same number'), findsOneWidget);
    });

    testWidgets('a mapping Go refuses is explained in the dialog', (
      tester,
    ) async {
      final go = FakeGo();
      types(go);
      go.noteTypeChange = (_) => throw '1 note would have no cards. Map the fields its card types use, or keep a card type.';
      FakeBrowse(go, [row(1, 'hola')]);
      await openBrowse(tester);
      await openChange(tester, 'hola');
      await tester.tap(find.byKey(const Key('change-note-type-apply')));
      await tester.pumpAndSettle();
      expect(find.textContaining('1 note would have no cards'), findsOneWidget);
      expect(find.text('Change note type'), findsOneWidget);
      expect(go.callsTo('browse.ChangeNoteType'), isEmpty);
    });

    testWidgets('a mixed selection is explained and nothing changes (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      types(go);
      go.cardNoteTypes = {2: 12};
      FakeBrowse(go, [row(1, 'hola'), row(2, '{{c1::x}}', cloze: true)]);
      await openBrowse(tester);
      await openChange(tester, 'hola');
      expect(
        find.text('All selected cards must use the same note type.'),
        findsOneWidget,
      );
      expect(find.text('Current note type: Basic'), findsNothing);
      expect(go.callsTo('browse.ChangeNoteTypePreview'), isEmpty);
      expect(go.callsTo('browse.ChangeNoteType'), isEmpty);
    });

    testWidgets(
      'the confirmation gives the counts; confirming applies (spec)',
      (tester) async {
        final go = FakeGo();
        types(go);
        go.cardNoteTypes = {for (var i = 1; i <= 5; i++) i: 11};
        go.noteTypeChange = (args) => noteTypeChangeJson(
          notes: 5,
          cardsDeleted: (args[3] as List).contains(1) ? 0 : 5,
          lostFields: const ['Extra'],
        );
        FakeBrowse(go, [for (var i = 1; i <= 5; i++) row(i, 'word $i')]);
        await openBrowse(tester);
        await openChange(tester, 'word 1');
        expect(
          find.text('Current note type: Basic (and reversed card)'),
          findsOneWidget,
        );
        // To Basic: Card 1 keeps Card 1, Card 2 is mapped to nothing.
        await tester.tap(find.byKey(const Key('change-note-type-apply')));
        await tester.pumpAndSettle();
        expect(find.text('Change 5 notes to Basic?'), findsOneWidget);
        expect(find.textContaining('5 cards will be deleted.'), findsOneWidget);
        expect(
          find.textContaining('The content of Extra will be lost.'),
          findsOneWidget,
        );
        expect(go.callsTo('browse.ChangeNoteType'), isEmpty);
        await tester.tap(find.widgetWithText(FilledButton, 'Change').last);
        await tester.pumpAndSettle();
        final call = go.callsTo('browse.ChangeNoteType').single;
        expect((call[0] as List).toSet(), {1, 2, 3, 4, 5});
        expect(call.sublist(1), [
          10,
          [0, 1],
          [0],
        ]);
        expect(find.text('Change note type'), findsNothing);
        expect(find.text('Changed note type of 5 notes'), findsOneWidget);
        expect(find.text('5 selected'), findsNothing);
      },
    );
  });

  group('Find duplicates', () {
    Future<void> openFindDuplicates(WidgetTester tester) async {
      await tester.tap(find.byTooltip('More'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Find duplicates'));
      await tester.pumpAndSettle();
    }

    Future<void> runFind(WidgetTester tester, {String query = ''}) async {
      await tester.enterText(
        find.byKey(const Key('find-duplicates-search')),
        query,
      );
      await tester.tap(find.byKey(const Key('find-duplicates-find')));
      await tester.pumpAndSettle();
    }

    testWidgets('lists sets in a field within a search and shows their '
        'notes (spec)', (tester) async {
      final go = FakeGo()
        ..duplicateSets = [
          {
            'content': 'perro',
            'noteIds': [1001, 1002],
          },
        ];
      FakeBrowse(go, [
        row(1, 'perro', deck: 'A'),
        row(2, 'Perro ', deck: 'A'),
        row(3, 'gato', deck: 'A'),
        row(4, 'perro', deck: 'B'),
      ]);
      // Rows have note ID card ID + 1000.
      go.handle(
        'notes.Get',
        (args) => {
          'id': args[0],
          'typeId': 10,
          'fields': ['perro', 'dog'],
          'tags': <String>[],
          'cardIds': [(args[0] as int) - 1000],
        },
      );
      await openBrowse(tester);
      await openFindDuplicates(tester);
      // Field names of all note types, each once.
      await tester.tap(find.byKey(const Key('find-duplicates-field')));
      await tester.pumpAndSettle();
      for (final f in ['Front', 'Back', 'Text', 'Back Extra']) {
        expect(find.text(f).last, findsOneWidget);
      }
      await tester.tap(find.text('Front').last);
      await tester.pumpAndSettle();

      await runFind(tester, query: 'deck:A');
      expect(go.callsTo('browse.FindDuplicates').single, ['Front', 'deck:A']);
      expect(find.text('perro'), findsOneWidget);
      expect(find.text('2 notes'), findsOneWidget);

      // Selecting the set lists its notes, whatever the default search.
      await tester.tap(find.byKey(const Key('duplicate-set-0')));
      await tester.pumpAndSettle();
      expect(find.text('Only 2 notes'), findsOneWidget);
      expect(find.text('2 cards'), findsOneWidget);
      expect(find.text('Perro '), findsOneWidget);
      expect(find.text('gato'), findsNothing);
      await tester.pageBack();
      await tester.pumpAndSettle();
      expect(find.text('Find duplicates'), findsOneWidget);

      // An invalid search says why; no duplicates says so.
      await runFind(tester, query: 'is:bogus');
      expect(find.textContaining('is:bogus: unknown state'), findsOneWidget);
      go.duplicateSets = [];
      await runFind(tester);
      expect(find.text('No duplicates found.'), findsOneWidget);
      expect(
        tester
            .widget<TextButton>(find.byKey(const Key('find-duplicates-tag')))
            .onPressed,
        isNull,
      );
    });

    testWidgets('Tag duplicates tags every note in the sets (spec)', (
      tester,
    ) async {
      final go = FakeGo()
        ..duplicateSets = [
          {
            'content': 'a',
            'noteIds': [1001, 1002, 1003],
          },
          {
            'content': 'b',
            'noteIds': [1004, 1005],
          },
          {
            'content': 'c',
            'noteIds': [1006, 1007],
          },
        ];
      FakeBrowse(go, [for (var i = 1; i <= 8; i++) row(i, 'note $i')]);
      go.handle('browse.AddTags', (_) => null);
      await openBrowse(tester);
      await openFindDuplicates(tester);
      await runFind(tester);
      expect(find.text('3 notes'), findsOneWidget);
      expect(find.text('2 notes'), findsNWidgets(2));
      await tester.tap(find.byKey(const Key('find-duplicates-tag')));
      await tester.pumpAndSettle();
      expect(go.callsTo('browse.NoteCards').single, [
        [1001, 1002, 1003, 1004, 1005, 1006, 1007],
      ]);
      expect(go.callsTo('browse.AddTags').single, [
        [1, 2, 3, 4, 5, 6, 7],
        ['duplicate'],
      ]);
      expect(find.text('Tagged 7 notes duplicate'), findsOneWidget);
    });
  });

  group('sidebar', () {
    /// A browser on a desktop with decks Spanish (with Verbs and "My deck")
    /// and French, and notes tagged [tags].
    Future<FakeGo> openSidebar(
      WidgetTester tester, {
      List<List<String>> tags = const [],
      // Tall enough to show every section.
      Size size = const Size(1280, 1400),
    }) async {
      final go = FakeGo()
        ..tree = [
          deckNode(1, 'Default'),
          deckNode(2, 'Spanish', hasChildren: true),
          deckNode(3, 'Spanish::My deck', parentId: 2),
          deckNode(4, 'Verbs'),
          deckNode(5, 'French'),
        ]
        ..noteTags = tags;
      FakeBrowse(go, [row(1, 'comer')]);
      await openBrowse(tester, size: size);
      return go;
    }

    String query(WidgetTester tester) => tester
        .widget<TextField>(find.byKey(const Key('search')))
        .controller!
        .text;

    Finder item(String key) => find.byKey(Key('sidebar-$key'));

    Future<void> withKey(
      WidgetTester tester,
      LogicalKeyboardKey key,
      Future<void> Function() action,
    ) async {
      await tester.sendKeyDownEvent(key);
      await action();
      await tester.sendKeyUpEvent(key);
    }

    Future<void> openMenu(WidgetTester tester, String key) async {
      await tester.longPress(item(key));
      await tester.pumpAndSettle();
    }

    /// Drags [from] onto [to] with a mouse, as on a desktop.
    Future<void> mouseDrag(WidgetTester tester, Finder from, Finder to) async {
      final mouse = await tester.createGesture(kind: PointerDeviceKind.mouse);
      await mouse.addPointer(location: Offset.zero);
      await tester.pump();
      await mouse.down(tester.getCenter(from));
      await tester.pump();
      final target = tester.getCenter(to);
      for (var i = 1; i <= 5; i++) {
        await mouse.moveTo(Offset.lerp(tester.getCenter(from), target, i / 5)!);
        await tester.pump();
      }
      await mouse.up();
      await tester.pumpAndSettle();
      await mouse.removePointer();
    }

    test('terms quote names with spaces and combine with the search', () {
      expect(sidebarTerm('deck', 'Spanish'), 'deck:Spanish');
      expect(sidebarTerm('deck', 'My deck'), 'deck:"My deck"');
      expect(
        sidebarTerm('note', 'Basic (and reversed card)'),
        'note:"Basic (and reversed card)"',
      );
      expect(sidebarTerm('deck', r'a\b'), r'deck:a\\b');
      expect(sidebarTerm('deck', 'say "hi"'), r'deck:"say \"hi\""');
      const term = 'tag:verb';
      expect(composeSearch('deck:A', term, SidebarSearch.replace), term);
      expect(composeSearch('', term, SidebarSearch.not), '-tag:verb');
      expect(
        composeSearch('deck:A', term, SidebarSearch.and),
        'deck:A tag:verb',
      );
      expect(
        composeSearch('deck:A', term, SidebarSearch.or),
        'deck:A or tag:verb',
      );
      // AND binds more tightly than OR: group a search with a top-level or.
      expect(
        composeSearch('deck:A or deck:B', term, SidebarSearch.not),
        '(deck:A or deck:B) -tag:verb',
      );
      expect(
        composeSearch('(deck:A or deck:B) is:due', term, SidebarSearch.and),
        '(deck:A or deck:B) is:due tag:verb',
      );
      expect(
        composeSearch('"black or white"', term, SidebarSearch.and),
        '"black or white" tag:verb',
      );
    });

    testWidgets('tags show as a tree (spec)', (tester) async {
      await openSidebar(
        tester,
        tags: [
          ['verb::irregular'],
          ['verb::regular', 'noun'],
        ],
      );
      expect(item('tag-verb'), findsOneWidget);
      expect(item('tag-noun'), findsOneWidget);
      expect(item('tag-verb::irregular'), findsNothing);
      await tester.tap(find.byKey(const Key('sidebar-expand-tag-verb')));
      await tester.pumpAndSettle();
      expect(
        find.descendant(
          of: item('tag-verb::irregular'),
          matching: find.text('irregular'),
        ),
        findsOneWidget,
      );
      expect(item('tag-verb::regular'), findsOneWidget);
      // Sections collapse.
      await tester.tap(find.byKey(const Key('sidebar-section-tags')));
      await tester.pumpAndSettle();
      expect(item('tag-verb'), findsNothing);
    });

    testWidgets('the filter narrows the sidebar, keeping parents', (
      tester,
    ) async {
      await openSidebar(
        tester,
        tags: [
          ['verb::irregular'],
          ['verb::regular'],
        ],
      );
      await tester.enterText(find.byKey(const Key('sidebar-filter')), 'IRR');
      await tester.pumpAndSettle();
      expect(item('tag-verb'), findsOneWidget);
      expect(item('tag-verb::irregular'), findsOneWidget);
      expect(item('tag-verb::regular'), findsNothing);
      expect(item('deck-spanish'), findsNothing);
      expect(item('state-is:new'), findsNothing);
      // A matching deck shows with its parent.
      await tester.enterText(find.byKey(const Key('sidebar-filter')), 'my');
      await tester.pumpAndSettle();
      expect(item('deck-spanish'), findsOneWidget);
      expect(item('deck-spanish::my deck'), findsOneWidget);
      expect(item('deck-french'), findsNothing);
      await tester.tap(find.byTooltip('Clear filter'));
      await tester.pumpAndSettle();
      expect(item('deck-french'), findsOneWidget);
      expect(item('state-is:new'), findsOneWidget);
    });

    testWidgets('clicking a deck searches it (spec)', (tester) async {
      final go = await openSidebar(tester);
      await tester.enterText(find.byKey(const Key('search')), 'comer');
      await tester.tap(item('deck-spanish'));
      await tester.pumpAndSettle();
      expect(query(tester), 'deck:Spanish');
      expect(go.callsTo('browse.Search').last[0], 'deck:Spanish');
      await tester.tap(find.byKey(const Key('sidebar-expand-deck-spanish')));
      await tester.pumpAndSettle();
      await tester.tap(item('deck-spanish::my deck'));
      await tester.pumpAndSettle();
      expect(query(tester), 'deck:"Spanish::My deck"');
      await tester.tap(item('noteType-basic (and reversed card)'));
      await tester.pumpAndSettle();
      expect(query(tester), 'note:"Basic (and reversed card)"');
      await tester.tap(item('state-is:due'));
      await tester.pumpAndSettle();
      expect(query(tester), 'is:due');
    });

    testWidgets('modifier clicks combine with the search (spec)', (
      tester,
    ) async {
      final go = await openSidebar(
        tester,
        tags: [
          ['verb'],
        ],
      );
      await tester.tap(item('deck-spanish'));
      await tester.pumpAndSettle();
      await withKey(tester, LogicalKeyboardKey.altLeft, () async {
        await tester.tap(item('tag-verb'));
      });
      await tester.pumpAndSettle();
      expect(query(tester), 'deck:Spanish -tag:verb');
      expect(go.callsTo('browse.Search').last[0], 'deck:Spanish -tag:verb');
      await withKey(tester, LogicalKeyboardKey.controlLeft, () async {
        await tester.tap(item('state-is:new'));
      });
      await tester.pumpAndSettle();
      expect(query(tester), 'deck:Spanish -tag:verb is:new');
      await withKey(tester, LogicalKeyboardKey.shiftLeft, () async {
        await tester.tap(item('deck-french'));
      });
      await tester.pumpAndSettle();
      expect(query(tester), 'deck:Spanish -tag:verb is:new or deck:French');
    });

    testWidgets('Cmd-click adds with AND on macOS', (tester) async {
      await openSidebar(tester);
      await tester.tap(item('deck-spanish'));
      await tester.pumpAndSettle();
      await withKey(tester, LogicalKeyboardKey.metaLeft, () async {
        await tester.tap(item('state-is:due'));
      });
      await tester.pumpAndSettle();
      expect(query(tester), 'deck:Spanish is:due');
    }, variant: TargetPlatformVariant.only(TargetPlatform.macOS));

    testWidgets('on a phone the sidebar is a drawer with a long-press menu', (
      tester,
    ) async {
      await openSidebar(
        tester,
        size: phone,
        tags: [
          ['verb'],
        ],
      );
      expect(item('deck-spanish'), findsNothing);
      await tester.tap(find.byTooltip('Open navigation menu'));
      await tester.pumpAndSettle();
      await tester.tap(item('deck-spanish'));
      await tester.pumpAndSettle();
      expect(query(tester), 'deck:Spanish');
      // Choosing closes the drawer.
      expect(item('deck-spanish'), findsNothing);
      await tester.tap(find.byTooltip('Open navigation menu'));
      await tester.pumpAndSettle();
      await openMenu(tester, 'tag-verb');
      await tester.tap(find.text('Exclude from search'));
      await tester.pumpAndSettle();
      expect(query(tester), 'deck:Spanish -tag:verb');
    });

    testWidgets('dragging a deck onto another moves it there (spec)', (
      tester,
    ) async {
      final go = await openSidebar(tester);
      go.handle('decks.Rename', (_) => null);
      await mouseDrag(tester, item('deck-verbs'), item('deck-spanish'));
      expect(go.callsTo('decks.Rename'), [
        [4, 'Spanish::Verbs'],
      ]);
      expect(find.text('Moved “Verbs” into “Spanish”'), findsOneWidget);
      // Not into itself, and to the top level from the section header.
      await tester.tap(find.byKey(const Key('sidebar-expand-deck-spanish')));
      await tester.pumpAndSettle();
      await mouseDrag(
        tester,
        item('deck-spanish'),
        item('deck-spanish::my deck'),
      );
      expect(go.callsTo('decks.Rename'), hasLength(1));
      await mouseDrag(
        tester,
        item('deck-spanish::my deck'),
        find.byKey(const Key('sidebar-section-decks')),
      );
      expect(go.callsTo('decks.Rename').last, [3, 'My deck']);
    });

    testWidgets('decks rename and delete from the menu as on the deck list', (
      tester,
    ) async {
      final go = await openSidebar(tester);
      go.handle('decks.Rename', (_) => null);
      go.handle(
        'decks.DeletionInfo',
        (_) => deletionInfoJson(decks: 2, cards: 12),
      );
      go.handle('decks.Delete', (_) => null);
      await openMenu(tester, 'deck-spanish');
      await tester.tap(find.text('Rename…'));
      await tester.pumpAndSettle();
      expect(find.text('Use :: to move it into another deck'), findsOneWidget);
      await tester.enterText(find.byType(TextField).last, 'Español');
      await tester.tap(find.text('Rename'));
      await tester.pumpAndSettle();
      expect(go.callsTo('decks.Rename').single, [2, 'Español']);
      await openMenu(tester, 'deck-spanish');
      await tester.tap(find.text('Delete…'));
      await tester.pumpAndSettle();
      expect(
        find.textContaining('This deletes 12 cards and 1 subdeck'),
        findsOneWidget,
      );
      await tester.tap(find.text('Delete'));
      await tester.pumpAndSettle();
      expect(go.callsTo('decks.Delete').single, [2]);
    });

    testWidgets('deleting a tag says how many notes have it (spec)', (
      tester,
    ) async {
      final go = await openSidebar(
        tester,
        tags: [
          for (var i = 0; i < 11; i++) ['old', 'keep'],
          ['old::draft'],
          ['older'],
        ],
      );
      await openMenu(tester, 'tag-old');
      await tester.tap(find.text('Delete…'));
      await tester.pumpAndSettle();
      expect(find.text('Delete the tag “old”?'), findsOneWidget);
      expect(
        find.text(
          'This removes it and its child tags from 12 notes. The notes are kept.',
        ),
        findsOneWidget,
      );
      await tester.tap(find.text('Delete'));
      await tester.pumpAndSettle();
      expect(go.callsTo('notes.DeleteTags').single, [
        ['old'],
      ]);
      expect(go.noteTags, hasLength(13));
      expect(
        go.noteTags.any((t) => t.any((x) => x.startsWith('old::'))),
        isFalse,
      );
      expect(item('tag-old'), findsNothing);
      expect(item('tag-older'), findsOneWidget);
      expect(find.text('Removed 1 tag from 12 notes'), findsOneWidget);
    });

    testWidgets('tags rename from the menu, children included', (tester) async {
      final go = await openSidebar(
        tester,
        tags: [
          ['verb::irregular'],
        ],
      );
      await openMenu(tester, 'tag-verb');
      await tester.tap(find.text('Rename…'));
      await tester.pumpAndSettle();
      await tester.enterText(find.byType(TextField).last, 'verbs');
      await tester.tap(find.text('Rename'));
      await tester.pumpAndSettle();
      expect(go.callsTo('notes.RenameTag').single, ['verb', 'verbs']);
      expect(go.noteTags, [
        ['verbs::irregular'],
      ]);
      expect(item('tag-verbs'), findsOneWidget);
    });

    testWidgets('tags move by dragging, alone or several selected', (
      tester,
    ) async {
      final go = await openSidebar(
        tester,
        tags: [
          ['noun', 'adj', 'grammar'],
        ],
      );
      await mouseDrag(tester, item('tag-noun'), item('tag-grammar'));
      expect(go.callsTo('notes.MoveTags').single, [
        ['noun'],
        'grammar',
      ]);
      expect(go.noteTags.single, ['grammar::noun', 'adj', 'grammar']);
      expect(find.text('Moved 1 tag under “grammar”'), findsOneWidget);
      // Select two and drag them together.
      go.noteTags = [
        ['noun', 'adj', 'grammar'],
      ];
      go.emitChange(notes: true);
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('sidebar-tags-menu')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Select tags'));
      await tester.pumpAndSettle();
      await tester.tap(item('tag-noun'));
      await tester.tap(item('tag-adj'));
      await tester.pumpAndSettle();
      expect(find.text('2 tags selected'), findsOneWidget);
      // Clicking picks instead of searching.
      expect(query(tester), '');
      await mouseDrag(tester, item('tag-adj'), item('tag-grammar'));
      expect((go.callsTo('notes.MoveTags').last[0] as List).toSet(), {
        'noun',
        'adj',
      });
      expect(go.noteTags.single, ['grammar::noun', 'grammar::adj', 'grammar']);
      expect(find.byKey(const Key('sidebar-tag-selection')), findsNothing);
    });

    testWidgets('selected tags move and delete from the selection bar', (
      tester,
    ) async {
      final go = await openSidebar(
        tester,
        tags: [
          ['noun', 'adj', 'grammar::word'],
        ],
      );
      await openMenu(tester, 'tag-noun');
      await tester.tap(find.text('Select tags'));
      await tester.pumpAndSettle();
      await tester.tap(item('tag-adj'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Move tags to'));
      await tester.pumpAndSettle();
      // Not into the selected tags themselves.
      expect(
        find.descendant(
          of: find.byType(SimpleDialog),
          matching: find.text('noun'),
        ),
        findsNothing,
      );
      await tester.tap(
        find.descendant(
          of: find.byType(SimpleDialog),
          matching: find.text('word'),
        ),
      );
      await tester.pumpAndSettle();
      expect(go.callsTo('notes.MoveTags').single[1], 'grammar::word');
      expect(go.noteTags.single, [
        'grammar::word::noun',
        'grammar::word::adj',
        'grammar::word',
      ]);
      await tester.tap(find.byKey(const Key('sidebar-tags-menu')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Select tags'));
      await tester.pumpAndSettle();
      await tester.tap(item('tag-grammar'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Delete tags'));
      await tester.pumpAndSettle();
      expect(find.text('Delete the tag “grammar”?'), findsOneWidget);
      await tester.tap(find.text('Delete'));
      await tester.pumpAndSettle();
      expect(go.noteTags.single, isEmpty);
      expect(find.text('No tags yet.'), findsOneWidget);
    });

    testWidgets('Clear unused tags reports what it cleared', (tester) async {
      final go = await openSidebar(
        tester,
        tags: [
          ['verb'],
        ],
      );
      await tester.tap(find.byKey(const Key('sidebar-tags-menu')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Clear unused tags'));
      await tester.pumpAndSettle();
      expect(go.callsTo('notes.ClearUnusedTags'), hasLength(1));
      expect(find.text('There are no unused tags.'), findsOneWidget);
    });
  });

  group('saved searches', () {
    String query(WidgetTester tester) => tester
        .widget<TextField>(find.byKey(const Key('search')))
        .controller!
        .text;

    Future<void> search(WidgetTester tester, String q) async {
      await tester.enterText(find.byKey(const Key('search')), q);
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pumpAndSettle();
    }

    Future<void> openMenu(WidgetTester tester) async {
      await tester.tap(find.byKey(const Key('saved-searches')));
      await tester.pumpAndSettle();
    }

    final spanish = [
      row(1, 'hablar', deck: 'Spanish', tags: 'leech'),
      row(2, 'comer', deck: 'Spanish'),
      row(3, 'manger', deck: 'French', tags: 'leech'),
    ];

    testWidgets('Save and run: a saved search fills the field and runs', (
      tester,
    ) async {
      final go = FakeGo();
      FakeBrowse(go, spanish);
      await openBrowse(tester);
      await search(tester, 'deck:Spanish tag:leech');
      await openMenu(tester);
      await tester.tap(find.byKey(const Key('save-search')));
      await tester.pumpAndSettle();
      expect(find.text('deck:Spanish tag:leech'), findsWidgets);
      await tester.enterText(
        find.byKey(const Key('saved-search-name')),
        'Spanish leeches',
      );
      await tester.tap(find.widgetWithText(FilledButton, 'Save'));
      await tester.pumpAndSettle();
      expect(go.savedSearches, [
        {'name': 'Spanish leeches', 'query': 'deck:Spanish tag:leech'},
      ]);

      await search(tester, '');
      expect(find.text('3 cards'), findsOneWidget);
      await openMenu(tester);
      await tester.tap(find.byKey(const Key('saved-search-Spanish leeches')));
      await tester.pumpAndSettle();
      expect(query(tester), 'deck:Spanish tag:leech');
      expect(go.callsTo('browse.Search').last[0], 'deck:Spanish tag:leech');
      expect(find.text('1 card'), findsOneWidget);
      expect(find.text('hablar'), findsOneWidget);
      expect(find.text('manger'), findsNothing);
    });

    testWidgets('Invalid search: saving is refused with the search error', (
      tester,
    ) async {
      final go = FakeGo();
      FakeBrowse(go, spanish);
      await openBrowse(tester);
      await search(tester, 'is:bogus');
      await openMenu(tester);
      await tester.tap(find.byKey(const Key('save-search')));
      await tester.pumpAndSettle();
      await tester.enterText(
        find.byKey(const Key('saved-search-name')),
        'Bogus',
      );
      await tester.tap(find.widgetWithText(FilledButton, 'Save'));
      await tester.pumpAndSettle();
      expect(go.callsTo('browse.SaveSearch').single, ['Bogus', 'is:bogus']);
      expect(go.savedSearches, isEmpty);
      // The dialog stays open with Go's error, spelled as Go spells it.
      expect(
        find.descendant(
          of: find.byType(AlertDialog),
          matching: find.text(FakeGo.bogusSearchError),
        ),
        findsOneWidget,
      );
    });

    testWidgets('an empty search cannot be saved; a taken name is refused', (
      tester,
    ) async {
      final go = FakeGo()
        ..savedSearches = [
          {'name': 'Leeches', 'query': 'tag:leech'},
        ];
      FakeBrowse(go, spanish);
      await openBrowse(tester);
      await openMenu(tester);
      // Nothing to save while the search is empty.
      expect(
        tester
            .widget<PopupMenuItem<Object?>>(
              find.byKey(const Key('save-search')),
            )
            .enabled,
        isFalse,
      );
      await tester.tapAt(Offset.zero);
      await tester.pumpAndSettle();

      await search(tester, 'deck:French');
      await openMenu(tester);
      await tester.tap(find.byKey(const Key('save-search')));
      await tester.pumpAndSettle();
      await tester.enterText(
        find.byKey(const Key('saved-search-name')),
        'leeches',
      );
      await tester.tap(find.widgetWithText(FilledButton, 'Save'));
      await tester.pumpAndSettle();
      expect(
        find.text('A saved search named "leeches" already exists.'),
        findsOneWidget,
      );
      expect(go.savedSearches, hasLength(1));
    });

    testWidgets('saved searches can be renamed and deleted', (tester) async {
      final go = FakeGo()
        ..savedSearches = [
          {'name': 'French', 'query': 'deck:French'},
          {'name': 'Leeches', 'query': 'tag:leech'},
        ];
      FakeBrowse(go, spanish);
      await openBrowse(tester);
      await openMenu(tester);
      await tester.tap(find.byKey(const Key('manage-saved-searches')));
      await tester.pumpAndSettle();

      await tester.tap(find.byKey(const Key('rename-Leeches')));
      await tester.pumpAndSettle();
      await tester.enterText(
        find.byKey(const Key('saved-search-name')),
        'All leeches',
      );
      await tester.tap(find.widgetWithText(FilledButton, 'Save'));
      await tester.pumpAndSettle();
      expect(go.callsTo('browse.RenameSavedSearch').single, [
        'Leeches',
        'All leeches',
      ]);
      expect(find.byKey(const Key('manage-All leeches')), findsOneWidget);

      await tester.tap(find.byKey(const Key('delete-French')));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(FilledButton, 'Delete'));
      await tester.pumpAndSettle();
      expect(go.savedSearches, [
        {'name': 'All leeches', 'query': 'tag:leech'},
      ]);
      expect(find.byKey(const Key('manage-French')), findsNothing);
      await tester.tap(find.text('Done'));
      await tester.pumpAndSettle();

      await openMenu(tester);
      expect(find.byKey(const Key('saved-search-All leeches')), findsOneWidget);
      expect(find.byKey(const Key('saved-search-French')), findsNothing);
    });

    testWidgets('Start filtered: the browser opens with the default search', (
      tester,
    ) async {
      final go = FakeGo()..settings['defaultSearch'] = 'is:due';
      FakeBrowse(go, [
        row(1, 'hablar', state: 'review'),
        row(2, 'comer'),
        row(3, 'vivir', state: 'review'),
      ]);
      await openBrowse(tester);
      expect(query(tester), 'is:due');
      expect(go.callsTo('browse.Search').first[0], 'is:due');
      expect(find.text('2 cards'), findsOneWidget);
      expect(find.text('comer'), findsNothing);
    });
  });

  group('selection commands', () {
    /// Whether row [id]'s checkbox is ticked in the wide table.
    bool ticked(WidgetTester tester, int id) => tester
        .widget<Checkbox>(
          find.descendant(
            of: find.byKey(Key('row-$id')),
            matching: find.byType(Checkbox),
          ),
        )
        .value!;

    Future<void> selectionMenu(WidgetTester tester, String item) async {
      await tester.tap(find.byTooltip('Selection'));
      await tester.pumpAndSettle();
      await tester.tap(find.text(item));
      await tester.pumpAndSettle();
    }

    testWidgets('Invert selects the other 7 of 10 cards (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      FakeBrowse(go, [for (var i = 1; i <= 10; i++) row(i, 'card $i')]);
      await openBrowse(tester, size: desktop);
      await tester.longPress(find.text('card 1'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('card 4'));
      await tester.tap(find.text('card 9'));
      await tester.pumpAndSettle();
      expect(find.text('3 selected'), findsOneWidget);

      await selectionMenu(tester, 'Invert selection');
      expect(find.text('7 selected'), findsOneWidget);
      expect(
        [
          for (var i = 1; i <= 10; i++)
            if (ticked(tester, i)) i,
        ],
        [2, 3, 5, 6, 7, 8, 10],
      );

      // Ctrl+Alt+S inverts it back.
      await tester.sendKeyDownEvent(LogicalKeyboardKey.controlLeft);
      await tester.sendKeyDownEvent(LogicalKeyboardKey.altLeft);
      await tester.sendKeyEvent(LogicalKeyboardKey.keyS);
      await tester.sendKeyUpEvent(LogicalKeyboardKey.altLeft);
      await tester.sendKeyUpEvent(LogicalKeyboardKey.controlLeft);
      await tester.pumpAndSettle();
      expect(find.text('3 selected'), findsOneWidget);
      expect(ticked(tester, 4), isTrue);
    });

    testWidgets('Select notes adds the reverse card of the note (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      FakeBrowse(go, [
        row(1, 'gato', noteId: 7),
        row(2, 'cat', noteId: 7),
        row(3, 'perro'),
      ]);
      await openBrowse(tester, size: desktop);
      await tester.longPress(find.text('gato'));
      await tester.pumpAndSettle();
      expect(find.text('1 selected'), findsOneWidget);

      await selectionMenu(tester, 'Select notes');
      expect(go.callsTo('browse.CardsOfNotes').single, [
        [1],
      ]);
      expect(find.text('2 selected'), findsOneWidget);
      expect(ticked(tester, 2), isTrue);
      expect(ticked(tester, 3), isFalse);
    });

    testWidgets('Select notes needs a selection and is not in Notes mode', (
      tester,
    ) async {
      final go = FakeGo();
      FakeBrowse(go, [row(1, 'gato')]);
      await openBrowse(tester, size: desktop);
      await tester.tap(find.byTooltip('Select'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Selection'));
      await tester.pumpAndSettle();
      expect(
        tester
            .widget<PopupMenuItem<VoidCallback>>(
              find.widgetWithText(PopupMenuItem<VoidCallback>, 'Select notes'),
            )
            .enabled,
        isFalse,
      );
      await tester.sendKeyEvent(LogicalKeyboardKey.escape);
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Clear selection'));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Notes'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Select'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Selection'));
      await tester.pumpAndSettle();
      expect(find.text('Invert selection'), findsOneWidget);
      // The title says "Select notes"; the menu has no such item.
      expect(
        find.widgetWithText(PopupMenuItem<VoidCallback>, 'Select notes'),
        findsNothing,
      );
    });
  });

  testWidgets('the list keeps the width it was dragged to (spec)', (
    tester,
  ) async {
    final go = FakeGo();
    FakeBrowse(go, [row(1, 'uno')]);
    await openBrowse(tester, size: desktop);
    double listWidth() =>
        tester.getSize(find.byKey(const Key('browse-list'))).width;
    final before = listWidth();
    await tester.drag(
      find.byKey(const Key('browse-split')),
      const Offset(120, 0),
    );
    await tester.pumpAndSettle();
    final after = listWidth();
    expect(after, closeTo(before + 120, 1));
    expect(go.callsTo('app.SetBrowserSplit'), hasLength(1));
    expect(go.settings['browserSplit'] as double, greaterThan(0.6));

    // Reopen the browser after a restart.
    await tester.pumpWidget(const SizedBox());
    await openBrowse(tester, size: desktop);
    expect(listWidth(), closeTo(after, 0.01));

    // It cannot be dragged past the edge.
    await tester.drag(
      find.byKey(const Key('browse-split')),
      const Offset(2000, 0),
    );
    await tester.pumpAndSettle();
    expect(go.settings['browserSplit'], 0.8);
  });

  group('keyboard navigation', () {
    List<Map<String, Object?>> five() => [
      for (var i = 1; i <= 5; i++) row(i, 'card $i'),
    ];

    /// The card the details show.
    int shown(FakeGo go) => go.callsTo('browse.Card').last[0] as int;

    Future<void> key(
      WidgetTester tester,
      LogicalKeyboardKey key, {
      bool shift = false,
      bool control = false,
    }) async {
      if (shift) await tester.sendKeyDownEvent(LogicalKeyboardKey.shiftLeft);
      if (control) {
        await tester.sendKeyDownEvent(LogicalKeyboardKey.controlLeft);
      }
      await tester.sendKeyEvent(key);
      if (control) await tester.sendKeyUpEvent(LogicalKeyboardKey.controlLeft);
      if (shift) await tester.sendKeyUpEvent(LogicalKeyboardKey.shiftLeft);
      await tester.pumpAndSettle();
    }

    testWidgets('Down twice from the first row shows the third (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      FakeBrowse(go, five());
      await openBrowse(tester, size: desktop);
      await tester.tap(find.text('card 1'));
      await tester.pumpAndSettle();
      expect(shown(go), 1);
      await key(tester, LogicalKeyboardKey.arrowDown);
      await key(tester, LogicalKeyboardKey.arrowDown);
      expect(shown(go), 3);
      expect(
        tester.widget<Material>(find.byKey(const Key('row-3'))).color,
        Theme.of(tester.element(find.byKey(const Key('row-3'))))
            .colorScheme
            .secondaryContainer,
      );
      await key(tester, LogicalKeyboardKey.arrowUp);
      expect(shown(go), 2);
      expect(find.text('Browse'), findsWidgets, reason: 'no selection mode');
    });

    testWidgets('Shift extends the selection; Home and End jump', (
      tester,
    ) async {
      final go = FakeGo();
      FakeBrowse(go, five());
      await openBrowse(tester, size: desktop);
      await tester.tap(find.text('card 2'));
      await tester.pumpAndSettle();
      await key(tester, LogicalKeyboardKey.arrowDown, shift: true);
      await key(tester, LogicalKeyboardKey.arrowDown, shift: true);
      expect(find.text('3 selected'), findsOneWidget);
      expect(shown(go), 4);
      await key(tester, LogicalKeyboardKey.arrowUp, shift: true);
      expect(find.text('2 selected'), findsOneWidget);

      await key(tester, LogicalKeyboardKey.end);
      expect(shown(go), 5);
      await key(tester, LogicalKeyboardKey.home);
      expect(shown(go), 1);
      // Moving without Shift keeps the selection.
      expect(find.text('2 selected'), findsOneWidget);
    });

    testWidgets('a long list scrolls to the current row', (tester) async {
      final go = FakeGo();
      FakeBrowse(go, [for (var i = 1; i <= 60; i++) row(i, 'card $i')]);
      await openBrowse(tester, size: desktop);
      await tester.tap(find.text('card 1'));
      await tester.pumpAndSettle();
      await key(tester, LogicalKeyboardKey.end);
      expect(shown(go), 60);
      expect(find.text('card 60'), findsOneWidget);
      expect(find.text('card 1'), findsNothing);
    });

    testWidgets('Ctrl+F focuses the search, Ctrl+Shift+F the list', (
      tester,
    ) async {
      final go = FakeGo();
      FakeBrowse(go, five());
      await openBrowse(tester, size: desktop);
      bool searchFocused() => tester
          .widget<TextField>(find.byKey(const Key('search')))
          .focusNode!
          .hasFocus;
      await tester.tap(find.text('card 1'));
      await tester.pumpAndSettle();
      expect(searchFocused(), isFalse);
      await key(tester, LogicalKeyboardKey.keyF, control: true);
      expect(searchFocused(), isTrue);
      // Arrows move in the search text, not the list.
      await key(tester, LogicalKeyboardKey.arrowDown);
      expect(shown(go), 1);

      await key(tester, LogicalKeyboardKey.keyF, control: true, shift: true);
      expect(searchFocused(), isFalse);
      await key(tester, LogicalKeyboardKey.arrowDown);
      expect(shown(go), 2);
    });

    testWidgets('Enter opens the editor of the current card', (tester) async {
      final go = FakeGo();
      FakeBrowse(go, five());
      await openBrowse(tester, size: desktop);
      await tester.tap(find.text('card 1'));
      await tester.pumpAndSettle();
      await key(tester, LogicalKeyboardKey.arrowDown);
      await key(tester, LogicalKeyboardKey.enter);
      final field = tester.widget<RichField>(
        find.byKey(const Key('details-field-0')),
      );
      expect(field.focusNode.hasFocus, isTrue);
      expect(
        fieldController(tester, 0, prefix: 'details-field').html,
        contains('card 2'),
      );
    });

    testWidgets('Delete deletes after confirmation', (tester) async {
      final go = FakeGo();
      FakeBrowse(go, five());
      await openBrowse(tester, size: desktop);
      await tester.tap(find.text('card 1'));
      await tester.pumpAndSettle();

      // The current card, when nothing is selected.
      await key(tester, LogicalKeyboardKey.delete);
      expect(find.text('Delete 1 card?'), findsOneWidget);
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(go.callsTo('browse.DeleteNotes'), isEmpty);

      // Else the selection.
      await tester.tap(find.text('card 1'));
      await tester.pumpAndSettle();
      await key(tester, LogicalKeyboardKey.arrowDown, shift: true);
      await key(tester, LogicalKeyboardKey.delete);
      expect(find.text('Delete 2 cards?'), findsOneWidget);
      await tester.tap(find.text('Delete'));
      await tester.pumpAndSettle();
      expect((go.callsTo('browse.DeleteNotes').single[0] as List).toSet(), {
        1,
        2,
      });
      expect(find.text('card 1'), findsNothing);
      expect(find.text('card 3'), findsOneWidget);
    });
  });
}
