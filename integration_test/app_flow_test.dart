import 'dart:io';
import 'dart:typed_data';

import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';

import 'package:cardflip/golang/golang.dart' as golang;

import 'flow.dart';

/// Finds a file of the repository, from wherever the test runs.
File repoFile(String relative) {
  final root = Platform.environment['CARDFLIP_ROOT'];
  if (root != null) return File('$root/$relative');
  var dir = Directory.current.absolute;
  while (true) {
    final f = File('${dir.path}/$relative');
    if (f.existsSync()) return f;
    if (dir.parent.path == dir.path) {
      throw StateError(
        '$relative not found above ${Directory.current.path}; set CARDFLIP_ROOT',
      );
    }
    dir = dir.parent;
  }
}

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('the whole app on Linux, with the real Go core', (tester) async {
    final dir = await Directory.systemTemp.createTemp('cardflip-it-');
    addTearDown(() => dir.delete(recursive: true));
    // Native platforms stage files by path: Go reads them directly.
    final picker = FlowPicker(
      stageFixture: () => golang.transfer.stagePath(
        repoFile('golang/importer/apkg/testdata/modern.apkg').path,
      ),
      stageBackup: (String name, Uint8List bytes) {
        final file = File('${dir.path}/$name')..writeAsBytesSync(bytes);
        return golang.transfer.stagePath(file.path);
      },
    );
    await runAppFlow(
      tester,
      picker: picker,
      dataDir: '${dir.path}/data',
      isWeb: false,
    );
  });
}
