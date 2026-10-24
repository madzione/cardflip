import 'package:flutter/widgets.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/data/sync_model.dart';

import 'support/fake_go.dart';
import 'support/fake_sync_platform.dart';

void main() {
  /// A started model on [platform], after Go's first status.
  Future<SyncModel> started(
    WidgetTester tester,
    FakeSyncPlatform platform,
  ) async {
    final model = SyncModel(isWeb: false, platform: platform)..start();
    addTearDown(model.dispose);
    await tester.pumpAndSettle();
    return model;
  }

  testWidgets('nothing is announced or searched for on an untrusted network', (
    tester,
  ) async {
    final go = FakeGo()
      ..syncStatus = syncStatusJson(wifi: 'Café', trusted: const []);
    final platform = FakeSyncPlatform(
      current: const NetworkState('wifi', wifi: 'Café', subnet: '10.0.0.0/24'),
    );
    final model = await started(tester, platform);
    expect(model.status!.listening, isFalse);
    expect(go.syncCalls, contains('SetNetwork wifi Café 10.0.0.0/24'));
    expect(
      platform.calls.where((c) => c.startsWith('advertise') || c == 'browse'),
      isEmpty,
    );

    // Trusted: announced and searching; untrusted again: both stop.
    go.setSyncStatus(
      syncStatusJson(
        wifi: 'Café',
        trusted: const [
          {'id': 'wifi:Café', 'wifi': 'Café'},
        ],
      ),
    );
    await tester.pumpAndSettle();
    expect(platform.advertised, {'id': 'LAPTOPDEVICEID01', 'v': '1'});
    expect(platform.browsing, isTrue);
    go.setSyncStatus(syncStatusJson(wifi: 'Café'));
    await tester.pumpAndSettle();
    expect(platform.advertised, isNull);
    expect(platform.browsing, isFalse);
  });

  testWidgets('the Wi-Fi name is read only with the permission', (
    tester,
  ) async {
    final go = FakeGo()..syncStatus = syncStatusJson();
    final platform = FakeSyncPlatform(needsLocation: true, granted: false);
    final model = await started(tester, platform);
    // Without the permission the network is known by its subnet, and
    // nothing asks for it.
    expect(go.syncCalls, contains('SetNetwork wifi  192.168.1.0/24'));
    expect(platform.calls, isNot(contains('requestLocation')));

    // Trusting asks; granted, the network is trusted by its name.
    await model.trustNetwork();
    expect(platform.calls, contains('requestLocation'));
    expect(go.syncCalls.sublist(go.syncCalls.length - 2), [
      'SetNetwork wifi Home 192.168.1.0/24',
      'TrustNetwork',
    ]);
  });

  testWidgets('a refused permission trusts the network by its subnet', (
    tester,
  ) async {
    final go = FakeGo()..syncStatus = syncStatusJson();
    final platform = FakeSyncPlatform(
      needsLocation: true,
      granted: false,
      grantLocation: false,
    );
    final model = await started(tester, platform);
    await model.trustNetwork();
    expect(platform.calls, contains('requestLocation'));
    expect(go.syncCalls.sublist(go.syncCalls.length - 2), [
      'SetNetwork wifi  192.168.1.0/24',
      'TrustNetwork',
    ]);
    final trusted = (go.syncStatus['trusted'] as List).single as Map;
    expect(trusted['subnet'], '192.168.1.0/24');
    expect(trusted['wifi'], isNull);
  });

  testWidgets('pairing mode announces pair=1', (tester) async {
    final go = FakeGo()
      ..syncStatus = syncStatusJson(
        trusted: const [
          {'id': 'wifi:Home', 'wifi': 'Home'},
        ],
      );
    final platform = FakeSyncPlatform();
    await started(tester, platform);
    expect(platform.advertised, {'id': 'LAPTOPDEVICEID01', 'v': '1'});
    go.setSyncStatus({...go.syncStatus, 'pairing': true});
    await tester.pumpAndSettle();
    expect(platform.advertised, {
      'id': 'LAPTOPDEVICEID01',
      'v': '1',
      'pair': '1',
    });
    expect(
      platform.calls.where((c) => c.startsWith('advertise Laptop 47812')),
      hasLength(2),
    );
  });

  testWidgets('found devices are passed to Go', (tester) async {
    final go = FakeGo()
      ..syncStatus = syncStatusJson(
        trusted: const [
          {'id': 'wifi:Home', 'wifi': 'Home'},
        ],
      );
    final platform = FakeSyncPlatform();
    await started(tester, platform);
    platform.find(
      const FoundService(
        name: 'Phone',
        host: '192.168.1.30',
        port: 47812,
        attributes: {'id': 'PHONEDEVICEID001', 'v': '1', 'pair': '1'},
      ),
    );
    // This device's own announcement is left out.
    platform.find(
      const FoundService(
        name: 'Laptop',
        host: '192.168.1.20',
        port: 47812,
        attributes: {'id': 'LAPTOPDEVICEID01', 'v': '1'},
      ),
    );
    platform.lose('Phone');
    await tester.pumpAndSettle();
    expect(
      go.syncCalls,
      contains('Seen PHONEDEVICEID001 Phone 192.168.1.30 47812 true'),
    );
    expect(go.syncCalls.where((c) => c.startsWith('Seen')), hasLength(1));
    expect(go.syncCalls, contains('Lost PHONEDEVICEID001'));
  });

  testWidgets('the app lifecycle goes to Go', (tester) async {
    final go = FakeGo()..syncStatus = syncStatusJson();
    await started(tester, FakeSyncPlatform());
    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.inactive);
    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.hidden);
    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.paused);
    await tester.pumpAndSettle();
    expect(go.syncCalls.where((c) => c == 'Background'), hasLength(1));
    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.hidden);
    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.inactive);
    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.resumed);
    await tester.pumpAndSettle();
    expect(go.syncCalls.where((c) => c == 'Foreground'), hasLength(1));
  });

  testWidgets('on the web nothing runs', (tester) async {
    final go = FakeGo();
    final platform = FakeSyncPlatform();
    final model = SyncModel(isWeb: true, platform: platform)..start();
    addTearDown(model.dispose);
    await tester.pumpAndSettle();
    expect(go.syncCalls, isEmpty);
    expect(platform.calls, isEmpty);
    expect(model.supported, isFalse);
  });

  test('subnets', () {
    expect(subnetOf('192.168.1.20', '255.255.255.0'), '192.168.1.0/24');
    expect(subnetOf('10.1.2.3', '255.255.0.0'), '10.1.0.0/16');
    expect(subnetOf('fe80::1', '255.255.255.0'), '');
    expect(subnetOf(null, null), '');
  });
}
