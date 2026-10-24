import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/features/decks/deck_overview_screen.dart';
import 'package:cardflip/features/decks/deck_picker.dart';
import 'package:cardflip/features/study/study_screen.dart';

import 'browse_test.dart' show FakeBrowse, row;
import 'study_test.dart' show FakeStudy, audioCard, studyCard;
import 'support/fake_go.dart';
import 'support/fake_media.dart';
import '../integration_test/rich_fields.dart';

void main() {
  layoutAndThemeTests();
  appearanceTests();
  undoTests();
  mediaTests();
  deckPickerTests();

  group('startup', () {
    testWidgets('ok: opens the app on the deck list', (tester) async {
      final go = FakeGo();
      await pumpApp(tester);
      expect(go.callsTo('app.Open'), [
        ['/data'],
      ]);
      expect(find.byType(NavigationBar), findsOneWidget);
      expect(find.text('Decks'), findsWidgets);
    });

    testWidgets('fresh: a new collection opens the app', (tester) async {
      final go = FakeGo();
      go.status = {...go.status, 'state': 'fresh'};
      await pumpApp(tester);
      expect(find.byType(NavigationBar), findsOneWidget);
    });

    testWidgets('corrupt: error screen, data kept until the user confirms', (
      tester,
    ) async {
      final go = FakeGo();
      go.status = {
        ...go.status,
        'state': 'corrupt',
        'message': 'Your collection could not be read: snapshot damaged',
      };
      await pumpApp(tester);
      expect(find.text('Your collection could not be opened'), findsOneWidget);
      expect(find.textContaining('snapshot damaged'), findsOneWidget);
      expect(find.text('Restore from backup'), findsOneWidget);
      expect(find.byType(NavigationBar), findsNothing);

      // Starting over needs an explicit confirmation; cancelling keeps data.
      await tester.tap(find.text('Start a new collection'));
      await tester.pumpAndSettle();
      expect(find.text('Start a new collection?'), findsOneWidget);
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.StartFresh'), isEmpty);

      await tester.tap(find.text('Start a new collection'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Delete and start over'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.StartFresh'), hasLength(1));
      expect(find.byType(NavigationBar), findsOneWidget);
    });

    testWidgets('corrupt: restore from a backup', (tester) async {
      final go = FakeGo();
      go.status = {...go.status, 'state': 'corrupt', 'message': 'unreadable'};
      final picker = FakePicker()
        ..next = stageInfo(7, 'backup.cardflip', 'backup');
      go.handle('transfer.Restore', (args) {
        go.status = {...go.status, 'state': 'ok', 'message': null};
        return null;
      });
      await pumpApp(tester, picker: picker);
      await tester.tap(find.text('Restore from backup'));
      await tester.pumpAndSettle();
      expect(go.callsTo('transfer.Restore'), [
        [7],
      ]);
      expect(picker.lastExtensions, ['cardflip']);
      expect(find.byType(NavigationBar), findsOneWidget);
    });

    testWidgets('opens the last used profile', (tester) async {
      final go = FakeGo()..lastProfile = 'p2';
      await pumpApp(tester);
      expect(go.callsTo('app.LoadProfiles'), [
        ['/data'],
      ]);
      expect(go.callsTo('app.Open'), [
        ['/data'],
      ]);
      expect(go.currentProfile, 'p2');
      expect(find.byType(NavigationBar), findsOneWidget);
    });

    testWidgets('the profile list first when set to show it', (tester) async {
      final go = FakeGo()
        ..showProfilesAtStartup = true
        ..profiles = [
          {'id': 'default', 'name': 'Anna'},
          {'id': 'p2', 'name': 'Ben'},
        ];
      await pumpApp(tester);
      expect(find.text('Profiles'), findsOneWidget);
      expect(find.byType(NavigationBar), findsNothing);
      expect(find.byType(BackButton), findsNothing);
      expect(go.callsTo('app.Open'), isEmpty);
      expect(find.text('Open now'), findsNothing);

      await tester.tap(find.text('Ben'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.SwitchProfile'), [
        ['p2'],
      ]);
      expect(find.byType(NavigationBar), findsOneWidget);
      expect(find.byKey(const Key('profile-name')), findsOneWidget);
    });

    testWidgets('a profile added on the start list opens fresh', (
      tester,
    ) async {
      final go = FakeGo()..showProfilesAtStartup = true;
      go.profileStatus['p2'] = {...go.status, 'state': 'fresh'};
      await pumpApp(tester);
      await tester.tap(find.text('Add profile'));
      await tester.pumpAndSettle();
      await tester.enterText(find.byKey(const Key('profile-name')), 'Ben');
      await tester.tap(find.widgetWithText(FilledButton, 'Add'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Ben'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.SwitchProfile'), [
        ['p2'],
      ]);
      expect(find.byType(NavigationBar), findsOneWidget);
    });

    testWidgets('corrupt: another profile can be opened instead', (
      tester,
    ) async {
      final go = FakeGo()
        ..profiles = [
          {'id': 'default', 'name': 'Anna'},
          {'id': 'p2', 'name': 'Ben'},
        ];
      go.status = {...go.status, 'state': 'corrupt', 'message': 'unreadable'};
      go.profileStatus['p2'] = {...go.status, 'state': 'ok', 'message': null};
      await pumpApp(tester);
      expect(find.text('Your collection could not be opened'), findsOneWidget);
      await tester.tap(find.text('Switch profile'));
      await tester.pumpAndSettle();
      expect(find.text('Profiles'), findsOneWidget);
      await tester.tap(find.text('Ben'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.SwitchProfile'), [
        ['p2'],
      ]);
      expect(find.byType(NavigationBar), findsOneWidget);
    });

    testWidgets('corrupt with one profile: no profile to switch to', (
      tester,
    ) async {
      final go = FakeGo();
      go.status = {...go.status, 'state': 'corrupt', 'message': 'unreadable'};
      await pumpApp(tester);
      expect(find.text('Your collection could not be opened'), findsOneWidget);
      expect(find.text('Switch profile'), findsNothing);
    });
  });
  group('switching profiles', () {
    testWidgets('from a note being added: the app starts over', (tester) async {
      final go = FakeGo()
        ..profiles = [
          {'id': 'default', 'name': 'Anna'},
          {'id': 'p2', 'name': 'Ben'},
        ];
      await pumpApp(tester);
      await tester.tap(
        find.descendant(
          of: find.byType(NavigationBar),
          matching: find.text('Add'),
        ),
      );
      await tester.pumpAndSettle();
      await enterField(tester, 0, 'hola');
      await tester.tap(
        find.descendant(
          of: find.byType(NavigationBar),
          matching: find.text('Settings'),
        ),
      );
      await tester.pumpAndSettle();
      await tester.scrollUntilVisible(
        find.text('Switch profile'),
        200,
        scrollable: find.byType(Scrollable).last,
      );
      await tester.tap(find.text('Switch profile'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Ben'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.SwitchProfile'), [
        ['p2'],
      ]);
      expect(tester.takeException(), isNull);
      expect(find.text('Ben'), findsOneWidget);
      expect(find.byKey(const Key('settings-list')), findsNothing);
    });
  });
  group('profile name on the deck list', () {
    testWidgets('hidden with one profile', (tester) async {
      FakeGo();
      await pumpApp(tester);
      expect(find.byKey(const Key('profile-name')), findsNothing);
    });

    testWidgets('shown with several, following renames', (tester) async {
      final go = FakeGo()
        ..profiles = [
          {'id': 'default', 'name': 'Anna'},
          {'id': 'p2', 'name': 'Ben'},
        ];
      await pumpApp(tester);
      final name = find.descendant(
        of: find.byType(AppBar),
        matching: find.byKey(const Key('profile-name')),
      );
      expect(tester.widget<Text>(name).data, 'Anna');

      final app = AppScope.read(tester.element(find.byType(NavigationBar)));
      await app.profiles.rename('default', 'Annika');
      await tester.pumpAndSettle();
      expect(go.callsTo('app.RenameProfile'), [
        ['default', 'Annika'],
      ]);
      expect(tester.widget<Text>(name).data, 'Annika');
    });
  });
  group('learning due reminder', () {
    List<Map<String, Object?>> tree({int spanish = 0, int biology = 0}) => [
      deckNode(1, 'Spanish', learning: spanish),
      deckNode(2, 'Biology', learning: biology),
    ];

    /// Starts the app, then lets learning cards become due in the decks of
    /// [after] as time passes.
    Future<void> becomeDue(
      WidgetTester tester,
      FakeGo go,
      List<Map<String, Object?>> after,
    ) async {
      // The counts change right now: the next time-driven reload is after
      // that moment.
      go.nextChange = DateTime.now;
      go.emitChange();
      await tester.pump();
      go.tree = after;
      go.nextChange = () => DateTime.now().add(const Duration(hours: 1));
      await tester.pump(const Duration(seconds: 2));
      await tester.pumpAndSettle();
    }

    void handleStudy(FakeGo go) {
      Map<String, Object?> state() => {
        'deckId': 1,
        'deckName': 'Spanish',
        'card': {
          'cardId': 7,
          'noteId': 70,
          'question': '<div>hola</div>',
          'answer': '<div>hola<hr>hello</div>',
          'kind': 'learning',
          'empty': false,
          'previews': <Object?>[],
          'autoplay': true,
          'replayQuestion': true,
          'showTimer': false,
          'stopTimerOnAnswer': true,
          'maxAnswerSeconds': 60,
          'secondsToShowQuestion': 0,
          'secondsToShowAnswer': 0,
          'questionAction': 'showAnswer',
          'answerAction': 'buryCard',
          'typeAnswer': false,
          'flag': 0,
          'marked': false,
        },
        'counts': {'new': 0, 'learning': 1, 'review': 0},
        'canUndo': false,
        'studied': 0,
        'timeSpent': 0,
        'nextDue': null,
        'learningDue': null,
        'leech': false,
        'leechSuspended': false,
      };
      go.handle('study.Start', (_) => state());
      // The card is fetched again when its render environment is known.
      go.handle('study.Current', (_) => state());
    }

    testWidgets('on another tab, Study opens the deck on the Decks tab', (
      tester,
    ) async {
      final go = FakeGo()..tree = tree();
      handleStudy(go);
      await pumpApp(tester);
      await tester.tap(find.text('Browse').last);
      await tester.pumpAndSettle();
      await becomeDue(tester, go, tree(spanish: 1));
      expect(find.text('Learning cards are due in Spanish'), findsOneWidget);

      await tester.tap(find.widgetWithText(SnackBarAction, 'Study'));
      await tester.pumpAndSettle();
      expect(go.callsTo('study.Start'), [
        [1],
      ]);
      expect(find.byType(StudyScreen), findsOneWidget);
      expect(
        tester.widget<NavigationBar>(find.byType(NavigationBar)).selectedIndex,
        0,
      );
    });

    testWidgets('several decks: one reminder that shows the deck list', (
      tester,
    ) async {
      final go = FakeGo()..tree = tree();
      await pumpApp(tester);
      await openSettings(tester);
      await becomeDue(tester, go, tree(spanish: 1, biology: 2));
      expect(find.text('Learning cards are due in 2 decks'), findsOneWidget);

      await tester.tap(find.widgetWithText(SnackBarAction, 'Show decks'));
      await tester.pumpAndSettle();
      expect(
        tester.widget<NavigationBar>(find.byType(NavigationBar)).selectedIndex,
        0,
      );
      expect(find.text('Theme'), findsNothing);
      expect(find.text('Biology'), findsOneWidget);
    });

    testWidgets('no reminder while a study screen is open', (tester) async {
      final go = FakeGo()..tree = tree();
      handleStudy(go);
      await pumpApp(tester);
      await studyDeck(tester, 'Spanish');
      expect(find.byType(StudyScreen), findsOneWidget);
      await becomeDue(tester, go, tree(spanish: 1));
      expect(find.textContaining('Learning cards are due'), findsNothing);
      expect(find.byType(SnackBar), findsNothing);
    });
  });
}

