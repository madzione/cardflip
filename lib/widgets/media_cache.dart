import 'dart:async';

import 'package:flutter/foundation.dart';

import 'package:cardflip/golang/golang.dart' as golang;

/// Card images from Go, with a small least-recently-used cache so that
/// flipping cards does not fetch the same image over and over.
class MediaCache {
  MediaCache({
    Future<Uint8List> Function(String name)? load,
    this.maxBytes = 32 << 20,
  }) : _load = load ?? golang.media.get_;

  /// The cache the card views use.
  static MediaCache instance = MediaCache();

  final Future<Uint8List> Function(String name) _load;
  final int maxBytes;
  // Map literals keep insertion order: the first key is the oldest.
  final _entries = <String, Uint8List>{};
  final _pending = <String, Future<Uint8List>>{};
  int _bytes = 0;

  /// Returns a media file's bytes. Fails when the file is missing.
  Future<Uint8List> get(String name) {
    final hit = _entries.remove(name);
    if (hit != null) {
      _entries[name] = hit; // most recently used last
      return SynchronousFuture(hit);
    }
    return _pending[name] ??= _load(name)
        .then((bytes) {
          _put(name, bytes);
          return bytes;
        })
        .whenComplete(() {
          // A block body: returning the removed future would wait on itself.
          _pending.remove(name);
        });
  }

  void _put(String name, Uint8List bytes) {
    _entries[name] = bytes;
    _bytes += bytes.length;
    while (_bytes > maxBytes && _entries.length > 1) {
      final oldest = _entries.keys.first;
      _bytes -= _entries.remove(oldest)!.length;
    }
  }

  /// Forgets everything (e.g. after restoring a backup).
  void clear() {
    _entries.clear();
    _bytes = 0;
  }
}

/// Loads a media file's bytes from [cache]; names in fields are sometimes
/// URL-encoded.
Future<Uint8List> loadMediaBytes(MediaCache cache, String name) async {
  try {
    return await cache.get(name);
  } catch (_) {
    final decoded = Uri.decodeFull(name);
    if (decoded == name) rethrow;
    return cache.get(decoded);
  }
}
