import 'dart:async';

import 'package:flutter/material.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/app/errors.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/data/sync_model.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/go_text.dart';
import 'package:cardflip/l10n/l10n.dart';
import 'package:cardflip/widgets/dialogs.dart';
import 'package:cardflip/widgets/time_format.dart';

/// How long "Pair a device" keeps this device visible
/// (syncnet.PairingWindow).
const pairingWindow = Duration(minutes: 5);

/// The port sync listens on unless it is taken; an address entered without
/// a port uses it.
const syncPort = 47812;

/// A trusted network as the Sync section lists it: its Wi-Fi name, or its
/// subnet (device-pairing spec: Trust a network).
String trustedNetworkLabel(golang.TrustedNetwork n, AppLocalizations l) {
  if (n.wiFi case final wifi? when wifi.isNotEmpty) return wifi;
  final subnet = n.subnet ?? '';
  return n.wired ?? false
      ? l.syncWiredNetwork(subnet)
      : l.syncSubnetNetwork(subnet);
}

/// The network the device is on, named like a trusted one.
String currentNetworkLabel(golang.NetworkInfo n, AppLocalizations l) {
  if (n.kind == 'cellular') return l.syncMobileNetwork;
  if (n.wiFi case final wifi? when wifi.isNotEmpty) return wifi;
  final subnet = n.subnet ?? '';
  if (n.kind == 'none' || subnet.isEmpty) return l.syncNoNetwork;
  return n.kind == 'ethernet'
      ? l.syncWiredNetwork(subnet)
      : l.syncSubnetNetwork(subnet);
}

/// The Sync section of Settings: this device's name, the current and the
/// trusted networks, the paired devices, "Pair a device" and the open
/// profile's links. Opening it gives the device its sync identity. Shown
/// on every platform but the web, below [heading].
class SyncSection extends StatefulWidget {
  const SyncSection({super.key, required this.heading});

  final Widget heading;

  @override
  State<SyncSection> createState() => _SyncSectionState();
}

class _SyncSectionState extends State<SyncSection> {
  Object? _error;

  /// Indents the entries of a list below its tile.
  static const _indent = EdgeInsetsDirectional.only(start: 72, end: 16);

  @override
  void initState() {
    super.initState();
    _init();
  }

  Future<void> _init() async {
    try {
      await AppScope.read(context).sync.init();
    } catch (e) {
      if (mounted) setState(() => _error = e);
    }
  }

  Future<void> _rename(String name) => askAndSave(
    context,
    title: context.l10n.syncRenameDevice,
    initial: name,
    label: context.l10n.syncDeviceName,
    helper: context.l10n.syncDeviceNameHelper,
    fieldKey: const Key('device-name-field'),
    save: golang.sync.setDeviceName,
  );

  /// Explains what trusting means, then trusts the current network
  /// (asking for the location permission where the Wi-Fi name needs it).
  Future<void> _trust() async {
    final sync = AppScope.read(context).sync;
    final l = context.l10n;
    final ok = await confirm(
      context,
      title: l.trustNetworkTitle,
      message: sync.platform.needsLocation
          ? '${l.trustNetworkBody}\n\n${l.trustNetworkLocation}'
          : l.trustNetworkBody,
      action: l.trust,
    );
    if (ok) await guard(sync.trustNetwork);
  }

  Future<void> _unpair(golang.PairedDeviceInfo device) async {
    final ok = await confirm(
      context,
      title: context.l10n.unpairTitle(device.name),
      message: context.l10n.unpairBody(device.name),
      action: context.l10n.unpair,
      destructive: true,
    );
    if (ok) await guard(() => golang.sync.unpair(device.id));
  }

  Future<void> _unlink(golang.LinkInfo link) async {
    final ok = await confirm(
      context,
      title: context.l10n.unlinkTitle(link.deviceName),
      message: context.l10n.unlinkBody(link.remoteName, link.deviceName),
      action: context.l10n.unlink,
    );
    if (ok) await guard(() => golang.sync.unlink(link.id));
  }

