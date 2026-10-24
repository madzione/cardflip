import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';

import 'package:cardflip/app/bootstrap.dart';
import 'package:cardflip/golang/golang.dart' as golang;

/// Browser responsiveness with 50,000 cards, against the real Go library:
/// first rows within a second, and smooth scrolling.
void main() {
  final binding = IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('browse 50,000 cards', (tester) async {
    final dir = await Directory.systemTemp.createTemp('cardflip-perf-');
    addTearDown(() => dir.delete(recursive: true));
    await tester.pumpWidget(
      CardflipBootstrap(dataDir: () async => dir.path, isWeb: false),
    );
    await tester.pumpAndSettle();

    // A synthetic collection: 50,000 Basic notes from one CSV import.
    final csv = StringBuffer();
    for (var i = 0; i < 50000; i++) {
      csv.writeln(
        'Question number $i about topic ${i % 97}\tThe answer $i, with a few more words',
      );
    }
    final stage = await golang.transfer.stageBytes(
      'synthetic.tsv',
      Uint8List.fromList(utf8.encode(csv.toString())),
    );
    final preview = (await golang.transfer.inspect(stage.id)).csv!;
    final started = DateTime.now();
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
    final summary = events.last.summary!;
    // ignore: avoid_print
    print(
      'imported ${summary.notesAdded} notes in ${DateTime.now().difference(started)}',
    );
    expect(summary.notesAdded, 50000);
    await tester.pumpAndSettle();

    // The editor's duplicate check, run on each pause in typing.
    final check = Stopwatch()..start();
    final dups = await golang.notes.duplicates(
      preview.typeID,
      'question number 4242 about topic ${4242 % 97} ',
      0,
    );
    check.stop();
    // ignore: avoid_print
    print('duplicate check in ${check.elapsedMicroseconds} µs');
    expect(dups, hasLength(1));
    expect(check.elapsed, lessThan(const Duration(milliseconds: 50)));

    // Open the browser: the first rows appear within a second.
    final opened = Stopwatch()..start();
    await tester.tap(find.text('Browse').last);
    await tester.pump();
    while (find.textContaining('Question number').evaluate().isEmpty) {
      await tester.pump(const Duration(milliseconds: 16));
      if (opened.elapsed > const Duration(seconds: 10)) {
        fail('rows did not appear');
      }
    }
    opened.stop();
    // ignore: avoid_print
    print('first rows after ${opened.elapsedMilliseconds} ms');
    expect(opened.elapsed, lessThan(const Duration(seconds: 1)));
    expect(find.text('50000 cards'), findsOneWidget);

    // Scroll through thousands of rows while recording frame timings.
    final list = find.byKey(const Key('browse-list'));
    await binding.watchPerformance(() async {
      for (var i = 0; i < 6; i++) {
        await tester.fling(list, const Offset(0, -6000), 8000);
        await tester.pumpAndSettle();
      }
      for (var i = 0; i < 3; i++) {
        await tester.fling(list, const Offset(0, 6000), 8000);
        await tester.pumpAndSettle();
      }
    }, reportKey: 'browse_scrolling');
    final report =
        binding.reportData!['browse_scrolling'] as Map<String, dynamic>;
    // ignore: avoid_print
    print('scrolling: ${const JsonEncoder.withIndent('  ').convert(report)}');
    expect(find.textContaining('Question number'), findsWidgets);

    // Sorting by a column computed from every card (the answer, rendered
    // for the first time) still shows the first rows within a second.
    final sorted = Stopwatch()..start();
    final ids = await golang.browse.search('', 'answer', true);
    final rows = await golang.browse.rows(ids.sublist(0, 100), [
      'question',
      'answer',
    ]);
    sorted.stop();
    // ignore: avoid_print
    print('sorted by answer in ${sorted.elapsedMilliseconds} ms');
    expect(rows.first.answer, 'The answer 9999, with a few more words');
    expect(sorted.elapsed, lessThan(const Duration(seconds: 1)));

    // Notes mode: switching shows the first note rows within a second too.
    final notes = Stopwatch()..start();
    await tester.tap(
      find.descendant(
        of: find.byKey(const Key('browse-mode')),
        matching: find.text('Notes'),
      ),
    );
    await tester.pump();
    while (find.text('50000 notes').evaluate().isEmpty ||
        find.textContaining('Question number').evaluate().isEmpty) {
      await tester.pump(const Duration(milliseconds: 16));
      if (notes.elapsed > const Duration(seconds: 10)) {
        fail('note rows did not appear');
      }
    }
    notes.stop();
    // ignore: avoid_print
    print('first note rows after ${notes.elapsedMilliseconds} ms');
    expect(notes.elapsed, lessThan(const Duration(seconds: 1)));

    await binding.watchPerformance(() async {
      for (var i = 0; i < 6; i++) {
        await tester.fling(list, const Offset(0, -6000), 8000);
        await tester.pumpAndSettle();
      }
    }, reportKey: 'browse_notes_scrolling');
    final notesReport =
        binding.reportData!['browse_notes_scrolling'] as Map<String, dynamic>;
    // ignore: avoid_print
    print(
      'notes scrolling: ${const JsonEncoder.withIndent('  ').convert(notesReport)}',
    );

    // Sorting notes by their first field.
    final notesSorted = Stopwatch()..start();
    final noteIds = await golang.browse.searchNotes('', 'sortField', true);
    final noteRows = await golang.browse.noteRows(noteIds.sublist(0, 100));
    notesSorted.stop();
    // ignore: avoid_print
    print(
      'notes sorted by sort field in ${notesSorted.elapsedMilliseconds} ms',
    );
    expect(noteRows.first.sortField, 'Question number 9999 about topic 8');
    expect(noteRows.first.cards, 1);
    expect(notesSorted.elapsed, lessThan(const Duration(seconds: 1)));
  });
}
