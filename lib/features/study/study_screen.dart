import 'dart:async';

import 'package:clock/clock.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/app/errors.dart';
import 'package:cardflip/app/shortcuts.dart';
import 'package:cardflip/app/theme.dart';
import 'package:cardflip/data/changes.dart';
import 'package:cardflip/data/progress_model.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/features/browse/card_info_screen.dart';
import 'package:cardflip/features/browse/delete_cards.dart';
import 'package:cardflip/features/browse/reset_dialog.dart';
import 'package:cardflip/features/browse/set_due_date_dialog.dart';
import 'package:cardflip/features/decks/filtered_deck_screen.dart';
import 'package:cardflip/features/editor/math_insert.dart' show MenuEntryLabel;
import 'package:cardflip/features/editor/note_editor_screen.dart';
import 'package:cardflip/features/progress/achievements.dart';
import 'package:cardflip/features/settings/daily_goal_dialog.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/card_view.dart';
import 'package:cardflip/widgets/media_player.dart';
import 'package:cardflip/widgets/time_format.dart';
import 'package:cardflip/widgets/voice_recorder.dart';
import 'package:cardflip/l10n/l10n.dart';

class _RevealIntent extends Intent {
  const _RevealIntent();
}

class _RateIntent extends Intent {
  const _RateIntent(this.rating);
  final int rating;
}

class _UndoIntent extends Intent {
  const _UndoIntent();
}

class _LeaveIntent extends Intent {
  const _LeaveIntent();
}

class _ReplayIntent extends Intent {
  const _ReplayIntent();
}

class _PauseIntent extends Intent {
  const _PauseIntent();
}

class _SeekIntent extends Intent {
  const _SeekIntent(this.seconds);
  final int seconds;
}

class _RecordIntent extends Intent {
  const _RecordIntent();
}

class _PlayVoiceIntent extends Intent {
  const _PlayVoiceIntent();
}

class _FlagIntent extends Intent {
  const _FlagIntent(this.flag);
  final int flag;
}

class _MarkIntent extends Intent {
  const _MarkIntent();
}

class _CardInfoIntent extends Intent {
  const _CardInfoIntent();
}

class _PreviousCardInfoIntent extends Intent {
  const _PreviousCardInfoIntent();
}

class _SetDueDateIntent extends Intent {
  const _SetDueDateIntent();
}

class _AutoAdvanceIntent extends Intent {
  const _AutoAdvanceIntent();
}

/// An action of the card menu, by its menu value (e.g. 'buryNote').
class _CardActionIntent extends Intent {
  const _CardActionIntent(this.action);
  final String action;
}

/// The keys of card menu actions. The characters are typed with whatever
/// keys the keyboard layout puts them on.
final _cardActionShortcuts = <ShortcutActivator, Intent>{
  const SingleActivator(LogicalKeyboardKey.keyE): const _CardActionIntent(
    'edit',
  ),
  const CharacterActivator('-'): const _CardActionIntent('bury'),
  const CharacterActivator('@'): const _CardActionIntent('suspend'),
  const CharacterActivator('='): const _CardActionIntent('buryNote'),
  const CharacterActivator('!'): const _CardActionIntent('suspendNote'),
  for (final s in deleteNoteShortcuts) s: const _CardActionIntent('deleteNote'),
  const SingleActivator(LogicalKeyboardKey.keyO): const _CardActionIntent(
    'options',
  ),
};

const _digits = [
  LogicalKeyboardKey.digit1,
  LogicalKeyboardKey.digit2,
  LogicalKeyboardKey.digit3,
  LogicalKeyboardKey.digit4,
  LogicalKeyboardKey.digit5,
  LogicalKeyboardKey.digit6,
  LogicalKeyboardKey.digit7,
];

/// Ctrl+1…7 and Cmd+1…7 toggle flags 1 to 7.
final _flagShortcuts = <ShortcutActivator, Intent>{
  for (var i = 0; i < _digits.length; i++) ...{
    SingleActivator(_digits[i], control: true): _FlagIntent(i + 1),
    SingleActivator(_digits[i], meta: true): _FlagIntent(i + 1),
  },
};

/// A study shortcut that steps aside while the answer input has focus, so
/// the key types into it instead.
class _StudyAction<T extends Intent> extends CallbackAction<T> {
  _StudyAction(this.typing, {required super.onInvoke});

  final bool Function() typing;

  @override
  bool isEnabled(T intent) => !typing();
}

/// The names of ratings 1 to 4.
List<String> ratingNames(AppLocalizations l) => [
  l.again,
  l.hard,
  l.good,
  l.easy,
];

/// Studies the due cards of a deck and its subdecks, one card at a time.
class StudyScreen extends StatefulWidget {
  const StudyScreen({super.key, required this.deckId, required this.deckName});

  final int deckId;
  final String deckName;

  @override
  State<StudyScreen> createState() => _StudyScreenState();
}

class _StudyScreenState extends State<StudyScreen> {
  golang.StudyState? _state;
  Object? _error;
  bool _revealed = false;
  bool _busy = false;

  /// The card on screen is the card just answered, shown again: its
  /// question says so until the answer is shown (as it says when a card is
  /// shown early).
  bool _shownAgain = false;

  /// The time spent on the card on screen, which an answer records; paused
  /// while the session is left for another screen.
  final _timer = clock.stopwatch();

  /// The on-screen timer's reading, frozen when the answer was shown with
  /// "Stop timer on answer"; null while it runs.
  Duration? _timerFrozen;

  /// The current timebox: the time since it started, and the session's
  /// studied count then.
  final _timebox = clock.stopwatch()..start();
  int _timeboxStudied = 0;

  late final ValueNotifier<int> _openSessions;

  /// The study screens hiding the app's navigation, and whether this one is
  /// among them: while "Hide bars during review" is on and it is on show.
  late final ValueNotifier<int> _barsHidden;
  bool _hidingBars = false;

  /// Fires when the next learning card becomes due, or while waiting comes
  /// within the learn ahead limit (state.checkAt).
  Timer? _learningTimer;

  /// A state with a newly due learning card, fetched while waiting and
  /// shown when the user chooses Continue.
  golang.StudyState? _pending;

  /// Plays the current card's sounds.
  final _player = MediaPlayer();
  CardSounds _sounds = const CardSounds();

  /// Records the user's voice; the recording belongs to the card on screen
  /// and is discarded when it changes.
  VoiceRecorder? _recorder;
  bool _recording = false;
  String? _voice;

  /// The input of a question with `{{type:…}}`, and the answer comparing
  /// what was typed, from study.reveal.
  final _typed = TextEditingController();
  final _typedFocus = FocusNode(debugLabel: 'type answer');
  String? _comparedAnswer;

  /// Whether the revealed answer was scrolled to its `id=answer` element.
  bool _scrolledToAnswer = false;

  /// Auto advance, turned on for this session with Shift+A or the menu.
  /// While on, [_autoTimer] counts down the side shown (from the card's
  /// deck options) and runs its action; it starts again on user input.
  bool _autoAdvance = false;
  Timer? _autoTimer;

  /// Set while the countdown waits for the side's sounds to finish.
  bool _autoWaitsForSounds = false;

  /// The typed answer as last seen, so typing restarts the countdown.
  String _typedText = '';

  /// Whether the screen is on show: its destination is selected and no
  /// route (dialog, menu, editor) covers it. Set by build.
  bool _onScreen = true;

  /// The screen's own focus, which the shortcuts listen below.
  final _screenFocus = FocusNode(debugLabel: 'study');

  /// Collection changes; the card is rendered again when the render
  /// environment (theme, card size) changes, and the session follows the
  /// changes device sync receives.
  StreamSubscription<golang.Change>? _changes;

  /// Set when the environment changed, or sync changed the collection
  /// ([_syncPending]), while a call was running.
  bool _renderPending = false;
  bool _syncPending = false;

  /// Whether today was a study day when the session started, so the
  /// summary can say the streak was extended; null until known.
  bool? _studiedBefore;

