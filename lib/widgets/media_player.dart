import 'dart:async';

import 'package:flutter/material.dart';
import 'package:media_kit/media_kit.dart';
import 'package:media_kit_video/media_kit_video.dart';

import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';

import 'media_files.dart';
import 'speech.dart';

/// The side of a card a sound belongs to.
enum CardSide { question, answer }

/// One sound, video or utterance of a card side: what the side's
/// `<cf-sound idx=N>`, `<cf-video idx=N>` or `<cf-tts idx=N>` placeholder
/// refers to.
@immutable
class MediaItem {
  const MediaItem(this.side, this.idx, this.file, {this.video = false})
    : direct = false,
      utterance = null;

  /// A recording of the user's voice, played from [uri] as it is.
  const MediaItem.recording(String uri)
    : side = CardSide.question,
      idx = -1,
      file = uri,
      video = false,
      direct = true,
      utterance = null;

  /// Text to speech; [file] is its text.
  MediaItem.tts(this.side, this.idx, golang.Utterance this.utterance)
    : file = utterance.text,
      video = false,
      direct = false;

  final CardSide side;
  final int idx;
  final String file;
  final bool video;

  /// Whether [file] is a URI to open as it is, not a media file name.
  final bool direct;

  /// What to speak, for text to speech; null for files.
  final golang.Utterance? utterance;

  bool get isTts => utterance != null;

  /// The items of one side, from the media list Go rendered with it.
  static List<MediaItem> of(CardSide side, List<golang.MediaRef> refs) => [
    for (final (i, r) in refs.indexed)
      MediaItem(side, i, r.file, video: r.kind == 'video'),
  ];

  static final _placeholder = RegExp(r'<cf-(sound|video|tts) idx="(\d+)"');

  /// The sounds, videos and utterances of one side in the order their
  /// placeholders appear in its [html], so they play in document order.
  /// Any that [html] does not show come last.
  static List<MediaItem> inOrder(
    CardSide side,
    String html,
    List<golang.MediaRef> media,
    List<golang.Utterance> tts,
  ) {
    final files = of(side, media);
    final spoken = [
      for (final (i, u) in tts.indexed) MediaItem.tts(side, i, u),
    ];
    final out = <MediaItem>[];
    for (final m in _placeholder.allMatches(html)) {
      final list = m[1] == 'tts' ? spoken : files;
      final i = int.parse(m[2]!);
      if (i < list.length && !out.contains(list[i])) out.add(list[i]);
    }
    return [
      ...out,
      ...[...files, ...spoken].where((item) => !out.contains(item)),
    ];
  }

  @override
  bool operator ==(Object other) =>
      other is MediaItem &&
      other.side == side &&
      other.idx == idx &&
      other.file == file &&
      other.isTts == isTts;

  @override
  int get hashCode => Object.hash(side, idx, file, isTts);

  @override
  String toString() => '${side.name}[$idx] ${isTts ? 'tts ' : ''}$file';
}

/// A card's sounds and the deck options that say how they autoplay.
@immutable
class CardSounds {
  const CardSounds({
    this.question = const [],
    this.answer = const [],
    this.autoplay = true,
    this.replayQuestion = true,
  });

  final List<MediaItem> question;
  final List<MediaItem> answer;

  /// Off with the deck option "Don't play audio automatically".
  final bool autoplay;

  /// Off with the deck option "Skip question when replaying answer".
  final bool replayQuestion;

  bool get isEmpty => question.isEmpty && answer.isEmpty;
}

/// A file that could not be played.
class PlaybackFailed implements Exception {
  const PlaybackFailed(this.message);

  final String message;

  @override
  String toString() => message;
}

/// Plays one file at a time. [MediaKitBackend] is the real one; tests use a
/// fake.
abstract class PlaybackBackend {
  /// Plays [uri] to its end. Completes when it ended or was stopped, and
  /// fails with [PlaybackFailed] when it cannot be played.
  Future<void> play(String uri, {required bool video});

  Future<void> stop();
  Future<void> pause();
  Future<void> resume();

  /// Moves the position by [delta], not before the start.
  Future<void> seekBy(Duration delta);

