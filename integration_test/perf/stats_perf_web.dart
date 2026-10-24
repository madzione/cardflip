import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';

import 'stats_flow.dart';

/// Statistics of 50,000 cards in the browser, with Go as WebAssembly and the
/// collection in IndexedDB. Run it like tool/web_integration.sh runs the app
/// flow, with this file as the --target.
void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('statistics of 50,000 cards in the browser', (tester) async {
    await runStatsFlow(tester, dataDir: '', isWeb: true);
  });
}
