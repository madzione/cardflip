import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/media_files.dart';
import 'package:cardflip/widgets/media_player.dart';
import 'package:cardflip/widgets/speech.dart';

import 'support/fake_media.dart';
import 'support/fake_speech.dart';

golang.Utterance utterance(String lang, String text) =>
    golang.Utterance(text: text, lang: lang, voices: const [], speed: 1);

const q1 = MediaItem(CardSide.question, 0, 'q1.mp3');
const q2 = MediaItem(CardSide.question, 1, 'q2.mp3');
const a1 = MediaItem(CardSide.answer, 0, 'a1.mp3');

CardSounds sounds({bool autoplay = true, bool replayQuestion = true}) =>
    CardSounds(
      question: const [q1, q2],
      answer: const [a1],
      autoplay: autoplay,
      replayQuestion: replayQuestion,
    );

void main() {
  late FakeBackend backend;
  late FakeMediaFiles files;
  late MediaPlayer player;

  setUp(() {
    backend = FakeBackend();
    files = FakeMediaFiles(['q1.mp3', 'q2.mp3', 'a1.mp3']);
    player = MediaPlayer(backend: () => backend, files: files);
  });
  tearDown(() => player.dispose());

  /// Lets the queue move on after the current file ended.
  Future<void> next() async {
    backend.finish();
    await pumpEventQueue();
  }

  test('a list plays in order, one file after the other', () async {
    player.play(const [q1, q2]);
    await pumpEventQueue();
    expect(backend.played, ['q1.mp3']);
    expect(player.current, q1);
    await next();
    expect(backend.played, ['q1.mp3', 'q2.mp3']);
    await next();
    expect(player.current, isNull);
  });

  test('showing the question autoplays it unless autoplay is off', () async {
    player.showQuestion(sounds());
    await pumpEventQueue();
    expect(backend.played, ['q1.mp3']);

    player.leave();
    backend.played.clear();
    player.showQuestion(sounds(autoplay: false));
    await pumpEventQueue();
    expect(backend.played, isEmpty);
  });

  test('revealing interrupts the question and plays the answer', () async {
    player.showQuestion(sounds());
    await pumpEventQueue();
    player.showAnswer(sounds(), interrupt: true);
    await pumpEventQueue();
    expect(backend.played, ['q1.mp3', 'a1.mp3']);
    expect(backend.log, contains('stop'));
  });

  test(
    'without interrupting, the answer is queued after the question',
    () async {
      player.showQuestion(sounds());
      await pumpEventQueue();
      player.showAnswer(sounds(), interrupt: false);
      await pumpEventQueue();
      expect(backend.played, ['q1.mp3']);
      await next();
      await next();
      expect(backend.played, ['q1.mp3', 'q2.mp3', 'a1.mp3']);
    },
  );

  test('replaying the answer plays the question first', () async {
    player.replay(sounds(), answerShown: true);
    await pumpEventQueue();
    await next();
    await next();
    expect(backend.played, ['q1.mp3', 'q2.mp3', 'a1.mp3']);
  });

  test('replaying the answer can skip the question', () async {
    player.replay(sounds(replayQuestion: false), answerShown: true);
    await pumpEventQueue();
    expect(backend.played, ['a1.mp3']);
  });

  test('replaying the question plays only the question', () async {
    player.replay(sounds(), answerShown: false);
    await pumpEventQueue();
    await next();
    await next();
    expect(backend.played, ['q1.mp3', 'q2.mp3']);
  });

  test('pause, resume and seek act on the current file', () async {
    player.togglePause(); // nothing playing: ignored
    player.play(const [q1]);
    await pumpEventQueue();
    player.togglePause();
    expect(player.paused, isTrue);
    player.togglePause();
    player.seekBy(const Duration(seconds: -5));
    player.seekBy(const Duration(seconds: 5));
    expect(backend.log, ['pause', 'resume', 'seek -5', 'seek 5']);
  });

  test('missing and unplayable files are skipped and explained', () async {
    files.available.remove('q1.mp3');
    backend.unplayable.add('q2.mp3');
    player.play(const [q1, q2, a1]);
    await pumpEventQueue();
    expect(backend.played, ['q2.mp3', 'a1.mp3']);
    expect(player.problem('q1.mp3'), 'q1.mp3 is missing');
    expect(player.problem('q2.mp3'), 'q2.mp3 cannot be played on this device');
    expect(await player.check('a1.mp3'), isNull);
    expect(await player.check('q1.mp3'), 'q1.mp3 is missing');
  });

  test('leaving the card stops playback and frees its files', () async {
    player.play(const [q1, q2]);
    await pumpEventQueue();
    player.leave();
    await pumpEventQueue();
    expect(player.current, isNull);
    expect(backend.played, ['q1.mp3']);
    expect(files.releases, 1);
  });

  test('disposing while playing only closes the backend', () async {
    final own = FakeBackend();
    final disposed = MediaPlayer(backend: () => own, files: files);
    disposed.play(const [q1, q2]);
    await pumpEventQueue();
    expect(own.playing, 'q1.mp3');
    disposed.dispose();
    await pumpEventQueue();
    expect(own.log, ['close']);
    expect(own.playing, isNull);
    // The queue does not move on to the next file.
    expect(own.played, ['q1.mp3']);
  });

  test('stopAll stops every player and forgets their queues', () async {
    final other = FakeBackend();
    final second = MediaPlayer(backend: () => other, files: files);
    addTearDown(second.dispose);
    player.play(const [q1, q2]);
    second.play(const [a1]);
    await pumpEventQueue();
    expect(backend.playing, 'q1.mp3');
    expect(other.playing, 'a1.mp3');

    MediaPlayer.stopAll();
    await pumpEventQueue();
    expect(backend.log, ['stop']);
    expect(other.log, ['stop']);
    expect(player.current, isNull);
    expect(second.current, isNull);
    // Nothing queued plays afterwards.
    expect(backend.played, ['q1.mp3']);
  });

  test('stopAll leaves disposed players alone', () async {
    final disposed = _CountingPlayer(files);
    disposed.play(const [q1]);
    await pumpEventQueue();
    disposed.dispose();
    final before = disposed.stops;

    MediaPlayer.stopAll();
    expect(disposed.stops, before);
  });

  group('text to speech', () {
    late FakeSpeechEngine engine;
    late MediaPlayer speaking;

    setUp(() {
      engine = FakeSpeechEngine();
      speaking = MediaPlayer(
        backend: () => backend,
        files: files,
        speech: Speech(engine),
      );
    });
    tearDown(() => speaking.dispose());

    test('items are in the order of their placeholders', () {
      final items = MediaItem.inOrder(
        CardSide.answer,
        '<cf-tts idx="0"></cf-tts> x <cf-sound idx="0"></cf-sound>'
        '<cf-tts idx="1"></cf-tts>',
        [golang.MediaRef(file: 'a1.mp3', kind: 'audio')],
        [utterance('es_ES', 'uno'), utterance('es_ES', 'dos')],
      );
      expect(items.map((i) => '${i.isTts ? 'tts' : 'file'} ${i.file}'), [
        'tts uno',
        'file a1.mp3',
        'tts dos',
      ]);
    });

    test('sounds and utterances share one queue', () async {
      final hola = MediaItem.tts(
        CardSide.question,
        0,
        utterance('es_ES', 'hola'),
      );
      speaking.play([q1, hola, q2]);
      await pumpEventQueue();
      expect(backend.played, ['q1.mp3']);
      await next();
      expect(engine.spoken, ['es_ES:hola']);
      expect(speaking.current, hola);
      expect(backend.played, ['q1.mp3']);
      engine.finish();
      await pumpEventQueue();
      expect(backend.played, ['q1.mp3', 'q2.mp3']);
    });

    test('stop silences speech', () async {
      speaking.play([
        MediaItem.tts(CardSide.question, 0, utterance('es_ES', 'hola')),
      ]);
      await pumpEventQueue();
      speaking.stop();
      expect(engine.stops, 1);
      expect(speaking.current, isNull);
    });

    test('revealing without interrupting still ends the question\'s '
        'speech', () async {
      final hola = MediaItem.tts(
        CardSide.question,
        0,
        utterance('es_ES', 'hola'),
      );
      final adios = MediaItem.tts(
        CardSide.question,
        1,
        utterance('es_ES', 'adiós'),
      );
      speaking.play([hola, adios, q1]);
      await pumpEventQueue();
      expect(engine.speaking, 'es_ES:hola');
      speaking.showAnswer(const CardSounds(answer: [a1]), interrupt: false);
      await pumpEventQueue();
      // The question's sound goes on, then the answer's.
      expect(engine.spoken, ['es_ES:hola']);
      expect(backend.played, ['q1.mp3']);
      await next();
      expect(backend.played, ['q1.mp3', 'a1.mp3']);
    });

    test('an utterance without a voice is skipped', () async {
      speaking.play([
        MediaItem.tts(CardSide.question, 0, utterance('ja_JP', '猫')),
        q1,
      ]);
      await pumpEventQueue();
      expect(engine.spoken, isEmpty);
      expect(backend.played, ['q1.mp3']);
    });
  });
}

/// Counts the calls to [stop].
class _CountingPlayer extends MediaPlayer {
  _CountingPlayer(MediaFiles files)
    : super(backend: FakeBackend.new, files: files);

  int stops = 0;

  @override
  void stop() {
    stops++;
    super.stop();
  }
}
