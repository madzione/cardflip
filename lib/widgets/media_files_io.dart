import 'dart:io';
import 'dart:typed_data';

import 'package:path/path.dart' as p;
import 'package:path_provider/path_provider.dart';

// Media is written once per run under a short generated name: names in
// fields can contain characters that are awkward in paths.
final _paths = <String, String>{};
Future<Directory>? _dir;

Future<Directory> _cacheDir() => _dir ??= () async {
  final dir = Directory(
    p.join((await getApplicationCacheDirectory()).path, 'media'),
  );
  await dir.create(recursive: true);
  return dir;
}();

Future<String> materialise(String name, Uint8List bytes) async {
  final known = _paths[name];
  if (known != null && File(known).existsSync()) return known;
  final dir = await _cacheDir();
  final path = p.join(dir.path, '${_paths.length}${p.extension(name)}');
  await File(path).writeAsBytes(bytes, flush: true);
  return _paths[name] = path;
}

void release(String uri) {
  // Files stay for the rest of the run; the cache is cleared at startup.
}

Future<void> clearCache() async {
  final dir = Directory(
    p.join((await getApplicationCacheDirectory()).path, 'media'),
  );
  if (await dir.exists()) await dir.delete(recursive: true);
}

void forget() => _paths.clear();

Future<void> deleteFile(String path) async {
  final f = File(path);
  if (await f.exists()) await f.delete();
}

/// Reads a recording: its bytes and its extension.
Future<(Uint8List, String)> readRecording(String path) async =>
    (await File(path).readAsBytes(), p.extension(path));
