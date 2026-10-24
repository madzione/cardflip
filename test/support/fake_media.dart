import 'dart:async';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/widgets/media_files.dart';
import 'package:cardflip/widgets/media_player.dart';
import 'package:cardflip/widgets/voice_recorder.dart';

/// Media files for tests: the names in [available] exist.
class FakeMediaFiles implements MediaFiles {
  FakeMediaFiles([Iterable<String> available = const []])
    : available = {...available};

  final Set<String> available;
  int releases = 0;

  @override
  Future<String> resolve(String name) async {
    if (!available.contains(name)) throw MediaMissing(name);
    return 'mem:$name';
  }

  @override
  void release() => releases++;
}

/// A player backend for tests. Each file plays until [finish] is called,
/// or fails at once when it is in [unplayable].
class FakeBackend implements PlaybackBackend {
  /// The files played, in order, without the `mem:` prefix.
  final played = <String>[];
  final unplayable = <String>{};
  final log = <String>[];
  Completer<void>? _done;

  /// The file playing now, or null.
  String? get playing => _done == null ? null : played.last;

  @override
  Future<void> play(String uri, {required bool video}) {
    final name = uri.substring('mem:'.length);
    played.add(name);
    if (unplayable.contains(name)) {
      return Future.error(const PlaybackFailed('unsupported codec'));
    }
    final done = _done = Completer<void>();
    return done.future;
  }

  /// Ends the file playing now.
  void finish() {
    final done = _done;
    _done = null;
    done?.complete();
  }

  @override
  Future<void> stop() async {
    log.add('stop');
    finish();
  }

  @override
  Future<void> pause() async => log.add('pause');

  @override
  Future<void> resume() async => log.add('resume');

  @override
  Future<void> seekBy(Duration delta) async =>
      log.add('seek ${delta.inSeconds}');

  @override
  Widget videoView() => const SizedBox(key: Key('fake-video'));

  @override
  Future<void> close() async {
    log.add('close');
    finish();
  }
}

/// Makes new players use [backend] and [files]; undone after the test.
void useFakeMedia(FakeBackend backend, FakeMediaFiles files) {
  final oldBackend = MediaPlayer.createBackend;
  final oldFiles = MediaPlayer.createFiles;
  MediaPlayer.createBackend = () => backend;
  MediaPlayer.createFiles = () => files;
  addTearDown(() {
    MediaPlayer.createBackend = oldBackend;
    MediaPlayer.createFiles = oldFiles;
  });
}

/// A recorder for tests: each recording is `mem:voice-N`.
class FakeRecorder implements VoiceRecorder {
  bool refuse = false;
  bool recording = false;
  int count = 0;
  final discarded = <String>[];

  @override
  Future<void> start() async {
    if (refuse) throw const MicrophoneRefused();
    recording = true;
  }

  @override
  Future<String?> stop() async {
    if (!recording) return null;
    recording = false;
    return 'mem:voice-${count++}';
  }

  @override
  Future<RecordedAudio> read(String uri) async =>
      (bytes: Uint8List.fromList(uri.codeUnits), extension: '.m4a');

  @override
  Future<void> discard(String uri) async => discarded.add(uri);

  @override
  void dispose() {}
}

/// Makes the study screen record with [recorder]; undone after the test.
void useFakeRecorder(FakeRecorder recorder) {
  final old = VoiceRecorder.create;
  VoiceRecorder.create = () => recorder;
  addTearDown(() => VoiceRecorder.create = old);
}