Future<void> openSettings(WidgetTester tester) async {
  await tester.tap(find.text('Settings').last);
  await tester.pumpAndSettle();
}

void layoutAndThemeTests() {
  group('layout', () {
    testWidgets('phone: bottom bar with Decks, Browse, Add and Settings', (
      tester,
    ) async {
      FakeGo();
      await pumpApp(tester, size: phone);
      final bar = find.byType(NavigationBar);
      expect(bar, findsOneWidget);
      expect(find.byType(NavigationRail), findsNothing);
      for (final label in ['Decks', 'Browse', 'Add', 'Settings']) {
        expect(
          find.descendant(of: bar, matching: find.text(label)),
          findsOneWidget,
        );
      }
    });

    testWidgets('wide: side rail from 840 px', (tester) async {
      FakeGo();
      await pumpApp(tester, size: const Size(840, 700));
      final rail = find.byType(NavigationRail);
      expect(rail, findsOneWidget);
      expect(find.byType(NavigationBar), findsNothing);
      for (final label in ['Decks', 'Browse', 'Add', 'Settings']) {
        expect(
          find.descendant(of: rail, matching: find.text(label)),
          findsOneWidget,
        );
      }
      await tester.tap(
        find.descendant(of: rail, matching: find.text('Settings')),
      );
      await tester.pumpAndSettle();
      expect(find.text('Theme'), findsOneWidget);
    });

    testWidgets('destinations keep their own screens', (tester) async {
      FakeGo();
      await pumpApp(tester);
      await openSettings(tester);
      expect(find.text('Theme'), findsOneWidget);
      await tester.tap(find.text('Decks').last);
      await tester.pumpAndSettle();
      expect(find.text('Theme'), findsNothing);
    });
  });

  group('theme', () {
    testWidgets('starts in the stored theme', (tester) async {
      final go = FakeGo();
      go.settings = {...go.settings, 'theme': 'dark'};
      await pumpApp(tester);
      final context = tester.element(find.byType(NavigationBar));
      expect(Theme.of(context).brightness, Brightness.dark);
    });

    testWidgets('the chosen theme is sent to Go and applied', (tester) async {
      final go = FakeGo();
      await pumpApp(tester);
      await openSettings(tester);
      await tester.tap(find.text('Dark'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.SetTheme'), [
        ['dark'],
      ]);
      final context = tester.element(find.byType(NavigationBar));
      expect(Theme.of(context).brightness, Brightness.dark);
    });

    testWidgets('a failed save shows an error and keeps the old theme', (
      tester,
    ) async {
      final go = FakeGo();
      go.handle(
        'app.SetTheme',
        (_) => throw 'could not save the change: disk full',
      );
      await pumpApp(tester);
      await openSettings(tester);
      await tester.tap(find.text('Light'));
      await tester.pumpAndSettle();
      expect(find.text('Could not save the change: disk full'), findsOneWidget);
      final context = tester.element(find.byType(NavigationBar));
      expect(Theme.of(context).brightness, Brightness.light);
      expect(find.byType(SnackBar), findsOneWidget);
    });
  });
}