  /// Links the open profile with a profile on one of [devices], or with a
  /// new one there (device-sync spec: Profile links).
  Future<void> _link(List<golang.PairedDeviceInfo> devices) async {
    final l = context.l10n;
    final device = devices.length == 1
        ? devices.single
        : await showDialog<golang.PairedDeviceInfo>(
            context: context,
            animationStyle: popupAnimation(context),
            builder: (context) => SimpleDialog(
              title: Text(l.chooseDevice),
              children: [
                for (final d in devices)
                  SimpleDialogOption(
                    onPressed: () => Navigator.pop(context, d),
                    child: Text(d.name),
                  ),
              ],
            ),
          );
    if (device == null || !mounted) return;
    var profiles = const <golang.Profile>[];
    final listed = await guard(
      () async => profiles = await golang.sync.remoteProfiles(device.id),
    );
    if (!listed || !mounted) return;
    final name = AppScope.read(context).profiles.currentName ?? '';
    final remote = await showDialog<String>(
      context: context,
      animationStyle: popupAnimation(context),
      builder: (context) => SimpleDialog(
        title: Text(l.chooseRemoteProfile(device.name)),
        children: [
          for (final p in profiles)
            SimpleDialogOption(
              key: ValueKey('remote-${p.id}'),
              onPressed: () => Navigator.pop(context, p.id),
              child: Text(p.name),
            ),
          SimpleDialogOption(
            key: const Key('remote-new'),
            onPressed: () => Navigator.pop(context, ''),
            child: Text(l.newProfileNamed(name)),
          ),
        ],
      ),
    );
    if (remote == null) return;
    await guard(() => golang.sync.link(device.id, remote));
  }

  static String _lastSynced(DateTime? t, AppLocalizations l) => t == null
      ? l.neverSynced
      : l.lastSyncedAt(formatDateTime(t.toLocal(), l10n: l));

  static String _linkState(golang.LinkInfo link, AppLocalizations l) {
    if (link.needsChoice) return l.linkNeedsChoice;
    if (link.error case final e? when e.isNotEmpty) return goText(e, l);
    return _lastSynced(link.lastSynced, l);
  }

  static IconData _networkIcon(String kind) => switch (kind) {
    'wifi' => Icons.wifi,
    'ethernet' => Icons.settings_ethernet,
    'cellular' => Icons.signal_cellular_alt,
    'none' => Icons.signal_wifi_off_outlined,
    _ => Icons.lan_outlined,
  };

