import 'dart:async';

import 'package:clock/clock.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/app/errors.dart';
import 'package:cardflip/app/theme.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/features/decks/deck_options_screen.dart';
import 'package:cardflip/features/study/study_screen.dart';
import 'package:cardflip/l10n/l10n.dart';
import 'package:cardflip/widgets/speech.dart';
import 'package:cardflip/widgets/time_format.dart';

import 'support/fake_go.dart';
import 'support/fake_media.dart';
import 'support/fake_speech.dart';
import '../integration_test/rich_fields.dart';

Map<String, Object?> studyCard(
  int id,
  String front,
  String back, {
  String kind = 'new',
  bool empty = false,
  List<Map<String, Object?>> questionMedia = const [],
  List<Map<String, Object?>> answerMedia = const [],
  List<Map<String, Object?>> questionTts = const [],
  List<Map<String, Object?>> answerTts = const [],
  bool autoplay = true,
  bool replayQuestion = true,
  bool typeAnswer = false,
  int flag = 0,
  bool marked = false,
  int? noteId,
  bool showTimer = false,
  bool stopTimerOnAnswer = true,
  int maxAnswerSeconds = 60,
  int secondsToShowQuestion = 0,
  int secondsToShowAnswer = 0,
  String questionAction = 'showAnswer',
  String answerAction = 'buryCard',
  DateTime? dueEarly,
}) => {
  'cardId': id,
  'noteId': noteId ?? id + 100,
  'question': '<div class="card">$front</div>',
  'answer': '<div class="card">$front<hr id="answer">$back</div>',
  'kind': kind,
  'empty': empty,
  'questionMedia': questionMedia,
  'answerMedia': answerMedia,
  'questionTts': questionTts,
  'answerTts': answerTts,
  'autoplay': autoplay,
  'replayQuestion': replayQuestion,
  'showTimer': showTimer,
  'stopTimerOnAnswer': stopTimerOnAnswer,
  'maxAnswerSeconds': maxAnswerSeconds,
  'secondsToShowQuestion': secondsToShowQuestion,
  'secondsToShowAnswer': secondsToShowAnswer,
  'questionAction': questionAction,
  'answerAction': answerAction,
  'typeAnswer': typeAnswer,
  'flag': flag,
  'marked': marked,
  if (dueEarly != null) 'dueEarly': dueEarly.toUtc().toIso8601String(),
  'previews': [
    {'rating': 1, 'interval': _ns(const Duration(minutes: 1)), 'ends': false},
    {'rating': 2, 'interval': _ns(const Duration(minutes: 6)), 'ends': false},
    {'rating': 3, 'interval': _ns(const Duration(days: 3)), 'ends': false},
    {'rating': 4, 'interval': _ns(const Duration(days: 8)), 'ends': false},
  ],
};

/// A study session on the Go side: cards in order, undo restores the last.
class FakeStudy {
  FakeStudy(this.go, this.cards) {
    go.handle('study.Start', (_) => state());
    go.handle('study.Current', (_) => state());
    go.handle('study.Answer', (args) {
      answers.add(args);
      history.add(pos);
      pos++;
      studied++;
      spentMs += args[2] as int;
      final leech = leechAnswers[answers.length];
      return {
        ...state(),
        'leech': leech != null,
        'leechSuspended': leech ?? false,
      };
    });
    go.handle('study.Undo', (_) {
      pos = history.removeLast();
      studied--;
      return state();
    });
    go.handle('study.Bury', (_) {
      pos++;
      return state();
    });
    go.handle('study.Suspend', (_) {
      pos++;
      return state();
    });
    for (final action in ['BuryNote', 'SuspendNote', 'DeleteNote']) {
      go.handle('study.$action', (args) {
        dropNote(args[0] as int);
        return state();
      });
    }
    go.handle('browse.DeleteImpact', (args) {
      final notes = {for (final id in args[0] as List) noteOf(id as int)};
      return {
        'notes': notes.length,
        'cards': cards.where((c) => notes.contains(c['noteId'])).length,
      };
    });
    // Like Go: an invalid input changes nothing; a card due later leaves.
    go.handle('study.SetDueDate', (args) {
      final spec = args[1] as String;
      if (int.tryParse(spec.replaceAll('!', '')) == null) {
        throw 'Enter days from today: 0 for today, 1 for tomorrow, 60-90 for a random day in that range, with ! to set the interval too.';
      }
      if (spec != '0') pos++;
      return state();
    });
    go.handle('study.SetFlag', (args) {
      final i = cards.indexWhere((c) => c['cardId'] == args[0]);
      final flag = args[1] as int;
      cards[i] = {...cards[i], 'flag': cards[i]['flag'] == flag ? 0 : flag};
      return state();
    });
    go.handle('study.ToggleMark', (args) {
      final i = cards.indexWhere((c) => c['cardId'] == args[0]);
      cards[i] = {...cards[i], 'marked': cards[i]['marked'] != true};
      return state();
    });
    go.handle(
      'study.PreviousCardID',
      (_) => history.isEmpty ? 0 : cards[history.last]['cardId'],
    );
    go.handle('study.Reveal', (args) {
      final card = cards.firstWhere((c) => c['cardId'] == args[0]);
      final typed = args[1] as String;
      return {
        ...card,
        'answer':
            '<div class="card"><code id="typeans">'
            '<span class="typeBad">$typed</span></code></div>',
      };
    });
  }

  final FakeGo go;
  final List<Map<String, Object?>> cards;
  final answers = <List<dynamic>>[];
  final history = <int>[];
  int pos = 0;
  int studied = 0;
  int spentMs = 0;
  String? nextDue;
  String? learningDue;

  /// The answers (1 for the first) that make the card a leech, and whether
  /// they suspend it.
  final leechAnswers = <int, bool>{};

  int noteOf(int cardId) =>
      cards.firstWhere((c) => c['cardId'] == cardId)['noteId']! as int;

  /// The card's note leaves the session with all its cards still to come.
  void dropNote(int cardId) {
    final note = noteOf(cardId);
    for (var i = cards.length - 1; i >= pos; i--) {
      if (cards[i]['noteId'] == note) cards.removeAt(i);
    }
  }

  /// New cards per day of deck 1, from its options as `decks.SaveSettings`
  /// left them.
  int get newPerDay {
    final limit = go.deckLimits[1]?['newPerDay'];
    if (limit is int) return limit;
    final preset = go.presets[go.deckPresets[1] ?? 1]!;
    return (preset['options']! as Map)['newPerDay'] as int;
  }

  Map<String, Object?> state() {
    // Like Go: new cards beyond the day's limit are left out.
    var newLeft =
        newPerDay - history.where((i) => cards[i]['kind'] == 'new').length;
    final left = [
      for (final c in cards.skip(pos))
        if (c['kind'] != 'new' || newLeft-- > 0) c,
    ];
    // As Go: while waiting, the screen checks again when the learning card
    // comes within the learn ahead limit.
    var checkAt = learningDue;
    if (checkAt != null && left.isEmpty) {
      final limit = go.settings['learnAheadMinutes']! as int;
      checkAt = DateTime.parse(checkAt)
          .subtract(Duration(minutes: limit))
          .toUtc()
          .toIso8601String();
    }
    return {
      'deckId': 1,
      'deckName': 'Default',
      'card': left.isEmpty ? null : left.first,
      'counts': {
        'new': left.where((c) => c['kind'] == 'new').length,
        'learning': left.where((c) => c['kind'] == 'learning').length,
        'review': left.where((c) => c['kind'] == 'review').length,
      },
      'canUndo': history.isNotEmpty,
      'studied': studied,
      'timeSpent': spentMs * 1000000,
      'nextDue': left.isEmpty ? nextDue : null,
      'learningDue': learningDue,
      'checkAt': checkAt,
      'leech': false,
      'leechSuspended': false,
    };
  }
}

/// The study screen of deck 1. Its app state is disposed after the test,
/// or by the test itself when [disposedByTest] (with [endStudy]), as a test
/// that sends collection changes must: the deck list then waits for its
/// next reload.
Future<AppState> pumpStudy(
  WidgetTester tester, {
  bool loadSettings = false,
  bool disposedByTest = false,
  Size size = phone,
  Locale? locale,
}) async {
  setSize(tester, size);
  final app = AppState(isWeb: false);
  if (!disposedByTest) addTearDown(app.dispose);
  if (loadSettings) await app.settings.load();
  // The card's size is known, as after the first card: showing one does not
  // change the render environment and fetch it again.
  app.renderEnv.setCardSize(size);
  await tester.pump();
  await tester.pumpWidget(
    AppScope(
      state: app,
      child: MaterialApp(
        scaffoldMessengerKey: messengerKey,
        locale: locale,
        supportedLocales: AppLocalizations.supportedLocales,
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        home: const StudyScreen(deckId: 1, deckName: 'Default'),
      ),
    ),
  );
  await tester.pumpAndSettle();
  return app;
}

/// Turns learning ahead off, so a session waits for learning cards.
void noLearnAhead(FakeGo go) =>
    go.settings = {...go.settings, 'learnAheadMinutes': 0};

/// A learning card due one minute from now, as Go reports it.
String inAMinute() =>
    DateTime.now().add(const Duration(minutes: 1)).toUtc().toIso8601String();

/// Lets the minute pass until the pending learning card is due.
Future<void> waitForLearningCard(WidgetTester tester) async {
  await tester.pump(const Duration(seconds: 61));
  await tester.pumpAndSettle();
}

/// Lets a snackbar time out and leave.
Future<void> waitForSnackBar(WidgetTester tester) async {
  await tester.pumpAndSettle();
  await tester.pump(const Duration(seconds: 5));
  await tester.pumpAndSettle();
}

Map<String, Object?> sound(String file) => {'file': file, 'kind': 'audio'};

/// A card with `hola.mp3` on the question and `hello.mp3` on the answer.
Map<String, Object?> audioCard(
  int id, {
  bool autoplay = true,
  bool replayQuestion = true,
}) => {
  ...studyCard(
    id,
    '<cf-sound idx="0"></cf-sound> hola',
    '',
    questionMedia: [sound('hola.mp3')],
    answerMedia: [sound('hello.mp3')],
    autoplay: autoplay,
    replayQuestion: replayQuestion,
  ),
  // As Go renders it: {{FrontSide}} without the question's sounds.
  'answer': '<div class="card">hola<hr id="answer"><cf-sound idx="0"></cf-sound> hello</div>',
};

Map<String, Object?> tts(String lang, String text) => {
  'text': text,
  'lang': lang,
  'voices': const [],
  'speed': 1.0,
};

/// A card that speaks `perro` in Spanish on the question and `dog` in
/// [answerLang] on the answer.
Map<String, Object?> ttsCard(int id, {String answerLang = 'en_US'}) => {
  ...studyCard(
    id,
    'perro<cf-tts idx="0"></cf-tts>',
    '',
    questionTts: [tts('es_ES', 'perro')],
    answerTts: [tts(answerLang, 'dog')],
  ),
  // As Go renders it: {{FrontSide}} without the question's utterances.
  'answer': '<div class="card">perro<hr id="answer">dog<cf-tts idx="0"></cf-tts></div>',
};

/// A fake speech engine with Spanish and English voices.
FakeSpeechEngine fakeSpeech() {
  final engine = FakeSpeechEngine(const [
    Voice('Mónica', 'es_ES'),
    Voice('Samantha', 'en_US'),
  ]);
  useFakeSpeech(engine);
  return engine;
}

/// Fake media with the files of [audioCard].
FakeBackend fakeMedia() {
  final backend = FakeBackend();
  useFakeMedia(backend, FakeMediaFiles(['hola.mp3', 'hello.mp3']));
  return backend;
}

Finder richText(String text) => find.textContaining(text, findRichText: true);

/// The remaining count of [kind] ('new', 'learning' or 'review').
String? remaining(WidgetTester tester, String kind) =>
    tester.widget<Text>(find.byKey(Key('remaining-$kind'))).data;

/// Closes the study screen and disposes [app], for [pumpStudy] with
/// `disposedByTest`.
Future<void> endStudy(WidgetTester tester, AppState app) async {
  await tester.pumpWidget(const SizedBox());
  app.dispose();
}