  /// The surface showing the video being played.
  Widget videoView();

  /// Stops playback and frees the player, in that order; the backend is not
  /// used again. Completes the pending [play].
  Future<void> close();
}

/// Plays a card's sounds and text to speech in order, one side's list at a
/// time, and follows the autoplay and replay rules of the study screen.
class MediaPlayer extends ChangeNotifier {
  MediaPlayer({
    PlaybackBackend Function()? backend,
    MediaFiles? files,
    this._speech,
  }) : _makeBackend = backend ?? createBackend,
       files = files ?? createFiles() {
    _live.add(this);
  }

  /// What new players use; tests replace these.
  static PlaybackBackend Function() createBackend = MediaKitBackend.new;
  static MediaFiles Function() createFiles = StoredMediaFiles.new;

  // The players not yet disposed, wherever they are on screen.
  static final _live = <MediaPlayer>{};

  /// Stops every player, as when the user switches to another destination
  /// and can no longer see what is playing.
  static void stopAll() {
    // A copy: a listener may dispose a player while it is told to stop.
    for (final p in _live.toList()) {
      p.stop();
    }
  }

  final PlaybackBackend Function() _makeBackend;
  final MediaFiles files;

  // Created on first use, so cards without sounds never load the player.
  PlaybackBackend? _backend;
  PlaybackBackend get backend => _backend ??= _makeBackend();

  final Speech? _speech;

  /// What speaks the utterances.
  Speech get speech => _speech ?? Speech.shared;

  final _queue = <MediaItem>[];
  final _problems = <String, String>{};
  final _checks = <String, Future<String?>>{};
  int _generation = 0;
  bool _running = false;
  bool _disposed = false;

  /// The item playing (or paused), or null.
  MediaItem? get current => _current;
  MediaItem? _current;

  bool get paused => _paused;
  bool _paused = false;

  /// Why [file] cannot be played, or null when it can or was not tried.
  String? problem(String file) => _problems[file];

  /// Checks that [file] exists; completes with the reason it cannot be
  /// played, or null.
  Future<String?> check(String file) => _checks[file] ??= files
      .resolve(file)
      .then<String?>((_) => _problems[file])
      .catchError((Object e) => _fail(file, e));

  /// Stops what is playing and plays [items] in order.
  void play(List<MediaItem> items) {
    stop();
    enqueue(items);
  }

  /// Plays [items] after what is playing or queued.
  void enqueue(List<MediaItem> items) {
    if (items.isEmpty) return;
    _queue.addAll(items);
    if (!_running) unawaited(_run());
  }

  Future<void> _run() async {
    _running = true;
    final gen = _generation;
    while (_queue.isNotEmpty && gen == _generation && !_disposed) {
      final item = _queue.removeAt(0);
      if (item.utterance case final u?) {
        // Without a voice it is left out, as an unplayable file is.
        if (await speech.voiceFor(u) == null) continue;
        if (gen != _generation || _disposed) return;
        _current = item;
        _paused = false;
        notifyListeners();
        await speech.speak(u);
        continue;
      }
      if (_problems.containsKey(item.file)) continue;
      _current = item;
      _paused = false;
      notifyListeners();
      try {
        final uri = item.direct ? item.file : await files.resolve(item.file);
        if (gen != _generation || _disposed) return;
        await backend.play(uri, video: item.video);
      } catch (e) {
        if (gen != _generation || _disposed) return;
        // A file that cannot be played never stops the others.
        _fail(item.file, e);
        _checks[item.file] = Future.value(_problems[item.file]);
      }
    }
    if (gen != _generation || _disposed) return;
    _running = false;
    _current = null;
    _paused = false;
    notifyListeners();
  }

  String _fail(String file, Object error) {
    final reason = error is MediaMissing
        ? appL10n.fileMissing(file)
        : appL10n.fileCannotPlay(file);
    _problems[file] = reason;
    if (!_disposed) notifyListeners();
    return reason;
  }

  /// Stops playback and forgets the queue.
  void stop() {
    _generation++;
    _queue.clear();
    _running = false;
    final was = _current;
    _current = null;
    _paused = false;
    if (was != null) {
      unawaited(was.isTts ? speech.stop() : _backend?.stop());
      notifyListeners();
    }
  }