void undoTests() {
  group('undo', () {
    testWidgets('the menu names the step: Undo Delete 3 notes', (tester) async {
      final go = FakeGo()..undoStep = 'Delete 3 notes';
      await pumpApp(tester);
      await tester.tap(find.byTooltip('More'));
      await tester.pumpAndSettle();
      expect(find.text('Undo Delete 3 notes'), findsOneWidget);
      final redo = tester.widget<PopupMenuItem<Object>>(
        find.ancestor(
          of: find.text('Redo'),
          matching: find.byWidgetPredicate((w) => w is PopupMenuItem),
        ),
      );
      expect(redo.enabled, isFalse);
      await tester.tap(find.text('Undo Delete 3 notes'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.Undo'), hasLength(1));
      expect(find.text('Undone: Delete 3 notes'), findsOneWidget);

      await tester.tap(find.byTooltip('More'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Redo Delete 3 notes'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.Redo'), hasLength(1));
      expect(find.text('Redone: Delete 3 notes'), findsOneWidget);
    });

    testWidgets('undo menu on iOS (spec)', (tester) async {
      FakeGo().undoStep = 'Delete 3 notes';
      await pumpApp(tester);
      await tester.tap(find.byTooltip('More'));
      await tester.pumpAndSettle();
      expect(find.text('Undo Delete 3 notes'), findsOneWidget);
      expect(find.text('⌘Z'), findsNothing);
      expect(find.text('Ctrl+Z'), findsNothing);
      expect(find.text('Ctrl+Shift+Z'), findsNothing);
    }, variant: TargetPlatformVariant.only(TargetPlatform.iOS));

    testWidgets('the undo menu shows the shortcuts on Linux', (tester) async {
      FakeGo().undoStep = 'Delete 3 notes';
      await pumpApp(tester);
      await tester.tap(find.byTooltip('More'));
      await tester.pumpAndSettle();
      expect(find.text('Undo Delete 3 notes'), findsOneWidget);
      expect(find.text('Ctrl+Z'), findsOneWidget);
      expect(find.text('Ctrl+Shift+Z'), findsOneWidget);
    }, variant: TargetPlatformVariant.only(TargetPlatform.linux));

    testWidgets('Ctrl+Z undoes, Ctrl+Shift+Z and Ctrl+Y redo', (tester) async {
      final go = FakeGo()..undoStep = 'Move 30 cards';
      await pumpApp(tester, size: desktop);
      await pressKeys(tester, [
        LogicalKeyboardKey.controlLeft,
        LogicalKeyboardKey.keyZ,
      ]);
      expect(go.callsTo('app.Undo'), hasLength(1));
      await pressKeys(tester, [
        LogicalKeyboardKey.controlLeft,
        LogicalKeyboardKey.shiftLeft,
        LogicalKeyboardKey.keyZ,
      ]);
      expect(go.callsTo('app.Redo'), hasLength(1));
      go.redoStep = 'Move 30 cards';
      await pressKeys(tester, [
        LogicalKeyboardKey.controlLeft,
        LogicalKeyboardKey.keyY,
      ]);
      expect(go.callsTo('app.Redo'), hasLength(2));
    });

    testWidgets('in a text field Ctrl+Z undoes the typing only', (
      tester,
    ) async {
      final go = FakeGo()..undoStep = 'Add note';
      go.handle('browse.Search', (_) => <int>[]);
      await pumpApp(tester, size: desktop);
      await tester.tap(find.text('Browse').last);
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('search')));
      await tester.enterText(find.byKey(const Key('search')), 'gato');
      await pressKeys(tester, [
        LogicalKeyboardKey.controlLeft,
        LogicalKeyboardKey.keyZ,
      ]);
      expect(go.callsTo('app.Undo'), isEmpty);
    });
  });
}

