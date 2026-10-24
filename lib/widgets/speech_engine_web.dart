import 'speech.dart';

/// Browsers speak through flutter_tts (the Web Speech API).
SpeechEngine createEngine() => FlutterTtsEngine();
