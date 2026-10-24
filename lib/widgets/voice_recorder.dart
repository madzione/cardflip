import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:path/path.dart' as p;
import 'package:path_provider/path_provider.dart';
import 'package:record/record.dart';
import 'package:cardflip/l10n/l10n.dart';

import 'media_files_io.dart'
    if (dart.library.js_interop) 'media_files_web.dart'
    as platform;

/// Microphone access was refused.
class MicrophoneRefused implements Exception {
  const MicrophoneRefused();

  @override
  String toString() => appL10n.microphoneRefused;
}

/// A recording's content and file extension (".m4a").
typedef RecordedAudio = ({Uint8List bytes, String extension});

/// Records the user's voice: for comparing pronunciation in study, and for
/// adding to a note in the editor. Recordings are temporary: a file in the
/// temporary directory on native platforms, a Blob URL on the web, part of
/// the collection only once the editor adds them to its media.
abstract class VoiceRecorder {
  /// What the study screen and the editor use; tests replace it.
  static VoiceRecorder Function() create = RecordVoiceRecorder.new;

  /// Starts recording. Fails with [MicrophoneRefused] without permission.
  Future<void> start();

  /// Stops recording and returns what the player opens to play it back.
  Future<String?> stop();

  /// Reads a recording [stop] returned.
  Future<RecordedAudio> read(String uri);

  /// Deletes a recording [stop] returned.
  Future<void> discard(String uri);

  void dispose();
}

/// Records with the `record` package.
class RecordVoiceRecorder implements VoiceRecorder {
  AudioRecorder? _recorder;
  AudioRecorder get _r => _recorder ??= AudioRecorder();
  int _count = 0;

  /// The extension of the encoder [start] chose.
  String _ext = '.wav';

  @override
  Future<void> start() async {
    if (!await _r.hasPermission()) throw const MicrophoneRefused();
    var encoder = AudioEncoder.wav;
    for (final e in [AudioEncoder.aacLc, AudioEncoder.opus]) {
      if (await _r.isEncoderSupported(e)) {
        encoder = e;
        break;
      }
    }
    final ext = _ext = switch (encoder) {
      AudioEncoder.aacLc => '.m4a',
      AudioEncoder.opus => '.ogg',
      _ => '.wav',
    };
    var path = '';
    if (!kIsWeb) {
      final dir = await getTemporaryDirectory();
      path = p.join(dir.path, 'cardflip-voice-${_count++}$ext');
    }
    await _r.start(RecordConfig(encoder: encoder), path: path);
  }

  @override
  Future<String?> stop() => _r.stop();

  @override
  Future<RecordedAudio> read(String uri) async {
    final (bytes, ext) = await platform.readRecording(uri);
    return (bytes: bytes, extension: ext.isEmpty ? _ext : ext);
  }

  @override
  Future<void> discard(String uri) async {
    if (kIsWeb) return platform.release(uri);
    await platform.deleteFile(uri);
  }

  @override
  void dispose() {
    final r = _recorder;
    if (r != null) unawaited(r.dispose());
  }
}
