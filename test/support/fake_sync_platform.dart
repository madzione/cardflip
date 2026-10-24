import 'dart:async';

import 'package:cardflip/data/sync_model.dart';

/// A [SyncPlatform] for tests: the network is [current] (whose Wi-Fi name
/// is only given when it may be read), the location permission is
/// [locationGranted] and requesting it answers [grantLocation]. Every
/// call is recorded in [calls]; [advertised] is what is announced now.
class FakeSyncPlatform implements SyncPlatform {
  FakeSyncPlatform({
    this.current = const NetworkState(
      'wifi',
      wifi: 'Home',
      subnet: '192.168.1.0/24',
    ),
    this.needsLocation = false,
    this.granted = false,
    this.grantLocation = true,
    this.name = 'Pixel 9',
  });

  NetworkState current;
  @override
  bool needsLocation;
  bool granted;
  bool grantLocation;
  String name;

  final calls = <String>[];
  Map<String, String>? advertised;
  bool browsing = false;
  final _changes = StreamController<void>.broadcast();
  void Function(FoundService)? _onFound;
  void Function(String)? _onLost;

  /// Changes the network and reports the change.
  void setNetwork(NetworkState n) {
    current = n;
    _changes.add(null);
  }

  /// A device appears on the network.
  void find(FoundService s) => _onFound?.call(s);

  /// A device leaves the network.
  void lose(String name) => _onLost?.call(name);

  @override
  Stream<void> get networkChanges => _changes.stream;

  @override
  Future<NetworkState> network({required bool readWifiName}) async {
    calls.add('network read=$readWifiName');
    final n = current;
    return readWifiName ? n : NetworkState(n.kind, subnet: n.subnet);
  }

  @override
  Future<bool> locationGranted() async => granted;

  @override
  Future<bool> requestLocation() async {
    calls.add('requestLocation');
    granted = grantLocation;
    return granted;
  }

  @override
  Future<String> deviceName() async => name;

  @override
  Future<void> advertise(String name, int port, Map<String, String> txt) async {
    calls.add('advertise $name $port ${txt.entries.join(',')}');
    advertised = txt;
  }

  @override
  Future<void> stopAdvertising() async {
    calls.add('stopAdvertising');
    advertised = null;
  }

  @override
  Future<void> browse(
    void Function(FoundService) onFound,
    void Function(String) onLost,
  ) async {
    calls.add('browse');
    browsing = true;
    _onFound = onFound;
    _onLost = onLost;
  }

  @override
  Future<void> stopBrowsing() async {
    calls.add('stopBrowsing');
    browsing = false;
    _onFound = null;
    _onLost = null;
  }
}
