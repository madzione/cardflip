import 'dart:async';
import 'dart:ui' show PointerDeviceKind;

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/fake_go.dart';

Map<String, Object?> event({
  String stage = 'notes',
  int done = 0,
  int total = 0,
  Map<String, Object?>? summary,
  String error = '',
}) => {
  'stage': stage,
  'done': done,
  'total': total,
  'summary': summary,
  'error': error,
};

Map<String, Object?> summary({
  int added = 0,
  int skipped = 0,
  int skippedOlder = 0,
  int cards = 0,
  int typesAdded = 0,
  int typesUpdated = 0,
  int typesMerged = 0,
  int reviews = 0,
  int media = 0,
  int comments = 0,
  List<String> warnings = const [],
}) => {
  'notesAdded': added,
  'notesUpdated': 0,
  'notesSkipped': skipped,
  'notesSkippedOlder': skippedOlder,
  'cardsAdded': cards,
  'noteTypesAdded': typesAdded,
  'noteTypesUpdated': typesUpdated,
  'noteTypesMerged': typesMerged,
  'reviews': reviews,
  'media': media,
  'commentLines': comments,
  'warnings': warnings,
};

/// Opens the import screen from the first-run hint and picks [stage].
Future<StreamController<Object?>> openImport(
  WidgetTester tester,
  FakeGo go,
  Map<String, Object?> inspection, {
  Size size = desktop,
  FakePicker? picker,
}) async {
  picker = (picker ?? FakePicker())
    ..next = stageInfo(
      4,
      (inspection['stage'] as Map)['name'] as String,
      (inspection['stage'] as Map)['kind'] as String,
    );
  go.handle('transfer.Inspect', (_) => inspection);
  final events = StreamController<Object?>();
  go.bridge.receiver.stream('transfer.Import', (data) {
    go.calls
        .putIfAbsent('transfer.Import', () => [])
        .add(data as List<dynamic>);
    return events.stream;
  });
  await pumpApp(tester, picker: picker, size: size);
  await tester.tap(find.text('Import a file'));
  await tester.pumpAndSettle();
  await tester.tap(find.text('Choose a file'));
  await tester.pumpAndSettle();
  expect(picker.lastExtensions, ['csv', 'tsv', 'txt', 'apkg', 'colpkg']);
  return events;
}

Map<String, Object?> apkgInspection() => {
  'stage': {'id': 4, 'name': 'Spanish.apkg', 'kind': 'apkg', 'size': 3000},
  'csv': null,
  'apkg': {
    'format': 'anki21b',
    'notes': 2,
    'cards': 2,
    'decks': ['Spanish'],
    'media': 0,
    'unsupported': 0,
  },
  'backup': null,
};

Map<String, Object?> csvPreview({
  String delimiter = 'tab',
  List<List<String>> rows = const [
    ['hola', 'hello', 'verb'],
    ['perro', 'dog', 'noun'],
  ],
  bool allowHtml = true,
  List<String> tags = const [],
  List<int>? mapping,
  String deckName = '',
  Map<String, Object?>? noteTypes,
  List<String> warnings = const [],
}) {
  final columns = rows.fold(0, (n, r) => r.length > n ? r.length : n);
  return {
    'delimiter': delimiter,
    'columns': columns,
    'rows': rows,
    'totalRows': rows.length,
    'headerRow': false,
    'columnNames': <String>[],
    'typeId': 10,
    'deckId': 1,
    'deckName': deckName,
    'mapping': mapping ?? [for (var i = 0; i < columns; i++) i < 2 ? i : -1],
    'noteTypes':
        noteTypes ??
        {
          'unknown': <String>[],
          'unknownRows': 0,
          'fields': ['Front', 'Back'],
        },
    'warnings': warnings,
    'allowHtml': allowHtml,
    'tags': tags,
  };
}

Map<String, Object?> csvInspection([Map<String, Object?>? csv]) => {
  'stage': {'id': 4, 'name': 'words.tsv', 'kind': 'csv', 'size': 30},
  'csv': csv ?? csvPreview(),
  'apkg': null,
  'backup': null,
};

Map<String, Object?> colpkgInspection() => {
  'stage': {
    'id': 4,
    'name': 'collection-2026-10-01.colpkg',
    'kind': 'colpkg',
    'size': 3000,
  },
  'csv': null,
  'apkg': {
    'format': 'anki21b',
    'notes': 8,
    'cards': 12,
    'decks': ['Default', 'Fixture', 'Fixture::Geo', 'Inbox'],
    'media': 4,
    'unsupported': 0,
  },
  'backup': null,
};