  @override
  void initState() {
    super.initState();
    final progress = AppScope.read(context).progress.progress;
    if (progress != null) {
      _studiedBefore = progress.todayStudied;
    } else if (AppScope.read(context).settings.showProgress) {
      golang.progress.today().then<void>(
        (p) => _studiedBefore ??= p.todayStudied,
        onError: (Object _) {},
      );
    }
    _openSessions = AppScope.read(context).openSessions;
    _openSessions.value++;
    _barsHidden = AppScope.read(context).barsHidden;
    // A leftover message (e.g. "Card added") must not cover the buttons.
    WidgetsBinding.instance.addPostFrameCallback(
      (_) => messengerKey.currentState?.hideCurrentSnackBar(),
    );
    _changes = AppScope.read(context).changes.stream.listen((c) {
      if (c.kind == ChangeHub.environment) _renderAgain();
      if (c.kind == ChangeHub.synced) _renderAgain(synced: true);
    });
    _player.addListener(_playerChanged);
    _typed.addListener(_typedChanged);
    _run(() => golang.study.start(widget.deckId));
  }

  @override
  void dispose() {
    _changes?.cancel();
    _autoTimer?.cancel();
    unawaited(_dropVoice());
    _recorder?.dispose();
    _player.dispose();
    _typed.dispose();
    _typedFocus.dispose();
    _screenFocus.dispose();
    _learningTimer?.cancel();
    _openSessions.value--;
    if (_hidingBars) _barsHidden.value--;
    super.dispose();
  }

  /// Runs a study call and shows its result; a new card starts unrevealed.
  Future<void> _run(
    Future<golang.StudyState> Function() call, {
    bool keepRevealed = false,
    bool answer = false,
  }) async {
    if (_busy) return;
    setState(() => _busy = true);
    try {
      final state = await call();
      if (!mounted) return;
      _apply(state, keepRevealed: keepRevealed, answer: answer);
    } catch (e) {
      if (!mounted) return;
      if (_state == null) {
        setState(() => _error = e);
      } else {
        showError(e);
      }
    } finally {
      if (mounted) _idle();
    }
  }

  /// Ends a call; a change it held back is handled now.
  void _idle() {
    setState(() => _busy = false);
    if (_renderPending) {
      final synced = _syncPending;
      _renderPending = _syncPending = false;
      _renderAgain(synced: synced);
    }
  }

  /// The render environment changed, or device sync changed the collection
  /// ([synced]): fetches the session's state again, keeping the card on
  /// screen and its side while the queue still has it, so that the counts
  /// and the queue follow synced changes and an answer being given is kept
  /// (device-sync spec: Sync keeps working while studying). The old HTML
  /// stays until the new one arrives. Without a card on screen only synced
  /// changes are fetched, as they can bring cards to study.
  Future<void> _renderAgain({bool synced = false}) async {
    final card = _state?.card;
    if (card == null && !(synced && _state != null && _pending == null)) {
      return;
    }
    if (_busy) {
      _renderPending = true;
      _syncPending |= synced;
      return;
    }
    final compared = _comparedAnswer != null;
    await _run(golang.study.current, keepRevealed: true);
    if (!compared || !mounted || _busy || !_revealed) return;
    if (card == null || _state?.card?.cardID != card.cardID) return;
    final answer = await _compare(card);
    if (mounted && answer != null && _state?.card?.cardID == card.cardID) {
      setState(() => _comparedAnswer = answer);
    }
  }

  /// Shows [state] and sets the timer for its next learning card. An answer
  /// that made its card a leech says so, and one whose card comes straight
  /// back says that it is shown again.
  void _apply(
    golang.StudyState state, {
    bool keepRevealed = false,
    bool answer = false,
  }) {
    final sameCard = state.card?.cardID == _state?.card?.cardID;
    final newCard = !(keepRevealed && sameCard);
    setState(() {
      _state = state;
      _pending = null;
      _error = null;
      // The session's cards are done: auto advance ends with it.
      if (state.card == null) _autoAdvance = false;
      if (newCard) {
        _revealed = false;
        _shownAgain = answer && sameCard && state.card != null;
        _comparedAnswer = null;
        _typed.clear();
        _timerFrozen = null;
        _timer
          ..reset()
          ..start();
      }
    });
    if (newCard) {
      _showQuestion(state.card);
      _restartAutoAdvance();
    }
    if (state.leech) {
      showInfo(
        state.leechSuspended ? context.l10n.leechSuspended : context.l10n.leech,
      );
    }
    _learningTimer?.cancel();
    final due = state.checkAt;
    if (due == null) return;
    // Go counts a card as due (or within the limit) at Due <= now (plus the
    // limit); the slack absorbs rounding.
    var wait = due.add(const Duration(seconds: 1)).difference(DateTime.now());
    if (wait < const Duration(seconds: 1)) wait = const Duration(seconds: 1);
    _learningTimer = Timer(wait, _learningBecameDue);
  }

  /// A learning card became due: refresh the counts of the card on screen,
  /// or, while waiting, a learning card came within the learn ahead limit:
  /// show the card or offer to continue.
  Future<void> _learningBecameDue() async {
    if (!mounted) return;
    if (_busy) {
      _learningTimer = Timer(const Duration(seconds: 1), _learningBecameDue);
      return;
    }
    if (_state?.card != null) {
      return _run(golang.study.current, keepRevealed: true);
    }
    if (AppScope.read(context).settings.autoContinueLearning) {
      return _run(golang.study.current);
    }
    setState(() => _busy = true);
    try {
      final state = await golang.study.current();
      if (!mounted) return;
      if (state.card != null) {
        setState(() => _pending = state);
      } else {
        _apply(state);
      }
    } catch (e) {
      if (mounted) showError(e);
    } finally {
      if (mounted) _idle();
    }
  }

  /// Shows the learning card that became due while waiting.
  void _continue() {
    final pending = _pending;
    if (pending != null) _apply(pending);
  }

