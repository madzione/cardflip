import 'dart:io';

Map<String, String> get environment => Platform.environment;

/// The file's text, or null when it can't be read.
Future<String?> readText(String path) async {
  try {
    return await File(path).readAsString();
  } on FileSystemException {
    return null;
  }
}