void main() {
  testWidgets('reveal shows the answer and then the rating buttons', (
    tester,
  ) async {
    final go = FakeGo();
    FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
    await pumpStudy(tester);
    expect(richText('hola'), findsOneWidget);
    expect(richText('hello'), findsNothing);
    expect(find.text('Good'), findsNothing);
    await tester.tap(find.text('Show answer'));
    await tester.pumpAndSettle();
    expect(richText('hello'), findsOneWidget);
    for (final label in ['Again', 'Hard', 'Good', 'Easy']) {
      expect(find.text(label), findsOneWidget);
    }
    expect(find.text('3d'), findsOneWidget);
  });

  group('typing the answer', () {
    Map<String, Object?> typeCard(int id) => studyCard(
      id,
      'dog <cf-typeans></cf-typeans>',
      'perro',
      typeAnswer: true,
    );
    final field = find.byKey(const Key('type-answer'));
    bool fieldFocused(WidgetTester tester) =>
        tester.widget<TextField>(field).focusNode!.hasFocus;

    testWidgets('the input is focused and keys type instead of studying '
        '(spec)', (tester) async {
      final go = FakeGo();
      final study = FakeStudy(go, [typeCard(1)]);
      await pumpStudy(tester);
      expect(field, findsOneWidget);
      expect(fieldFocused(tester), isTrue);
      // Unhandled, the keys reach the input as text.
      for (final key in [
        LogicalKeyboardKey.digit3,
        LogicalKeyboardKey.space,
        LogicalKeyboardKey.keyR,
        LogicalKeyboardKey.keyI,
      ]) {
        expect(await tester.sendKeyEvent(key), isFalse, reason: '$key');
      }
      await tester.pumpAndSettle();
      expect(study.answers, isEmpty);
      expect(go.callsTo('browse.CardInfo'), isEmpty);
      expect(find.text('Show answer'), findsOneWidget);
      expect(go.callsTo('study.Reveal'), isEmpty);
    });

    testWidgets('Enter reveals the comparison, then rating keys work (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      final study = FakeStudy(go, [typeCard(1), studyCard(2, 'hola', 'hi')]);
      await pumpStudy(tester);
      await tester.enterText(field, 'pero');
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pumpAndSettle();
      expect(go.callsTo('study.Reveal'), [
        [1, 'pero'],
      ]);
      expect(richText('pero'), findsOneWidget);
      expect(find.text('Good'), findsOneWidget);
      expect(field, findsNothing);
      await pressKeys(tester, [LogicalKeyboardKey.digit3]);
      expect(study.answers.single.take(2), [1, 3]);
      expect(richText('hola'), findsOneWidget);
    });

    testWidgets('Show answer sends the typed text too', (tester) async {
      final go = FakeGo();
      FakeStudy(go, [typeCard(1)]);
      await pumpStudy(tester);
      await tester.enterText(field, 'perro');
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      expect(go.callsTo('study.Reveal'), [
        [1, 'perro'],
      ]);
      expect(find.text('Good'), findsOneWidget);
    });

    testWidgets('Enter that confirms a composition does not reveal', (
      tester,
    ) async {
      final go = FakeGo();
      FakeStudy(go, [typeCard(1)]);
      await pumpStudy(tester);
      await tester.showKeyboard(field);
      tester.testTextInput.updateEditingValue(
        const TextEditingValue(
          text: 'いぬ',
          selection: TextSelection.collapsed(offset: 2),
          composing: TextRange(start: 0, end: 2),
        ),
      );
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pumpAndSettle();
      expect(go.callsTo('study.Reveal'), isEmpty);
      expect(find.text('Show answer'), findsOneWidget);
    });

    testWidgets('a card without type: keeps the plain answer', (tester) async {
      final go = FakeGo();
      FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
      await pumpStudy(tester);
      expect(field, findsNothing);
      await pressKeys(tester, [LogicalKeyboardKey.space]);
      expect(richText('hello'), findsOneWidget);
      expect(go.callsTo('study.Reveal'), isEmpty);
    });
  });

  testWidgets('an empty card says so and can still be rated', (tester) async {
    final go = FakeGo();
    final study = FakeStudy(go, [
      studyCard(1, '', 'Paris is the capital of France', empty: true),
      studyCard(2, 'hola', 'hello'),
    ]);
    await pumpStudy(tester);
    // Empty cloze card (spec): the screen says the card is empty and how to
    // remove it, and the user can still rate it.
    expect(find.byKey(const Key('emptyCardNotice')), findsOneWidget);
    expect(find.textContaining('This card is empty'), findsOneWidget);
    expect(find.textContaining('Settings › Empty cards'), findsOneWidget);
    await tester.tap(find.text('Show answer'));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('emptyCardNotice')), findsOneWidget);
    await tester.tap(find.byKey(const Key('rate-3')));
    await tester.pumpAndSettle();
    expect(study.answers.single.take(2), [1, 3]);
    // The next card is not empty: no notice.
    expect(richText('hola'), findsOneWidget);
    expect(find.byKey(const Key('emptyCardNotice')), findsNothing);
  });

  group('card shown early or again', () {
    final notice = find.byKey(const Key('returningCardNotice'));
    const early =
        'Shown early: due in 11 minutes, because nothing else is due.';

    Future<void> answer(WidgetTester tester, int rating) async {
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(Key('rate-$rating')));
      await tester.pumpAndSettle();
    }

    testWidgets('shown early: says when it is due; gone with the answer', (
      tester,
    ) async {
      final go = FakeGo();
      // Go: Good made the new card due in about 11 minutes, and with nothing
      // else due it comes back right away.
      final due = clock.now().add(const Duration(minutes: 10, seconds: 30));
      FakeStudy(go, [
        studyCard(1, 'hola', 'hello'),
        studyCard(1, 'hola', 'hello', kind: 'learning', dueEarly: due),
      ]);
      await pumpStudy(tester);
      expect(notice, findsNothing);
      await answer(tester, 3);
      // Shown early (spec).
      expect(richText('hola'), findsOneWidget);
      expect(
        find.text(
          'This card again: due in 11 minutes, shown early because nothing '
          'else is due.',
        ),
        findsOneWidget,
      );
      expect(find.text('Show answer'), findsOneWidget);
      // Notice goes away (spec).
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      expect(notice, findsNothing);
    });

    testWidgets('due again now', (tester) async {
      final go = FakeGo();
      FakeStudy(go, [
        studyCard(1, 'hola', 'hello', kind: 'learning'),
        studyCard(1, 'hola', 'hello', kind: 'learning'),
      ]);
      await pumpStudy(tester);
      await answer(tester, 1);
      // Due again now (spec).
      expect(
        find.text('This card again: it is due again now.'),
        findsOneWidget,
      );
    });

    testWidgets('another card shown early says so', (tester) async {
      final go = FakeGo();
      final due = clock.now().add(const Duration(minutes: 10, seconds: 30));
      FakeStudy(go, [
        studyCard(1, 'hola', 'hello'),
        studyCard(2, 'perro', 'dog', kind: 'learning', dueEarly: due),
      ]);
      await pumpStudy(tester);
      expect(notice, findsNothing);
      await answer(tester, 3);
      // Shown early (spec): another card than the one just answered.
      expect(richText('perro'), findsOneWidget);
      expect(find.text(early), findsOneWidget);
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      expect(notice, findsNothing);
    });

    testWidgets('the first card of a session shown early says so', (
      tester,
    ) async {
      final go = FakeGo();
      final due = clock.now().add(const Duration(minutes: 10, seconds: 30));
      FakeStudy(go, [
        studyCard(1, 'hola', 'hello', kind: 'learning', dueEarly: due),
      ]);
      await pumpStudy(tester);
      expect(find.text(early), findsOneWidget);
    });

    testWidgets('another card that is due, also a sibling: no notice', (
      tester,
    ) async {
      final go = FakeGo();
      FakeStudy(go, [
        studyCard(1, 'hola', 'hello', noteId: 100),
        studyCard(2, 'hello', 'hola', noteId: 100),
      ]);
      await pumpStudy(tester);
      await answer(tester, 3);
      // Another card (spec).
      expect(richText('hello'), findsOneWidget);
      expect(notice, findsNothing);
    });

    testWidgets('undo brings the card back without the notice', (tester) async {
      final go = FakeGo();
      final due = clock.now().add(const Duration(minutes: 10, seconds: 30));
      FakeStudy(go, [
        studyCard(1, 'hola', 'hello'),
        studyCard(1, 'hola', 'hello', kind: 'learning', dueEarly: due),
      ]);
      await pumpStudy(tester);
      await answer(tester, 3);
      expect(notice, findsOneWidget);
      // Undo (spec).
      await tester.tap(find.widgetWithIcon(IconButton, Icons.undo));
      await tester.pumpAndSettle();
      expect(go.callsTo('study.Undo'), hasLength(1));
      expect(richText('hola'), findsOneWidget);
      expect(find.text('Show answer'), findsOneWidget);
      expect(notice, findsNothing);
    });
  });

  testWidgets('rating Good schedules the card and shows the next question', (
    tester,
  ) async {
    final go = FakeGo();
    final study = FakeStudy(go, [
      studyCard(1, 'hola', 'hello', kind: 'review'),
      studyCard(2, 'perro', 'dog', kind: 'review'),
    ]);
    await pumpStudy(tester);
    expect(find.byKey(const Key('remaining-review')), findsOneWidget);
    expect(
      tester.widget<Text>(find.byKey(const Key('remaining-review'))).data,
      '2',
    );
    await tester.tap(find.text('Show answer'));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('rate-3')));
    await tester.pumpAndSettle();
    expect(study.answers.single.take(2), [1, 3]);
    expect(study.answers.single[2], isA<int>());
    expect(richText('perro'), findsOneWidget);
    expect(find.text('Show answer'), findsOneWidget);
    expect(
      tester.widget<Text>(find.byKey(const Key('remaining-review'))).data,
      '1',
    );
  });

  testWidgets('a synced deletion of the next card is skipped, with the '
      'counts updated and the answer kept', (tester) async {
    final go = FakeGo();
    final study = FakeStudy(go, [
      studyCard(1, 'hola', 'hello', kind: 'review'),
      studyCard(2, 'perro', 'dog', kind: 'review'),
      studyCard(3, 'gato', 'cat', kind: 'review'),
    ]);
    final app = await pumpStudy(tester, disposedByTest: true);
    String remaining() =>
        tester.widget<Text>(find.byKey(const Key('remaining-review'))).data!;
    expect(remaining(), '3');
    await tester.tap(find.text('Show answer'));
    await tester.pumpAndSettle();

    // Another device deleted the note of the next card; sync applies it.
    final current = go.callsTo('study.Current').length;
    study.dropNote(2);
    go.emitChange(kind: 'sync', notes: true);
    await tester.pumpAndSettle();
    expect(go.callsTo('study.Current'), hasLength(current + 1));
    expect(remaining(), '2');
    // The card being answered stays, revealed.
    expect(richText('hola'), findsOneWidget);
    expect(find.byKey(const Key('rate-3')), findsOneWidget);

    await tester.tap(find.byKey(const Key('rate-3')));
    await tester.pumpAndSettle();
    expect(study.answers.single.take(2), [1, 3]);
    expect(richText('gato'), findsOneWidget);
    expect(remaining(), '1');
    await endStudy(tester, app);
  });

  testWidgets('a synced change during an answer waits for it', (tester) async {
    final go = FakeGo();
    final study = FakeStudy(go, [
      studyCard(1, 'hola', 'hello', kind: 'review'),
      studyCard(2, 'perro', 'dog', kind: 'review'),
      studyCard(3, 'gato', 'cat', kind: 'review'),
    ]);
    final answering = Completer<void>();
    go.handle('study.Answer', (args) async {
      await answering.future;
      study.answers.add(args);
      study.history.add(study.pos);
      study.pos++;
      return study.state();
    });
    final app = await pumpStudy(tester, disposedByTest: true);
    await tester.tap(find.text('Show answer'));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('rate-3')));
    await tester.pump();
    final current = go.callsTo('study.Current').length;
    go.emitChange(kind: 'sync', notes: true);
    await tester.pump();
    expect(go.callsTo('study.Current'), hasLength(current));

    answering.complete();
    await tester.pumpAndSettle();
    expect(study.answers.single.take(2), [1, 3]);
    expect(go.callsTo('study.Current'), hasLength(current + 1));
    expect(richText('perro'), findsOneWidget);
    expect(find.text('Show answer'), findsOneWidget);
    await endStudy(tester, app);
  });

  testWidgets('undo shows the previous card again, unrevealed', (tester) async {
    final go = FakeGo();
    FakeStudy(go, [
      studyCard(1, 'hola', 'hello'),
      studyCard(2, 'perro', 'dog'),
    ]);
    await pumpStudy(tester);
    final undo = find.widgetWithIcon(IconButton, Icons.undo);
    expect(tester.widget<IconButton>(undo).onPressed, isNull);
    await tester.tap(find.text('Show answer'));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('rate-1')));
    await tester.pumpAndSettle();
    expect(richText('perro'), findsOneWidget);
    await tester.tap(undo);
    await tester.pumpAndSettle();
    expect(go.callsTo('study.Undo'), hasLength(1));
    expect(richText('hola'), findsOneWidget);
    expect(find.text('Show answer'), findsOneWidget);
    expect(tester.widget<IconButton>(undo).onPressed, isNull);
  });

  testWidgets('an answer that makes a leech says so; undo shows the card', (
    tester,
  ) async {
    final go = FakeGo();
    final study = FakeStudy(go, [
      studyCard(1, 'hola', 'hello', kind: 'review'),
      studyCard(2, 'perro', 'dog', kind: 'review'),
      studyCard(3, 'gato', 'cat', kind: 'review'),
    ]);
    study.leechAnswers[1] = true;
    study.leechAnswers[2] = false;
    await pumpStudy(tester);
    Future<void> again() async {
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('rate-1')));
      await tester.pump();
    }

    // Message (spec): suspend action.
    await again();
    expect(find.text('The card is a leech and was suspended.'), findsOneWidget);
    await waitForSnackBar(tester);
    expect(find.textContaining('leech'), findsNothing);
    // Undo leech (spec): the card is shown again. (Go reverts the
    // suspension and the tag.)
    await tester.tap(find.widgetWithIcon(IconButton, Icons.undo));
    await tester.pumpAndSettle();
    expect(go.callsTo('study.Undo'), hasLength(1));
    expect(richText('hola'), findsOneWidget);
    expect(find.textContaining('leech'), findsNothing);
    // Tag only: no mention of a suspension.
    await again();
    expect(find.text('The card is a leech.'), findsOneWidget);
    await waitForSnackBar(tester);
    // An ordinary answer shows no message.
    await again();
    expect(find.textContaining('leech'), findsNothing);
  });

  testWidgets(
    'keyboard: Space reveals, 3 rates Good, Ctrl+Z undoes, Esc leaves',
    (tester) async {
      final go = FakeGo();
      final study = FakeStudy(go, [
        studyCard(1, 'hola', 'hello'),
        studyCard(2, 'perro', 'dog'),
      ]);
      setSize(tester, desktop);
      final app = AppState(isWeb: false);
      addTearDown(app.dispose);
      await tester.pumpWidget(
        AppScope(
          state: app,
          child: MaterialApp(
            home: Builder(
              builder: (context) => Scaffold(
                body: Center(
                  child: TextButton(
                    onPressed: () => Navigator.of(context).push(
                      MaterialPageRoute<void>(
                        builder: (_) =>
                            const StudyScreen(deckId: 1, deckName: 'Default'),
                      ),
                    ),
                    child: const Text('open'),
                  ),
                ),
              ),
            ),
          ),
        ),
      );
      await tester.tap(find.text('open'));
      await tester.pumpAndSettle();

      // Keys 1–4 do nothing before the answer is shown.
      await tester.sendKeyEvent(LogicalKeyboardKey.digit3);
      await tester.pumpAndSettle();
      expect(study.answers, isEmpty);

      await tester.sendKeyEvent(LogicalKeyboardKey.space);
      await tester.pumpAndSettle();
      expect(richText('hello'), findsOneWidget);
      await tester.sendKeyEvent(LogicalKeyboardKey.digit3);
      await tester.pumpAndSettle();
      expect(study.answers.single.take(2), [1, 3]);
      expect(richText('perro'), findsOneWidget);

      // Space twice: reveal, then Good.
      await tester.sendKeyEvent(LogicalKeyboardKey.space);
      await tester.pumpAndSettle();
      await tester.sendKeyEvent(LogicalKeyboardKey.enter);
      await tester.pumpAndSettle();
      expect(study.answers.last.take(2), [2, 3]);

      await tester.sendKeyDownEvent(LogicalKeyboardKey.controlLeft);
      await tester.sendKeyEvent(LogicalKeyboardKey.keyZ);
      await tester.sendKeyUpEvent(LogicalKeyboardKey.controlLeft);
      await tester.pumpAndSettle();
      expect(go.callsTo('study.Undo'), hasLength(1));
      expect(richText('perro'), findsOneWidget);

      await tester.sendKeyEvent(LogicalKeyboardKey.escape);
      await tester.pumpAndSettle();
      expect(find.byType(StudyScreen), findsNothing);
    },
  );

  testWidgets('completion shows cards studied and time spent', (tester) async {
    final go = FakeGo();
    final study = FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
    await pumpStudy(tester);
    await tester.tap(find.text('Show answer'));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('rate-4')));
    await tester.pumpAndSettle();
    expect(find.text('Session complete'), findsOneWidget);
    expect(
      find.text(
        'You studied 1 card in ${formatSpent(Duration(milliseconds: study.spentMs))}.',
      ),
      findsOneWidget,
    );
    expect(find.textContaining('learning cards'), findsNothing);
  });

  testWidgets('completion mentions learning cards that are due soon', (
    tester,
  ) async {
    final go = FakeGo();
    final study = FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
    final soon = DateTime.now().add(const Duration(minutes: 10));
    study.learningDue = soon.toUtc().toIso8601String();
    await pumpStudy(tester);
    await tester.tap(find.text('Show answer'));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('rate-3')));
    await tester.pumpAndSettle();
    expect(
      find.text('More learning cards will be due ${formatDue(soon)}.'),
      findsOneWidget,
    );
    expect(find.text('Check again'), findsOneWidget);
  });

  testWidgets('nothing due says when the next card is due', (tester) async {
    final go = FakeGo();
    final study = FakeStudy(go, []);
    final tomorrow = DateTime.now().add(const Duration(days: 1));
    study.nextDue = DateTime(
      tomorrow.year,
      tomorrow.month,
      tomorrow.day,
      4,
    ).toUtc().toIso8601String();
    await pumpStudy(tester);
    expect(
      find.text('Congratulations! Nothing is due right now.'),
      findsOneWidget,
    );
    expect(find.text('The next card is due tomorrow.'), findsOneWidget);
  });

  testWidgets('edit, bury and suspend from the card menu', (tester) async {
    final go = FakeGo();
    final cards = [
      studyCard(1, 'hola', 'hello'),
      studyCard(2, 'perro', 'dog'),
      studyCard(3, 'gato', 'cat'),
    ];
    FakeStudy(go, cards);
    go.handle(
      'notes.Get',
      (args) => {
        'id': args[0],
        'typeId': 10,
        'fields': ['hola', 'hello'],
        'tags': <String>[],
        'cardIds': [1],
      },
    );
    await pumpStudy(tester);
    await tester.tap(find.byTooltip('Card actions'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Edit card'));
    await tester.pumpAndSettle();
    expect(find.text('Edit card'), findsOneWidget); // the editor
    // The note was corrected meanwhile.
    cards[0] = studyCard(1, 'hola!', 'hello');
    await tester.pageBack();
    await tester.pumpAndSettle();
    expect(go.callsTo('study.Current'), hasLength(1));
    expect(richText('hola!'), findsOneWidget);

    await tester.tap(find.byTooltip('Card actions'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Bury card until tomorrow'));
    await tester.pumpAndSettle();
    expect(go.callsTo('study.Bury'), [
      [1],
    ]);
    expect(richText('perro'), findsOneWidget);
    await tester.tap(find.byTooltip('Card actions'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Suspend card'));
    await tester.pumpAndSettle();
    expect(go.callsTo('study.Suspend'), [
      [2],
    ]);
    expect(richText('gato'), findsOneWidget);
  });

  group('note actions', () {
    /// The forward and reverse card of note 7, then a card of note 8.
    List<Map<String, Object?>> reversedNote() => [
      studyCard(1, 'hola', 'hello', noteId: 7),
      studyCard(2, 'hello', 'hola', noteId: 7),
      studyCard(3, 'gato', 'cat', noteId: 8),
    ];

    Future<void> chooseFromMenu(WidgetTester tester, String item) async {
      await tester.tap(find.byTooltip('Card actions'));
      await tester.pumpAndSettle();
      await tester.tap(find.text(item));
      await tester.pumpAndSettle();
    }

    Future<void> pressDelete(WidgetTester tester) async {
      await tester.sendKeyDownEvent(LogicalKeyboardKey.controlLeft);
      await tester.sendKeyEvent(LogicalKeyboardKey.delete);
      await tester.sendKeyUpEvent(LogicalKeyboardKey.controlLeft);
      await tester.pumpAndSettle();
    }

    testWidgets('suspend note suspends both cards and moves on (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      FakeStudy(go, reversedNote());
      await pumpStudy(tester);
      expect(remaining(tester, 'new'), '3');
      await chooseFromMenu(tester, 'Suspend note');
      expect(go.callsTo('study.SuspendNote'), [
        [1],
      ]);
      expect(richText('gato'), findsOneWidget);
      expect(remaining(tester, 'new'), '1');
    });

    testWidgets('bury note buries both cards and moves on', (tester) async {
      final go = FakeGo();
      FakeStudy(go, reversedNote());
      await pumpStudy(tester);
      await chooseFromMenu(tester, 'Bury note until tomorrow');
      expect(go.callsTo('study.BuryNote'), [
        [1],
      ]);
      expect(richText('gato'), findsOneWidget);
    });

    testWidgets('delete note asks with the card count, then deletes (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      FakeStudy(go, reversedNote());
      await pumpStudy(tester);
      await chooseFromMenu(tester, 'Delete note');
      expect(find.text('Delete 2 cards?'), findsOneWidget);
      expect(
        find.textContaining('Including 1 other card from the same note.'),
        findsOneWidget,
      );
      // Cancelled: nothing is deleted and the card stays.
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(go.callsTo('study.DeleteNote'), isEmpty);
      expect(richText('hola'), findsOneWidget);

      // With the key, confirmed: the note goes, the next note's card shows.
      await pressDelete(tester);
      expect(find.text('Delete 2 cards?'), findsOneWidget);
      await tester.tap(find.widgetWithText(FilledButton, 'Delete'));
      await tester.pumpAndSettle();
      expect(go.callsTo('study.DeleteNote'), [
        [1],
      ]);
      expect(richText('gato'), findsOneWidget);
      expect(remaining(tester, 'new'), '1');
    });

    Future<void> typeCharacter(
      WidgetTester tester,
      LogicalKeyboardKey key,
      String character, {
      bool shift = false,
    }) async {
      if (shift) await tester.sendKeyDownEvent(LogicalKeyboardKey.shiftLeft);
      await tester.sendKeyDownEvent(key, character: character);
      await tester.sendKeyUpEvent(key);
      if (shift) await tester.sendKeyUpEvent(LogicalKeyboardKey.shiftLeft);
      await tester.pumpAndSettle();
    }

    for (final (key, character, shift, channel) in [
      (LogicalKeyboardKey.minus, '-', false, 'study.Bury'),
      (LogicalKeyboardKey.digit2, '@', true, 'study.Suspend'),
      (LogicalKeyboardKey.equal, '=', false, 'study.BuryNote'),
      (LogicalKeyboardKey.digit1, '!', true, 'study.SuspendNote'),
    ]) {
      testWidgets('$character runs $channel and moves on (spec)', (
        tester,
      ) async {
        final go = FakeGo();
        FakeStudy(go, reversedNote());
        await pumpStudy(tester);
        await tester.tap(find.text('Show answer'));
        await tester.pumpAndSettle();
        await typeCharacter(tester, key, character, shift: shift);
        expect(go.callsTo(channel), [
          [1],
        ]);
        // The shifted digit is no rating.
        expect(go.callsTo('study.Answer'), isEmpty);
        expect(find.text('Show answer'), findsOneWidget);
        expect(richText('hola'), findsNothing);
      });
    }

    testWidgets('E opens the editor (spec)', (tester) async {
      final go = FakeGo();
      FakeStudy(go, reversedNote());
      go.handle(
        'notes.Get',
        (args) => {
          'id': 7,
          'typeId': 10,
          'fields': ['hola', 'hello'],
          'tags': <String>[],
          'cardIds': [1, 2],
        },
      );
      await pumpStudy(tester);
      await tester.sendKeyEvent(LogicalKeyboardKey.keyE);
      await tester.pumpAndSettle();
      expect(find.text('Edit card'), findsOneWidget);
      expect(go.callsTo('notes.Get'), [
        [7],
      ]);
    });

    testWidgets('create copy adds a new note and returns to the card (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      FakeStudy(go, reversedNote());
      go.handle(
        'browse.Card',
        (args) => {
          'cardId': args[0],
          'deckId': 2,
          'deck': 'Spanish',
          'cardName': 'Card 1',
          'question': '<div>hola</div>',
          'answer': '<div>hola<hr>hello</div>',
          'questionMedia': const [],
          'answerMedia': const [],
          'state': 'new',
          'intervalDays': 0,
          'reviews': 0,
          'lapses': 0,
          'suspended': false,
          'flag': 0,
          'note': {
            'id': 7,
            'typeId': 11,
            'fields': ['hola', 'hello'],
            'tags': ['greeting'],
            'cardIds': [1, 2],
          },
          'noteType': noteTypeJson(
            11,
            'Basic (and reversed card)',
            ['Front', 'Back'],
            templates: ['Card 1', 'Card 2'],
          ),
          'noteCards': 2,
        },
      );
      go.handle('notes.Add', (_) => 99);
      await pumpStudy(tester, loadSettings: true);
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      await chooseFromMenu(tester, 'Create copy');
      expect(go.callsTo('browse.Card'), [
        [1],
      ]);
      expect(find.text('Add card'), findsWidgets);
      expect(fieldHtml(tester, 0), 'hola');

      await enterField(tester, 0, 'buenos días');
      await tester.tap(find.byKey(const Key('editor-submit')));
      await tester.pumpAndSettle();
      final note =
          go.callsTo('notes.Add').single.single as Map<String, dynamic>;
      expect(note['typeId'], 11);
      expect(note['deckId'], 2);
      expect(note['fields'], ['buenos días', 'hello']);
      expect(note['tags'], ['greeting']);

      // Closing the add screen returns to the same card, answer still shown.
      await tester.pageBack();
      await tester.pumpAndSettle();
      expect(find.byType(StudyScreen), findsOneWidget);
      expect(go.callsTo('study.Current'), hasLength(1));
      expect(richText('hello'), findsOneWidget);
      expect(find.text('Good'), findsOneWidget);
      expect(go.callsTo('study.Answer'), isEmpty);
    });

    testWidgets('lowering the new card limit in the options ends new cards '
        '(spec)', (tester) async {
      final go = FakeGo();
      FakeStudy(go, [
        studyCard(1, 'hola', 'hello'),
        studyCard(2, 'perro', 'dog'),
        studyCard(3, 'gato', 'cat'),
      ]);
      // Saved without announcing a change: the deck list, which is not
      // shown, would reload and leave its timer running past the test.
      go.handle('decks.SaveSettings', (args) {
        final [_, preset as int, options as Map, limits as Map] = args;
        go.presets[preset] = {
          ...go.presets[preset]!,
          'options': options.cast<String, Object?>(),
        };
        go.deckLimits[1] = limits.cast<String, Object?>();
        return null;
      });
      await pumpStudy(tester);
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('rate-3')));
      await tester.pumpAndSettle();
      expect(richText('perro'), findsOneWidget);

      // From the menu, closed without changes: the same card.
      await chooseFromMenu(tester, 'Deck options');
      expect(find.byType(DeckOptionsScreen), findsOneWidget);
      expect(find.text('Options for Default'), findsOneWidget);
      await tester.pageBack();
      await tester.pumpAndSettle();
      expect(go.callsTo('study.Current'), hasLength(1));
      expect(richText('perro'), findsOneWidget);

      // With O: one new card was introduced today, the limit becomes 1.
      await tester.sendKeyEvent(LogicalKeyboardKey.keyO);
      await tester.pumpAndSettle();
      expect(find.byType(DeckOptionsScreen), findsOneWidget);
      await tester.enterText(find.byKey(const Key('newPerDay')), '1');
      await tester.tap(find.text('Save'));
      await tester.pumpAndSettle();
      expect(go.callsTo('decks.SaveSettings'), hasLength(1));
      expect(find.byType(DeckOptionsScreen), findsNothing);
      expect(go.callsTo('study.Current'), hasLength(2));
      expect(richText('perro'), findsNothing);
      expect(find.text('Session complete'), findsOneWidget);
    });

    testWidgets('the keys type into the answer input', (tester) async {
      final go = FakeGo();
      FakeStudy(go, [
        studyCard(
          1,
          'dog <cf-typeans></cf-typeans>',
          'perro',
          typeAnswer: true,
        ),
      ]);
      await pumpStudy(tester);
      await typeCharacter(tester, LogicalKeyboardKey.minus, '-');
      await typeCharacter(tester, LogicalKeyboardKey.digit1, '!', shift: true);
      await pressDelete(tester);
      expect(go.callsTo('study.Bury'), isEmpty);
      expect(go.callsTo('study.SuspendNote'), isEmpty);
      expect(find.text('Delete 1 card?'), findsNothing);
    });
  });

  testWidgets('set due date pushes the card to next week (spec)', (
    tester,
  ) async {
    final go = FakeGo();
    FakeStudy(go, [
      studyCard(1, 'uno', 'one', kind: 'review'),
      studyCard(2, 'dos', 'two', kind: 'review'),
      studyCard(3, 'tres', 'three', kind: 'review'),
    ]);
    await pumpStudy(tester);
    String reviews() =>
        tester.widget<Text>(find.byKey(const Key('remaining-review'))).data!;
    expect(reviews(), '3');

    await tester.tap(find.byTooltip('Card actions'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Set due date'));
    await tester.pumpAndSettle();
    final input = find.byKey(const Key('due-date-input'));
    // An invalid input is explained in the dialog; the card stays.
    await tester.enterText(input, 'next week');
    await tester.tap(find.widgetWithText(FilledButton, 'Set due date'));
    await tester.pumpAndSettle();
    expect(find.textContaining('Enter days from today'), findsOneWidget);
    await tester.enterText(input, '7');
    await tester.tap(find.widgetWithText(FilledButton, 'Set due date'));
    await tester.pumpAndSettle();
    expect(go.callsTo('study.SetDueDate').last, [1, '7']);
    expect(input, findsNothing);
    expect(richText('dos'), findsOneWidget);
    expect(reviews(), '2');

    // Ctrl+Shift+D opens it too; Cancel leaves the card on screen.
    Future<void> shortcut() async {
      await tester.sendKeyDownEvent(LogicalKeyboardKey.controlLeft);
      await tester.sendKeyDownEvent(LogicalKeyboardKey.shiftLeft);
      await tester.sendKeyEvent(LogicalKeyboardKey.keyD);
      await tester.sendKeyUpEvent(LogicalKeyboardKey.shiftLeft);
      await tester.sendKeyUpEvent(LogicalKeyboardKey.controlLeft);
      await tester.pumpAndSettle();
    }

    await shortcut();
    expect(input, findsOneWidget);
    await tester.tap(find.text('Cancel'));
    await tester.pumpAndSettle();
    expect(go.callsTo('study.SetDueDate'), hasLength(2));
    expect(richText('dos'), findsOneWidget);
    await shortcut();
    await tester.enterText(input, '30!');
    await tester.testTextInput.receiveAction(TextInputAction.done);
    await tester.pumpAndSettle();
    expect(go.callsTo('study.SetDueDate').last, [2, '30!']);
    expect(richText('tres'), findsOneWidget);
    expect(reviews(), '1');
  });

  testWidgets('reset card from the study menu (spec)', (tester) async {
    final go = FakeGo();
    final study = FakeStudy(go, [
      studyCard(1, 'uno', 'one', kind: 'review'),
      studyCard(2, 'dos', 'two', kind: 'review'),
    ]);
    // Like Go: the card is New again and the session moves on.
    go.handle('study.Reset', (_) {
      study.pos++;
      return study.state();
    });
    await pumpStudy(tester);
    Future<void> openReset() async {
      await tester.tap(find.byTooltip('Card actions'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Reset card'));
      await tester.pumpAndSettle();
    }

    bool checked(String key) =>
        tester.widget<CheckboxListTile>(find.byKey(Key(key))).value!;

    // The options start with their defaults; Cancel changes nothing.
    await openReset();
    expect(find.text('Reset 1 card to new?'), findsOneWidget);
    expect(checked('reset-restore-position'), isTrue);
    expect(checked('reset-counts'), isFalse);
    await tester.tap(find.text('Cancel'));
    await tester.pumpAndSettle();
    expect(go.callsTo('study.Reset'), isEmpty);
    expect(richText('uno'), findsOneWidget);

    await openReset();
    await tester.tap(find.widgetWithText(FilledButton, 'Reset'));
    await tester.pumpAndSettle();
    expect(go.callsTo('study.Reset').single, [
      1,
      {'restorePosition': true, 'resetCounts': false},
    ]);
    expect(richText('dos'), findsOneWidget);
  });

  group('card info', () {
    /// Card info as Go reports it: card 1 with the answers in [study].
    void fakeCardInfo(FakeGo go, FakeStudy study) {
      go.handle('browse.CardInfo', (args) {
        final id = args[0] as int;
        final ratings = [
          for (final (i, a) in study.answers.indexed)
            if (a[0] == id) (i, a[1] as int),
        ];
        return cardInfoJson(
          id,
          history: [
            for (final (i, r) in ratings.reversed)
              historyEntry(
                DateTime.utc(2026, 10, 13, 12, i),
                r,
                kind: 'learning',
              ),
          ],
        );
      });
    }

    testWidgets('Ctrl+Alt+I opens the previously answered card (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      final study = FakeStudy(go, [
        studyCard(1, 'hola', 'hello'),
        studyCard(2, 'perro', 'dog'),
      ]);
      fakeCardInfo(go, study);
      await pumpStudy(tester);
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      await pressKeys(tester, [LogicalKeyboardKey.digit1]);
      expect(richText('perro'), findsOneWidget);

      await pressKeys(tester, [
        LogicalKeyboardKey.controlLeft,
        LogicalKeyboardKey.altLeft,
        LogicalKeyboardKey.keyI,
      ]);
      expect(go.callsTo('browse.CardInfo'), [
        [1],
      ]);
      expect(find.text('Card info'), findsOneWidget);
      expect(find.text('Again · Learning', findRichText: true), findsOneWidget);

      // Closing returns to the next card, unrevealed; nothing was answered.
      await tester.pageBack();
      await tester.pumpAndSettle();
      expect(richText('perro'), findsOneWidget);
      expect(find.text('Show answer'), findsOneWidget);
      expect(go.callsTo('study.Answer'), hasLength(1));
    });

    testWidgets('closing card info returns to the same revealed answer '
        '(spec)', (tester) async {
      final go = FakeGo();
      final study = FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
      fakeCardInfo(go, study);
      await pumpStudy(tester);
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();

      await pressKeys(tester, [LogicalKeyboardKey.keyI]);
      expect(go.callsTo('browse.CardInfo'), [
        [1],
      ]);
      expect(find.text('No reviews yet.'), findsOneWidget);
      await tester.pageBack();
      await tester.pumpAndSettle();
      expect(richText('hello'), findsOneWidget);
      for (final label in ['Again', 'Hard', 'Good', 'Easy']) {
        expect(find.text(label), findsOneWidget);
      }

      // The same from the card menu; it calls nothing in the session.
      await tester.tap(find.byTooltip('Card actions'));
      await tester.pumpAndSettle();
      expect(find.text('Previous card info'), findsNothing);
      await tester.tap(find.text('Card info'));
      await tester.pumpAndSettle();
      expect(go.callsTo('browse.CardInfo'), hasLength(2));
      await tester.pageBack();
      await tester.pumpAndSettle();
      expect(richText('hello'), findsOneWidget);
      expect(find.text('Good'), findsOneWidget);
      expect(go.callsTo('study.Current'), isEmpty);
      expect(go.callsTo('study.Answer'), isEmpty);
    });

    testWidgets('previous card info before any answer says so', (tester) async {
      final go = FakeGo();
      FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
      await pumpStudy(tester);
      await pressKeys(tester, [
        LogicalKeyboardKey.controlLeft,
        LogicalKeyboardKey.altLeft,
        LogicalKeyboardKey.keyI,
      ]);
      expect(find.text('No card has been answered yet.'), findsOneWidget);
      expect(go.callsTo('browse.CardInfo'), isEmpty);
    });
  });

  group('flags', () {
    final marker = find.byKey(const Key('flag-marker'));
    Color? markerColor(WidgetTester tester) =>
        tester.widget<Icon>(marker).color;

    testWidgets('Ctrl+1 toggles the red flag, Ctrl+3 changes it (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
      await pumpStudy(tester);
      expect(marker, findsNothing);
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();

      // Toggle with keyboard (spec).
      await pressKeys(tester, [
        LogicalKeyboardKey.controlLeft,
        LogicalKeyboardKey.digit1,
      ]);
      expect(go.callsTo('study.SetFlag'), [
        [1, 1],
      ]);
      expect(markerColor(tester), FlagColors.light.flag(1));
      // Flagging is no answer: the answer stays shown, nothing was rated.
      expect(find.text('Good'), findsOneWidget);
      expect(go.callsTo('study.Answer'), isEmpty);

      // Change flag (spec).
      await pressKeys(tester, [
        LogicalKeyboardKey.controlLeft,
        LogicalKeyboardKey.digit3,
      ]);
      expect(markerColor(tester), FlagColors.light.flag(3));
      await pressKeys(tester, [
        LogicalKeyboardKey.metaLeft,
        LogicalKeyboardKey.digit3,
      ]);
      expect(go.callsTo('study.SetFlag').last, [1, 3]);
      expect(marker, findsNothing);
      expect(find.text('Good'), findsOneWidget);
    });

    testWidgets('the card menu sets and removes flags by their names', (
      tester,
    ) async {
      final go = FakeGo();
      go.settings['flagNames'] = ['Fix later', '', '', '', '', '', ''];
      FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
      await pumpStudy(tester, loadSettings: true);

      await tester.tap(find.byTooltip('Card actions'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Flag'));
      await tester.pumpAndSettle();
      // Rename a flag (spec): the study menu shows the custom name.
      expect(find.text('Fix later'), findsOneWidget);
      expect(find.text('Red'), findsNothing);
      expect(find.text('Turquoise'), findsOneWidget);
      expect(find.text('Remove flag'), findsNothing);
      expect(find.text('Green'), findsOneWidget);
      await tester.tap(find.byKey(const Key('flag-item-3')));
      await tester.pumpAndSettle();
      expect(go.callsTo('study.SetFlag'), [
        [1, 3],
      ]);
      expect(markerColor(tester), FlagColors.light.flag(3));

      await tester.tap(find.byTooltip('Card actions'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Flag'));
      await tester.pumpAndSettle();
      expect(
        tester
            .widget<CheckedPopupMenuItem<int>>(
              find.byKey(const Key('flag-item-3')),
            )
            .checked,
        isTrue,
      );
      await tester.tap(find.text('Remove flag'));
      await tester.pumpAndSettle();
      expect(go.callsTo('study.SetFlag').last, [1, 0]);
      expect(marker, findsNothing);
    });
  });

  group('marked notes', () {
    final star = find.byKey(const Key('marked-star'));

    /// Types `*` as Shift+8 does on a US keyboard.
    Future<void> pressStar(WidgetTester tester) async {
      await tester.sendKeyDownEvent(LogicalKeyboardKey.shiftLeft);
      await tester.sendKeyDownEvent(LogicalKeyboardKey.digit8, character: '*');
      await tester.sendKeyUpEvent(LogicalKeyboardKey.digit8);
      await tester.sendKeyUpEvent(LogicalKeyboardKey.shiftLeft);
      await tester.pumpAndSettle();
    }

    testWidgets('* toggles the mark and the star (spec)', (tester) async {
      final go = FakeGo();
      FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
      await pumpStudy(tester);
      expect(star, findsNothing);
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();

      // Toggle with key (spec).
      await pressStar(tester);
      expect(go.callsTo('study.ToggleMark'), [
        [1],
      ]);
      expect(star, findsOneWidget);
      expect(tester.widget<Icon>(star).color, FlagColors.light.marked);
      // Marking is no answer: the answer stays shown, nothing was rated.
      expect(find.text('Good'), findsOneWidget);
      expect(go.callsTo('study.Answer'), isEmpty);

      await pressStar(tester);
      expect(go.callsTo('study.ToggleMark'), hasLength(2));
      expect(star, findsNothing);
      expect(find.text('Good'), findsOneWidget);
    });

    testWidgets('the card menu marks and unmarks the note', (tester) async {
      final go = FakeGo();
      FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
      await pumpStudy(tester);

      await tester.tap(find.byTooltip('Card actions'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Mark note'));
      await tester.pumpAndSettle();
      expect(go.callsTo('study.ToggleMark'), [
        [1],
      ]);
      expect(star, findsOneWidget);

      await tester.tap(find.byTooltip('Card actions'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Unmark note'));
      await tester.pumpAndSettle();
      expect(star, findsNothing);
    });

    testWidgets('* types into the answer input', (tester) async {
      final go = FakeGo();
      FakeStudy(go, [
        studyCard(
          1,
          'dog <cf-typeans></cf-typeans>',
          'perro',
          typeAnswer: true,
        ),
      ]);
      await pumpStudy(tester);
      await pressStar(tester);
      expect(go.callsTo('study.ToggleMark'), isEmpty);
    });
  });

  testWidgets('an open study screen counts as an open session', (tester) async {
    final go = FakeGo();
    FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
    setSize(tester, phone);
    final app = AppState(isWeb: false);
    addTearDown(app.dispose);
    await tester.pumpWidget(
      AppScope(
        state: app,
        child: MaterialApp(
          home: Builder(
            builder: (context) => TextButton(
              onPressed: () => Navigator.of(context).push(
                MaterialPageRoute<void>(
                  builder: (_) =>
                      const StudyScreen(deckId: 1, deckName: 'Default'),
                ),
              ),
              child: const Text('Study'),
            ),
          ),
        ),
      ),
    );
    expect(app.openSessions.value, 0);
    await tester.tap(find.text('Study'));
    await tester.pumpAndSettle();
    expect(app.openSessions.value, 1);
    await tester.pageBack();
    await tester.pumpAndSettle();
    expect(find.byType(StudyScreen), findsNothing);
    expect(app.openSessions.value, 0);
  });

  group('learning cards becoming due', () {
    testWidgets('counts update and the revealed answer stays', (tester) async {
      final go = FakeGo();
      final cards = [studyCard(1, 'hola', 'hello', kind: 'review')];
      final study = FakeStudy(go, cards)..learningDue = inAMinute();
      await pumpStudy(tester);
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      expect(
        tester.widget<Text>(find.byKey(const Key('remaining-learning'))).data,
        '0',
      );
      // The learning card is now due, after the card on screen.
      cards.add(studyCard(2, 'perro', 'dog', kind: 'learning'));
      study.learningDue = null;
      await waitForLearningCard(tester);
      expect(go.callsTo('study.Current'), hasLength(1));
      expect(
        tester.widget<Text>(find.byKey(const Key('remaining-learning'))).data,
        '1',
      );
      expect(richText('hello'), findsOneWidget);
      expect(find.text('Good'), findsOneWidget);
    });

    testWidgets('auto-continue shows the due card without a tap', (
      tester,
    ) async {
      final go = FakeGo();
      final cards = [studyCard(1, 'hola', 'hello')];
      noLearnAhead(go);
      final study = FakeStudy(go, cards);
      final app = await pumpStudy(tester);
      app.settings.autoContinueLearning = true;
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      study.learningDue = inAMinute();
      await tester.tap(find.byKey(const Key('rate-1')));
      await tester.pumpAndSettle();
      expect(
        find.textContaining('More learning cards will be due'),
        findsOneWidget,
      );
      cards.add(studyCard(1, 'hola', 'hello', kind: 'learning'));
      study.learningDue = null;
      await waitForLearningCard(tester);
      expect(richText('hola'), findsOneWidget);
      expect(find.text('Show answer'), findsOneWidget);
      expect(find.text('Session complete'), findsNothing);
    });

    testWidgets('by default the summary asks, and Continue shows the card', (
      tester,
    ) async {
      final go = FakeGo();
      noLearnAhead(go);
      final cards = [studyCard(1, 'hola', 'hello')];
      final study = FakeStudy(go, cards);
      await pumpStudy(tester);
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      study.learningDue = inAMinute();
      await tester.tap(find.byKey(const Key('rate-1')));
      await tester.pumpAndSettle();
      cards.add(studyCard(2, 'perro', 'dog', kind: 'learning'));
      study.learningDue = null;
      await waitForLearningCard(tester);
      expect(find.text('Learning cards are due now'), findsOneWidget);
      expect(find.textContaining('You studied 1 card'), findsOneWidget);
      expect(richText('perro'), findsNothing);

      await tester.tap(find.text('Continue'));
      await tester.pumpAndSettle();
      expect(richText('perro'), findsOneWidget);
      expect(find.text('Show answer'), findsOneWidget);
      expect(go.callsTo('study.Current'), hasLength(1));
    });

    testWidgets('nothing due right now asks when the card becomes due', (
      tester,
    ) async {
      final go = FakeGo();
      noLearnAhead(go);
      final cards = <Map<String, Object?>>[];
      final study = FakeStudy(go, cards)..learningDue = inAMinute();
      await pumpStudy(tester);
      expect(
        find.text('Congratulations! Nothing is due right now.'),
        findsOneWidget,
      );
      cards.add(studyCard(1, 'hola', 'hello', kind: 'learning'));
      study.learningDue = null;
      await waitForLearningCard(tester);
      expect(find.text('Learning cards are due now'), findsOneWidget);
      await tester.tap(find.text('Continue'));
      await tester.pumpAndSettle();
      expect(richText('hola'), findsOneWidget);
    });

    testWidgets('a learning card within the limit is shown without waiting', (
      tester,
    ) async {
      final go = FakeGo();
      // Go shows the learning card due in 5 minutes after the last review.
      FakeStudy(go, [
        studyCard(1, 'hola', 'hello', kind: 'review'),
        studyCard(2, 'perro', 'dog', kind: 'learning'),
      ]);
      await pumpStudy(tester);
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('rate-3')));
      await tester.pumpAndSettle();
      expect(richText('perro'), findsOneWidget);
      expect(find.text('Show answer'), findsOneWidget);
      expect(find.text('Session complete'), findsNothing);
    });

    testWidgets('the summary reacts when a card comes within the limit', (
      tester,
    ) async {
      final go = FakeGo();
      final cards = [studyCard(1, 'hola', 'hello')];
      final study = FakeStudy(go, cards);
      await pumpStudy(tester);
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      // Due in 21 minutes: within the limit of 20 in one minute.
      final due = DateTime.now().add(const Duration(minutes: 21));
      study.learningDue = due.toUtc().toIso8601String();
      await tester.tap(find.byKey(const Key('rate-1')));
      await tester.pumpAndSettle();
      expect(
        find.text('More learning cards will be due ${formatDue(due)}.'),
        findsOneWidget,
      );
      await tester.pump(const Duration(seconds: 30));
      expect(go.callsTo('study.Current'), isEmpty);
      cards.add(studyCard(2, 'perro', 'dog', kind: 'learning'));
      study.learningDue = null;
      await waitForLearningCard(tester);
      expect(find.text('Learning cards are due now'), findsOneWidget);
      await tester.tap(find.text('Continue'));
      await tester.pumpAndSettle();
      expect(richText('perro'), findsOneWidget);
    });

    testWidgets('nothing happens when no learning card is pending', (
      tester,
    ) async {
      final go = FakeGo();
      FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
      await pumpStudy(tester);
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('rate-3')));
      await tester.pumpAndSettle();
      await tester.pump(const Duration(minutes: 10));
      await tester.pumpAndSettle();
      expect(find.text('Session complete'), findsOneWidget);
      expect(go.callsTo('study.Current'), isEmpty);
    });
  });

  test('due times read naturally', () {
    final now = DateTime(2026, 10, 5, 12);
    expect(formatDue(DateTime(2026, 10, 5, 14, 5), now: now), 'at 14:05');
    expect(formatDue(DateTime(2026, 10, 6, 4), now: now), 'tomorrow');
    expect(formatDue(DateTime(2026, 10, 9, 4), now: now), 'in 4 days');
    expect(formatDue(DateTime(2026, 11, 2, 4), now: now), 'on 2026-11-02');
  });

  group('audio', () {
    testWidgets('the question autoplays (spec)', (tester) async {
      final backend = fakeMedia();
      final go = FakeGo();
      FakeStudy(go, [audioCard(1)]);
      await pumpStudy(tester);
      expect(backend.played, ['hola.mp3']);
      expect(find.byTooltip('Stop'), findsOneWidget);
    });

    testWidgets('with autoplay off, nothing plays until replay (spec)', (
      tester,
    ) async {
      final backend = fakeMedia();
      final go = FakeGo();
      FakeStudy(go, [audioCard(1, autoplay: false)]);
      await pumpStudy(tester);
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      expect(backend.played, isEmpty);
      await tester.tap(find.byTooltip('Play hello.mp3'));
      await tester.pump();
      expect(backend.played, ['hello.mp3']);
    });

    testWidgets('revealing stops the question and plays the answer', (
      tester,
    ) async {
      final backend = fakeMedia();
      final go = FakeGo();
      FakeStudy(go, [audioCard(1)]);
      await pumpStudy(tester);
      await tester.sendKeyEvent(LogicalKeyboardKey.space);
      await tester.pumpAndSettle();
      expect(backend.log, ['stop']);
      expect(backend.played, ['hola.mp3', 'hello.mp3']);
    });

    testWidgets('without interrupting, the answer waits for the question', (
      tester,
    ) async {
      final backend = fakeMedia();
      final go = FakeGo();
      go.settings = {...go.settings, 'interruptAudio': false};
      FakeStudy(go, [audioCard(1)]);
      await pumpStudy(tester, loadSettings: true);
      await tester.sendKeyEvent(LogicalKeyboardKey.space);
      await tester.pumpAndSettle();
      expect(backend.played, ['hola.mp3']);
      backend.finish();
      await tester.pumpAndSettle();
      expect(backend.played, ['hola.mp3', 'hello.mp3']);
    });

    testWidgets('R on the answer replays the question, then the answer '
        '(spec)', (tester) async {
      final backend = fakeMedia();
      final go = FakeGo();
      FakeStudy(go, [audioCard(1)]);
      await pumpStudy(tester);
      await tester.sendKeyEvent(LogicalKeyboardKey.space);
      await tester.pumpAndSettle();
      backend.played.clear();
      await tester.sendKeyEvent(LogicalKeyboardKey.keyR);
      await tester.pumpAndSettle();
      expect(backend.played, ['hola.mp3']);
      backend.finish();
      await tester.pumpAndSettle();
      expect(backend.played, ['hola.mp3', 'hello.mp3']);
      // F5 does the same; the deck option skips the question.
      backend.finish();
      await tester.pumpAndSettle();
      backend.played.clear();
      await tester.sendKeyEvent(LogicalKeyboardKey.f5);
      await tester.pumpAndSettle();
      expect(backend.played, ['hola.mp3']);
    });

    testWidgets('replaying can skip the question', (tester) async {
      final backend = fakeMedia();
      final go = FakeGo();
      FakeStudy(go, [audioCard(1, replayQuestion: false)]);
      await pumpStudy(tester);
      await tester.sendKeyEvent(LogicalKeyboardKey.space);
      await tester.pumpAndSettle();
      backend.played.clear();
      await tester.sendKeyEvent(LogicalKeyboardKey.keyR);
      await tester.pumpAndSettle();
      expect(backend.played, ['hello.mp3']);
    });

    testWidgets('5 pauses and resumes, 6 and 7 seek (spec)', (tester) async {
      final backend = fakeMedia();
      final go = FakeGo();
      FakeStudy(go, [audioCard(1)]);
      await pumpStudy(tester);
      await tester.sendKeyEvent(LogicalKeyboardKey.digit5);
      await tester.sendKeyEvent(LogicalKeyboardKey.digit6);
      await tester.sendKeyEvent(LogicalKeyboardKey.digit7);
      await tester.sendKeyEvent(LogicalKeyboardKey.digit5);
      expect(backend.log, ['pause', 'seek -5', 'seek 5', 'resume']);
    });

    testWidgets('answering stops the card and plays the next one', (
      tester,
    ) async {
      final backend = fakeMedia();
      final go = FakeGo();
      FakeStudy(go, [audioCard(1), audioCard(2)]);
      await pumpStudy(tester);
      await tester.sendKeyEvent(LogicalKeyboardKey.space);
      await tester.pumpAndSettle();
      await tester.sendKeyEvent(LogicalKeyboardKey.digit3);
      await tester.pumpAndSettle();
      expect(backend.log, ['stop', 'stop']);
      expect(backend.played, ['hola.mp3', 'hello.mp3', 'hola.mp3']);
    });

    testWidgets('leaving the session stops playback', (tester) async {
      final backend = fakeMedia();
      final go = FakeGo();
      FakeStudy(go, [audioCard(1)]);
      await pumpStudy(tester);
      await tester.pumpWidget(const SizedBox());
      expect(backend.log, ['close']);
    });

    testWidgets('back during a video stops it before the player closes '
        '(spec)', (tester) async {
      final backend = FakeBackend();
      useFakeMedia(backend, FakeMediaFiles(['clip.mp4']));
      final go = FakeGo();
      FakeStudy(go, [
        studyCard(
          1,
          '<cf-video idx="0"></cf-video>',
          'hello',
          questionMedia: [
            {'file': 'clip.mp4', 'kind': 'video'},
          ],
        ),
      ]);
      await pumpApp(tester);
      await studyDeck(tester, 'Default');
      expect(backend.playing, 'clip.mp4');
      expect(find.byKey(const Key('fake-video')), findsOneWidget);

      await tester.pageBack();
      await tester.pump();
      // The video is gone as soon as the pop starts.
      expect(backend.log, ['stop']);
      expect(find.byKey(const Key('fake-video')), findsNothing);
      await tester.pumpAndSettle();
      expect(find.byType(StudyScreen), findsNothing);
      expect(find.text('Default'), findsOneWidget);
      expect(backend.log, ['stop', 'close']);

      await studyDeck(tester, 'Default');
      expect(backend.played, ['clip.mp4', 'clip.mp4']);
      expect(backend.playing, 'clip.mp4');
    });

    group('text to speech', () {
      testWidgets('the question is spoken without user action (spec)', (
        tester,
      ) async {
        final engine = fakeSpeech();
        final go = FakeGo();
        FakeStudy(go, [ttsCard(1)]);
        await pumpStudy(tester);
        expect(engine.spoken, ['es_ES:perro']);
        expect(find.byTooltip('Stop'), findsOneWidget);
      });

      testWidgets('revealing stops the question and speaks the answer', (
        tester,
      ) async {
        final engine = fakeSpeech();
        final go = FakeGo();
        FakeStudy(go, [ttsCard(1)]);
        await pumpStudy(tester);
        expect(engine.speaking, 'es_ES:perro');
        await tester.sendKeyEvent(LogicalKeyboardKey.space);
        await tester.pumpAndSettle();
        expect(engine.stops, 1);
        expect(engine.spoken, ['es_ES:perro', 'en_US:dog']);
      });

      testWidgets('R on the answer speaks the side again (spec)', (
        tester,
      ) async {
        final engine = fakeSpeech();
        final go = FakeGo();
        FakeStudy(go, [ttsCard(1)]);
        await pumpStudy(tester);
        await tester.sendKeyEvent(LogicalKeyboardKey.space);
        await tester.pumpAndSettle();
        engine.finish();
        await tester.pumpAndSettle();
        engine.spoken.clear();
        await tester.sendKeyEvent(LogicalKeyboardKey.keyR);
        await tester.pumpAndSettle();
        expect(engine.spoken, ['es_ES:perro']);
        engine.finish();
        await tester.pumpAndSettle();
        expect(engine.spoken, ['es_ES:perro', 'en_US:dog']);
        // F5 does the same.
        engine.finish();
        await tester.pumpAndSettle();
        engine.spoken.clear();
        await tester.sendKeyEvent(LogicalKeyboardKey.f5);
        await tester.pumpAndSettle();
        expect(engine.spoken, ['es_ES:perro']);
      });

      testWidgets('leaving the card stops speaking', (tester) async {
        final engine = fakeSpeech();
        final go = FakeGo();
        FakeStudy(go, [ttsCard(1), ttsCard(2)]);
        await pumpStudy(tester);
        await tester.sendKeyEvent(LogicalKeyboardKey.space);
        await tester.pumpAndSettle();
        expect(engine.speaking, 'en_US:dog');
        await tester.sendKeyEvent(LogicalKeyboardKey.digit3);
        await tester.pumpAndSettle();
        expect(engine.stops, 2);
        expect(engine.speaking, 'es_ES:perro');
        await tester.pumpWidget(const SizedBox());
        expect(engine.stops, 3);
        expect(engine.speaking, isNull);
      });

      testWidgets('no voice: the button explains, studying goes on (spec)', (
        tester,
      ) async {
        final engine = fakeSpeech();
        final go = FakeGo();
        final study = FakeStudy(go, [ttsCard(1, answerLang: 'ja_JP')]);
        await pumpStudy(tester);
        await tester.sendKeyEvent(LogicalKeyboardKey.space);
        await tester.pumpAndSettle();
        expect(engine.spoken, ['es_ES:perro']);
        final button = tester.widget<IconButton>(
          find.byKey(const Key('tts-answer-0')),
        );
        expect(button.onPressed, isNull);
        expect(button.tooltip, 'No voice is installed for ja_JP');
        await tester.sendKeyEvent(LogicalKeyboardKey.digit3);
        await tester.pumpAndSettle();
        expect(study.answers, hasLength(1));
      });
    });

    testWidgets('a recording plays back with V and is dropped on the next '
        'card (spec)', (tester) async {
      final backend = fakeMedia();
      final recorder = FakeRecorder();
      useFakeRecorder(recorder);
      final go = FakeGo();
      FakeStudy(go, [
        studyCard(1, 'hola', 'hello'),
        studyCard(2, 'adiós', 'bye'),
      ]);
      await pumpStudy(tester);
      // Compare pronunciation: record, stop, play back.
      await tester.sendKeyDownEvent(LogicalKeyboardKey.shiftLeft);
      await tester.sendKeyEvent(LogicalKeyboardKey.keyV);
      await tester.sendKeyUpEvent(LogicalKeyboardKey.shiftLeft);
      await tester.pumpAndSettle();
      expect(recorder.recording, isTrue);
      expect(find.byTooltip('Stop recording'), findsOneWidget);
      await tester.tap(find.byTooltip('Stop recording'));
      await tester.pumpAndSettle();
      expect(recorder.recording, isFalse);
      await tester.sendKeyEvent(LogicalKeyboardKey.keyV);
      await tester.pump();
      expect(backend.played, ['voice-0']);
      backend.finish();
      // Recording not kept: answering moves on and discards it.
      await tester.sendKeyEvent(LogicalKeyboardKey.space);
      await tester.pumpAndSettle();
      await tester.sendKeyEvent(LogicalKeyboardKey.digit3);
      await tester.pumpAndSettle();
      expect(richText('adiós'), findsOneWidget);
      expect(recorder.discarded, ['mem:voice-0']);
      await tester.sendKeyEvent(LogicalKeyboardKey.keyV);
      await tester.pump();
      expect(backend.played, ['voice-0']);
      // Nothing was stored in the collection.
      expect(
        go.calls.keys.where(
          (c) => !c.startsWith('study.') && !c.startsWith('app.'),
        ),
        isEmpty,
      );
    });

    testWidgets('the menu records too', (tester) async {
      fakeMedia();
      final recorder = FakeRecorder();
      useFakeRecorder(recorder);
      final go = FakeGo();
      FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
      await pumpStudy(tester);
      await tester.tap(find.byTooltip('Card actions'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Record own voice'));
      await tester.pumpAndSettle();
      expect(recorder.recording, isTrue);
      await tester.tap(find.byTooltip('Card actions'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Stop recording'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Card actions'));
      await tester.pumpAndSettle();
      expect(find.text('Play own voice'), findsOneWidget);
    });

    testWidgets('a refused microphone is explained and studying goes on', (
      tester,
    ) async {
      fakeMedia();
      final recorder = FakeRecorder()..refuse = true;
      useFakeRecorder(recorder);
      final go = FakeGo();
      final study = FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
      await pumpStudy(tester);
      await tester.sendKeyDownEvent(LogicalKeyboardKey.shiftLeft);
      await tester.sendKeyEvent(LogicalKeyboardKey.keyV);
      await tester.sendKeyUpEvent(LogicalKeyboardKey.shiftLeft);
      await tester.pumpAndSettle();
      expect(find.textContaining('may not use the microphone'), findsOneWidget);
      await tester.tap(find.text('OK'));
      await tester.pumpAndSettle();
      expect(find.byTooltip('Stop recording'), findsNothing);
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Good'));
      await tester.pumpAndSettle();
      expect(study.answers, hasLength(1));
    });
  });

  group('timers', () {
    final timer = find.byKey(const Key('answerTimer'));
    String shown(WidgetTester tester) => tester.widget<Text>(timer).data!;
    Color? colour(WidgetTester tester) =>
        tester.widget<Text>(timer).style?.color;
    ColorScheme scheme(WidgetTester tester) =>
        Theme.of(tester.element(timer)).colorScheme;

    Future<void> answer(
      WidgetTester tester, {
      Duration after = Duration.zero,
    }) async {
      await tester.pump(after);
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('rate-3')));
      await tester.pumpAndSettle();
    }

    testWidgets('the timer is hidden unless the deck shows it', (tester) async {
      final go = FakeGo();
      FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
      await pumpStudy(tester);
      expect(timer, findsNothing);
    });

    testWidgets('the timer stops on answer (spec)', (tester) async {
      final go = FakeGo();
      final study = FakeStudy(go, [
        studyCard(1, 'hola', 'hello', showTimer: true),
        studyCard(2, 'perro', 'dog', showTimer: true),
      ]);
      await pumpStudy(tester);
      expect(shown(tester), '0:00');
      await tester.pump(const Duration(seconds: 3));
      expect(shown(tester), '0:03');
      // Timer stops on answer (spec): revealed after 8 seconds, it shows 8
      // seconds until the card is rated.
      await tester.pump(const Duration(seconds: 5));
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      expect(shown(tester), '0:08');
      await tester.pump(const Duration(seconds: 30));
      expect(shown(tester), '0:08');
      await tester.tap(find.byKey(const Key('rate-3')));
      await tester.pumpAndSettle();
      // The timer does not change what is recorded.
      expect(study.answers.single[2] as int, greaterThanOrEqualTo(38000));
      expect(richText('perro'), findsOneWidget);
      expect(shown(tester), '0:00');
    });

    testWidgets(
      'the timer changes colour at the maximum and can keep running',
      (tester) async {
        final go = FakeGo();
        FakeStudy(go, [
          studyCard(
            1,
            'hola',
            'hello',
            showTimer: true,
            stopTimerOnAnswer: false,
            maxAnswerSeconds: 10,
          ),
        ]);
        await pumpStudy(tester);
        await tester.pump(const Duration(seconds: 9));
        expect(shown(tester), '0:09');
        expect(colour(tester), scheme(tester).onSurfaceVariant);
        await tester.pump(const Duration(seconds: 1));
        expect(shown(tester), '0:10');
        expect(colour(tester), scheme(tester).error);
        await tester.tap(find.text('Show answer'));
        await tester.pumpAndSettle();
        await tester.pump(const Duration(seconds: 65));
        expect(shown(tester), '1:15');
      },
    );

    List<Map<String, Object?>> cards(int n) => [
      for (var i = 1; i <= n; i++) studyCard(i, 'q$i', 'a$i', kind: 'review'),
    ];

    /// Answers 42 cards in a little over 10 minutes, the last one after
    /// the time has passed.
    Future<void> studyTenMinutes(WidgetTester tester) async {
      for (var i = 0; i < 41; i++) {
        await answer(tester, after: const Duration(seconds: 14));
        expect(find.byKey(const Key('timeboxMessage')), findsNothing);
      }
      await answer(tester, after: const Duration(seconds: 30));
    }

    testWidgets('a timebox reports the cards studied (spec)', (tester) async {
      final go = FakeGo();
      go.settings = {...go.settings, 'timeboxMinutes': 10};
      final study = FakeStudy(go, cards(50));
      await pumpStudy(tester, loadSettings: true);
      // Ten-minute timebox (spec): 42 cards answered in 10 minutes.
      await studyTenMinutes(tester);
      expect(find.byKey(const Key('timeboxMessage')), findsOneWidget);
      expect(find.text('You studied 42 cards in 10 minutes.'), findsOneWidget);
      // Continue starts the next timebox.
      await tester.tap(find.text('Continue'));
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('timeboxMessage')), findsNothing);
      expect(richText('q43'), findsOneWidget);
      await answer(tester, after: const Duration(minutes: 5));
      expect(find.byKey(const Key('timeboxMessage')), findsNothing);
      await answer(tester, after: const Duration(minutes: 5));
      expect(find.text('You studied 2 cards in 10 minutes.'), findsOneWidget);
      expect(study.answers, hasLength(44));
    });

    testWidgets('Finish ends the session with its summary', (tester) async {
      final go = FakeGo();
      go.settings = {...go.settings, 'timeboxMinutes': 10};
      final study = FakeStudy(go, cards(50));
      await pumpStudy(tester, loadSettings: true);
      await studyTenMinutes(tester);
      await tester.tap(find.text('Finish'));
      await tester.pumpAndSettle();
      expect(find.text('Session complete'), findsOneWidget);
      expect(find.textContaining('You studied 42 cards in'), findsOneWidget);
      expect(find.text('Show answer'), findsNothing);
      expect(find.text('Check again'), findsNothing);
      // Nothing more is answered or undone.
      await tester.sendKeyEvent(LogicalKeyboardKey.space);
      await tester.pumpAndSettle();
      expect(study.answers, hasLength(42));
      final undo = find.widgetWithIcon(IconButton, Icons.undo);
      expect(tester.widget<IconButton>(undo).onPressed, isNull);
    });

    testWidgets('no timebox message while it is off', (tester) async {
      final go = FakeGo();
      FakeStudy(go, cards(3));
      await pumpStudy(tester, loadSettings: true);
      await answer(tester, after: const Duration(hours: 1));
      expect(find.byKey(const Key('timeboxMessage')), findsNothing);
    });
  });

  group('auto advance', () {
    final indicator = find.byKey(const Key('autoAdvanceOn'));
    final showAnswer = find.text('Show answer');
    final good = find.byKey(const Key('rate-3'));

    Map<String, Object?> timed(
      int id,
      String front,
      String back, {
      String questionAction = 'showAnswer',
      String answerAction = 'answerGood',
    }) => studyCard(
      id,
      front,
      back,
      secondsToShowQuestion: 5,
      secondsToShowAnswer: 3,
      questionAction: questionAction,
      answerAction: answerAction,
    );

    Future<void> toggle(WidgetTester tester) async {
      await tester.sendKeyDownEvent(LogicalKeyboardKey.shiftLeft);
      await tester.sendKeyEvent(LogicalKeyboardKey.keyA);
      await tester.sendKeyUpEvent(LogicalKeyboardKey.shiftLeft);
      await tester.pump();
    }

    testWidgets('hands-free review (spec)', (tester) async {
      final go = FakeGo();
      final study = FakeStudy(go, [
        timed(1, 'hola', 'hello'),
        timed(2, 'perro', 'dog'),
        timed(3, 'gato', 'cat'),
      ]);
      await pumpStudy(tester);
      await toggle(tester);
      expect(indicator, findsOneWidget);
      // The answer appears after 5 seconds ...
      await tester.pump(const Duration(milliseconds: 4900));
      expect(showAnswer, findsOneWidget);
      await tester.pump(const Duration(milliseconds: 100));
      await tester.pump();
      expect(good, findsOneWidget);
      // ... and 3 seconds later the card is rated Good and the next appears.
      await tester.pump(const Duration(milliseconds: 2900));
      expect(study.answers, isEmpty);
      await tester.pump(const Duration(milliseconds: 100));
      await tester.pump();
      expect(study.answers.single[1], 3);
      expect(richText('perro'), findsOneWidget);
      expect(showAnswer, findsOneWidget);
      await tester.pump(const Duration(seconds: 5));
      await tester.pump();
      await tester.pump(const Duration(seconds: 3));
      await tester.pump();
      expect(study.answers.map((a) => a[1]), [3, 3]);
      expect(richText('gato'), findsOneWidget);
      // Turned off, nothing happens any more.
      await toggle(tester);
      expect(indicator, findsNothing);
      await tester.pump(const Duration(seconds: 30));
      expect(showAnswer, findsOneWidget);
      expect(study.answers, hasLength(2));
    });

    testWidgets('off by default (spec)', (tester) async {
      final go = FakeGo();
      final study = FakeStudy(go, [timed(1, 'hola', 'hello')]);
      await pumpStudy(tester);
      expect(indicator, findsNothing);
      await tester.pump(const Duration(seconds: 30));
      expect(showAnswer, findsOneWidget);
      await tester.tap(showAnswer);
      await tester.pumpAndSettle();
      await tester.pump(const Duration(seconds: 30));
      expect(study.answers, isEmpty);
    });

    testWidgets('show reminder only reminds; the menu turns it on and off', (
      tester,
    ) async {
      final go = FakeGo();
      final study = FakeStudy(go, [
        timed(
          1,
          'hola',
          'hello',
          questionAction: 'showReminder',
          answerAction: 'showReminder',
        ),
      ]);
      await pumpStudy(tester);
      await tester.tap(find.byTooltip('Card actions'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Auto advance'));
      await tester.pumpAndSettle();
      expect(indicator, findsOneWidget);
      await tester.pump(const Duration(seconds: 5));
      await tester.pump();
      expect(find.text('Time to show the answer.'), findsOneWidget);
      expect(showAnswer, findsOneWidget);
      await waitForSnackBar(tester);
      await tester.tap(showAnswer);
      await tester.pump();
      await tester.pump(const Duration(seconds: 3));
      await tester.pump();
      expect(find.text('Time to answer.'), findsOneWidget);
      expect(study.answers, isEmpty);
      expect(good, findsOneWidget);
      await waitForSnackBar(tester);
      await tester.tap(find.byTooltip('Card actions'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Stop auto advance'));
      await tester.pumpAndSettle();
      expect(indicator, findsNothing);
    });

    testWidgets('the default answer action buries the card', (tester) async {
      final go = FakeGo();
      FakeStudy(go, [
        studyCard(1, 'hola', 'hello', secondsToShowAnswer: 2),
        studyCard(2, 'perro', 'dog'),
      ]);
      await pumpStudy(tester);
      await toggle(tester);
      // 0 seconds for the question: it waits for the user.
      await tester.pump(const Duration(seconds: 30));
      expect(showAnswer, findsOneWidget);
      await tester.tap(showAnswer);
      await tester.pump();
      await tester.pump(const Duration(seconds: 2));
      await tester.pump();
      expect(go.callsTo('study.Bury').single[0], 1);
      expect(richText('perro'), findsOneWidget);
    });

    testWidgets('input restarts the countdown', (tester) async {
      final go = FakeGo();
      FakeStudy(go, [timed(1, 'hola', 'hello')]);
      await pumpStudy(tester);
      await toggle(tester);
      await tester.pump(const Duration(seconds: 4));
      // Marking the note (* as Shift+8 types it) is input on the card.
      await tester.sendKeyDownEvent(LogicalKeyboardKey.shiftLeft);
      await tester.sendKeyDownEvent(LogicalKeyboardKey.digit8, character: '*');
      await tester.sendKeyUpEvent(LogicalKeyboardKey.digit8);
      await tester.sendKeyUpEvent(LogicalKeyboardKey.shiftLeft);
      await tester.pump();
      expect(find.byKey(const Key('marked-star')), findsOneWidget);
      await tester.pump(const Duration(seconds: 4));
      expect(showAnswer, findsOneWidget);
      await tester.pump(const Duration(seconds: 1));
      await tester.pump();
      expect(good, findsOneWidget);
      await toggle(tester);
    });

    testWidgets('the countdown starts once the sounds have finished', (
      tester,
    ) async {
      final backend = fakeMedia();
      final go = FakeGo();
      FakeStudy(go, [
        {...audioCard(1), 'secondsToShowQuestion': 5, 'secondsToShowAnswer': 0},
      ]);
      await pumpStudy(tester);
      await toggle(tester);
      expect(backend.played, ['hola.mp3']);
      await tester.pump(const Duration(seconds: 20));
      expect(showAnswer, findsOneWidget);
      backend.finish();
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 4900));
      expect(showAnswer, findsOneWidget);
      await tester.pump(const Duration(milliseconds: 100));
      await tester.pump();
      expect(good, findsOneWidget);
      await toggle(tester);
    });
  });

  group('review display options', () {
    testWidgets('Space does not rate when the setting is off (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      go.settings = {...go.settings, 'spaceAnswersGood': false};
      final study = FakeStudy(go, [
        studyCard(1, 'hola', 'hello'),
        studyCard(2, 'perro', 'dog'),
      ]);
      await pumpStudy(tester, loadSettings: true);
      for (final key in [
        LogicalKeyboardKey.space,
        LogicalKeyboardKey.space,
        LogicalKeyboardKey.enter,
      ]) {
        await tester.sendKeyEvent(key);
        await tester.pumpAndSettle();
      }
      expect(richText('hello'), findsOneWidget);
      expect(study.answers, isEmpty);
      // The rating keys still work.
      await tester.sendKeyEvent(LogicalKeyboardKey.digit3);
      await tester.pumpAndSettle();
      expect(study.answers.single.take(2), [1, 3]);
    });

    testWidgets('intervals hidden shows only the rating names (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      go.settings = {...go.settings, 'showIntervals': false};
      FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
      await pumpStudy(tester, loadSettings: true);
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      for (final label in ['Again', 'Hard', 'Good', 'Easy']) {
        expect(find.text(label), findsOneWidget);
      }
      for (final interval in ['<1m', '<6m', '3d', '8d']) {
        expect(find.text(interval), findsNothing);
      }
    });

    testWidgets('counts hidden (spec)', (tester) async {
      final go = FakeGo();
      go.settings = {...go.settings, 'showRemainingCounts': false};
      FakeStudy(go, [studyCard(1, 'hola', 'hello', kind: 'review')]);
      await pumpStudy(tester, loadSettings: true);
      expect(richText('hola'), findsOneWidget);
      for (final kind in ['new', 'learning', 'review']) {
        expect(find.byKey(Key('remaining-$kind')), findsNothing);
      }
    });

    testWidgets('the timer stays when the counts are hidden', (tester) async {
      final go = FakeGo();
      go.settings = {...go.settings, 'showRemainingCounts': false};
      FakeStudy(go, [studyCard(1, 'hola', 'hello', showTimer: true)]);
      await pumpStudy(tester, loadSettings: true);
      expect(find.byKey(const Key('answerTimer')), findsOneWidget);
      expect(find.byKey(const Key('remaining-new')), findsNothing);
    });

    group('hidden bars', () {
      /// Opens a study session of Default through the app, on [size].
      Future<FakeGo> studyInApp(
        WidgetTester tester, {
        required bool hide,
        Size size = phone,
      }) async {
        final go = FakeGo();
        go.tree = [deckNode(1, 'Default', newCount: 2)];
        go.settings = {...go.settings, 'hideBarsDuringReview': hide};
        FakeStudy(go, [
          studyCard(1, 'hola', 'hello'),
          studyCard(2, 'perro', 'dog'),
        ]);
        await pumpApp(tester, size: size);
        await tester.tap(find.text('Default'));
        await tester.pumpAndSettle();
        await tester.tap(find.text('Study now'));
        await tester.pumpAndSettle();
        expect(find.byType(StudyScreen), findsOneWidget);
        return go;
      }

      Finder studyAppBar() => find.descendant(
        of: find.byType(StudyScreen),
        matching: find.byType(AppBar),
      );

      testWidgets('distraction-free on a phone (spec)', (tester) async {
        await studyInApp(tester, hide: true);
        expect(find.byType(NavigationBar), findsNothing);
        expect(studyAppBar(), findsNothing);
        // The card, counts and buttons stay.
        expect(richText('hola'), findsOneWidget);
        expect(find.byKey(const Key('remaining-new')), findsOneWidget);
        expect(find.text('Show answer'), findsOneWidget);
        // The top edge leaves; the bars are back.
        await tester.tap(find.byKey(const Key('leave-session')));
        await tester.pumpAndSettle();
        expect(find.byType(StudyScreen), findsNothing);
        expect(find.byType(NavigationBar), findsOneWidget);
      });

      testWidgets('Escape leaves and the rail returns', (tester) async {
        await studyInApp(tester, hide: true, size: desktop);
        expect(find.byType(NavigationRail), findsNothing);
        await tester.sendKeyEvent(LogicalKeyboardKey.escape);
        await tester.pumpAndSettle();
        expect(find.byType(StudyScreen), findsNothing);
        expect(find.byType(NavigationRail), findsOneWidget);
      });

      testWidgets('the bars stay by default', (tester) async {
        await studyInApp(tester, hide: false);
        expect(find.byType(NavigationBar), findsOneWidget);
        expect(studyAppBar(), findsOneWidget);
        expect(find.byKey(const Key('leave-session')), findsNothing);
      });
    });

    group('question fade', () {
      final fade = find.byKey(const Key('question-fade'));

      /// Answers the first card Good and shows the next one for a frame.
      Future<void> next(WidgetTester tester) async {
        await tester.tap(find.text('Show answer'));
        await tester.pumpAndSettle();
        await tester.tap(find.byKey(const Key('rate-3')));
        await tester.pump();
        await tester.pump();
        expect(richText('perro'), findsOneWidget);
      }

      testWidgets('no fade (spec)', (tester) async {
        final go = FakeGo();
        go.settings = {...go.settings, 'questionFadeMs': 0};
        FakeStudy(go, [
          studyCard(1, 'hola', 'hello'),
          studyCard(2, 'perro', 'dog'),
        ]);
        await pumpStudy(tester, loadSettings: true);
        await next(tester);
        expect(fade, findsNothing);
      });

      testWidgets('by default the question fades in over 100 ms', (
        tester,
      ) async {
        final go = FakeGo();
        FakeStudy(go, [
          studyCard(1, 'hola', 'hello'),
          studyCard(2, 'perro', 'dog'),
        ]);
        await pumpStudy(tester, loadSettings: true);
        await next(tester);
        expect(tester.widget<Opacity>(fade).opacity, lessThan(1));
        await tester.pump(const Duration(milliseconds: 100));
        expect(tester.widget<Opacity>(fade).opacity, 1);
        // The answer shows at once.
        await tester.tap(find.text('Show answer'));
        await tester.pump();
        expect(fade, findsNothing);
        expect(richText('dog'), findsOneWidget);
      });
    });

    group('scroll to answer', () {
      final anchor = find.byKey(const Key('answer-anchor'));
      double pixels(WidgetTester tester) => tester
          .state<ScrollableState>(
            find.descendant(
              of: find.byType(SingleChildScrollView),
              matching: find.byType(Scrollable),
            ),
          )
          .position
          .pixels;

      ScrollPosition position(WidgetTester tester) => tester
          .state<ScrollableState>(
            find.descendant(
              of: find.byType(SingleChildScrollView),
              matching: find.byType(Scrollable),
            ),
          )
          .position;

      String lines(String word) => List.filled(60, word).join('<br>');

      testWidgets('long front side (spec)', (tester) async {
        final go = FakeGo();
        FakeStudy(go, [
          studyCard(1, lines('front'), 'hello<br>${lines('back')}'),
          studyCard(2, 'perro', 'dog'),
        ]);
        await pumpStudy(tester, loadSettings: true);
        await tester.tap(find.text('Show answer'));
        await tester.pumpAndSettle();
        // The divider is at the top of the view, the Back field below it.
        expect(pixels(tester), greaterThan(0));
        final view = tester.getRect(find.byType(SingleChildScrollView));
        expect(tester.getTopLeft(anchor).dy, moreOrLessEquals(view.top));
        final back = tester.getRect(richText('hello'));
        expect(view.contains(back.topLeft + const Offset(1, 1)), isTrue);
        // The next card starts at the top.
        await tester.tap(find.byKey(const Key('rate-3')));
        await tester.pumpAndSettle();
        expect(pixels(tester), 0);
      });

      testWidgets('a short back side scrolls as far as the card goes', (
        tester,
      ) async {
        final go = FakeGo();
        FakeStudy(go, [studyCard(1, lines('front'), 'hello')]);
        await pumpStudy(tester, loadSettings: true);
        await tester.tap(find.text('Show answer'));
        await tester.pumpAndSettle();
        final p = position(tester);
        expect(p.maxScrollExtent, greaterThan(0));
        expect(p.pixels, p.maxScrollExtent);
        final view = tester.getRect(find.byType(SingleChildScrollView));
        expect(view.contains(tester.getRect(richText('hello')).center), isTrue);
      });

      testWidgets('a card that fits does not scroll', (tester) async {
        final go = FakeGo();
        FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
        await pumpStudy(tester, loadSettings: true);
        await tester.tap(find.text('Show answer'));
        await tester.pumpAndSettle();
        expect(anchor, findsOneWidget);
        expect(pixels(tester), 0);
      });

      testWidgets('the question has no anchor', (tester) async {
        final go = FakeGo();
        FakeStudy(go, [
          {
            ...studyCard(1, 'hola', 'hello'),
            'question': '<div>hola<hr id="answer"></div>',
          },
        ]);
        await pumpStudy(tester, loadSettings: true);
        expect(anchor, findsNothing);
      });
    });
  });

  group('shortcut hints', () {
    Future<void> openMenu(
      WidgetTester tester, [
      String tooltip = 'Card actions',
    ]) async {
      await tester.tap(find.byTooltip(tooltip));
      await tester.pumpAndSettle();
    }

    Future<void> openFlagMenu(WidgetTester tester) async {
      await openMenu(tester);
      await tester.tap(find.text('Flag'));
      await tester.pumpAndSettle();
    }

    testWidgets('study menu on a phone (spec)', (tester) async {
      final go = FakeGo();
      FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
      await pumpStudy(tester);
      await openMenu(tester);
      for (final label in ['Edit card', 'Card info', 'Set due date']) {
        expect(find.text(label), findsOneWidget);
      }
      for (final hint in ['(E)', '(I)', 'Ctrl+Shift+D']) {
        expect(find.textContaining(hint), findsNothing);
      }
    }, variant: TargetPlatformVariant.only(TargetPlatform.android));

    testWidgets('study menu on desktop (spec)', (tester) async {
      final go = FakeGo();
      FakeStudy(go, [
        studyCard(1, 'hola', 'hello'),
        studyCard(2, 'perro', 'dog'),
      ]);
      await pumpStudy(tester);
      await tester.tap(find.text('Show answer'));
      await tester.pumpAndSettle();
      await pressKeys(tester, [LogicalKeyboardKey.digit1]);
      await openMenu(tester);
      for (final label in [
        'Edit card (E)',
        'Card info (I)',
        'Set due date (Ctrl+Shift+D)',
        'Previous card info (Ctrl+Alt+I)',
      ]) {
        expect(find.text(label), findsOneWidget);
      }
      await tester.tap(find.text('Flag'));
      await tester.pumpAndSettle();
      expect(find.text('Red'), findsOneWidget);
      for (var n = 1; n <= 7; n++) {
        expect(find.text('Ctrl+$n'), findsOneWidget);
      }
    }, variant: TargetPlatformVariant.only(TargetPlatform.linux));

    testWidgets('flag menu on a phone (spec)', (tester) async {
      final go = FakeGo();
      FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
      await pumpStudy(tester);
      await openFlagMenu(tester);
      expect(find.text('Red'), findsOneWidget);
      expect(
        find.descendant(
          of: find.byKey(const Key('flag-item-1')),
          matching: find.byIcon(Icons.flag),
        ),
        findsOneWidget,
      );
      for (var n = 1; n <= 7; n++) {
        expect(find.text('Ctrl+$n'), findsNothing);
      }
    }, variant: TargetPlatformVariant.only(TargetPlatform.android));

    testWidgets('hardware keyboard on a tablet (spec)', (tester) async {
      final go = FakeGo();
      FakeStudy(go, [studyCard(1, 'hola', 'hello', noteId: 7)]);
      go.handle(
        'notes.Get',
        (args) => {
          'id': 7,
          'typeId': 10,
          'fields': ['hola', 'hello'],
          'tags': <String>[],
          'cardIds': [1],
        },
      );
      await pumpStudy(tester, size: const Size(800, 1280));
      await openMenu(tester);
      expect(find.text('Edit card'), findsOneWidget);
      expect(find.text('Edit card (E)'), findsNothing);
      await tester.tapAt(Offset.zero);
      await tester.pumpAndSettle();

      await tester.sendKeyEvent(LogicalKeyboardKey.keyE);
      await tester.pumpAndSettle();
      expect(go.callsTo('notes.Get'), [
        [7],
      ]);
      expect(find.text('Edit card'), findsOneWidget); // the editor
    }, variant: TargetPlatformVariant.only(TargetPlatform.android));

    testWidgets('German key names (spec)', (tester) async {
      final go = FakeGo();
      FakeStudy(go, [studyCard(1, 'hola', 'hello')]);
      appL10n = lookupAppLocalizations(const Locale('de'));
      addTearDown(() => appL10n = english);
      await pumpStudy(tester, locale: const Locale('de'));
      await openMenu(tester, 'Kartenaktionen');
      expect(find.text('Eigene Stimme aufnehmen (Umschalt+V)'), findsOneWidget);
    }, variant: TargetPlatformVariant.only(TargetPlatform.windows));
  });
}

/// A Go time.Duration as JSON: nanoseconds.
int _ns(Duration d) => d.inMicroseconds * 1000;
