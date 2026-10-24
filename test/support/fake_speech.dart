import 'dart:async';

import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/widgets/speech.dart';

/// A speech engine for tests with [voiceList]. Each text is spoken until
/// [finish] is called.
class FakeSpeechEngine implements SpeechEngine {
  FakeSpeechEngine([this.voiceList = const [Voice('Mónica', 'es_ES')]]);

  List<Voice> voiceList;

  /// The texts spoken, in order, as `lang:text`.
  final spoken = <String>[];
  int stops = 0;
  Completer<void>? _done;

  /// The text being spoken now, or null.
  String? get speaking => _done == null ? null : spoken.last;

  @override
  Future<List<Voice>> voices() async => voiceList;

  @override
  Future<void> speak(
    String text, {
    required Voice voice,
    required double speed,
  }) {
    _done?.complete();
    spoken.add('${voice.lang}:$text');
    final done = _done = Completer<void>();
    return done.future;
  }

  /// Ends the text being spoken.
  void finish() {
    final done = _done;
    _done = null;
    done?.complete();
  }

  @override
  Future<void> stop() async {
    stops++;
    finish();
  }
}

/// Makes the app speak with [engine]; undone after the test.
void useFakeSpeech(FakeSpeechEngine engine) {
  final old = Speech.create;
  Speech.create = () => Speech(engine);
  Speech.reset();
  addTearDown(() {
    Speech.create = old;
    Speech.reset();
  });
}