  /// Pauses or resumes the current item; speech cannot be paused.
  void togglePause() {
    if (_current == null || _current!.isTts) return;
    _paused = !_paused;
    unawaited(_paused ? backend.pause() : backend.resume());
    notifyListeners();
  }

  /// Seeks the current item by [delta].
  void seekBy(Duration delta) {
    if (_current == null || _current!.isTts) return;
    unawaited(backend.seekBy(delta));
  }

  // The study screen's rules (design decision 4).

  /// A side was shown: play its sounds unless autoplay is off.
  void showQuestion(CardSounds sounds) {
    stop();
    if (sounds.autoplay) enqueue(sounds.question);
  }

  /// The answer was revealed: stop the question's sounds when [interrupt]
  /// is on, then play the answer's (after the question's otherwise). The
  /// question's speech always stops.
  void showAnswer(CardSounds sounds, {required bool interrupt}) {
    if (interrupt) {
      stop();
    } else {
      _queue.removeWhere((item) => item.isTts);
      // Ends the speaking, and the queue goes on.
      if (_current?.isTts == true) unawaited(speech.stop());
    }
    if (sounds.autoplay) enqueue(sounds.answer);
  }

  /// Replays the current side; on the answer, the question's sounds come
  /// first unless the deck skips them.
  void replay(CardSounds sounds, {required bool answerShown}) {
    if (!answerShown) return play(sounds.question);
    play([if (sounds.replayQuestion) ...sounds.question, ...sounds.answer]);
  }

  /// The card is left: stop everything and free its files.
  void leave() {
    stop();
    _checks.clear();
    _problems.clear();
    files.release();
  }

  @override
  void dispose() {
    _live.remove(this);
    _disposed = true;
    // The backend stops itself before it frees the player: stopping it here
    // as well would overlap the two (design decision 1).
    _generation++;
    _queue.clear();
    _running = false;
    if (_current?.isTts == true) unawaited(speech.stop());
    _current = null;
    _paused = false;
    files.release();
    unawaited(_backend?.close());
    super.dispose();
  }
}

/// Plays audio and video with media_kit (libmpv on native platforms, the
/// browser's media element on the web).
class MediaKitBackend implements PlaybackBackend {
  Player? _player;
  VideoController? _video;
  Completer<void>? _done;
  final _subscriptions = <StreamSubscription<Object?>>[];

  Player get _p => _player ??= _create();

  Player _create() {
    MediaKit.ensureInitialized();
    final p = Player();
    _subscriptions
      ..add(
        p.stream.completed.listen((done) {
          if (done) _finish();
        }),
      )
      ..add(p.stream.error.listen((e) => _finish(error: e)));
    return p;
  }

  void _finish({String? error}) {
    final done = _done;
    _done = null;
    if (done == null || done.isCompleted) return;
    if (error == null) {
      done.complete();
    } else {
      done.completeError(PlaybackFailed(error));
    }
  }

  @override
  Future<void> play(String uri, {required bool video}) async {
    _finish();
    final done = _done = Completer<void>();
    if (video) _video ??= VideoController(_p);
    await _p.open(Media(uri));
    return done.future;
  }

  @override
  Future<void> stop() async {
    _finish();
    await _player?.stop();
  }

  @override
  Future<void> pause() async => _player?.pause();

  @override
  Future<void> resume() async => _player?.play();

  @override
  Future<void> seekBy(Duration delta) async {
    final p = _player;
    if (p == null) return;
    final to = p.state.position + delta;
    await p.seek(to.isNegative ? Duration.zero : to);
  }

  @override
  Widget videoView() => Video(controller: _video ??= VideoController(_p));

  @override
  Future<void> close() async {
    _finish();
    for (final s in _subscriptions) {
      unawaited(s.cancel());
    }
    final p = _player;
    if (p == null) return;
    // A stop that never finishes must not keep the player alive.
    try {
      await p.stop().timeout(const Duration(seconds: 3));
    } catch (_) {}
    await p.dispose();
  }
}
