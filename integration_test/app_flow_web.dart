import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';

import 'package:cardflip/golang/golang.dart' as golang;

import 'flow.dart';
import 'fixture.g.dart';

/// The same flow in the browser: Go runs as WebAssembly, the collection is
/// stored in IndexedDB, and files are staged as bytes. Run it with
/// tool/web_integration.sh, which also generates fixture.g.dart.
void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('the whole app in the browser, with the real Go core', (
    tester,
  ) async {
    final picker = FlowPicker(
      stageFixture: () => golang.transfer.stageBytes(
        'modern.apkg',
        base64Decode(modernApkgBase64),
      ),
      stageBackup: (String name, Uint8List bytes) =>
          golang.transfer.stageBytes(name, bytes),
    );
    await runAppFlow(tester, picker: picker, dataDir: '', isWeb: true);
  });
}
