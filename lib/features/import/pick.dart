import 'package:file_picker/file_picker.dart';
import 'package:flutter/foundation.dart';

import 'package:cardflip/golang/golang.dart' as golang;

/// A file the user picked, with its content.
typedef PickedFile = ({String name, Uint8List bytes});

/// The extensions of the images, audio and video the editor attaches: the
/// formats every platform shows or plays.
const mediaExtensions = [
  'jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp', // images
  'mp3', 'ogg', 'oga', 'opus', 'm4a', 'aac', 'wav', 'flac', 'weba', // audio
  'mp4', 'webm', 'm4v', // video
];

/// Picks files to import or restore, picks media to attach, and saves
/// exported files. Tests replace it with a fake.
class FilePickerService {
  const FilePickerService();

  /// Lets the user pick a file and stages it in Go: by path on native
  /// platforms (Go reads it directly), by content on the web (nothing is
  /// uploaded anywhere). Returns null when the user cancels.
  Future<golang.StageInfo?> pick(List<String> extensions) async {
    final files = await FilePicker.pickFiles(
      type: FileType.custom,
      allowedExtensions: extensions,
    );
    if (files.isEmpty) return null;
    final file = files.first;
    final path = file.path;
    if (!kIsWeb && path != null) {
      return golang.transfer.stagePath(path);
    }
    return golang.transfer.stageBytes(file.name, await file.readAsBytes());
  }

  /// Lets the user pick images, audio and video files to attach to a
  /// note. Returns an empty list when the user cancels.
  Future<List<PickedFile>> pickMedia() async {
    final files = await FilePicker.pickFiles(
      type: FileType.custom,
      allowedExtensions: mediaExtensions,
    );
    return [
      for (final f in files) (name: f.name, bytes: await f.readAsBytes()),
    ];
  }

  /// Saves bytes under a suggested name: a save dialog on native platforms,
  /// a download on the web. Returns false when the user cancels.
  Future<bool> save(String fileName, Uint8List bytes) async {
    final uri = await FilePicker.saveFile(fileName: fileName, bytes: bytes);
    return uri != null;
  }
}
