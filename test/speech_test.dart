import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/speech.dart';
import 'package:cardflip/widgets/speech_engine_io.dart';

import 'support/fake_speech.dart';

golang.Utterance utterance(
  String lang, {
  List<String> voices = const [],
  String text = 'hola',
  double speed = 1,
}) => golang.Utterance(text: text, lang: lang, voices: voices, speed: speed);

void main() {
  test('languages are written as Anki writes them', () {
    expect(normalizeLang('en-us'), 'en_US');
    expect(normalizeLang('EN_us'), 'en_US');
    expect(normalizeLang('de'), 'de');
    expect(normalizeLang(''), '');
  });

  group('picking a voice', () {
    const voices = [
      Voice('Samantha', 'en_US'),
      Voice('Daniel', 'en_GB'),
      Voice('Microsoft Zira Desktop - English (United States)', 'en_US'),
      Voice('Paulina', 'es_MX'),
      Voice('Mónica', 'es_ES'),
    ];

    test('language and region first', () {
      expect(pickVoice(voices, utterance('en_GB')), voices[1]);
      expect(pickVoice(voices, utterance('es_ES')), voices[4]);
      expect(pickVoice(voices, utterance('es-es')), voices[4]);
    });

    test('then the language in another region', () {
      expect(pickVoice(voices, utterance('en_AU')), voices[0]);
      expect(pickVoice(voices, utterance('es_AR')), voices[3]);
    });

    test('none for another language (spec: no voice)', () {
      expect(pickVoice(voices, utterance('ja_JP')), isNull);
      expect(pickVoice(const [], utterance('en_US')), isNull);
    });

    test('voices= picks the first one the device has, by Anki name', () {
      expect(
        pickVoice(
          voices,
          utterance('en_US', voices: ['Apple_Karen', 'Apple_Daniel']),
        ),
        voices[1],
      );
      expect(
        pickVoice(voices, utterance('en_US', voices: ['Microsoft_Zira'])),
        voices[2],
      );
      // None of them: a voice for the language.
      expect(
        pickVoice(voices, utterance('es_ES', voices: ['Apple_Jorge'])),
        voices[4],
      );
    });

    test('no language: the first voice', () {
      expect(pickVoice(voices, utterance('')), voices[0]);
    });
  });

  group('Speech', () {
    test('speaks with the voice for the language', () async {
      final engine = FakeSpeechEngine();
      final speech = Speech(engine);
      final done = speech.speak(utterance('es_ES', text: 'perro'));
      await pumpEventQueue();
      expect(engine.spoken, ['es_ES:perro']);
      engine.finish();
      await done;
    });

    test('says nothing without a voice', () async {
      final engine = FakeSpeechEngine();
      await Speech(engine).speak(utterance('ja_JP'));
      expect(engine.spoken, isEmpty);
    });

    test('asks for the voices again while there are none', () async {
      final engine = FakeSpeechEngine(const []);
      final speech = Speech(engine);
      expect(await speech.voices(), isEmpty);
      engine.voiceList = const [Voice('Anna', 'de_DE')];
      expect(await speech.voices(), [const Voice('Anna', 'de_DE')]);
    });
  });

  group('speech-dispatcher', () {
    test('reads the voice list of spd-say -L', () {
      const output = '''
                     NAME                 LANGUAGE                  VARIANT
                Afrikaans                       af                     none
       English (America)                    en-us                     none
                  German                       de                     none
''';
      expect(SpdSayEngine.parseVoiceList(output), const [
        Voice('Afrikaans', 'af'),
        Voice('English (America)', 'en_US'),
        Voice('German', 'de'),
      ]);
    });

    test('speed as rate', () {
      expect(SpdSayEngine.rate(1), 0);
      expect(SpdSayEngine.rate(1.4), 40);
      expect(SpdSayEngine.rate(0.5), -50);
      expect(SpdSayEngine.rate(3), 100);
    });

    test('no spd-say, no voices', () async {
      final engine = SpdSayEngine(command: 'spd-say-not-installed');
      expect(await engine.voices(), isEmpty);
    });
  });
}
