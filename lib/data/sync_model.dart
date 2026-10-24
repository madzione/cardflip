import 'dart:async';
import 'dart:io' show Platform;

import 'package:bonsoir/bonsoir.dart';
import 'package:connectivity_plus/connectivity_plus.dart';
import 'package:device_info_plus/device_info_plus.dart';
import 'package:flutter/widgets.dart';
import 'package:network_info_plus/network_info_plus.dart';
import 'package:permission_handler/permission_handler.dart';

import 'package:cardflip/golang/golang.dart' as golang;

/// The DNS-SD service type devices announce themselves with (sync design
/// decision 11).
const syncServiceType = '_cardflip._tcp';

/// The version of the announcement's TXT record.
const syncTxtVersion = '1';

/// The network the device is on, as the platform reports it.
class NetworkState {
  const NetworkState(this.kind, {this.wifi = '', this.subnet = ''});

  /// "wifi", "ethernet", "cellular", "other" or "none" (see Go's
  /// SetNetwork).
  final String kind;

  /// The Wi-Fi name, when the app may read it.
  final String wifi;

  /// The subnet in CIDR notation, when known; Go finds it otherwise.
  final String subnet;
}

/// A service found on the network.
class FoundService {
  const FoundService({
    required this.name,
    required this.host,
    required this.port,
    required this.attributes,
  });

  final String name;
  final String host;
  final int port;
  final Map<String, String> attributes;
}

/// What device sync needs of the platform: the network, the location
/// permission that reading the Wi-Fi name needs, the device's name, and
/// DNS-SD. Tests fake it.
abstract class SyncPlatform {
  /// Network changes; the current network first.
  Stream<void> get networkChanges;

  /// The network now. With [readWifiName] the Wi-Fi name is read, which
  /// needs the location permission where [needsLocation].
  Future<NetworkState> network({required bool readWifiName});

  /// Whether reading the Wi-Fi name needs the location permission.
  bool get needsLocation;

  /// Whether the location permission is granted.
  Future<bool> locationGranted();

  /// Asks for the location permission and reports whether it is granted.
  Future<bool> requestLocation();

  /// The device's name to start with: the phone's model on Android and
  /// iOS, empty on desktops (Go takes the host name).
  Future<String> deviceName();

  /// Announces this device until [stopAdvertising].
  Future<void> advertise(String name, int port, Map<String, String> txt);
  Future<void> stopAdvertising();

  /// Looks for devices: [onFound] gets each one once it is resolved, and
  /// [onLost] the name of each that leaves.
  Future<void> browse(
    void Function(FoundService service) onFound,
    void Function(String name) onLost,
  );
  Future<void> stopBrowsing();
}

/// Device sync for the app: it tells Go the network and the app's
/// lifecycle, announces the device and looks for others while Go listens,
/// and follows Go's sync status. It does nothing on the web.
class SyncModel extends ChangeNotifier {
  SyncModel({required this.isWeb, SyncPlatform? platform})
    : platform = platform ?? PluginSyncPlatform();

  final bool isWeb;
  final SyncPlatform platform;

  /// The latest status from Go; null before the first one.
  golang.SyncStatus? status;

  bool get supported => !isWeb && (status?.supported ?? true);

  StreamSubscription<golang.SyncStatus>? _statusSub;
  StreamSubscription<void>? _networkSub;
  AppLifecycleListener? _lifecycle;
  bool _started = false;

  /// What is announced now, as "name|port|txt", or null.
  String? _advertised;
  bool _browsing = false;

  /// The names of the services found, to their device IDs.
  final _found = <String, String>{};

  /// Starts following the network, the lifecycle and Go's status.
  void start() {
    if (isWeb || _started) return;
    _started = true;
    _statusSub = golang.sync.statusChanges().listen(_onStatus, onError: (_) {});
    _networkSub = platform.networkChanges.listen((_) => reportNetwork());
    _lifecycle = AppLifecycleListener(onStateChange: _onLifecycle);
    reportNetwork();
  }

  /// Whether Go was last told the app is in the foreground. Inactive
  /// (a dialog of the system, the app switcher) changes nothing.
  bool _foreground = true;