/// Taps "Replace collection" and answers the confirmation with [answer].
Future<void> replaceCollection(WidgetTester tester, String answer) async {
  final button = find.widgetWithText(FilledButton, 'Replace collection');
  await tester.ensureVisible(button);
  await tester.pumpAndSettle();
  await tester.tap(button);
  await tester.pumpAndSettle();
  expect(find.text('Replace your collection?'), findsOneWidget);
  expect(
    find.textContaining('replaces your current collection entirely'),
    findsOneWidget,
  );
  await tester.tap(
    find.descendant(of: find.byType(AlertDialog), matching: find.text(answer)),
  );
  await tester.pump();
  await tester.pump();
}

Future<void> tapImport(WidgetTester tester) async {
  final button = find.widgetWithText(FilledButton, 'Import');
  await tester.ensureVisible(button);
  await tester.pumpAndSettle();
  await tester.tap(button);
  await tester.pump();
  await tester.pump();
}

void main() {
  testWidgets(
    'CSV: preview, default mapping to Front and Back, change mapping',
    (tester) async {
      final go = FakeGo();
      final events = await openImport(tester, go, csvInspection());
      expect(find.byKey(const Key('csv-preview')), findsOneWidget);
      expect(find.text('hola'), findsOneWidget);
      expect(find.text('dog'), findsOneWidget);
      // Two columns map to Front and Back of Basic by default.
      expect(
        find.descendant(
          of: find.byKey(const Key('map-0')),
          matching: find.text('Front'),
        ),
        findsOneWidget,
      );
      expect(
        find.descendant(
          of: find.byKey(const Key('map-1')),
          matching: find.text('Back'),
        ),
        findsOneWidget,
      );
      expect(
        find.descendant(
          of: find.byKey(const Key('map-2')),
          matching: find.text('Ignore'),
        ),
        findsOneWidget,
      );
      expect(find.text('Basic'), findsOneWidget);

      await tester.tap(find.byKey(const Key('map-2')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Tags').last);
      await tester.pumpAndSettle();
      await tester.tap(find.text('Update the existing note'));
      await tester.pumpAndSettle();
      await tapImport(tester);
      final options =
          go.callsTo('transfer.Import').single[1] as Map<String, dynamic>;
      expect(options['typeId'], 10);
      expect(options['deckId'], 1);
      expect(options['mapping'], [0, 1, -2]);
      expect(options['duplicates'], 'update');
      // The file's separator, HTML allowed, no extra tags.
      expect(options['separator'], 'tab');
      expect(options['allowHtml'], true);
      expect(options['tags'], isEmpty);
      await events.close();
    },
  );

  testWidgets('CSV: note type, deck and GUID columns from the file', (
    tester,
  ) async {
    // An Anki notes export: the header lines name the GUID, note type, deck
    // and tags columns, and a default deck that does not exist yet.
    final go = FakeGo();
    final events = await openImport(
      tester,
      go,
      csvInspection(
        csvPreview(
          rows: [
            ['g1', 'Basic', 'Spanish', 'hola', 'hello', 'verb'],
            ['g2', 'Vocab', '', 'gato', 'cat', ''],
          ],
          mapping: [-5, -3, -4, 0, 1, -2],
          deckName: 'Imported::Words',
          noteTypes: {
            'unknown': ['Vocab'],
            'unknownRows': 1,
            'fields': ['Field 1', 'Field 2'],
          },
          warnings: ['The file\'s note type "Missing" does not exist.'],
        ),
      ),
    );
    expect(
      find.text('The file\'s note type "Missing" does not exist.'),
      findsOneWidget,
    );
    // Unknown note type (spec): its rows will be skipped.
    expect(
      find.text('1 row will be skipped: there is no note type named "Vocab".'),
      findsOneWidget,
    );
    for (final (col, label) in [
      (0, 'GUID'),
      (1, 'Note type'),
      (2, 'Deck'),
      (3, 'Field 1'),
      (4, 'Field 2'),
      (5, 'Tags'),
    ]) {
      expect(
        find.descendant(
          of: find.byKey(Key('map-$col')),
          matching: find.text(label),
        ),
        findsOneWidget,
        reason: 'column $col',
      );
    }
    expect(find.text('Note type of rows without one'), findsOneWidget);
    expect(find.text('Imported::Words (new)'), findsOneWidget);
    expect(
      find.text(
        'A row whose GUID matches a note of its note type updates that note.',
      ),
      findsOneWidget,
    );

    final scope = find.byKey(const Key('import-scope'));
    await tester.ensureVisible(scope);
    await tester.pumpAndSettle();
    await tester.tap(scope);
    await tester.pumpAndSettle();
    await tester.tap(find.text('In its note type and deck').last);
    await tester.pumpAndSettle();
    await tapImport(tester);
    final options =
        go.callsTo('transfer.Import').single[1] as Map<String, dynamic>;
    expect(options['mapping'], [-5, -3, -4, 0, 1, -2]);
    expect(options['deckId'], 0);
    expect(options['deckName'], 'Imported::Words');
    expect(options['duplicates'], 'skip');
    expect(options['duplicateScope'], 'deck');
    // The summary names the note type of the skipped rows.
    events.add(
      event(
        stage: 'done',
        done: 1,
        total: 1,
        summary: summary(
          added: 1,
          skipped: 1,
          cards: 1,
          warnings: ['Skipped 1 row: there is no note type named "Vocab".'],
        ),
      ),
    );
    await events.close();
    await tester.pumpAndSettle();
    expect(
      find.text('Skipped 1 row: there is no note type named "Vocab".'),
      findsOneWidget,
    );
  });

  testWidgets('CSV: mapping a column to the note type checks its names', (
    tester,
  ) async {
    final go = FakeGo();
    go.handle('transfer.CsvNoteTypes', (args) {
      final options = args[1] as Map<String, dynamic>;
      final col = (options['mapping'] as List).indexOf(-3);
      return {
        'unknown': [
          if (col == 2) 'verb' else if (col == 1) 'hello',
          if (col == 1) 'dog',
        ],
        'unknownRows': 2,
        'fields': ['Front', 'Back'],
      };
    });
    final events = await openImport(tester, go, csvInspection());
    expect(find.byKey(const Key('import-unknown-types')), findsNothing);

    Future<void> map(int col, String target) async {
      await tester.tap(find.byKey(Key('map-$col')));
      await tester.pumpAndSettle();
      await tester.tap(find.text(target).last);
      await tester.pumpAndSettle();
    }

    await map(2, 'Note type');
    final call = go.callsTo('transfer.CsvNoteTypes').single;
    expect(call[0], 4);
    expect((call[1] as Map)['mapping'], [0, 1, -3]);
    expect(
      find.text('2 rows will be skipped: there is no note type named "verb".'),
      findsOneWidget,
    );
    expect(find.text('Note type of rows without one'), findsOneWidget);
    // Only one column holds the note type.
    await map(1, 'Note type');
    expect(
      find.descendant(
        of: find.byKey(const Key('map-2')),
        matching: find.text('Ignore'),
      ),
      findsOneWidget,
    );
    expect(
      find.text(
        '2 rows will be skipped: there is no note type named "hello", "dog".',
      ),
      findsOneWidget,
    );
    await map(1, 'Ignore');
    expect(find.byKey(const Key('import-unknown-types')), findsNothing);
    expect(find.text('Note type'), findsOneWidget);

    await tapImport(tester);
    final options =
        go.callsTo('transfer.Import').single[1] as Map<String, dynamic>;
    expect(options['mapping'], [0, -1, -1]);
    expect(options['deckName'], '');
    expect(options['duplicateScope'], 'noteType');
    await events.close();
  });

  testWidgets('CSV: a preview wider than the window scrolls sideways', (
    tester,
  ) async {
    final go = FakeGo();
    await openImport(
      tester,
      go,
      csvInspection(
        csvPreview(
          rows: [
            ['g1', 'Basic', 'Spanish', 'hola', 'hello', 'last column'],
          ],
        ),
      ),
      size: phone,
    );
    final last = find.text('last column');
    final table = find.byKey(const Key('csv-preview'));
    await tester.ensureVisible(table);
    await tester.pumpAndSettle();
    expect(tester.getRect(last).left, greaterThan(phone.width));
    // The scrollbar is always shown, so the overflow can be seen.
    final bar = tester.widget<Scrollbar>(
      find.byKey(const Key('csv-preview-scrollbar')),
    );
    expect(bar.thumbVisibility, isTrue);
    // A mouse drags the table sideways, as touch does.
    final mouse = await tester.startGesture(
      tester.getCenter(find.text('g1')),
      kind: PointerDeviceKind.mouse,
    );
    for (var i = 0; i < 10; i++) {
      await mouse.moveBy(const Offset(-100, 0));
      await tester.pump();
    }
    await mouse.up();
    await mouse.removePointer();
    await tester.pumpAndSettle();
    // The last column has come into view.
    expect(tester.getRect(last).left, lessThan(phone.width - 100));
  });

  testWidgets('CSV: choosing another separator splits the preview again', (
    tester,
  ) async {
    // Override detection (spec): detected as comma-separated, split at "|".
    final go = FakeGo();
    go.handle(
      'transfer.PreviewCsv',
      (_) => csvPreview(
        delimiter: 'pipe',
        rows: [
          ['a, x', 'b'],
          ['c, y', 'd'],
        ],
      ),
    );
    final events = await openImport(
      tester,
      go,
      csvInspection(
        csvPreview(
          delimiter: 'comma',
          rows: [
            ['a', ' x|b'],
            ['c', ' y|d'],
          ],
        ),
      ),
    );
    expect(find.text(' x|b'), findsOneWidget);
    expect(
      find.descendant(
        of: find.byKey(const Key('import-separator')),
        matching: find.text('Comma'),
      ),
      findsOneWidget,
    );
    await tester.tap(find.byKey(const Key('import-separator')));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Pipe').last);
    await tester.pumpAndSettle();
    expect(go.callsTo('transfer.PreviewCsv').single, [4, 'pipe']);
    expect(find.text(' x|b'), findsNothing);
    expect(find.text('a, x'), findsOneWidget);
    expect(find.text('d'), findsOneWidget);
    expect(find.byKey(const Key('map-1')), findsOneWidget);
    expect(find.byKey(const Key('map-2')), findsNothing);

    await tapImport(tester);
    final options =
        go.callsTo('transfer.Import').single[1] as Map<String, dynamic>;
    expect(options['separator'], 'pipe');
    expect(options['mapping'], [0, 1]);
    await events.close();
  });

  testWidgets('CSV: HTML toggle and header tags from the file', (tester) async {
    final go = FakeGo();
    final events = await openImport(
      tester,
      go,
      csvInspection(csvPreview(allowHtml: false, tags: ['chapter1', 'vocab'])),
    );
    // The file declares "#html:false": the toggle starts off.
    final toggle = find.byKey(const Key('import-allow-html'));
    expect(tester.widget<SwitchListTile>(toggle).value, isFalse);
    expect(
      find.text('Tags added to every note: chapter1 vocab'),
      findsOneWidget,
    );
    // Turning it on imports the fields as HTML.
    await tester.ensureVisible(toggle);
    await tester.pumpAndSettle();
    await tester.tap(toggle);
    await tester.pumpAndSettle();
    expect(tester.widget<SwitchListTile>(toggle).value, isTrue);
    await tapImport(tester);
    final options =
        go.callsTo('transfer.Import').single[1] as Map<String, dynamic>;
    expect(options['allowHtml'], true);
    expect(options['tags'], ['chapter1', 'vocab']);
    // Comment lines skipped are in the summary.
    events.add(
      event(
        stage: 'done',
        done: 1,
        total: 1,
        summary: summary(added: 2, cards: 2, comments: 1),
      ),
    );
    await events.close();
    await tester.pumpAndSettle();
    expect(find.text('1 comment line skipped'), findsOneWidget);
  });

  testWidgets('Anki package: the scheduling choice is required', (
    tester,
  ) async {
    final go = FakeGo();
    final events = await openImport(tester, go, {
      'stage': {'id': 4, 'name': 'Spanish.apkg', 'kind': 'apkg', 'size': 3000},
      'csv': null,
      'apkg': {
        'format': 'anki21b',
        'notes': 7,
        'cards': 9,
        'decks': ['Fixture::Spanish'],
        'media': 4,
        'unsupported': 2,
      },
      'backup': null,
    });
    expect(
      find.text('7 notes, 9 cards, 4 media files in Fixture::Spanish'),
      findsOneWidget,
    );
    expect(
      find.textContaining('2 audio or video files will be skipped'),
      findsOneWidget,
    );
    final import = find.widgetWithText(FilledButton, 'Import');
    expect(tester.widget<FilledButton>(import).onPressed, isNull);
    await tester.tap(find.text('Import all cards as new'));
    await tester.pumpAndSettle();
    expect(tester.widget<FilledButton>(import).onPressed, isNotNull);
    await tester.ensureVisible(import);
    await tester.pumpAndSettle();
    await tester.tap(import);
    await tester.pump();
    await tester.pump();
    final options = go.callsTo('transfer.Import').single[1] as Map;
    expect(options['keepScheduling'], false);
    // Existing notes and note types: updated if newer, not merged.
    expect(options['updateNotes'], 'ifNewer');
    expect(options['updateNoteTypes'], 'ifNewer');
    expect(options['mergeNoteTypes'], false);
    await events.close();
  });

  testWidgets('Anki package: update and merge options', (tester) async {
    final go = FakeGo();
    final events = await openImport(tester, go, apkgInspection(), size: phone);
    expect(find.text('Notes already in your collection'), findsOneWidget);
    expect(find.text('If the package version is newer'), findsNWidgets(2));
    await tester.tap(find.text('Keep scheduling'));
    await tester.pumpAndSettle();

    Future<void> choose(String key, String policy) async {
      await tester.ensureVisible(find.byKey(Key(key)));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(Key(key)));
      await tester.pumpAndSettle();
      await tester.tap(find.text(policy).last);
      await tester.pumpAndSettle();
    }

    await choose('update-notes', 'Never');
    await choose('update-note-types', 'Always');
    final merge = find.byKey(const Key('merge-note-types'));
    await tester.ensureVisible(merge);
    await tester.pumpAndSettle();
    await tester.tap(merge);
    await tester.pumpAndSettle();
    expect(tester.widget<CheckboxListTile>(merge).value, isTrue);
    await tapImport(tester);
    final options =
        go.callsTo('transfer.Import').single[1] as Map<String, dynamic>;
    expect(options['keepScheduling'], true);
    expect(options['updateNotes'], 'never');
    expect(options['updateNoteTypes'], 'always');
    expect(options['mergeNoteTypes'], true);
    expect(tester.takeException(), isNull);
    await events.close();
  });

  testWidgets('Anki package: the summary reports updates and note types', (
    tester,
  ) async {
    final go = FakeGo();
    final events = await openImport(tester, go, apkgInspection());
    await tester.tap(find.text('Import all cards as new'));
    await tester.pumpAndSettle();
    await tapImport(tester);
    events.add(
      event(
        stage: 'done',
        done: 1,
        total: 1,
        summary: summary(
          skipped: 3,
          skippedOlder: 2,
          typesAdded: 1,
          typesUpdated: 2,
          typesMerged: 1,
        ),
      ),
    );
    await events.close();
    await tester.pumpAndSettle();
    expect(find.text('Import complete'), findsOneWidget);
    expect(find.text('3 notes skipped'), findsOneWidget);
    expect(
      find.text('2 notes not updated: your version is newer'),
      findsOneWidget,
    );
    expect(find.text('1 note type added'), findsOneWidget);
    expect(find.text('2 note types updated'), findsOneWidget);
    expect(find.text('1 note type merged'), findsOneWidget);
    // Changing only note types is still an import to undo.
    expect(find.widgetWithText(SnackBarAction, 'Undo'), findsOneWidget);
  });

  testWidgets(
    'Anki collection: replace after confirming, settings kept, no undo',
    (tester) async {
      final go = FakeGo();
      final events = await openImport(tester, go, colpkgInspection());
      expect(
        find.text(
          'Anki collection: 8 notes, 12 cards, 4 media files in 4 decks',
        ),
        findsOneWidget,
      );
      expect(find.byKey(const Key('replace-warning')), findsOneWidget);
      expect(find.textContaining('export a backup'), findsOneWidget);
      // No scheduling choice: a collection keeps its scheduling.
      expect(find.text('Import all cards as new'), findsNothing);

      // Cancel: nothing changes.
      await replaceCollection(tester, 'Cancel');
      await tester.pumpAndSettle();
      expect(go.callsTo('transfer.Import'), isEmpty);
      expect(find.byKey(const Key('replace-warning')), findsOneWidget);

      await replaceCollection(tester, 'Replace');
      expect(go.callsTo('transfer.Import').single[0], 4);
      expect(find.text('Reading the file'), findsOneWidget);
      events.add(
        event(
          stage: 'done',
          done: 1,
          total: 1,
          summary: summary(added: 8, cards: 12, reviews: 24, media: 4),
        ),
      );
      await events.close();
      await tester.pumpAndSettle();
      expect(find.text('Collection replaced'), findsOneWidget);
      expect(find.text('8 notes added'), findsOneWidget);
      expect(find.text('24 reviews of history imported'), findsOneWidget);
      // Replacing the collection clears the history: no undo offered.
      expect(find.widgetWithText(SnackBarAction, 'Undo'), findsNothing);
    },
  );

  testWidgets('Anki collection: export a backup first', (tester) async {
    final go = FakeGo();
    final picker = FakePicker();
    go.handle(
      'transfer.ExportBackup',
      (_) => {'name': 'cardflip-2026-10-13.cardflip', 'data': 'UEsDBA=='},
    );
    await openImport(tester, go, colpkgInspection(), picker: picker);
    final export = find.text('Export a backup first');
    await tester.ensureVisible(export);
    await tester.pumpAndSettle();
    await tester.tap(export);
    await tester.pumpAndSettle();
    expect(picker.saved.keys, ['cardflip-2026-10-13.cardflip']);
    expect(go.callsTo('transfer.Import'), isEmpty);
    // The replace button is still there.
    expect(find.text('Replace collection'), findsOneWidget);
  });

  testWidgets('Anki collection: a damaged package leaves the collection', (
    tester,
  ) async {
    final go = FakeGo();
    final events = await openImport(tester, go, colpkgInspection());
    await replaceCollection(tester, 'Replace');
    events.add(
      event(error: 'not a readable Anki package: the collection is damaged'),
    );
    await events.close();
    await tester.pumpAndSettle();
    expect(find.text('The import failed'), findsOneWidget);
    expect(find.textContaining('the collection is damaged'), findsOneWidget);
    expect(find.text('Your collection was not changed.'), findsOneWidget);
  });

  testWidgets('progress follows the import stream, then the summary', (
    tester,
  ) async {
    final go = FakeGo();
    final events = await openImport(tester, go, csvInspection());
    await tapImport(tester);
    expect(find.text('Reading the file'), findsOneWidget);
    events.add(event(stage: 'notes', done: 800, total: 2000));
    await tester.pump();
    await tester.pump();
    expect(find.text('Adding notes'), findsOneWidget);
    expect(find.text('800 of 2000'), findsOneWidget);
    expect(
      tester
          .widget<LinearProgressIndicator>(
            find.byKey(const Key('import-progress')),
          )
          .value,
      closeTo(0.4, 1e-9),
    );

    events.add(
      event(
        stage: 'done',
        done: 1,
        total: 1,
        summary: summary(
          added: 2000,
          cards: 2000,
          warnings: [
            '3 audio/video files were skipped, because Cardflip cannot play their format: a.amr, b.amr, c.amr',
          ],
        ),
      ),
    );
    await events.close();
    await tester.pump();
    await tester.pump();
    await tester.pumpAndSettle();
    expect(find.text('Import complete'), findsOneWidget);
    expect(find.text('2,000 notes added'), findsOneWidget);
    expect(find.text('2,000 cards added'), findsOneWidget);
    expect(find.textContaining('review'), findsNothing);
    expect(
      find.textContaining('3 audio/video files were skipped'),
      findsOneWidget,
    );
  });

  testWidgets('review history: its stage, then its count in the summary', (
    tester,
  ) async {
    final go = FakeGo();
    final events = await openImport(tester, go, csvInspection());
    await tapImport(tester);
    events.add(event(stage: 'history', done: 3, total: 60));
    await tester.pump();
    await tester.pump();
    expect(find.text('Adding review history'), findsOneWidget);
    events.add(
      event(
        stage: 'done',
        done: 1,
        total: 1,
        summary: summary(added: 8, cards: 12, reviews: 24),
      ),
    );
    await events.close();
    await tester.pumpAndSettle();
    expect(find.text('24 reviews of history imported'), findsOneWidget);
  });

  testWidgets('a failed import explains why; nothing changed', (tester) async {
    final go = FakeGo();
    final events = await openImport(tester, go, csvInspection());
    await tapImport(tester);
    events.add(
      event(
        error: 'not a readable Anki package: the file is damaged or not a zip archive',
      ),
    );
    await events.close();
    await tester.pump();
    await tester.pump();
    await tester.pumpAndSettle();
    expect(find.text('The import failed'), findsOneWidget);
    expect(find.textContaining('the file is damaged'), findsOneWidget);
    expect(find.text('Your collection was not changed.'), findsOneWidget);
  });
}