void mediaTests() {
  group('switching destinations stops playback', () {
    late FakeBackend backend;

    setUp(() {
      backend = FakeBackend();
    });

    /// Taps [label] in the side rail; the browser's sidebar has its own
    /// Decks heading.
    Future<void> selectDestination(WidgetTester tester, String label) async {
      await tester.tap(
        find.descendant(
          of: find.byType(NavigationRail),
          matching: find.text(label),
        ),
      );
      await tester.pumpAndSettle();
    }

    /// Opens study on a card whose question autoplays `hola.mp3`.
    Future<FakeGo> studyAudio(WidgetTester tester) async {
      useFakeMedia(backend, FakeMediaFiles(['hola.mp3', 'hello.mp3']));
      final go = FakeGo()..tree = [deckNode(1, 'Spanish')];
      FakeStudy(go, [audioCard(1)]);
      await pumpApp(tester, size: desktop);
      await studyDeck(tester, 'Spanish');
      expect(backend.playing, 'hola.mp3');
      return go;
    }

    testWidgets('study audio stops on switching to Browse (spec)', (
      tester,
    ) async {
      await studyAudio(tester);
      await selectDestination(tester, 'Browse');
      expect(backend.log, ['stop']);
      expect(backend.playing, isNull);
    });

    testWidgets('the browser preview stops on switching to Decks (spec)', (
      tester,
    ) async {
      useFakeMedia(backend, FakeMediaFiles(['gato.mp3']));
      final go = FakeGo();
      FakeBrowse(go, [
        row(1, 'gato')
          ..['question'] = '<cf-sound idx="0"></cf-sound> gato'
          ..['questionMedia'] = [
            {'file': 'gato.mp3', 'kind': 'audio'},
          ],
      ]);
      await pumpApp(tester, size: desktop);
      await selectDestination(tester, 'Browse');
      await tester.tap(find.byKey(const Key('row-1')));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Play gato.mp3'));
      await tester.pumpAndSettle();
      expect(backend.playing, 'gato.mp3');

      await selectDestination(tester, 'Decks');
      expect(backend.log, ['stop']);
      expect(backend.playing, isNull);
    });

    testWidgets('nothing resumes on return; R plays the side again (spec)', (
      tester,
    ) async {
      final go = await studyAudio(tester);
      await selectDestination(tester, 'Browse');
      await selectDestination(tester, 'Decks');
      // The same session and card, silent.
      expect(go.callsTo('study.Start'), hasLength(1));
      expect(find.byType(StudyScreen), findsOneWidget);
      expect(find.text('Show answer'), findsOneWidget);
      expect(backend.played, ['hola.mp3']);
      expect(backend.playing, isNull);

      await tester.sendKeyEvent(LogicalKeyboardKey.keyR);
      await tester.pumpAndSettle();
      expect(backend.played, ['hola.mp3', 'hola.mp3']);
      expect(backend.playing, 'hola.mp3');
    });

    testWidgets('re-selecting the current destination still goes back to '
        'its first screen', (tester) async {
      await studyAudio(tester);
      await selectDestination(tester, 'Decks');
      expect(find.byType(StudyScreen), findsNothing);
      expect(backend.playing, isNull);
    });
  });
}

