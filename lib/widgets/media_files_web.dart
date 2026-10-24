import 'dart:js_interop';
import 'dart:typed_data';

import 'package:web/web.dart' as web;

const _types = {
  '.mp3': 'audio/mpeg',
  '.ogg': 'audio/ogg',
  '.oga': 'audio/ogg',
  '.opus': 'audio/ogg',
  '.m4a': 'audio/mp4',
  '.aac': 'audio/aac',
  '.wav': 'audio/wav',
  '.weba': 'audio/webm',
  '.flac': 'audio/flac',
  '.webm': 'video/webm',
  '.mp4': 'video/mp4',
};

Future<String> materialise(String name, Uint8List bytes) async {
  final dot = name.lastIndexOf('.');
  final type = dot < 0 ? null : _types[name.substring(dot).toLowerCase()];
  final blob = web.Blob(
    [bytes.toJS].toJS,
    web.BlobPropertyBag(type: type ?? ''),
  );
  return web.URL.createObjectURL(blob);
}

void release(String uri) => web.URL.revokeObjectURL(uri);

Future<void> clearCache() async {}

void forget() {}

Future<void> deleteFile(String path) async {}

/// File extensions of the audio types browsers record.
const _recordedExtensions = {
  'audio/webm': '.weba',
  'audio/ogg': '.ogg',
  'audio/mp4': '.m4a',
  'audio/aac': '.aac',
  'audio/mpeg': '.mp3',
  'audio/wav': '.wav',
  'audio/x-wav': '.wav',
};

/// Reads a recording from its Blob URL: its bytes and the extension of
/// its type ("" when unknown).
Future<(Uint8List, String)> readRecording(String uri) async {
  final response = await web.window.fetch(uri.toJS).toDart;
  final blob = await response.blob().toDart;
  final buffer = await blob.arrayBuffer().toDart;
  final type = blob.type.split(';').first.trim().toLowerCase();
  return (buffer.toDart.asUint8List(), _recordedExtensions[type] ?? '');
}
