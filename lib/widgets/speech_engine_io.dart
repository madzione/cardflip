import 'dart:async';
import 'dart:io';

import 'speech.dart';

/// flutter_tts on native platforms; speech-dispatcher on Linux, which the
/// plugin does not support.
SpeechEngine createEngine() =>
    Platform.isLinux ? SpdSayEngine() : FlutterTtsEngine();

/// Speaks with speech-dispatcher's `spd-say`. Without it, there are no
/// voices.
class SpdSayEngine implements SpeechEngine {
  SpdSayEngine({this.command = 'spd-say'});

  final String command;
  Process? _speaking;

  @override
  Future<List<Voice>> voices() async {
    final ProcessResult result;
    try {
      result = await Process.run(command, ['-L']);
    } on ProcessException {
      return const [];
    }
    if (result.exitCode != 0) return const [];
    return parseVoiceList('${result.stdout}');
  }

  /// The voices in the output of `spd-say -L`: a header line, then one
  /// line per voice with its name, language and variant.
  static List<Voice> parseVoiceList(String output) {
    final voices = <Voice>[];
    for (final line in output.split('\n').skip(1)) {
      final words = line.trim().split(RegExp(r'\s+'));
      if (words.length < 3) continue;
      final name = words.sublist(0, words.length - 2).join(' ');
      voices.add(Voice(name, normalizeLang(words[words.length - 2])));
    }
    return voices;
  }

  /// Anki's speed (1 is normal) as speech-dispatcher's rate, -100 to 100.
  static int rate(double speed) => ((speed - 1) * 100).round().clamp(-100, 100);

  @override
  Future<void> speak(
    String text, {
    required Voice voice,
    required double speed,
  }) async {
    await stop();
    final Process p;
    try {
      p = _speaking = await Process.start(command, [
        '-w',
        '-l',
        voice.lang.replaceAll('_', '-').toLowerCase(),
        '-y',
        voice.name,
        '-r',
        '${rate(speed)}',
        '--',
        text,
      ]);
    } on ProcessException {
      return;
    }
    unawaited(p.stdout.drain<void>());
    unawaited(p.stderr.drain<void>());
    await p.exitCode;
    if (identical(_speaking, p)) _speaking = null;
  }

  @override
  Future<void> stop() async {
    final p = _speaking;
    if (p == null) return;
    _speaking = null;
    p.kill();
    // The message goes on after its client is gone; stop it.
    try {
      await Process.run(command, ['-S']);
    } on ProcessException {
      // Not installed any more: nothing to stop.
    }
  }
}