  void _onLifecycle(AppLifecycleState state) {
    switch (state) {
      case AppLifecycleState.resumed when !_foreground:
        _foreground = true;
        golang.sync.foreground().catchError((Object _) {});
      case AppLifecycleState.hidden || AppLifecycleState.paused
          when _foreground:
        _foreground = false;
        golang.sync.background().catchError((Object _) {});
      default:
    }
  }

  /// Tells Go the current network. The Wi-Fi name is read only if that
  /// needs no permission or it was granted already.
  Future<void> reportNetwork() async {
    if (isWeb) return;
    final read = !platform.needsLocation || await platform.locationGranted();
    final net = await platform.network(readWifiName: read);
    await golang.sync.setNetwork(net.kind, net.wifi, net.subnet);
  }

  Future<void> _onStatus(golang.SyncStatus st) async {
    status = st;
    notifyListeners();
    await _discover(st);
  }

  /// Announces the device and looks for others exactly while Go listens on
  /// a trusted network (device-pairing spec: Trusted networks only).
  Future<void> _discover(golang.SyncStatus st) async {
    final port = st.port ?? 0;
    final on = st.ready && st.listening && port > 0;
    if (!on) {
      if (_advertised != null) {
        _advertised = null;
        await platform.stopAdvertising();
      }
      if (_browsing) {
        _browsing = false;
        _found.clear();
        await platform.stopBrowsing();
      }
      return;
    }
    final txt = {
      'id': st.deviceID ?? '',
      'v': syncTxtVersion,
      if (st.pairing) 'pair': '1',
    };
    final name = st.deviceName ?? '';
    final key = '$name|$port|${txt.entries.join(',')}';
    if (_advertised != key) {
      if (_advertised != null) await platform.stopAdvertising();
      _advertised = key;
      await platform.advertise(name, port, txt);
    }
    if (!_browsing) {
      _browsing = true;
      await platform.browse(_onFound, _onLost);
    }
  }

  void _onFound(FoundService s) {
    final id = s.attributes['id'];
    if (id == null || id.isEmpty || id == status?.deviceID) return;
    _found[s.name] = id;
    golang.sync
        .seen(id, s.name, s.host, s.port, s.attributes['pair'] == '1')
        .catchError((Object _) {});
  }

  void _onLost(String name) {
    final id = _found.remove(name);
    if (id != null) golang.sync.lost(id).catchError((Object _) {});
  }

  /// Sets up sync on this device, when the Sync section opens.
  Future<void> init() async {
    status = await golang.sync.init(await platform.deviceName());
    notifyListeners();
  }

  /// Trusts the current network. Where reading the Wi-Fi name needs the
  /// location permission, it is asked for now; refused, the network is
  /// trusted by its subnet (device-pairing spec: Wi-Fi name permission).
  Future<void> trustNetwork() async {
    var read = true;
    if (platform.needsLocation && !await platform.locationGranted()) {
      read = await platform.requestLocation();
    }
    final net = await platform.network(readWifiName: read);
    await golang.sync.setNetwork(net.kind, net.wifi, net.subnet);
    await golang.sync.trustNetwork();
  }

  @override
  void dispose() {
    _statusSub?.cancel();
    _networkSub?.cancel();
    _lifecycle?.dispose();
    if (_advertised != null) platform.stopAdvertising();
    if (_browsing) platform.stopBrowsing();
    super.dispose();
  }
}

/// The subnet of [ip] with [mask] in CIDR notation, or '' if either is not
/// an IPv4 address.
String subnetOf(String? ip, String? mask) {
  List<int>? parse(String? s) {
    final parts = s?.split('.');
    if (parts == null || parts.length != 4) return null;
    final out = parts.map(int.tryParse).toList();
    if (out.any((p) => p == null || p < 0 || p > 255)) return null;
    return out.cast<int>();
  }

  final a = parse(ip), m = parse(mask);
  if (a == null || m == null) return '';
  var bits = 0;
  for (final b in m) {
    bits += b.toRadixString(2).replaceAll('0', '').length;
  }
  return '${[for (var i = 0; i < 4; i++) a[i] & m[i]].join('.')}/$bits';
}