  /// A card is shown: the previous card's sounds stop, the question's play,
  /// and its answer input, if any, gets the focus.
  void _showQuestion(golang.StudyCard? card) {
    if (card?.typeAnswer == true) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted && !_revealed) _typedFocus.requestFocus();
      });
    } else if (_typedFocus.hasFocus) {
      _screenFocus.requestFocus();
    }
    _player.leave();
    unawaited(_dropVoice());
    _sounds = card == null
        ? const CardSounds()
        : CardSounds(
            question: MediaItem.inOrder(
              CardSide.question,
              card.question,
              card.questionMedia,
              card.questionTTS,
            ),
            answer: MediaItem.inOrder(
              CardSide.answer,
              card.answer,
              card.answerMedia,
              card.answerTTS,
            ),
            autoplay: card.autoplay,
            replayQuestion: card.replayQuestion,
          );
    _player.showQuestion(_sounds);
  }

  /// Shows the answer; with an answer input, compared with what was typed.
  Future<void> _reveal() async {
    final card = _state?.card;
    if (card == null || _revealed || _busy) return;
    if (card.typeAnswer) {
      setState(() => _busy = true);
      final compared = await _compare(card);
      if (!mounted) return;
      if (_state?.card?.cardID != card.cardID || _revealed) {
        _idle();
        return;
      }
      _comparedAnswer = compared;
      // Back to the screen, so the rating keys work.
      _screenFocus.requestFocus();
    }
    setState(() {
      _revealed = true;
      _shownAgain = false;
      _scrolledToAnswer = false;
      if (card.showTimer && card.stopTimerOnAnswer) {
        _timerFrozen = _timer.elapsed;
      }
    });
    _player.showAnswer(
      _sounds,
      interrupt: AppScope.read(context).settings.interruptAudio,
    );
    _restartAutoAdvance();
    // Last, so a held-back render compares the answer again.
    if (_busy) _idle();
  }

  /// The answer of [card] comparing the typed text, or null (the expected
  /// answer) when Go could not render it.
  Future<String?> _compare(golang.StudyCard card) async {
    try {
      return (await golang.study.reveal(card.cardID, _typed.text)).answer;
    } catch (e) {
      if (mounted) showError(e);
      return null;
    }
  }

  /// Enter in the answer input reveals, unless it confirms an input method
  /// composition (e.g. Japanese).
  void _submitTyped() {
    final composing = _typed.value.composing;
    if (composing.isValid && !composing.isCollapsed) return;
    _reveal();
  }

  void _replay() {
    if (_state?.card == null) return;
    _player.replay(_sounds, answerShown: _revealed);
    _restartAutoAdvance();
  }

  /// Shift+A: turns auto advance on or off for this session.
  void _toggleAutoAdvance() {
    final card = _state?.card;
    if (card == null) return;
    setState(() => _autoAdvance = !_autoAdvance);
    _restartAutoAdvance();
    if (_autoAdvance &&
        card.secondsToShowQuestion == 0 &&
        card.secondsToShowAnswer == 0) {
      showInfo(context.l10n.autoAdvanceOff);
    }
  }

  /// Starts the auto advance countdown for the side shown, once its sounds
  /// have finished. Does nothing while auto advance is off or the side's
  /// time is 0.
  void _restartAutoAdvance() {
    _autoTimer?.cancel();
    _autoTimer = null;
    _autoWaitsForSounds = false;
    final card = _state?.card;
    if (!_autoAdvance || card == null) return;
    final seconds = _revealed
        ? card.secondsToShowAnswer
        : card.secondsToShowQuestion;
    if (seconds <= 0) return;
    if (_player.current != null) {
      _autoWaitsForSounds = true;
      return;
    }
    _autoTimer = Timer(Duration(seconds: seconds), _autoAdvanceDue);
  }

  void _playerChanged() {
    if (_autoWaitsForSounds && _player.current == null) _restartAutoAdvance();
  }

  void _typedChanged() {
    if (_typed.text == _typedText) return;
    _typedText = _typed.text;
    if (!_revealed) _restartAutoAdvance();
  }

  /// The countdown ended: runs the side's action. While a call runs or the
  /// screen is covered (a dialog, a menu, the editor), it waits and counts
  /// again once the screen is back.
  void _autoAdvanceDue() {
    _autoTimer = null;
    final card = _state?.card;
    if (!mounted || !_autoAdvance || card == null) return;
    if (_busy || !_onScreen) {
      _autoTimer = Timer(const Duration(milliseconds: 500), () {
        _autoTimer = null;
        if (!mounted) return;
        _busy || !_onScreen ? _autoAdvanceDue() : _restartAutoAdvance();
      });
      return;
    }
    if (!_revealed) {
      if (card.questionAction == 'showReminder') {
        showInfo(context.l10n.timeToShowAnswer);
      } else {
        _reveal();
      }
      return;
    }
    switch (card.answerAction) {
      case 'answerAgain':
        _rate(1);
      case 'answerHard':
        _rate(2);
      case 'answerGood':
        _rate(3);
      case 'showReminder':
        showInfo(context.l10n.timeToAnswer);
      default:
        _cardAction('bury');
    }
  }

  /// Shift+V: starts recording, or stops it.
  Future<void> _toggleRecording() async {
    final recorder = _recorder ??= VoiceRecorder.create();
    if (_recording) {
      final uri = await recorder.stop();
      if (!mounted || !_recording) {
        // The card changed meanwhile: the recording belonged to it.
        if (uri != null) unawaited(recorder.discard(uri));
        return;
      }
      final old = _voice;
      setState(() {
        _recording = false;
        _voice = uri;
      });
      if (old != null) unawaited(recorder.discard(old));
      return;
    }
    if (_state?.card == null) return;
    _player.stop();
    try {
      await recorder.start();
      if (mounted) setState(() => _recording = true);
    } catch (e) {
      if (!mounted) return;
      // A dialog, not a message that would cover the answer buttons.
      await showDialog<void>(
        context: context,
        animationStyle: popupAnimation(context),
        builder: (context) => AlertDialog(
          title: Text(context.l10n.recordingUnavailable),
          content: Text(
            e is MicrophoneRefused ? '$e' : context.l10n.recordingFailed('$e'),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: Text(context.l10n.ok),
            ),
          ],
        ),
      );
    }
  }

  /// V: plays the recording of this card.
  void _playVoice() {
    final voice = _voice;
    if (voice != null) _player.play([MediaItem.recording(voice)]);
    _restartAutoAdvance();
  }

  /// Discards the recording; it is never stored.
  Future<void> _dropVoice() async {
    final recorder = _recorder;
    if (recorder == null) return;
    final voice = _voice;
    final recording = _recording;
    _voice = null;
    _recording = false;
    if (voice != null) await recorder.discard(voice);
    if (recording) {
      final unfinished = await recorder.stop();
      if (unfinished != null) await recorder.discard(unfinished);
    }
  }

  Future<void> _rate(int rating) async {
    final card = _state?.card;
    if (card == null || !_revealed) return;
    _player.stop();
    final ms = _timer.elapsedMilliseconds;
    final studied = _state!.studied;
    await _run(
      () => golang.study.answer(card.cardID, rating, ms),
      answer: true,
    );
    if (mounted && (_state?.studied ?? 0) > studied) await _checkTimebox();
  }

  /// After an answer: once the timebox time limit has passed, says how many
  /// cards were studied in it. Continue starts the next timebox; Finish ends
  /// the session with its summary. A finished session shows its summary
  /// anyway.
  Future<void> _checkTimebox() async {
    final minutes = AppScope.read(context).settings.timeboxMinutes;
    final state = _state;
    if (minutes <= 0 || state?.card == null) return;
    if (_timebox.elapsed < Duration(minutes: minutes)) return;
    final studied = state!.studied - _timeboxStudied;
    // The next card's time starts once the message is closed.
    final timing = _timer.isRunning;
    _timer.stop();
    final finish = await showDialog<bool>(
      context: context,
      animationStyle: popupAnimation(context),
      barrierDismissible: false,
      builder: (context) => AlertDialog(
        key: const Key('timeboxMessage'),
        title: Text(context.l10n.timeboxReached),
        content: Text(
          context.l10n.timeboxStudied(studied < 0 ? 0 : studied, minutes),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text(context.l10n.finish),
          ),
          FilledButton(
            autofocus: true,
            onPressed: () => Navigator.pop(context, false),
            child: Text(context.l10n.continueAction),
          ),
        ],
      ),
    );
    if (!mounted) return;
    if (finish == true) {
      _finish();
      return;
    }
    _timebox
      ..reset()
      ..start();
    _timeboxStudied = _state?.studied ?? 0;
    if (timing) _timer.start();
  }

  /// Ends the session before its cards are done: shows its summary, with
  /// nothing left to study or undo.
  void _finish() {
    final s = _state;
    if (s == null) return;
    _learningTimer?.cancel();
    _showQuestion(null);
    setState(() {
      _revealed = false;
      _pending = null;
      _autoAdvance = false;
      _state = golang.StudyState(
        deckID: s.deckID,
        deckName: s.deckName,
        counts: s.counts,
        canUndo: false,
        studied: s.studied,
        timeSpent: s.timeSpent,
        leech: false,
        leechSuspended: false,
      );
    });
    _restartAutoAdvance();
  }

  void _undo() {
    if (_state?.canUndo != true) return;
    _run(golang.study.undo);
  }

  Future<void> _edit() async {
    final card = _state?.card;
    if (card == null) return;
    _timer.stop();
    _player.stop();
    await Navigator.of(context).push(
      MaterialPageRoute<void>(
        builder: (_) => NoteEditorScreen(noteId: card.noteID),
      ),
    );
    _timer.start();
    // Show the corrected content of the same card.
    await _run(golang.study.current, keepRevealed: true);
    if (mounted) _restartAutoAdvance();
    // A comparison is rendered again against the edited note.
    final shown = _state?.card;
    if (!mounted || !_revealed || _comparedAnswer == null) return;
    if (shown == null || shown.cardID != card.cardID) return;
    final compared = await _compare(shown);
    if (mounted && _state?.card?.cardID == card.cardID) {
      setState(() => _comparedAnswer = compared);
    }
  }

  /// Opens the add screen with a copy of the note of the card on screen.
  /// The session shows the same card when it closes.
  Future<void> _createCopy() async {
    final card = _state?.card;
    if (card == null || _busy) return;
    final golang.CardDetails details;
    try {
      details = await golang.browse.card(card.cardID);
    } catch (e) {
      showError(e);
      return;
    }
    if (!mounted) return;
    await _away(
      () => Navigator.of(context).push(
        MaterialPageRoute<void>(
          builder: (_) => NoteEditorScreen(
            initial: NoteDraft(
              typeId: details.note.typeID,
              deckId: details.deckID,
              fields: details.note.fields,
              tags: details.note.tags,
            ),
          ),
        ),
      ),
    );
  }

  /// Opens the options of the deck being studied. Changed options apply to
  /// the rest of the session, e.g. a lowered new card limit.
  Future<void> _options() async {
    if (_busy) return;
    final golang.DeckNode deck;
    try {
      final tree = await golang.decks.tree();
      final found = tree.where((d) => d.id == widget.deckId).firstOrNull;
      if (found == null) throw appL10n.goDeckGone;
      deck = found;
    } catch (e) {
      showError(e);
      return;
    }
    if (!mounted) return;
    await _away(() => openDeckOptions(context, deck));
  }

  /// Leaves the session for [open] (another screen), pausing the timer and
  /// sounds, and shows the session's state when it is back.
  Future<void> _away(Future<void> Function() open) async {
    final timing = _timer.isRunning;
    _timer.stop();
    _player.stop();
    await open();
    if (!mounted) return;
    if (timing) _timer.start();
    await _run(golang.study.current, keepRevealed: true);
    if (mounted) _restartAutoAdvance();
  }

  /// Asks before deleting the note of the card on screen with all its
  /// cards, then deletes it and moves on.
  Future<void> _deleteNote() async {
    final card = _state?.card;
    if (card == null || _busy) return;
    final timing = _timer.isRunning;
    _timer.stop();
    final impact = await confirmDeleteCards(context, [card.cardID]);
    if (!mounted) return;
    if (impact == null) {
      if (timing) _timer.start();
      return;
    }
    await _run(() => golang.study.deleteNote(card.cardID));
  }

  /// Runs the card menu action [action] on the card on screen.
  void _cardAction(String action) {
    final card = _state?.card;
    if (action == 'options') {
      _options();
      return;
    }
    if (card == null) return;
    switch (action) {
      case 'edit':
        _edit();
      case 'copy':
        _createCopy();
      case 'bury':
        _run(() => golang.study.bury(card.cardID));
      case 'suspend':
        _run(() => golang.study.suspend(card.cardID));
      case 'buryNote':
        _run(() => golang.study.buryNote(card.cardID));
      case 'suspendNote':
        _run(() => golang.study.suspendNote(card.cardID));
      case 'deleteNote':
        _deleteNote();
    }
  }

  /// Opens card info for [cardId]. The session stays as it is: the same
  /// card and side are shown when it closes.
  Future<void> _cardInfo(int cardId) async {
    final timing = _timer.isRunning;
    _timer.stop();
    _player.stop();
    await showCardInfo(context, cardId);
    if (mounted && timing) _timer.start();
  }

  /// Opens card info for the previously answered card.
  Future<void> _previousCardInfo() async {
    try {
      final id = await golang.study.previousCardID();
      if (!mounted) return;
      if (id == 0) {
        showInfo(context.l10n.noCardAnsweredYet);
        return;
      }
      await _cardInfo(id);
    } catch (e) {
      if (mounted) showError(e);
    }
  }

  bool _typing() => _typedFocus.hasFocus;

  /// The card actions menu button, where the flag menu opens.
  final _menuKey = GlobalKey();

  /// Sets [flag] on the card on screen, or clears it when the card has it;
  /// 0 clears it. The side shown stays.
  Future<void> _flag(int flag) async {
    final card = _state?.card;
    if (card == null) return;
    await _run(
      () => golang.study.setFlag(card.cardID, flag),
      keepRevealed: true,
    );
    if (mounted) _restartAutoAdvance();
  }

  /// Asks for the due date of the card on screen and sets it. The session
  /// moves on, unless the card is still due today.
  Future<void> _setDueDate() async {
    final card = _state?.card;
    if (card == null || _busy) return;
    final timing = _timer.isRunning;
    _timer.stop();
    final state = await askSetDueDate(
      context,
      cards: 1,
      apply: (spec) => golang.study.setDueDate(card.cardID, spec),
    );
    if (!mounted) return;
    if (state == null) {
      if (timing) _timer.start();
      return;
    }
    _player.stop();
    _apply(state);
  }

  /// Asks how to reset the card on screen to new, resets it and moves on.
  Future<void> _reset() async {
    final card = _state?.card;
    if (card == null || _busy) return;
    final timing = _timer.isRunning;
    _timer.stop();
    final opts = await askReset(context, cards: 1);
    if (!mounted) return;
    if (opts == null) {
      if (timing) _timer.start();
      return;
    }
    await _run(() => golang.study.reset(card.cardID, opts));
  }

  /// Toggles the mark on the note of the card on screen. The side shown
  /// stays.
  Future<void> _toggleMark() async {
    final card = _state?.card;
    if (card == null) return;
    await _run(() => golang.study.toggleMark(card.cardID), keepRevealed: true);
    if (mounted) _restartAutoAdvance();
  }

  /// The Flag submenu of the card actions, opened where that menu was.
  Future<void> _flagMenu(golang.StudyCard card) async {
    final box = _menuKey.currentContext?.findRenderObject() as RenderBox?;
    final overlay =
        Overlay.of(context).context.findRenderObject() as RenderBox?;
    if (box == null || overlay == null) return;
    final position = RelativeRect.fromRect(
      Rect.fromPoints(
        box.localToGlobal(Offset.zero, ancestor: overlay),
        box.localToGlobal(box.size.bottomRight(Offset.zero), ancestor: overlay),
      ),
      Offset.zero & overlay.size,
    );
    final settings = AppScope.read(context).settings;
    final colors = FlagColors.of(context);
    final flag = await showMenu<int>(
      context: context,
      popUpAnimationStyle: popupAnimation(context),
      position: position,
      items: [
        for (var n = 1; n <= flagCount; n++)
          CheckedPopupMenuItem(
            key: Key('flag-item-$n'),
            value: n,
            checked: card.flag == n,
            child: MenuEntryLabel(
              settings.flagName(n),
              flagShortcutLabel(n),
              leading: Icon(Icons.flag, color: colors.flag(n)),
            ),
          ),
        if (card.flag != 0) ...[
          const PopupMenuDivider(),
          PopupMenuItem(value: 0, child: Text(context.l10n.removeFlag)),
        ],
      ],
    );
    if (flag != null && mounted) _flag(flag);
  }

  /// Hides the app's navigation while the screen is on show (its
  /// destination is selected and nothing covers it) and the setting is on.
  /// The navigation belongs to an ancestor, so it is told after the frame.
  void _syncBars(bool hide) {
    if (hide == _hidingBars) return;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted || hide == _hidingBars) return;
      _hidingBars = hide;
      _barsHidden.value += hide ? 1 : -1;
    });
  }

  @override
  Widget build(BuildContext context) {
    final card = _state?.card;
    final hideBars = AppScope.of(context).settings.hideBarsDuringReview;
    final l = context.l10n;
    _onScreen =
        Visibility.of(context) && (ModalRoute.of(context)?.isCurrent ?? true);
    _syncBars(hideBars && _onScreen);
    return PopScope(
      // Stop as soon as the pop starts, so the video is gone well before the
      // player is freed (design decision 2).
      onPopInvokedWithResult: (didPop, _) {
        if (didPop) _player.stop();
      },
      child: Shortcuts(
        shortcuts: {
          SingleActivator(LogicalKeyboardKey.space): _RevealIntent(),
          SingleActivator(LogicalKeyboardKey.enter): _RevealIntent(),
          SingleActivator(LogicalKeyboardKey.numpadEnter): _RevealIntent(),
          SingleActivator(LogicalKeyboardKey.digit1): _RateIntent(1),
          SingleActivator(LogicalKeyboardKey.digit2): _RateIntent(2),
          SingleActivator(LogicalKeyboardKey.digit3): _RateIntent(3),
          SingleActivator(LogicalKeyboardKey.digit4): _RateIntent(4),
          SingleActivator(LogicalKeyboardKey.numpad1): _RateIntent(1),
          SingleActivator(LogicalKeyboardKey.numpad2): _RateIntent(2),
          SingleActivator(LogicalKeyboardKey.numpad3): _RateIntent(3),
          SingleActivator(LogicalKeyboardKey.numpad4): _RateIntent(4),
          SingleActivator(LogicalKeyboardKey.keyZ, control: true):
              _UndoIntent(),
          SingleActivator(LogicalKeyboardKey.keyZ, meta: true): _UndoIntent(),
          SingleActivator(LogicalKeyboardKey.escape): _LeaveIntent(),
          SingleActivator(LogicalKeyboardKey.keyR): _ReplayIntent(),
          SingleActivator(LogicalKeyboardKey.f5): _ReplayIntent(),
          SingleActivator(LogicalKeyboardKey.digit5): _PauseIntent(),
          SingleActivator(LogicalKeyboardKey.numpad5): _PauseIntent(),
          SingleActivator(LogicalKeyboardKey.digit6): _SeekIntent(-5),
          SingleActivator(LogicalKeyboardKey.numpad6): _SeekIntent(-5),
          SingleActivator(LogicalKeyboardKey.digit7): _SeekIntent(5),
          SingleActivator(LogicalKeyboardKey.numpad7): _SeekIntent(5),
          SingleActivator(LogicalKeyboardKey.keyV, shift: true):
              _RecordIntent(),
          SingleActivator(LogicalKeyboardKey.keyV): _PlayVoiceIntent(),
          SingleActivator(LogicalKeyboardKey.keyI): _CardInfoIntent(),
          previousCardInfoShortcut: _PreviousCardInfoIntent(),
          setDueDateShortcut: _SetDueDateIntent(),
          CharacterActivator('*'): _MarkIntent(),
          SingleActivator(LogicalKeyboardKey.keyA, shift: true):
              _AutoAdvanceIntent(),
          ..._flagShortcuts,
          ..._cardActionShortcuts,
        },
        child: Actions(
          actions: {
            // Space/Enter reveal the answer, then choose Good unless the
            // setting is off. While the answer input has focus, keys other
            // than Escape type into it.
            _RevealIntent: _StudyAction<_RevealIntent>(
              _typing,
              onInvoke: (_) {
                if (!_revealed) return _reveal();
                if (AppScope.read(context).settings.spaceAnswersGood) {
                  return _rate(3);
                }
                return null;
              },
            ),
            _RateIntent: _StudyAction<_RateIntent>(
              _typing,
              onInvoke: (i) => _rate(i.rating),
            ),
            _UndoIntent: _StudyAction<_UndoIntent>(
              _typing,
              onInvoke: (_) => _undo(),
            ),
            _LeaveIntent: CallbackAction<_LeaveIntent>(
              onInvoke: (_) => Navigator.of(context).maybePop(),
            ),
            _ReplayIntent: _StudyAction<_ReplayIntent>(
              _typing,
              onInvoke: (_) => _replay(),
            ),
            _PauseIntent: _StudyAction<_PauseIntent>(
              _typing,
              onInvoke: (_) => _player.togglePause(),
            ),
            _SeekIntent: _StudyAction<_SeekIntent>(
              _typing,
              onInvoke: (i) => _player.seekBy(Duration(seconds: i.seconds)),
            ),
            _RecordIntent: _StudyAction<_RecordIntent>(
              _typing,
              onInvoke: (_) => _toggleRecording(),
            ),
            _PlayVoiceIntent: _StudyAction<_PlayVoiceIntent>(
              _typing,
              onInvoke: (_) => _playVoice(),
            ),
            _CardInfoIntent: _StudyAction<_CardInfoIntent>(
              _typing,
              onInvoke: (_) {
                final card = _state?.card;
                if (card != null) _cardInfo(card.cardID);
                return null;
              },
            ),
            // Ctrl+Alt is AltGr on some keyboards, so it can type too.
            _PreviousCardInfoIntent: _StudyAction<_PreviousCardInfoIntent>(
              _typing,
              onInvoke: (_) => _previousCardInfo(),
            ),
            _MarkIntent: _StudyAction<_MarkIntent>(
              _typing,
              onInvoke: (_) => _toggleMark(),
            ),
            _AutoAdvanceIntent: _StudyAction<_AutoAdvanceIntent>(
              _typing,
              onInvoke: (_) => _toggleAutoAdvance(),
            ),
            // Ctrl+Shift+D types nothing either.
            _SetDueDateIntent: CallbackAction<_SetDueDateIntent>(
              onInvoke: (_) => _setDueDate(),
            ),
            // Ctrl+Delete deletes a word while typing an answer.
            _CardActionIntent: _StudyAction<_CardActionIntent>(
              _typing,
              onInvoke: (i) => _cardAction(i.action),
            ),
            // Ctrl+N types nothing, so it flags also while typing an answer.
            _FlagIntent: CallbackAction<_FlagIntent>(
              onInvoke: (i) => _flag(i.flag),
            ),
          },
          child: Focus(
            focusNode: _screenFocus,
            autofocus: true,
            child: Scaffold(
              appBar: hideBars
                  ? null
                  : AppBar(
                      title: Text(widget.deckName),
                      actions: [
                        if (_recording)
                          IconButton(
                            tooltip: withShortcut(
                              context.l10n.stopRecording,
                              recordVoiceShortcutLabel,
                            ),
                            icon: Icon(
                              Icons.mic,
                              color: Theme.of(context).colorScheme.error,
                            ),
                            onPressed: _toggleRecording,
                          ),
                        IconButton(
                          tooltip: withShortcut(
                            context.l10n.undo,
                            undoShortcutLabel,
                          ),
                          icon: const Icon(Icons.undo),
                          onPressed: _state?.canUndo == true && !_busy
                              ? _undo
                              : null,
                        ),
                        if (card != null)
                          PopupMenuButton<String>(
                            popUpAnimationStyle: popupAnimation(context),
                            key: _menuKey,
                            tooltip: context.l10n.cardActions,
                            onSelected: (v) {
                              switch (v) {
                                case 'replay':
                                  _replay();
                                case 'record':
                                  _toggleRecording();
                                case 'voice':
                                  _playVoice();
                                case 'flag':
                                  _flagMenu(card);
                                case 'mark':
                                  _toggleMark();
                                case 'info':
                                  _cardInfo(card.cardID);
                                case 'previous':
                                  _previousCardInfo();
                                case 'dueDate':
                                  _setDueDate();
                                case 'reset':
                                  _reset();
                                case 'autoAdvance':
                                  _toggleAutoAdvance();
                                default:
                                  _cardAction(v);
                              }
                            },
                            itemBuilder: (_) => [
                              PopupMenuItem(
                                value: 'edit',
                                child: Text(
                                  withShortcut(
                                    l.editCard,
                                    editCardShortcutLabel,
                                  ),
                                ),
                              ),
                              PopupMenuItem(
                                value: 'copy',
                                child: Text(l.createCopy),
                              ),
                              if (!_sounds.isEmpty)
                                PopupMenuItem(
                                  value: 'replay',
                                  child: Text(
                                    withShortcut(
                                      l.replayAudio,
                                      replayAudioShortcutLabel,
                                    ),
                                  ),
                                ),
                              PopupMenuItem(
                                value: 'record',
                                child: Text(
                                  withShortcut(
                                    _recording
                                        ? l.stopRecording
                                        : l.recordVoice,
                                    recordVoiceShortcutLabel,
                                  ),
                                ),
                              ),
                              if (_voice != null)
                                PopupMenuItem(
                                  value: 'voice',
                                  child: Text(
                                    withShortcut(
                                      l.playVoice,
                                      playVoiceShortcutLabel,
                                    ),
                                  ),
                                ),
                              PopupMenuItem(
                                value: 'flag',
                                child: Row(
                                  children: [
                                    Expanded(child: Text(l.flag)),
                                    const Icon(Icons.arrow_right),
                                  ],
                                ),
                              ),
                              PopupMenuItem(
                                value: 'mark',
                                child: Text(
                                  withShortcut(
                                    card.marked ? l.unmarkNote : l.markNote,
                                    markNoteShortcutLabel,
                                  ),
                                ),
                              ),
                              PopupMenuItem(
                                value: 'info',
                                child: Text(
                                  withShortcut(
                                    l.cardInfo,
                                    cardInfoShortcutLabel,
                                  ),
                                ),
                              ),
                              if ((_state?.studied ?? 0) > 0)
                                PopupMenuItem(
                                  value: 'previous',
                                  child: Text(
                                    withShortcut(
                                      l.previousCardInfo,
                                      previousCardInfoShortcutLabel,
                                    ),
                                  ),
                                ),
                              PopupMenuItem(
                                value: 'dueDate',
                                child: Text(
                                  withShortcut(
                                    l.setDueDate,
                                    setDueDateShortcutLabel,
                                  ),
                                ),
                              ),
                              PopupMenuItem(
                                value: 'reset',
                                child: Text(l.resetCard),
                              ),
                              PopupMenuItem(
                                value: 'bury',
                                child: Text(
                                  withShortcut(
                                    l.buryCardUntilTomorrow,
                                    buryCardShortcutLabel,
                                  ),
                                ),
                              ),
                              PopupMenuItem(
                                value: 'buryNote',
                                child: Text(
                                  withShortcut(
                                    l.buryNoteUntilTomorrow,
                                    buryNoteShortcutLabel,
                                  ),
                                ),
                              ),
                              PopupMenuItem(
                                value: 'suspend',
                                child: Text(
                                  withShortcut(
                                    l.suspendCard,
                                    suspendCardShortcutLabel,
                                  ),
                                ),
                              ),
                              PopupMenuItem(
                                value: 'suspendNote',
                                child: Text(
                                  withShortcut(
                                    l.suspendNote,
                                    suspendNoteShortcutLabel,
                                  ),
                                ),
                              ),
                              PopupMenuItem(
                                value: 'deleteNote',
                                child: Text(
                                  withShortcut(
                                    l.deleteNote,
                                    deleteNoteShortcutLabel,
                                  ),
                                ),
                              ),
                              const PopupMenuDivider(),
                              PopupMenuItem(
                                value: 'autoAdvance',
                                child: Text(
                                  withShortcut(
                                    _autoAdvance
                                        ? l.stopAutoAdvance
                                        : l.autoAdvance,
                                    autoAdvanceShortcutLabel,
                                  ),
                                ),
                              ),
                              PopupMenuItem(
                                value: 'options',
                                child: Text(
                                  withShortcut(
                                    l.deckOptions,
                                    deckOptionsShortcutLabel,
                                  ),
                                ),
                              ),
                            ],
                          ),
                      ],
                    ),
              body: hideBars
                  ? SafeArea(
                      bottom: false,
                      child: Column(
                        children: [
                          _LeaveBar(deckName: widget.deckName),
                          Expanded(child: _body(context)),
                        ],
                      ),
                    )
                  : _body(context),
            ),
          ),
        ),
      ),
    );
  }

  Widget _body(BuildContext context) {
    final state = _state;
    if (state == null) {
      if (_error != null) {
        return _Message(
          icon: Icons.error_outline,
          title: context.l10n.deckCouldNotOpen,
          lines: [errorMessage(_error!)],
        );
      }
      return const Center(child: CircularProgressIndicator());
    }
    final card = state.card;
    if (card == null) {
      final onContinue = _pending != null ? _continue : null;
      return state.studied == 0
          ? _NothingDue(
              state: state,
              studiedBefore: _studiedBefore,
              onContinue: onContinue,
            )
          : _Summary(
              state: state,
              studiedBefore: _studiedBefore,
              onCheck: () => _run(golang.study.current),
              onContinue: onContinue,
            );
    }
    final settings = AppScope.of(context).settings;
    final showCounts = settings.showRemainingCounts;
    return Column(
      children: [
        if (showCounts || card.showTimer || _autoAdvance) ...[
          Stack(
            alignment: Alignment.center,
            children: [
              if (showCounts)
                _Counts(counts: state.counts, kind: card.kind)
              else
                const SizedBox(height: _Counts.height),
              if (_autoAdvance)
                PositionedDirectional(
                  start: 4,
                  child: IconButton(
                    key: const Key('autoAdvanceOn'),
                    tooltip: withShortcut(
                      context.l10n.autoAdvanceIsOn,
                      autoAdvanceShortcutLabel,
                    ),
                    visualDensity: VisualDensity.compact,
                    color: Theme.of(context).colorScheme.primary,
                    icon: const Icon(Icons.fast_forward),
                    onPressed: _toggleAutoAdvance,
                  ),
                ),
              if (card.showTimer)
                PositionedDirectional(
                  end: 16,
                  child: _AnswerTimer(
                    elapsed: () => _timerFrozen ?? _timer.elapsed,
                    running: _timerFrozen == null,
                    limit: Duration(seconds: card.maxAnswerSeconds),
                  ),
                ),
            ],
          ),
          const Divider(height: 1),
        ],
        if (card.empty) const _EmptyCardNotice(),
        if (!_revealed && (_shownAgain || card.dueEarly != null))
          _ReturningCardNotice(dueEarly: card.dueEarly, again: _shownAgain),
        Expanded(
          child: Stack(
            children: [
              Positioned.fill(
                // A new card starts at the top.
                child: SingleChildScrollView(
                  key: ValueKey('scroll-${card.cardID}'),
                  child: _questionFade(
                    card,
                    settings.questionFadeMs,
                    CardView(
                      key: ValueKey('${card.cardID}-$_revealed'),
                      html: _revealed
                          ? _comparedAnswer ?? card.answer
                          : card.question,
                      background: card.background,
                      color: card.color,
                      fonts: card.fonts,
                      sounds: _revealed ? _sounds.answer : _sounds.question,
                      player: _player,
                      showPlayButtons: settings.showPlayButtons,
                      typeAnswer: TypeAnswerField(
                        controller: _typed,
                        focusNode: _typedFocus,
                        onSubmitted: _submitTyped,
                      ),
                      onAnswerAnchor: _revealed ? _scrollToAnswer : null,
                    ),
                  ),
                ),
              ),
              if (card.flag != 0)
                PositionedDirectional(
                  top: 8,
                  end: 8,
                  child: _FlagMarker(flag: card.flag),
                ),
              if (card.marked)
                const PositionedDirectional(
                  top: 8,
                  start: 8,
                  child: _MarkedStar(),
                ),
            ],
          ),
        ),
        const Divider(height: 1),
        SafeArea(
          top: false,
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: _revealed ? _ratingButtons(card) : _showAnswerButton(),
          ),
        ),
      ],
    );
  }

  /// Fades in the question of [card] over [ms] milliseconds; the answer, a
  /// fade of 0 and reduce motion show at once.
  Widget _questionFade(golang.StudyCard card, int ms, Widget child) {
    if (_revealed || ms <= 0 || reduceMotion(context)) return child;
    return TweenAnimationBuilder<double>(
      key: ValueKey('fade-${card.cardID}'),
      tween: Tween(begin: 0, end: 1),
      duration: Duration(milliseconds: ms),
      builder: (context, opacity, child) => Opacity(
        key: const Key('question-fade'),
        opacity: opacity,
        child: child,
      ),
      child: child,
    );
  }

  /// The answer's element with id `answer` is laid out: when the card is
  /// taller than the view, scrolls it to the top (as far as the card
  /// goes), once per reveal.
  void _scrollToAnswer(BuildContext anchor) {
    if (!_revealed || _scrolledToAnswer) return;
    _scrolledToAnswer = true;
    final position = Scrollable.maybeOf(anchor)?.position;
    if (position == null || position.maxScrollExtent <= 0) return;
    Scrollable.ensureVisible(anchor);
  }

  Widget _showAnswerButton() => SizedBox(
    width: double.infinity,
    height: 52,
    child: FilledButton(
      onPressed: _busy ? null : _reveal,
      child: Text(context.l10n.showAnswerButton),
    ),
  );

  Widget _ratingButtons(golang.StudyCard card) {
    final scheme = Theme.of(context).colorScheme;
    final showIntervals = AppScope.of(context).settings.showIntervals;
    final l = context.l10n;
    return Row(
      children: [
        for (final p in card.previews)
          Expanded(
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 4),
              child: SizedBox(
                height: 60,
                child: FilledButton.tonal(
                  key: Key('rate-${p.rating}'),
                  style: FilledButton.styleFrom(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 4,
                      vertical: 6,
                    ),
                    backgroundColor: p.rating == 1
                        ? scheme.errorContainer
                        : null,
                    foregroundColor: p.rating == 1
                        ? scheme.onErrorContainer
                        : null,
                  ),
                  onPressed: _busy ? null : () => _rate(p.rating),
                  child: FittedBox(
                    fit: BoxFit.scaleDown,
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        if (showIntervals)
                          Text(
                            p.ends
                                ? l.previewEnd
                                : formatIntervalLabel(p.interval, l10n: l),
                            style: const TextStyle(fontSize: 12),
                          ),
                        Text(ratingNames(l)[p.rating - 1]),
                      ],
                    ),
                  ),
                ),
              ),
            ),
          ),
      ],
    );
  }
}