  @override
  Widget build(BuildContext context) {
    final sync = AppScope.of(context).sync;
    return ListenableBuilder(
      listenable: sync,
      builder: (context, _) {
        if (!sync.supported) return const SizedBox.shrink();
        final l = context.l10n;
        final st = sync.status;
        final error = _error;
        if (error != null || st == null || !st.ready) {
          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              widget.heading,
              if (error != null)
                ListTile(
                  leading: const Icon(Icons.error_outline),
                  title: Text(errorMessage(error)),
                ),
            ],
          );
        }
        final net = st.network;
        final trust = net.kind == 'cellular'
            ? l.syncMobileNeverTrusted
            : net.trusted
            ? l.syncNetworkTrusted
            : l.syncNetworkNotTrusted;
        final unlinked = [
          for (final d in st.paired)
            if (!st.links.any((link) => link.device == d.id)) d,
        ];
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            widget.heading,
            ListTile(
              key: const Key('sync-device-name'),
              leading: const Icon(Icons.badge_outlined),
              title: Text(l.syncDeviceName),
              subtitle: Text(st.deviceName ?? ''),
              trailing: const Icon(Icons.edit_outlined),
              onTap: () => _rename(st.deviceName ?? ''),
            ),
            // The button goes below the text: beside it, it leaves the
            // text a word or two per line on a phone.
            ListTile(
              key: const Key('sync-network'),
              leading: Icon(_networkIcon(net.kind)),
              title: Text(l.syncCurrentNetwork),
              subtitle: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text([currentNetworkLabel(net, l), trust].join('\n')),
                  if ((net.id ?? '').isNotEmpty && net.kind != 'cellular')
                    Padding(
                      padding: const EdgeInsets.only(top: 8),
                      child: net.trusted
                          ? OutlinedButton(
                              key: const Key('untrust-network'),
                              onPressed: () => guard(
                                () => golang.sync.untrustNetwork(net.id!),
                              ),
                              child: Text(l.stopTrusting),
                            )
                          : FilledButton.tonal(
                              key: const Key('trust-network'),
                              onPressed: _trust,
                              child: Text(l.trustThisNetwork),
                            ),
                    ),
                ],
              ),
              isThreeLine: true,
            ),
            ListTile(
              leading: const Icon(Icons.verified_user_outlined),
              title: Text(l.trustedNetworks),
              subtitle: st.trusted.isEmpty ? Text(l.noTrustedNetworks) : null,
            ),
            for (final n in st.trusted)
              ListTile(
                key: ValueKey('trusted-${n.id}'),
                contentPadding: _indent,
                title: Text(trustedNetworkLabel(n, l)),
                trailing: IconButton(
                  tooltip: l.stopTrusting,
                  icon: const Icon(Icons.remove_circle_outline),
                  onPressed: () =>
                      guard(() => golang.sync.untrustNetwork(n.id)),
                ),
              ),
            ListTile(
              leading: const Icon(Icons.devices_other_outlined),
              title: Text(l.pairedDevices),
              subtitle: st.paired.isEmpty ? Text(l.noPairedDevices) : null,
            ),
            for (final d in st.paired)
              ListTile(
                key: ValueKey('paired-${d.id}'),
                contentPadding: _indent,
                title: Text(d.name),
                subtitle: Text(
                  '${d.reachable ? l.deviceReachable : l.deviceNotReachable}'
                  ' · ${_lastSynced(d.lastSynced, l)}',
                ),
                trailing: TextButton(
                  onPressed: () => _unpair(d),
                  child: Text(l.unpair),
                ),
              ),
            ListTile(
              key: const Key('pair-device'),
              leading: const Icon(Icons.add_circle_outline),
              title: Text(l.pairDevice),
              subtitle: Text(
                st.listening ? l.pairDeviceHelp : l.goSyncUntrusted,
              ),
              enabled: st.listening,
              onTap: () => Navigator.of(context).push(
                MaterialPageRoute<void>(builder: (_) => const SyncPairScreen()),
              ),
            ),
            ListTile(
              leading: const Icon(Icons.link),
              title: Text(l.profileLinks),
              subtitle: st.links.isEmpty ? Text(l.noProfileLinks) : null,
            ),
            for (final link in st.links)
              ListTile(
                key: ValueKey('link-${link.id}'),
                contentPadding: _indent,
                title: Text(l.linkTarget(link.remoteName, link.deviceName)),
                subtitle: Text(_linkState(link, l)),
                trailing: TextButton(
                  onPressed: () => _unlink(link),
                  child: Text(l.unlink),
                ),
              ),
            ListTile(
              key: const Key('link-profile'),
              leading: const Icon(Icons.add_link),
              title: Text(l.linkProfile),
              subtitle: Text(l.linkProfileHelp),
              enabled: unlinked.isNotEmpty,
              onTap: () => _link(unlinked),
            ),
          ],
        );
      },
    );
  }
}

/// "Pair a device": keeps this device visible for pairing for 5 minutes
/// or until it is closed, and lists the devices visible for pairing, this
/// device's address and a field for another device's address
/// (device-pairing spec: Pair a device, Pair by address).
class SyncPairScreen extends StatefulWidget {
  const SyncPairScreen({super.key});

  @override
  State<SyncPairScreen> createState() => _SyncPairScreenState();
}

class _SyncPairScreenState extends State<SyncPairScreen> {
  late final SyncModel _sync = AppScope.read(context).sync;
  final _address = TextEditingController();
  DateTime? _visibleUntil;
  Timer? _tick;
  bool _busy = false;

  @override
  void initState() {
    super.initState();
    _startPairing();
  }

  @override
  void dispose() {
    _tick?.cancel();
    _address.dispose();
    golang.sync.stopPairing().catchError((Object _) {});
    super.dispose();
  }

  Future<void> _startPairing() async {
    if (!await guard(golang.sync.startPairing) || !mounted) return;
    setState(() => _visibleUntil = DateTime.now().add(pairingWindow));
    // Counts down the time left.
    _tick?.cancel();
    _tick = Timer.periodic(const Duration(seconds: 1), (t) {
      if (!mounted) return;
      if (_left() == Duration.zero) t.cancel();
      setState(() {});
    });
  }

