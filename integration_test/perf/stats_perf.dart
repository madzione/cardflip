import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';

import 'stats_flow.dart';

/// Statistics of 50,000 cards on Linux, against the real Go library:
///
///   fvm flutter test integration_test/perf/stats_perf.dart -d linux
void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('statistics of 50,000 cards', (tester) async {
    final dir = await Directory.systemTemp.createTemp('cardflip-stats-');
    addTearDown(() => dir.delete(recursive: true));
    await runStatsFlow(tester, dataDir: dir.path, isWeb: false);
  });
}
