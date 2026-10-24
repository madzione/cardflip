import 'dart:typed_data';

import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';

import 'media_files_io.dart'
    if (dart.library.js_interop) 'media_files_web.dart'
    as platform;

/// A media file that is not in the collection.
class MediaMissing implements Exception {
  const MediaMissing(this.name);

  final String name;

  @override
  String toString() => appL10n.fileMissing(name);
}

/// Turns media files of the collection into something the player can open:
/// a file in the cache directory on native platforms, a Blob URL on the web.
abstract class MediaFiles {
  /// Returns a URI or path the player opens for [name]. Fails with
  /// [MediaMissing] when the collection has no such file.
  Future<String> resolve(String name);

  /// Frees what [resolve] made for the current card (Blob URLs on the web).
  void release();
}

/// The media files of the collection, read through Go.
class StoredMediaFiles implements MediaFiles {
  StoredMediaFiles({Future<Uint8List> Function(String name)? load})
    : _load = load ?? golang.media.get_;

  final Future<Uint8List> Function(String name) _load;
  final _resolved = <String, Future<String>>{};

  @override
  Future<String> resolve(String name) =>
      _resolved[name] ??= _bytes(name)
          .then((b) => platform.materialise(name, b));

  Future<Uint8List> _bytes(String name) async {
    try {
      return await _load(name);
    } catch (_) {
      // Names in fields are sometimes URL-encoded.
      final decoded = Uri.decodeFull(name);
      if (decoded != name) {
        try {
          return await _load(decoded);
        } catch (_) {}
      }
      throw MediaMissing(name);
    }
  }

  @override
  void release() {
    final made = _resolved.values.toList();
    _resolved.clear();
    for (final f in made) {
      f.then(platform.release, onError: (_) {});
    }
  }

  /// Removes the files materialised in earlier runs; called at startup.
  static Future<void> clearCache() => platform.clearCache();

  /// Forgets materialised files, so they are read again (e.g. after a
  /// backup was restored).
  static void forget() => platform.forget();
}