/// With hidden bars, the top edge of the study screen: tapping it leaves
/// the session, as Escape does.
class _LeaveBar extends StatelessWidget {
  const _LeaveBar({required this.deckName});

  final String deckName;

  @override
  Widget build(BuildContext context) {
    final muted = Theme.of(context).colorScheme.onSurfaceVariant;
    return Semantics(
      button: true,
      label: context.l10n.leaveDeck(deckName),
      excludeSemantics: true,
      child: InkWell(
        key: const Key('leave-session'),
        onTap: () => Navigator.of(context).maybePop(),
        child: SizedBox(
          height: 28,
          width: double.infinity,
          child: Tooltip(
            message: withShortcut(context.l10n.leave, escShortcutLabel),
            child: Icon(Icons.keyboard_arrow_up, size: 20, color: muted),
          ),
        ),
      ),
    );
  }
}

/// The current card's flag, in the flag's colour.
class _FlagMarker extends StatelessWidget {
  const _FlagMarker({required this.flag});

  final int flag;

  @override
  Widget build(BuildContext context) {
    final name = AppScope.of(context).settings.flagName(flag);
    return Tooltip(
      message: context.l10n.flagNamed(name),
      child: Icon(
        Icons.flag,
        key: const Key('flag-marker'),
        color: FlagColors.of(context).flag(flag),
        semanticLabel: context.l10n.flagNamed(name),
      ),
    );
  }
}

