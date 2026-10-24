import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter_tts/flutter_tts.dart';

import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';

import 'speech_engine_io.dart'
    if (dart.library.js_interop) 'speech_engine_web.dart'
    as platform;

/// A voice of the device's speech engine. [lang] is written as Anki writes
/// languages, e.g. `en_US`.
@immutable
class Voice {
  const Voice(this.name, this.lang);

  final String name;
  final String lang;

  @override
  bool operator ==(Object other) =>
      other is Voice && other.name == name && other.lang == lang;

  @override
  int get hashCode => Object.hash(name, lang);

  @override
  String toString() => '$lang $name';
}

/// Writes a language code like `en-us` or `EN_US` as `en_US`.
String normalizeLang(String lang) {
  final parts = lang.trim().split(RegExp('[-_]'));
  if (parts.first.isEmpty) return '';
  return [
    parts.first.toLowerCase(),
    if (parts.length > 1) parts[1].toUpperCase(),
    ...parts.skip(2),
  ].join('_');
}

/// The voice that speaks [u]: the first of its `voices=` the device has,
/// else one for its language and region, else one for its language. An
/// utterance without a language gets the first voice. Null when none fits.
Voice? pickVoice(List<Voice> voices, golang.Utterance u) {
  String simple(String name) => name.toLowerCase().replaceAll('_', ' ').trim();
  for (final wanted in u.voices.map(simple)) {
    // Anki's names carry the vendor, e.g. Apple_Samantha for Samantha, and
    // Windows adds a description: Microsoft Zira Desktop - English (…).
    final bare = wanted.contains(' ')
        ? wanted.substring(wanted.indexOf(' ') + 1)
        : wanted;
    for (final v in voices) {
      final name = simple(v.name);
      if (name == wanted || name == bare || name.startsWith('$wanted ')) {
        return v;
      }
    }
  }
  final lang = normalizeLang(u.lang);
  if (lang.isEmpty) return voices.firstOrNull;
  final language = lang.split('_').first;
  return voices.where((v) => normalizeLang(v.lang) == lang).firstOrNull ??
      voices
          .where((v) => normalizeLang(v.lang).split('_').first == language)
          .firstOrNull;
}

/// Why an utterance in [lang] has no voice, for its disabled button.
String noVoiceMessage(String lang) => lang.isEmpty
    ? appL10n.noVoiceForThisLanguage
    : appL10n.noVoiceForLanguage(lang);

/// A platform's speech engine. Tests use a fake.
abstract class SpeechEngine {
  /// The voices the device has; empty when it cannot speak.
  Future<List<Voice>> voices();

  /// Speaks [text] with [voice] at [speed] (1 is normal). Completes when it
  /// was spoken or [stop]ped.
  Future<void> speak(
    String text, {
    required Voice voice,
    required double speed,
  });

  Future<void> stop();
}

/// Speaks utterances with the device's speech engine: flutter_tts, or
/// speech-dispatcher on Linux.
class Speech {
  Speech(this.engine);

  /// What [shared] is made with; tests replace it.
  static Speech Function() create = () => Speech(platform.createEngine());

  static Speech? _shared;

  /// The app's speech, made on first use.
  static Speech get shared => _shared ??= create();

  /// Forgets [shared], so the next use makes it with [create].
  @visibleForTesting
  static void reset() => _shared = null;

  final SpeechEngine engine;
  Future<List<Voice>>? _voices;

  /// The device's voices. An empty list is asked for again next time:
  /// browsers load their voices after the page.
  Future<List<Voice>> voices() {
    final voices = _voices ??= engine.voices().catchError(
      (Object _) => const <Voice>[],
    );
    unawaited(
      voices.then((v) {
        if (v.isEmpty && identical(_voices, voices)) _voices = null;
      }),
    );
    return voices;
  }

  /// The voice that speaks [u], or null when the device has none for it.
  Future<Voice?> voiceFor(golang.Utterance u) async =>
      pickVoice(await voices(), u);

  /// Speaks [u]; completes when it was spoken or stopped. Does nothing when
  /// the device has no voice for it.
  Future<void> speak(golang.Utterance u) async {
    final voice = await voiceFor(u);
    if (voice == null) return;
    await engine.speak(u.text, voice: voice, speed: u.speed);
  }

  Future<void> stop() => engine.stop();
}

/// The flutter_tts engine (Android, iOS, macOS, Windows, web).
class FlutterTtsEngine implements SpeechEngine {
  final _tts = FlutterTts();
  Completer<void>? _done;
  bool _ready = false;

  @override
  Future<List<Voice>> voices() async {
    final raw = await _tts.getVoices;
    if (raw is! List) return const [];
    return [
      for (final v in raw)
        if (v is Map && v['name'] != null)
          Voice('${v['name']}', normalizeLang('${v['locale'] ?? ''}')),
    ];
  }

  /// Anki's speed (1 is normal) on the plugin's scale for this platform.
  static double _rate(double speed) {
    if (kIsWeb) return speed;
    return switch (defaultTargetPlatform) {
      TargetPlatform.iOS ||
      TargetPlatform.macOS ||
      TargetPlatform.windows => 0.5 * speed,
      _ => speed,
    };
  }

  @override
  Future<void> speak(
    String text, {
    required Voice voice,
    required double speed,
  }) async {
    _finish();
    final done = _done = Completer<void>();
    if (!_ready) {
      await _tts.awaitSpeakCompletion(true);
      _ready = true;
    }
    await _tts.setVoice({
      'name': voice.name,
      'locale': voice.lang.replaceAll('_', '-'),
    });
    await _tts.setSpeechRate(_rate(speed));
    if (done.isCompleted) return;
    // Some platforms never complete speak() after stop(): stop() ends it.
    unawaited(_tts.speak(text).whenComplete(() => _finish(done)));
    return done.future;
  }

  void _finish([Completer<void>? which]) {
    final done = _done;
    if (done == null || which != null && which != done) return;
    _done = null;
    if (!done.isCompleted) done.complete();
  }

  @override
  Future<void> stop() async {
    _finish();
    await _tts.stop();
  }
}
