// The end-to-end flow shared by the Linux and web integration checks. It
// avoids dart:io so that it also runs in the browser.
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/app/bootstrap.dart';
import 'package:cardflip/features/import/pick.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/card_view.dart';

import 'rich_fields.dart';

/// A file picker that "chooses" the Anki fixture or the backup the flow
/// saved, staging them the way each platform does.
class FlowPicker extends FilePickerService {
  FlowPicker({required this.stageFixture, required this.stageBackup});

  final Future<golang.StageInfo> Function() stageFixture;
  final Future<golang.StageInfo> Function(String name, Uint8List bytes)
  stageBackup;
  Future<golang.StageInfo> Function()? _next;
  String? savedName;
  Uint8List? saved;

  void nextFixture() => _next = stageFixture;

  void nextBackup() => _next = () => stageBackup(savedName!, saved!);

  @override
  Future<golang.StageInfo?> pick(List<String> extensions) async {
    final next = _next;
    _next = null;
    return next == null ? null : next();
  }

  @override
  Future<bool> save(String fileName, Uint8List bytes) async {
    savedName = fileName;
    saved = bytes;
    return true;
  }
}

/// First launch, Basic and Cloze notes, studying with undo, the Anki
/// fixture import with media, backup export and restore, and a restart that
/// reads everything back from storage.
Future<void> runAppFlow(
  WidgetTester tester, {
  required FlowPicker picker,
  required String dataDir,
  required bool isWeb,
}) async {
  // A wide layout (side rail, details pane). In the browser the window
  // size is set by the runner instead: overriding the view there breaks
  // pointer hit testing.
  if (!isWeb) {
    tester.view.physicalSize =
        const Size(1280, 860) * tester.view.devicePixelRatio;
    addTearDown(tester.view.resetPhysicalSize);
  }

  // Go answers asynchronously (over FFI, or the browser's event loop for
  // wasm), so settling frames alone can finish before a reply arrived.
  Future<void> idle() async {
    for (var i = 0; i < 3; i++) {
      await Future<void>.delayed(const Duration(milliseconds: 40));
      await tester.pumpAndSettle();
    }
  }

  Future<void> start() async {
    await tester.pumpWidget(
      CardflipBootstrap(
        dataDir: () async => dataDir,
        isWeb: isWeb,
        picker: picker,
      ),
    );
    await idle();
  }

  // The browser may not use the requested size, so navigation is a side
  // rail or a bottom bar.
  Future<void> go(String destination) async {
    final rail = find.byType(NavigationRail);
    final nav = rail.evaluate().isNotEmpty ? rail : find.byType(NavigationBar);
    await tester.tap(
      find.descendant(of: nav, matching: find.text(destination)),
    );
    await idle();
  }

  Finder rich(String text) => find.textContaining(text, findRichText: true);

  /// Scrolls the settings list until [finder] is built and visible: the
  /// list is lazy, so rows far down do not exist before.
  Future<void> scrollSettingsTo(Finder finder) async {
    await tester.scrollUntilVisible(
      finder,
      200,
      scrollable: find.descendant(
        of: find.byKey(const Key('settings-list')),
        matching: find.byType(Scrollable),
      ),
    );
    await idle();
  }

  Future<void> settle() async {
    for (var i = 0; i < 20; i++) {
      await tester.pump(const Duration(milliseconds: 50));
    }
  }

  // Failures name the step they happened in: on the web the stack trace
  // does not reach into this file.
  var step = 'start';
  try {
    // First launch: the Default deck and the empty-state hint.
    step = 'First launch: the Default deck and the empty-state hint.';
    await start();
    // A browser that refuses persistent storage gets a one-time notice.
    final opened = await golang.app.status();
    if (!opened.persistent) {
      expect(find.text('Your browser may clear Cardflip data'), findsOneWidget);
      await tester.tap(find.text('OK'));
      await idle();
    }
    expect(find.text('Default'), findsOneWidget);
    expect(find.text('Your collection is empty'), findsOneWidget);

    // Add a Basic note and a Cloze note.
    step = 'Add a Basic note and a Cloze note.';
    await go('Add');
    await enterField(tester, 0, 'hola');
    await enterField(tester, 1, 'hello');
    await tester.tap(find.byKey(const Key('editor-submit')));
    await idle();
    expect(find.text('Card added'), findsOneWidget);
    await tester.tap(find.byKey(const Key('type-picker')));
    await idle();
    await tester.tap(find.text('Cloze').last);
    await idle();
    await enterField(tester, 0, '{{c1::Paris}} is in {{c2::France}}');
    await tester.tap(find.byKey(const Key('editor-submit')));
    await idle();

    // A daily goal of 3 cards.
    step = 'Set a daily goal of 3 cards.';
    await go('Settings');
    final goal = find.byKey(const Key('daily-goal'));
    await scrollSettingsTo(goal);
    await tester.tap(goal);
    await idle();
    final dialog = find.byType(AlertDialog);
    await tester.tap(
      find.descendant(
        of: find.byKey(const Key('daily-goal-kind')),
        matching: find.text('Cards'),
      ),
    );
    await idle();
    await tester.enterText(find.byKey(const Key('daily-goal-value')), '3');
    await tester.tap(find.descendant(of: dialog, matching: find.text('Save')));
    await idle();
    expect(
      find.descendant(of: goal, matching: find.text('3 cards a day')),
      findsOneWidget,
    );

    // Study with undo: three new cards in Default.
    step = 'Study with undo: three new cards in Default.';
    await go('Decks');
    // The deck list header: no streak yet, nothing toward the goal.
    final header = find.byKey(const Key('progress-header'));
    expect(
      find.descendant(of: header, matching: find.text('0 days')),
      findsOneWidget,
    );
    expect(
      find.descendant(of: header, matching: find.text('0 / 3 cards')),
      findsOneWidget,
    );
    final counts = find.byKey(
      ValueKey('counts-${(await golang.decks.tree()).first.id}'),
    );
    expect(
      find.descendant(of: counts, matching: find.text('3')),
      findsOneWidget,
    );
    // Choosing the deck shows its overview first.
    await tester.tap(find.text('Default'));
    await idle();
    expect(find.text('Study now'), findsOneWidget);
    expect(
      tester.widget<Text>(find.byKey(const Key('overview-new'))).data,
      '3',
    );
    // Deck progress: three new cards, none mature.
    step = 'The overview shows the deck progress before studying.';
    final deckProgress = find.byKey(const Key('deck-progress'));
    expect(
      find.descendant(of: deckProgress, matching: find.text('0% mature')),
      findsOneWidget,
    );
    expect(
      find.descendant(of: deckProgress, matching: find.text('New: 3')),
      findsOneWidget,
    );
    await tester.tap(find.text('Study now'));
    await idle();
    expect(rich('hola'), findsOneWidget);
    await tester.sendKeyEvent(LogicalKeyboardKey.space);
    await idle();
    expect(rich('hello'), findsOneWidget);
    await tester.sendKeyEvent(LogicalKeyboardKey.digit3);
    await idle();
    expect(rich('[...] is in France'), findsOneWidget);
    // Undo brings back the first card, unrevealed.
    step = 'Undo brings back the first card, unrevealed.';
    await tester.sendKeyDownEvent(LogicalKeyboardKey.controlLeft);
    await tester.sendKeyEvent(LogicalKeyboardKey.keyZ);
    await tester.sendKeyUpEvent(LogicalKeyboardKey.controlLeft);
    await idle();
    expect(rich('hola'), findsOneWidget);
    expect(find.text('Show answer'), findsOneWidget);
    await tester.tap(find.text('Show answer'));
    await idle();
    await tester.tap(find.byKey(const Key('rate-4')));
    await idle();
    // Cloze 1, then its sibling cloze 2 the same day (siblings are not
    // buried by default): done.
    step = 'Cloze 1, then the sibling cloze 2 the same day: done.';
    expect(rich('[...] is in France'), findsOneWidget);
    await tester.tap(find.text('Show answer'));
    await idle();
    expect(rich('Paris'), findsWidgets);
    await tester.tap(find.byKey(const Key('rate-4')));
    await idle();
    expect(rich('[...] is in France'), findsNothing);
    expect(rich('Paris is in [...]'), findsOneWidget);
    await tester.tap(find.text('Show answer'));
    await idle();
    await tester.tap(find.byKey(const Key('rate-4')));
    await idle();
    expect(find.text('Session complete'), findsOneWidget);
    expect(find.textContaining('You studied 3 cards'), findsOneWidget);
    // The first session of the day starts the streak and reaches the goal.
    step = 'The summary shows the streak and the goal.';
    expect(find.text('Streak extended to 1 day'), findsOneWidget);
    expect(find.text('Daily goal reached'), findsOneWidget);
    await tester.tap(find.text('Back to decks'));
    await idle();
    // Back on the overview: nothing is due and nothing is buried.
    step = 'The overview shows nothing due and nothing buried.';
    expect(find.text('Nothing is due right now.'), findsOneWidget);
    expect(find.textContaining('buried'), findsNothing);
    // The progress bar follows the session: no new cards left, and the
    // parts are Go's counts of the deck.
    step = 'The overview progress bar shows the studied cards.';
    final afterStudy = (await golang.decks.tree()).first.progress;
    expect(afterStudy.new_, 0);
    expect(afterStudy.learning + afterStudy.young + afterStudy.mature, 3);
    for (final label in [
      'New: 0',
      'Learning: ${afterStudy.learning}',
      'Young: ${afterStudy.young}',
      'Mature: ${afterStudy.mature}',
    ]) {
      expect(
        find.descendant(of: deckProgress, matching: find.text(label)),
        findsOneWidget,
        reason: label,
      );
    }
    await tester.pageBack();
    await idle();
    step = 'The deck list header shows the new streak and goal progress.';
    expect(
      find.descendant(of: header, matching: find.text('1 day')),
      findsOneWidget,
    );
    expect(
      find.descendant(of: header, matching: find.text('3 / 3 cards')),
      findsOneWidget,
    );

    // The achievements: nothing earned yet, with the session's progress.
    step = 'The achievements screen shows the progress of the session.';
    await tester.tap(find.byTooltip('More'));
    await idle();
    await tester.tap(find.text('Achievements'));
    await idle();
    expect(
      tester
          .widget<Text>(find.byKey(const Key('achievement-answers-status')))
          .data,
      '3 / 100',
    );
    expect(
      tester
          .widget<Text>(find.byKey(const Key('achievement-streak-status')))
          .data,
      '1 / 7',
    );
    expect(find.byKey(const Key('new-achievements')), findsNothing);
    await tester.pageBack();
    await idle();

    // Device sync is offered everywhere but on the web.
    step = 'Settings has a Sync section, except on the web.';
    await go('Settings');
    if (isWeb) {
      await scrollSettingsTo(find.text('Import'));
      expect(find.text('Sync'), findsNothing);
      expect(find.byKey(const Key('pair-device')), findsNothing);
    } else {
      await scrollSettingsTo(find.byKey(const Key('pair-device')));
      expect(find.text('Sync'), findsOneWidget);
    }

    // Import the modern Anki fixture, keeping its scheduling.
    step = 'Import the modern Anki fixture, keeping its scheduling.';
    picker.nextFixture();
    await go('Settings');
    await scrollSettingsTo(find.text('Import'));
    await tester.tap(find.text('Import'));
    await idle();
    await tester.tap(find.text('Choose a file'));
    await idle();
    expect(
      find.textContaining('8 notes, 12 cards, 4 media files'),
      findsOneWidget,
    );
    await tester.tap(find.text('Keep scheduling'));
    await idle();
    await tester.tap(find.widgetWithText(FilledButton, 'Import'));
    await settle();
    await idle();
    expect(find.text('Import complete'), findsOneWidget);
    expect(find.text('8 notes added'), findsOneWidget);
    expect(find.text('4 media files imported'), findsOneWidget);
    expect(find.textContaining('files were skipped'), findsNothing);
    await tester.tap(find.text('Done'));
    await idle();
    await go('Decks');
    expect(find.text('Fixture'), findsOneWidget);
    expect(find.text('Spanish'), findsOneWidget);

    // The imported image shows on its card.
    step = 'The imported image shows on its card.';
    await go('Browse');
    await tester.enterText(find.byKey(const Key('search')), 'perro');
    await tester.testTextInput.receiveAction(TextInputAction.done);
    await idle();
    // The row, not the search field.
    await tester.tap(
      find.byWidgetPredicate((w) => w is Text && w.data == 'perro'),
    );
    await settle();
    await idle();
    expect(find.byType(MediaImage), findsWidgets);
    expect(find.byType(Image), findsWidgets);
    expect(find.byKey(const Key('missing-image')), findsNothing);

    // A filtered deck in preview mode: build it from the deck list, study
    // it, empty and rebuild it. The backup below is taken while its cards
    // are in it.
    step = 'Build a filtered deck that previews the Default deck.';
    Future<int> inFiltered() async =>
        (await golang.browse.search('deck:filtered', 'created', false)).length;
    await go('Decks');
    await tester.tap(find.byTooltip('Create filtered deck'));
    await idle();
    await tester.enterText(
      find.byKey(const Key('first-search')),
      'deck:Default',
    );
    await tester.tap(find.byKey(const Key('reschedule')));
    await idle();
    await tester.tap(find.byKey(const Key('build-filtered')));
    await idle();
    expect(find.text('Filtered Deck 1'), findsOneWidget);
    expect(await inFiltered(), 3);

    step = 'Preview a card: Good returns it home unchanged.';
    await tester.tap(find.text('Filtered Deck 1'));
    await idle();
    await tester.tap(find.text('Study now'));
    await idle();
    final previewed = (await golang.study.current()).card!.cardID;
    final before = await golang.browse.card(previewed);
    await tester.tap(find.text('Show answer'));
    await idle();
    expect(find.text('End'), findsWidgets);
    await tester.tap(find.byKey(const Key('rate-3')));
    await idle();
    expect(await inFiltered(), 2);
    final after = await golang.browse.card(previewed);
    expect(after.state, before.state);
    expect(after.reviews, before.reviews);
    expect(after.deck, 'Default');
    await tester.pageBack();
    await idle();

    step = 'Empty and rebuild the filtered deck from its overview.';
    await tester.tap(find.widgetWithText(OutlinedButton, 'Empty'));
    await idle();
    expect(await inFiltered(), 0);
    await tester.tap(find.widgetWithText(OutlinedButton, 'Rebuild'));
    await idle();
    expect(await inFiltered(), 3);
    await tester.pageBack();
    await idle();

    // Export a backup.
    step = 'Export a backup.';
    await go('Settings');
    await scrollSettingsTo(find.text('Export backup'));
    await tester.tap(find.text('Export backup'));
    await settle();
    await idle();
    expect(picker.savedName, endsWith('.cardflip'));
    expect(picker.saved!.length, greaterThan(1000));

    // Change something, then restore: the change is gone.
    step = 'Change something, then restore: the change is gone.';
    await golang.decks.create('Made after the backup');
    await idle();
    picker.nextBackup();
    await scrollSettingsTo(find.text('Restore backup'));
    await tester.tap(find.text('Restore backup'));
    await idle();
    expect(find.text('Replace your collection?'), findsOneWidget);
    await tester.tap(find.widgetWithText(FilledButton, 'Restore'));
    await settle();
    await idle();
    expect(find.text('Backup restored'), findsOneWidget);
    await go('Decks');
    expect(find.text('Made after the backup'), findsNothing);
    expect(find.text('Fixture'), findsOneWidget);
    // The cards are in the filtered deck again, and can go home.
    step = 'The restored filtered deck holds its cards and empties.';
    expect(find.text('Filtered Deck 1'), findsOneWidget);
    expect(await inFiltered(), 3);
    final filtered = (await golang.decks.tree()).firstWhere((d) => d.filtered);
    await golang.decks.empty(filtered.id);
    await idle();
    expect(await inFiltered(), 0);

    // Restoring made an automatic backup first; make another one now.
    step = 'Automatic backups: before a restore, and created now.';
    expect((await golang.app.backups()).backups, hasLength(1));
    await go('Settings');
    await scrollSettingsTo(find.text('Create backup now'));
    await idle();
    await tester.tap(find.text('Create backup now'));
    await settle();
    await idle();
    expect(find.text('2 automatic backups'), findsOneWidget);
    await golang.decks.create('Made after the automatic backup');
    await idle();

    // Restart: the collection is read again from disk.
    step = 'Restart: the collection is read again from disk.';
    await tester.pumpWidget(const SizedBox());
    final status = await golang.app.reload();
    expect(status.state, 'ok');
    await start();
    expect(find.text('Your collection is empty'), findsNothing);
    expect(find.text('Fixture'), findsOneWidget);
    final ids = await golang.browse.search('', 'created', false);
    expect(ids, hasLength(15)); // 3 own cards + 12 imported
    final studied = await golang.browse.search(
      'hola deck:Default -is:new',
      'created',
      false,
    );
    expect(studied, hasLength(1)); // the answer survived
    final image = await golang.media.get_('cat.jpg');
    expect(image, isNotEmpty);

    // Automatic backups survive the restart; restoring the newest from the
    // list undoes the later change and leaves the media alone.
    step = 'Automatic backups survive a restart and restore.';
    final backups = (await golang.app.backups()).backups;
    expect(backups, hasLength(2));
    await go('Settings');
    final row = find.byKey(ValueKey(backups.first.name));
    await scrollSettingsTo(row);
    await idle();
    await tester.tap(
      find.descendant(
        of: row,
        matching: find.widgetWithText(TextButton, 'Restore'),
      ),
    );
    await idle();
    expect(find.text('Replace your collection?'), findsOneWidget);
    await tester.tap(find.widgetWithText(FilledButton, 'Restore'));
    await settle();
    await idle();
    expect(find.text('Backup restored'), findsOneWidget);
    await go('Decks');
    expect(find.text('Made after the automatic backup'), findsNothing);
    expect(find.text('Fixture'), findsOneWidget);
    expect(await golang.media.get_('cat.jpg'), isNotEmpty);

    // The storage backend meets its contract (IndexedDB on the web).
    step = 'Storage backend contract';
    expect(await golang.app.checkStorage(), isEmpty);

    // Profiles: a second collection, separate from the first.
    // Settings keeps its scroll position, maybe below the profile section.
    Future<void> showProfiles() async {
      await go('Settings');
      tester
          .state<ScrollableState>(
            find.descendant(
              of: find.byKey(const Key('settings-list')),
              matching: find.byType(Scrollable),
            ),
          )
          .position
          .jumpTo(0);
      await idle();
      await scrollSettingsTo(find.text('Switch profile'));
      await tester.tap(find.text('Switch profile'));
      await idle();
    }

    Future<void> openProfile(String id) async {
      await showProfiles();
      await tester.tap(find.byKey(Key('profile-$id')));
      await settle();
      await idle();
      // Each collection shows the storage notice once.
      if (!(await golang.app.status()).persistent &&
          find.text('OK').evaluate().isNotEmpty) {
        await tester.tap(find.text('OK'));
        await idle();
      }
    }

    Future<int> cards(String query) async =>
        (await golang.browse.search(query, 'created', false)).length;

    // Each profile remembers its last note type: Basic in a new one, Cloze
    // in the first.
    Future<void> addCard(List<String> fields) async {
      await go('Add');
      for (final (i, text) in fields.indexed) {
        await enterField(tester, i, text);
      }
      await tester.tap(find.byKey(const Key('editor-submit')));
      await idle();
      expect(find.text('Card added'), findsOneWidget);
    }

    String? profileName() {
      final name = find.byKey(const Key('profile-name'));
      return name.evaluate().isEmpty ? null : tester.widget<Text>(name).data;
    }

    step = 'Add a profile and open it: a fresh collection.';
    await showProfiles();
    await tester.tap(find.text('Add profile'));
    await idle();
    await tester.enterText(find.byKey(const Key('profile-name')), 'Ben');
    await tester.tap(find.widgetWithText(FilledButton, 'Add'));
    await idle();
    await openProfile('p2');
    await go('Decks');
    expect(find.text('Your collection is empty'), findsOneWidget);
    expect(find.text('Fixture'), findsNothing);
    expect(profileName(), 'Ben');

    step = 'Add a card in each profile.';
    await addCard(['uno', 'one']);
    expect(await cards(''), 1);
    await openProfile('default');
    await go('Decks');
    expect(find.text('Fixture'), findsOneWidget);
    expect(profileName(), 'Default');
    expect(await cards('uno'), 0);
    await addCard(['{{c1::dos}} is two']);
    expect(await cards('dos'), 1);

    step = 'Switching back: each profile keeps its own cards.';
    await openProfile('p2');
    expect(await cards(''), 1);
    expect(await cards('uno'), 1);
    expect(await cards('dos'), 0);
    expect(await cards('hola'), 0);

    step = 'Restart: the last used profile opens again.';
    await tester.pumpWidget(const SizedBox());
    expect((await golang.app.reload()).state, 'ok');
    await start();
    expect(profileName(), 'Ben');
    expect(await cards(''), 1);
    expect(await cards('uno'), 1);

    step = 'Delete a profile from another one.';
    await openProfile('default');
    await showProfiles();
    await tester.tap(find.byKey(const Key('profile-menu-p2')));
    await idle();
    await tester.tap(find.text('Delete profile'));
    await idle();
    expect(find.text('Delete “Ben”?'), findsOneWidget);
    expect(find.textContaining('with 1 card '), findsOneWidget);
    await tester.tap(find.widgetWithText(FilledButton, 'Delete'));
    await idle();
    expect(find.byKey(const Key('profile-p2')), findsNothing);
    expect((await golang.app.profiles()).profiles, hasLength(1));
    expect(await cards('dos'), 1);
    expect(await cards('hola'), greaterThan(0));
  } catch (e, st) {
    final screen = [
      for (final t in tester.widgetList<Text>(find.byType(Text)))
        if (t.data != null) t.data!,
    ].take(40).join(' | ');
    Error.throwWithStackTrace(
      TestFailure('Failed during "$step": $e\nOn screen: $screen'),
      st,
    );
  }
}