/// The star on a card whose note is marked.
class _MarkedStar extends StatelessWidget {
  const _MarkedStar();

  @override
  Widget build(BuildContext context) => Tooltip(
    message: context.l10n.marked,
    child: Icon(
      Icons.star,
      key: const Key('marked-star'),
      color: FlagColors.of(context).marked,
      semanticLabel: context.l10n.marked,
    ),
  );
}

/// Remaining new, learning and review counts; the current card's kind is
/// underlined.
class _Counts extends StatelessWidget {
  const _Counts({required this.counts, required this.kind});

  final golang.Counts counts;
  final String kind;

  /// The height of the counts row, which an on-screen timer keeps when
  /// the counts are hidden.
  static const height = 40.0;

  @override
  Widget build(BuildContext context) {
    final colors = CountColors.of(context);
    Widget item(String k, int n, Color color, String label) => Padding(
      padding: const EdgeInsets.symmetric(horizontal: 10),
      child: Tooltip(
        message: label,
        child: Text(
          '$n',
          key: Key('remaining-$k'),
          style: TextStyle(
            color: color,
            fontWeight: FontWeight.w600,
            fontSize: 16,
            decoration: k == kind ? TextDecoration.underline : null,
          ),
        ),
      ),
    );
    return SizedBox(
      height: height,
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          item('new', counts.new_, colors.newCards, context.l10n.countNew),
          item(
            'learning',
            counts.learning,
            colors.learning,
            context.l10n.countLearning,
          ),
          item(
            'review',
            counts.review,
            colors.review,
            context.l10n.countToReview,
          ),
        ],
      ),
    );
  }
}

