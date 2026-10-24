import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';

import 'package:cardflip/golang/golang.dart' as golang;

/// Fonts from the collection's media, for the note type's `@font-face`
/// rules. The files of a CSS family are registered with the engine once,
/// under a family name of their own, so families of different note types
/// never clash. Flutter cannot unload fonts: [clear] only makes the next
/// request load the files again, under a new name.
class FontCache {
  FontCache({
    Future<Uint8List> Function(String name)? load,
    Future<void> Function(String family, List<Uint8List> fonts)? register,
  }) : _load = load ?? golang.media.get_,
       _register = register ?? _loadFonts;

  /// The cache the card views use.
  static FontCache instance = FontCache();

  final Future<Uint8List> Function(String name) _load;
  final Future<void> Function(String family, List<Uint8List> fonts) _register;
  final _pending = <String, Future<String?>>{};
  final _loaded = <String, String>{};
  int _next = 0;

  static Future<void> _loadFonts(String family, List<Uint8List> fonts) {
    final loader = FontLoader(family);
    for (final bytes in fonts) {
      loader.addFont(Future.value(ByteData.sublistView(bytes)));
    }
    return loader.load();
  }

  static String _key(Iterable<String> files) =>
      (files.toSet().toList()..sort()).join('\n');

  /// The family name [files] are registered under, when that is done.
  String? loaded(Iterable<String> files) => _loaded[_key(files)];

  /// Registers the font [files] of one CSS family and returns the family
  /// name to use, or null when none of them could be loaded. Missing files
  /// are tried again on the next call, in case they were added since.
  Future<String?> family(Iterable<String> files) {
    final key = _key(files);
    final hit = _loaded[key];
    if (hit != null) return SynchronousFuture(hit);
    return _pending[key] ??= _family(key).whenComplete(() {
      // A block body: returning the removed future would wait on itself.
      _pending.remove(key);
    });
  }

  Future<String?> _family(String key) async {
    final fonts = <Uint8List>[];
    for (final file in key.split('\n')) {
      final bytes = await _bytes(file);
      if (bytes != null) fonts.add(bytes);
    }
    if (fonts.isEmpty) return null;
    final name = 'cf-font-${_next++}';
    try {
      await _register(name, fonts);
    } catch (_) {
      return null; // not a font the engine can read
    }
    return _loaded[key] = name;
  }

  /// A media file's bytes, or null when it is missing. Names in CSS are
  /// sometimes URL-encoded.
  Future<Uint8List?> _bytes(String file) async {
    for (final name in {file, _decoded(file)}) {
      try {
        return await _load(name);
      } catch (_) {
        // try the next spelling
      }
    }
    return null;
  }

  static String _decoded(String name) {
    try {
      return Uri.decodeFull(name);
    } on ArgumentError {
      return name;
    }
  }

  /// Forgets the loaded fonts (e.g. after restoring a backup), so that
  /// changed files are loaded again.
  void clear() {
    _pending.clear();
    _loaded.clear();
  }
}
