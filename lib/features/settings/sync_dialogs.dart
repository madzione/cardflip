import 'dart:async';

import 'package:flutter/material.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/app/errors.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';

/// How long Go waits for an answer during pairing (syncnet.PairingTimeout);
/// a dialog still open then is closed.
const pairingTimeout = Duration(minutes: 2);

/// What Go asks the user during device sync, as dialogs over the app's
/// navigator: pairing requests, the pairing code and the first-sync choice.
class SyncDialogs implements golang.SyncUI {
  SyncDialogs(this._context);

  /// The navigator's context, or null while the app has none.
  final BuildContext? Function() _context;

  /// The pairing dialog that is open, which ends with the pairing.
  Route<Object?>? _pairingDialog;

  /// Shows a dialog built by [builder] and returns its result, or null when
  /// it was closed without one: after [timeout], or because pairing ended
  /// if it is a [pairing] dialog.
  Future<T?> _show<T>(
    WidgetBuilder builder, {
    Duration? timeout,
    bool pairing = true,
  }) async {
    final context = _context();
    if (context == null || !context.mounted) return null;
    final navigator = Navigator.of(context, rootNavigator: true);
    final route = DialogRoute<T>(
      context: context,
      builder: builder,
      barrierDismissible: false,
      animationStyle: popupAnimation(context),
    );
    if (pairing) _pairingDialog = route;
    final timer = timeout == null
        ? null
        : Timer(timeout, () {
            if (route.isActive) navigator.removeRoute(route);
          });
    try {
      return await navigator.push(route);
    } finally {
      timer?.cancel();
      if (_pairingDialog == route) _pairingDialog = null;
    }
  }

  @override
  Future<bool> pairingRequest(String deviceName) async {
    final accepted = await _show<bool>(
      (context) => AlertDialog(
        scrollable: true,
        icon: const Icon(Icons.devices_outlined),
        title: Text(context.l10n.pairingRequestTitle(deviceName)),
        content: Text(context.l10n.pairingRequestBody(deviceName)),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(context.l10n.decline),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text(context.l10n.accept),
          ),
        ],
      ),
      timeout: pairingTimeout,
    );
    return accepted ?? false;
  }

  @override
  Future<bool> confirmPairing(String deviceName, String code) async {
    final shown = code.length == 6
        ? '${code.substring(0, 3)} ${code.substring(3)}'
        : code;
    final match = await _show<bool>(
      (context) => AlertDialog(
        scrollable: true,
        title: Text(context.l10n.confirmCodeTitle),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(context.l10n.confirmCodeBody(deviceName)),
            const SizedBox(height: 16),
            Text(
              shown,
              key: const Key('pairing-code'),
              style: Theme.of(context).textTheme.displaySmall,
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(context.l10n.codesDiffer),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text(context.l10n.codesMatch),
          ),
        ],
      ),
      timeout: pairingTimeout,
    );
    if (match ?? false) showInfo(appL10n.waitingForConfirmation(deviceName));
    return match ?? false;
  }

  @override
  void pairingFinished(String deviceName, bool paired) {
    // The other side declined or gave up: its code dialog is pointless.
    final open = _pairingDialog;
    if (open != null && open.isActive) open.navigator?.removeRoute(open);
    showInfo(
      paired
          ? appL10n.pairedWith(deviceName)
          : appL10n.pairingFailed(deviceName),
    );
  }

  /// Asks how to combine the open profile with the linked one (device-sync
  /// spec: First sync of a link): "merge", "replaceOther", "replaceThis",
  /// or "" to decide later.
  @override
  Future<String> chooseFirstSync(golang.LinkInfo link) async {
    final device = link.deviceName;
    final choice = await _show<String>((context) {
      final l = context.l10n;
      Widget option(String value, String title, String help) => ListTile(
        key: Key('first-sync-$value'),
        contentPadding: EdgeInsets.zero,
        title: Text(title),
        subtitle: Text(help),
        onTap: () => Navigator.pop(context, value),
      );
      return AlertDialog(
        icon: const Icon(Icons.sync),
        title: Text(l.firstSyncTitle(device)),
        content: SizedBox(
          width: 480,
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Text(l.firstSyncBody(link.remoteName, device)),
                const SizedBox(height: 8),
                option('merge', l.firstSyncMerge, l.firstSyncMergeHelp),
                option(
                  'replaceOther',
                  l.firstSyncReplaceOther(device),
                  l.firstSyncReplaceOtherHelp(device),
                ),
                option(
                  'replaceThis',
                  l.firstSyncReplaceThis,
                  l.firstSyncReplaceThisHelp(link.remoteName, device),
                ),
              ],
            ),
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, ''),
            child: Text(l.decideLater),
          ),
        ],
      );
    }, pairing: false);
    return choice ?? '';
  }
}