void deckPickerTests() {
  group('deck picker', () {
    /// Types `/`; returns whether the app handled the key (when it did not,
    /// the platform types the character into the focused text field).
    Future<bool> slash(WidgetTester tester) async {
      final handled = await tester.sendKeyEvent(
        LogicalKeyboardKey.slash,
        character: '/',
      );
      await tester.pumpAndSettle();
      return handled;
    }

    /// Chooses the deck that [query] finds in the open picker.
    Future<void> pick(WidgetTester tester, String query) async {
      expect(find.byType(DeckPicker), findsOneWidget);
      await tester.enterText(
        find.descendant(
          of: find.byType(DeckPicker),
          matching: find.byType(TextField),
        ),
        query,
      );
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pumpAndSettle();
    }

    int selectedDestination(WidgetTester tester) => tester
        .widget<NavigationRail>(find.byType(NavigationRail))
        .selectedIndex!;

    List<Map<String, Object?>> decks() => [
      deckNode(1, 'Biology'),
      deckNode(2, 'Spanish', newCount: 1),
    ];

    testWidgets('/ opens the picker on the deck list, also where / needs '
        'Shift', (tester) async {
      FakeGo().tree = decks();
      await pumpApp(tester, size: desktop);
      expect(await slash(tester), isTrue);
      expect(find.byType(DeckPicker), findsOneWidget);
      // Another / types into the picker's search field.
      expect(await slash(tester), isFalse);
      expect(find.byType(DeckPicker), findsOneWidget);
      await tester.sendKeyEvent(LogicalKeyboardKey.escape);
      await tester.pumpAndSettle();
      expect(find.byType(DeckPicker), findsNothing);

      // German layout: Shift+7 types /.
      await tester.sendKeyDownEvent(LogicalKeyboardKey.shiftLeft);
      await tester.sendKeyEvent(LogicalKeyboardKey.digit7, character: '/');
      await tester.sendKeyUpEvent(LogicalKeyboardKey.shiftLeft);
      await tester.pumpAndSettle();
      expect(find.byType(DeckPicker), findsOneWidget);
    });

    testWidgets('while studying, the chosen deck replaces the session', (
      tester,
    ) async {
      final go = FakeGo()..tree = decks();
      FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
      await pumpApp(tester, size: desktop);
      await studyDeck(tester, 'Spanish');
      expect(find.byType(StudyScreen), findsOneWidget);

      expect(await slash(tester), isTrue);
      await pick(tester, 'bio');
      expect(find.byType(StudyScreen), findsNothing);
      expect(find.byType(DeckOverviewScreen), findsOneWidget);
      expect(go.callsTo('decks.Overview').last, [1]);
    });

    testWidgets('from the browser, the deck opens on the Decks tab', (
      tester,
    ) async {
      final go = FakeGo()..tree = decks();
      FakeBrowse(go, [row(1, 'gato')]);
      await pumpApp(tester, size: desktop);
      await tester.tap(
        find.descendant(
          of: find.byType(NavigationRail),
          matching: find.text('Browse'),
        ),
      );
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('row-1')));
      await tester.pumpAndSettle();

      expect(await slash(tester), isTrue);
      await pick(tester, 'span');
      expect(selectedDestination(tester), 0);
      expect(find.byType(DeckOverviewScreen), findsOneWidget);
      expect(go.callsTo('decks.Overview').last, [2]);
    });

    testWidgets('typing / in a note field or the search field types it '
        '(spec)', (tester) async {
      final go = FakeGo()..tree = decks();
      FakeBrowse(go, [row(1, 'gato')]);
      await pumpApp(tester, size: desktop);

      // A note field (the visual editor) on the Add screen.
      await tester.tap(
        find.descendant(
          of: find.byType(NavigationRail),
          matching: find.text('Add'),
        ),
      );
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('field-0-visual')));
      await tester.pump();
      expect(await slash(tester), isFalse);
      expect(find.byType(DeckPicker), findsNothing);

      // The same field as HTML.
      await tester.tap(find.byKey(const ValueKey('field-0-html')));
      await tester.pumpAndSettle();
      expect(fieldController(tester, 0).htmlMode, isTrue);
      await tester.tap(
        find.descendant(
          of: find.byKey(const Key('field-0')),
          matching: find.byType(TextField),
        ),
      );
      await tester.pump();
      expect(await slash(tester), isFalse);
      expect(find.byType(DeckPicker), findsNothing);

      // The browser's search field.
      await tester.tap(
        find.descendant(
          of: find.byType(NavigationRail),
          matching: find.text('Browse'),
        ),
      );
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('search')));
      await tester.pump();
      expect(await slash(tester), isFalse);
      expect(find.byType(DeckPicker), findsNothing);
    });
  });
}