/// The time spent on the card, as m:ss, in the error colour once it reaches
/// [limit], the most an answer records.
class _AnswerTimer extends StatefulWidget {
  const _AnswerTimer({
    required this.elapsed,
    required this.running,
    required this.limit,
  });

  final Duration Function() elapsed;

  /// Whether the reading changes; a frozen timer is not rebuilt.
  final bool running;
  final Duration limit;

  @override
  State<_AnswerTimer> createState() => _AnswerTimerState();
}

class _AnswerTimerState extends State<_AnswerTimer> {
  Timer? _tick;

  @override
  void initState() {
    super.initState();
    _schedule();
  }

  @override
  void didUpdateWidget(_AnswerTimer oldWidget) {
    super.didUpdateWidget(oldWidget);
    _schedule();
  }

  @override
  void dispose() {
    _tick?.cancel();
    super.dispose();
  }

  /// Rebuilds when the shown second changes, while running.
  void _schedule() {
    _tick?.cancel();
    _tick = null;
    if (!widget.running) return;
    final ms = widget.elapsed().inMilliseconds % 1000;
    _tick = Timer(Duration(milliseconds: 1000 - ms), () {
      if (mounted) setState(_schedule);
    });
  }

  @override
  Widget build(BuildContext context) {
    final elapsed = widget.elapsed();
    final scheme = Theme.of(context).colorScheme;
    final seconds = (elapsed.inSeconds % 60).toString().padLeft(2, '0');
    return Tooltip(
      message: context.l10n.timeOnCard,
      child: Text(
        '${elapsed.inMinutes}:$seconds',
        key: const Key('answerTimer'),
        style: TextStyle(
          color: elapsed >= widget.limit
              ? scheme.error
              : scheme.onSurfaceVariant,
          fontWeight: FontWeight.w600,
          fontSize: 16,
          fontFeatures: const [FontFeature.tabularFigures()],
        ),
      ),
    );
  }
}