  Duration _left() {
    final until = _visibleUntil;
    if (until == null) return Duration.zero;
    final left = until.difference(DateTime.now());
    return left.isNegative ? Duration.zero : left;
  }

  Future<void> _pair(String host, int port) async {
    setState(() => _busy = true);
    try {
      await golang.sync.pairWith(host, port);
    } catch (e) {
      // Unconfirmed codes are reported by SyncDialogs.pairingFinished.
      if (e is! golang.GolangException ||
          goCode(e.message) != 'syncPairFailed') {
        showError(e);
      }
    }
    if (mounted) setState(() => _busy = false);
  }

  /// Pairs with the address typed in, "host:port" or just the host.
  Future<void> _pairByAddress() {
    final uri = Uri.tryParse('//${_address.text.trim()}');
    final host = uri?.host ?? '';
    return _pair(host, uri != null && uri.hasPort ? uri.port : syncPort);
  }

  @override
  Widget build(BuildContext context) {
    final l = context.l10n;
    final theme = Theme.of(context);
    return Scaffold(
      appBar: AppBar(title: Text(l.pairDevice)),
      body: ListenableBuilder(
        listenable: _sync,
        builder: (context, _) {
          final st = _sync.status;
          if (st == null) return const SizedBox.shrink();
          final left = _left();
          final visible = st.pairing && left > Duration.zero;
          final seconds = left.inSeconds % 60;
          final address = st.address == null
              ? '${st.port ?? ''}'
              : '${st.address}:${st.port}';
          return ListView(
            children: [
              if (_busy) const LinearProgressIndicator(),
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
                child: Text(l.pairDeviceIntro),
              ),
              if (!st.listening)
                ListTile(
                  leading: const Icon(Icons.error_outline),
                  title: Text(l.goSyncUntrusted),
                )
              else ...[
                ListTile(
                  key: const Key('pairing-visibility'),
                  leading: Icon(
                    visible
                        ? Icons.visibility_outlined
                        : Icons.visibility_off_outlined,
                  ),
                  title: Text(
                    visible
                        ? l.visibleForPairing(
                            '${left.inMinutes}:'
                            '${seconds.toString().padLeft(2, '0')}',
                          )
                        : l.notVisibleForPairing,
                  ),
                  trailing: visible
                      ? null
                      : TextButton(
                          onPressed: _startPairing,
                          child: Text(l.makeVisible),
                        ),
                ),
                ListTile(
                  key: const Key('own-address'),
                  leading: const Icon(Icons.lan_outlined),
                  title: Text(l.thisDeviceAddress),
                  subtitle: SelectableText(address),
                ),
                Padding(
                  padding: const EdgeInsets.fromLTRB(16, 16, 16, 4),
                  child: Text(
                    l.devicesNearby,
                    style: theme.textTheme.titleSmall?.copyWith(
                      color: theme.colorScheme.primary,
                    ),
                  ),
                ),
                if (st.discovered.isEmpty)
                  ListTile(
                    leading: const Icon(Icons.search),
                    title: Text(l.lookingForDevices),
                  ),
                for (final d in st.discovered)
                  ListTile(
                    key: ValueKey('discovered-${d.id}'),
                    leading: const Icon(Icons.devices_outlined),
                    title: Text(d.name),
                    subtitle: Text('${d.host}:${d.port}'),
                    enabled: !_busy,
                    onTap: () => _pair(d.host, d.port),
                  ),
                Padding(
                  padding: const EdgeInsets.fromLTRB(16, 16, 16, 16),
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Expanded(
                        child: TextField(
                          key: const Key('pair-address'),
                          controller: _address,
                          keyboardType: TextInputType.url,
                          decoration: InputDecoration(
                            labelText: l.enterAddress,
                            helperText: l.enterAddressHelper,
                          ),
                          onSubmitted: (_) => _pairByAddress(),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Padding(
                        padding: const EdgeInsets.only(top: 8),
                        child: FilledButton(
                          key: const Key('pair-by-address'),
                          onPressed: _busy ? null : _pairByAddress,
                          child: Text(l.pair),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ],
          );
        },
      ),
    );
  }
}