void appearanceTests() {
  group('UI size', () {
    // Every visible element responds at every UI size (spec): a tap that
    // misses its target fails the test.
    setUp(() => WidgetController.hitTestWarningShouldBeFatal = true);
    tearDown(() => WidgetController.hitTestWarningShouldBeFatal = false);

    testWidgets('150 % makes text and icons 1.5 times larger at once (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      // Wide enough to keep the side rail at 150 %.
      await pumpApp(tester, size: const Size(1300, 800));
      final bar = find.byType(NavigationRail);
      final text = find.descendant(of: bar, matching: find.text('Decks'));
      final icon = find.descendant(of: bar, matching: find.byType(Icon)).first;
      final textBefore = tester.getRect(text).size;
      final iconBefore = tester.getRect(icon).size;

      await AppScope.read(tester.element(bar)).settings.setUiScale(150);
      await tester.pumpAndSettle();
      expect(go.callsTo('app.SetUIScale'), [
        [150],
      ]);
      final textAfter = tester.getRect(text).size;
      final iconAfter = tester.getRect(icon).size;
      expect(textAfter.width, closeTo(textBefore.width * 1.5, 0.5));
      expect(textAfter.height, closeTo(textBefore.height * 1.5, 0.5));
      expect(iconAfter.width, closeTo(iconBefore.width * 1.5, 0.5));
      expect(iconAfter.height, closeTo(iconBefore.height * 1.5, 0.5));
    });

    for (final percent in [75, 90, 100, 150]) {
      testWidgets('at $percent % taps reach the whole window', (tester) async {
        setSize(tester, phone);
        final scale = percent / 100;
        final taps = <Offset>[];
        await tester.pumpWidget(
          MediaQuery(
            data: MediaQueryData.fromView(tester.view),
            child: Appearance(
              scale: scale,
              reduceMotion: 'system',
              child: Listener(
                behavior: HitTestBehavior.opaque,
                onPointerDown: (event) => taps.add(event.localPosition),
              ),
            ),
          ),
        );
        const right = 399.0, bottom = 799.0;
        for (final point in const [
          Offset.zero,
          Offset(right, 0),
          Offset(0, bottom),
          Offset(right, bottom),
          Offset(200, 400),
        ]) {
          taps.clear();
          await tester.tapAt(point);
          expect(taps, hasLength(1), reason: 'tap at $point');
          expect(taps.single.dx, closeTo(point.dx / scale, 0.01));
          expect(taps.single.dy, closeTo(point.dy / scale, 0.01));
        }
      });
    }

    for (final percent in [75, 200]) {
      testWidgets('$percent % fits a phone on every destination', (
        tester,
      ) async {
        final go = FakeGo()..tree = [deckNode(1, 'Spanish', newCount: 3)];
        go.settings = {...go.settings, 'uiScale': percent};
        await pumpApp(tester, size: phone);
        for (final label in ['Browse', 'Add', 'Settings', 'Decks']) {
          await tester.tap(
            find.descendant(
              of: find.byType(NavigationBar),
              matching: find.text(label),
            ),
          );
          await tester.pumpAndSettle();
        }
        // Overflows fail the test with a layout error.
        expect(tester.takeException(), isNull);
      });
    }

    testWidgets('200 % fits a phone while studying and selecting cards', (
      tester,
    ) async {
      final go = FakeGo()..tree = [deckNode(1, 'Spanish', newCount: 1)];
      go.settings = {...go.settings, 'uiScale': 200};
      FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
      FakeBrowse(go, [row(1, 'gato'), row(2, 'perro')]);
      await pumpApp(tester, size: phone);
      await studyDeck(tester, 'Spanish');
      expect(find.byType(StudyScreen), findsOneWidget);
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      expect(find.text('Good'), findsOneWidget);

      await tester.tap(
        find.descendant(
          of: find.byType(NavigationBar),
          matching: find.text('Browse'),
        ),
      );
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('More'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Select').last);
      await tester.pumpAndSettle();
      await tester.tap(find.text('gato'));
      await tester.pumpAndSettle();
      // Overflows fail the test with a layout error.
      expect(tester.takeException(), isNull);
    });
  });

  group('minimalist mode', () {
    double elevationOf(WidgetTester tester, Finder of) => tester
        .widget<Material>(
          find.descendant(of: of, matching: find.byType(Material)).first,
        )
        .elevation;

    testWidgets('cards and bars are shown without shadows (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      await pumpApp(tester);
      // The deck list's hint for an empty collection is a card.
      expect(elevationOf(tester, find.byType(Card)), greaterThan(0));
      expect(elevationOf(tester, find.byType(NavigationBar)), greaterThan(0));

      await openSettings(tester);
      await tester.tap(find.text('Minimalist mode'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.SetMinimalist'), [
        [true],
      ]);
      await tester.tap(find.text('Decks').last);
      await tester.pumpAndSettle();
      expect(elevationOf(tester, find.byType(Card)), 0);
      expect(elevationOf(tester, find.byType(NavigationBar)), 0);
      final theme = Theme.of(tester.element(find.byType(Card)));
      expect(theme.dialogTheme.elevation, 0);
      expect(theme.popupMenuTheme.elevation, 0);
      expect(theme.appBarTheme.scrolledUnderElevation, 0);
    });

    testWidgets('dark mode is flat too', (tester) async {
      final go = FakeGo();
      go.settings = {...go.settings, 'theme': 'dark', 'minimalist': true};
      await pumpApp(tester);
      expect(
        Theme.of(tester.element(find.byType(Card))).brightness,
        Brightness.dark,
      );
      expect(elevationOf(tester, find.byType(Card)), 0);
    });
  });

  group('reduce motion', () {
    /// Starts studying Spanish and returns how far the study page's
    /// transition has come in the first frame that shows it.
    Future<double> studyTransition(WidgetTester tester) async {
      final row = find
          .ancestor(of: find.text('Spanish'), matching: find.byType(InkWell))
          .first;
      await tester.tap(
        find.descendant(of: row, matching: find.byTooltip('Deck actions')),
      );
      await tester.pumpAndSettle();
      await tester.tap(find.text('Study').last);
      for (
        var i = 0;
        i < 10 && find.byType(StudyScreen).evaluate().isEmpty;
        i++
      ) {
        await tester.pump();
      }
      final route = ModalRoute.of(tester.element(find.byType(StudyScreen)))!;
      final progress = route.animation!.value;
      await tester.pumpAndSettle();
      return progress;
    }

    FakeGo studyGo(String reduceMotion) {
      final go = FakeGo()..tree = [deckNode(1, 'Spanish', newCount: 1)];
      go.settings = {...go.settings, 'reduceMotion': reduceMotion};
      FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
      return go;
    }

    testWidgets('on: a study session opens without a transition', (
      tester,
    ) async {
      studyGo('on');
      await pumpApp(tester);
      expect(await studyTransition(tester), 1);
    });

    testWidgets('off: a study session opens with a transition', (tester) async {
      studyGo('off');
      await pumpApp(tester);
      expect(await studyTransition(tester), lessThan(1));
    });

    testWidgets('follow system: no transition when the system asks for '
        'reduced motion (spec)', (tester) async {
      tester.platformDispatcher.accessibilityFeaturesTestValue =
          const FakeAccessibilityFeatures(disableAnimations: true);
      addTearDown(
        tester.platformDispatcher.clearAccessibilityFeaturesTestValue,
      );
      studyGo('system');
      await pumpApp(tester);
      expect(await studyTransition(tester), 1);
    });

    testWidgets('follow system: transitions when the system does not ask', (
      tester,
    ) async {
      studyGo('system');
      await pumpApp(tester);
      expect(await studyTransition(tester), lessThan(1));
    });

    testWidgets('on: dialogs open at once', (tester) async {
      final go = FakeGo();
      go.settings = {...go.settings, 'reduceMotion': 'on'};
      await pumpApp(tester, size: desktop);
      await tester.tap(find.byTooltip('Create deck'));
      await tester.pump();
      final dialog = find.byType(Dialog);
      expect(dialog, findsOneWidget);
      expect(ModalRoute.of(tester.element(dialog))!.animation!.value, 1);
    });
  });
}