/// Says that the card's question is empty and how to remove such cards.
class _EmptyCardNotice extends StatelessWidget {
  const _EmptyCardNotice();

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return Container(
      key: const Key('emptyCardNotice'),
      width: double.infinity,
      color: scheme.tertiaryContainer,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
      child: Row(
        children: [
          Icon(Icons.info_outline, color: scheme.onTertiaryContainer),
          const SizedBox(width: 12),
          Expanded(
            child: Text(
              context.l10n.emptyCardNotice,
              style: TextStyle(color: scheme.onTertiaryContainer),
            ),
          ),
        ],
      ),
    );
  }
}

/// Says that the card on screen is shown before it is due, or is the card
/// just answered, shown again, and when it is due: so an answer does not
/// look like a missed click, nor the card like it came back too soon.
class _ReturningCardNotice extends StatelessWidget {
  const _ReturningCardNotice({required this.dueEarly, required this.again});

  /// When the card is due, when it is shown before then; null when it is
  /// due now.
  final DateTime? dueEarly;

  /// The card is the card just answered.
  final bool again;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final due = dueEarly;
    final String text;
    if (due == null) {
      text = context.l10n.shownAgainDueNow;
    } else {
      // In whole minutes, rounded up, at least one.
      final seconds = due.difference(clock.now()).inSeconds;
      final minutes = seconds <= 60 ? 1 : (seconds / 60).ceil();
      final time = formatSpent(Duration(minutes: minutes), l10n: context.l10n);
      text = again
          ? context.l10n.shownAgainEarly(time)
          : context.l10n.shownEarly(time);
    }
    return Semantics(
      liveRegion: true,
      child: Container(
        key: const Key('returningCardNotice'),
        width: double.infinity,
        color: scheme.tertiaryContainer,
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
        child: Row(
          children: [
            Icon(
              again ? Icons.replay : Icons.schedule,
              color: scheme.onTertiaryContainer,
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                text,
                style: TextStyle(color: scheme.onTertiaryContainer),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _Message extends StatelessWidget {
  const _Message({
    required this.icon,
    required this.title,
    required this.lines,
    this.extra,
    this.action,
  });

  final IconData icon;
  final String title;
  final List<String> lines;

  /// Shown below the lines, such as the streak and goal lines.
  final Widget? extra;
  final Widget? action;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 56, color: theme.colorScheme.primary),
            const SizedBox(height: 16),
            Text(
              title,
              style: theme.textTheme.headlineSmall,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 12),
            for (final l in lines)
              Padding(
                padding: const EdgeInsets.only(bottom: 6),
                child: Text(l, textAlign: TextAlign.center),
              ),
            ?extra,
            if (action != null) ...[const SizedBox(height: 16), action!],
          ],
        ),
      ),
    );
  }
}