/// [SyncPlatform] on the platform's own services, through connectivity_plus,
/// network_info_plus, permission_handler, device_info_plus and bonsoir (NSD
/// on Android, Bonjour on Apple platforms, DNS-SD on Windows, Avahi on
/// Linux).
class PluginSyncPlatform implements SyncPlatform {
  final _connectivity = Connectivity();
  final _info = NetworkInfo();
  BonsoirBroadcast? _broadcast;
  BonsoirDiscovery? _discovery;
  StreamSubscription<BonsoirDiscoveryEvent>? _discoverySub;

  @override
  Stream<void> get networkChanges => _connectivity.onConnectivityChanged;

  @override
  Future<NetworkState> network({required bool readWifiName}) async {
    final results = await _connectivity.checkConnectivity();
    final kind = results.contains(ConnectivityResult.wifi)
        ? 'wifi'
        : results.contains(ConnectivityResult.ethernet)
        ? 'ethernet'
        : results.contains(ConnectivityResult.mobile)
        ? 'cellular'
        : results.contains(ConnectivityResult.none) || results.isEmpty
        ? 'none'
        : 'other';
    if (kind != 'wifi') return NetworkState(kind);
    String wifi = '', subnet = '';
    try {
      if (readWifiName) {
        // iOS and Android quote the name.
        wifi = (await _info.getWifiName() ?? '').replaceAll('"', '');
      }
      subnet = subnetOf(await _info.getWifiIP(), await _info.getWifiSubmask());
    } catch (_) {}
    return NetworkState(kind, wifi: wifi, subnet: subnet);
  }

  @override
  bool get needsLocation => Platform.isAndroid || Platform.isIOS;

  @override
  Future<bool> locationGranted() async =>
      needsLocation && await Permission.locationWhenInUse.isGranted;

  @override
  Future<bool> requestLocation() async =>
      (await Permission.locationWhenInUse.request()).isGranted;

  @override
  Future<String> deviceName() async {
    final info = DeviceInfoPlugin();
    try {
      if (Platform.isAndroid) return (await info.androidInfo).model;
      if (Platform.isIOS) return (await info.iosInfo).modelName;
    } catch (_) {}
    return '';
  }

  @override
  Future<void> advertise(String name, int port, Map<String, String> txt) async {
    final broadcast = BonsoirBroadcast(
      service: BonsoirService(
        name: name,
        type: syncServiceType,
        port: port,
        attributes: txt,
      ),
    );
    _broadcast = broadcast;
    try {
      await broadcast.initialize();
      await broadcast.start();
    } catch (_) {
      // Pairing by address still works.
    }
  }

  @override
  Future<void> stopAdvertising() async {
    final b = _broadcast;
    _broadcast = null;
    await b?.stop().catchError((Object _) {});
  }

  @override
  Future<void> browse(
    void Function(FoundService service) onFound,
    void Function(String name) onLost,
  ) async {
    final discovery = BonsoirDiscovery(type: syncServiceType);
    _discovery = discovery;
    try {
      await discovery.initialize();
      _discoverySub = discovery.eventStream?.listen((event) {
        switch (event) {
          case BonsoirDiscoveryServiceFoundEvent():
            event.service.resolve(discovery.serviceResolver);
          case BonsoirDiscoveryServiceResolvedEvent():
          case BonsoirDiscoveryServiceUpdatedEvent():
            final s = event.service!;
            final host = s.hostAddresses.firstWhere(
              (a) => !a.contains(':'),
              orElse: () => s.hostAddress ?? '',
            );
            if (host.isNotEmpty) {
              onFound(
                FoundService(
                  name: s.name,
                  host: host,
                  port: s.port,
                  attributes: s.attributes,
                ),
              );
            }
          case BonsoirDiscoveryServiceLostEvent():
            onLost(event.service.name);
          default:
        }
      });
      await discovery.start();
    } catch (_) {
      // Pairing by address still works.
    }
  }

  @override
  Future<void> stopBrowsing() async {
    await _discoverySub?.cancel();
    _discoverySub = null;
    final d = _discovery;
    _discovery = null;
    await d?.stop().catchError((Object _) {});
  }
}