/// Says that learning cards are due now and offers to continue.
class _LearningDueNow extends StatelessWidget {
  const _LearningDueNow({required this.lines, required this.onContinue});

  final List<String> lines;
  final VoidCallback onContinue;

  @override
  Widget build(BuildContext context) => _Message(
    icon: Icons.schedule,
    title: context.l10n.learningDueNow,
    lines: lines,
    action: Wrap(
      spacing: 12,
      children: [
        FilledButton(
          onPressed: onContinue,
          child: Text(context.l10n.continueAction),
        ),
        OutlinedButton(
          onPressed: () => Navigator.of(context).maybePop(),
          child: Text(context.l10n.backToDecks),
        ),
      ],
    ),
  );
}

class _NothingDue extends StatelessWidget {
  const _NothingDue({
    required this.state,
    required this.studiedBefore,
    this.onContinue,
  });

  final golang.StudyState state;

  /// See [_ProgressLines.studiedBefore].
  final bool? studiedBefore;

  /// Set when a learning card became due while waiting.
  final VoidCallback? onContinue;

  @override
  Widget build(BuildContext context) {
    if (onContinue case final onContinue?) {
      return _LearningDueNow(lines: const [], onContinue: onContinue);
    }
    final next = state.learningDue ?? state.nextDue;
    return _Message(
      icon: Icons.celebration_outlined,
      title: context.l10n.congratulationsNothingDue,
      lines: [
        if (next != null)
          context.l10n.nextCardDue(formatDue(next))
        else
          context.l10n.deckHasNoCards,
      ],
      extra: _ProgressLines(studiedBefore: studiedBefore),
      action: OutlinedButton(
        onPressed: () => Navigator.of(context).maybePop(),
        child: Text(context.l10n.backToDecks),
      ),
    );
  }
}

class _Summary extends StatelessWidget {
  const _Summary({
    required this.state,
    required this.studiedBefore,
    required this.onCheck,
    this.onContinue,
  });

  final golang.StudyState state;

  /// See [_ProgressLines.studiedBefore].
  final bool? studiedBefore;
  final VoidCallback onCheck;

  /// Set when a learning card became due while waiting.
  final VoidCallback? onContinue;

  @override
  Widget build(BuildContext context) {
    final l = context.l10n;
    final studied = l.studiedInTime(
      state.studied,
      formatSpent(state.timeSpent),
    );
    if (onContinue case final onContinue?) {
      return _LearningDueNow(lines: [studied], onContinue: onContinue);
    }
    final learning = state.learningDue;
    return _Message(
      icon: Icons.check_circle_outline,
      title: l.sessionComplete,
      lines: [
        studied,
        if (learning != null) l.moreLearningDue(formatDue(learning)),
      ],
      extra: _ProgressLines(
        studiedBefore: studiedBefore,
        newAchievements: true,
      ),
      action: Wrap(
        spacing: 12,
        children: [
          if (learning != null)
            FilledButton(onPressed: onCheck, child: Text(l.checkAgain)),
          OutlinedButton(
            onPressed: () => Navigator.of(context).maybePop(),
            child: Text(l.backToDecks),
          ),
        ],
      ),
    );
  }
}

/// The streak and goal lines of the summary and nothing-due screens
/// (progress spec, "Summary progress lines"), the summary's line of new
/// achievements (study-session spec, "New achievements on the summary"),
/// and the one-time suggestion to set a daily goal. Nothing is shown while
/// streaks and goals are off.
class _ProgressLines extends StatefulWidget {
  const _ProgressLines({
    required this.studiedBefore,
    this.newAchievements = false,
  });

  /// Whether today was a study day when the session started: if not and it
  /// is now, the session extended the streak.
  final bool? studiedBefore;

  /// Whether to name the achievement tiers not shown yet, and mark them
  /// shown so the deck list does not tell again.
  final bool newAchievements;

  @override
  State<_ProgressLines> createState() => _ProgressLinesState();
}

class _ProgressLinesState extends State<_ProgressLines> {
  /// The new achievement tiers named here, kept once marked shown.
  final _new = <golang.NewAchievement>[];

  @override
  void initState() {
    super.initState();
    // The numbers after the session's last answer.
    AppScope.read(context).progress.load();
  }

  /// "New: 1,000 answers" for the tiers new so far, which are then marked
  /// shown; null while there are none.
  String? _newLine(ProgressModel progress, AppLocalizations l) {
    final fresh = progress.newAchievements;
    if (fresh.isNotEmpty) {
      _new.addAll(fresh);
      progress.markSeen(fresh);
    }
    return _new.isEmpty ? null : newAchievementsLine(l, _new);
  }

  @override
  Widget build(BuildContext context) {
    final app = AppScope.of(context);
    return ListenableBuilder(
      listenable: Listenable.merge([app.progress, app.settings]),
      builder: (context, _) {
        final settings = app.settings;
        final p = app.progress.progress;
        if (!settings.showProgress || p == null) {
          return const SizedBox.shrink();
        }
        final l = context.l10n;
        final lines = [
          if (widget.studiedBefore == false && p.todayStudied)
            l.streakExtended(p.current)
          else if (p.current > 0)
            l.streakCurrent(p.current),
          ?switch (p.goalKind) {
            _ when p.goalReached => l.goalReached,
            'cards' => l.goalCardsLeft(p.goal - p.todayCards),
            'minutes' => l.goalMinutesLeft(p.goal - p.todayMinutes),
            _ => null,
          },
          if (widget.newAchievements) ?_newLine(app.progress, l),
        ];
        final suggest =
            settings.dailyGoalKind.isEmpty && !settings.goalSuggestionShown;
        return Column(
          key: const Key('progress-lines'),
          mainAxisSize: MainAxisSize.min,
          children: [
            for (final line in lines)
              Padding(
                padding: const EdgeInsets.only(bottom: 6),
                child: Text(line, textAlign: TextAlign.center),
              ),
            if (suggest) ...[
              const SizedBox(height: 8),
              Text(
                l.goalSuggestion,
                textAlign: TextAlign.center,
                style: Theme.of(context).textTheme.bodySmall,
              ),
              Wrap(
                alignment: WrapAlignment.center,
                children: [
                  TextButton(
                    onPressed: () {
                      settings.markGoalSuggestionShown();
                      editDailyGoal(context);
                    },
                    child: Text(l.goalSuggestionSet),
                  ),
                  TextButton(
                    onPressed: settings.markGoalSuggestionShown,
                    child: Text(l.goalSuggestionDismiss),
                  ),
                ],
              ),
            ],
          ],
        );
      },
    );
  }
}
