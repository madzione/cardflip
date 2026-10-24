// Device sync on Linux, with the real Go core and a headless peer
// (golang/cmd/syncpeer) in another data directory: the user trusts the
// network, pairs with the peer by address, links the profile and merges.
// A note added in the app arrives at the peer, and a review given on the
// peer lowers the app's due count (device-pairing and device-sync specs).
//
// The network is a wired one that only this computer is on, and nothing is
// announced over mDNS: the test pairs by address and sends no traffic to
// the real network.
import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';

import 'package:cardflip/app/bootstrap.dart';
import 'package:cardflip/data/sync_model.dart';
import 'package:cardflip/features/import/pick.dart';
import 'package:cardflip/golang/golang.dart' as golang;

import 'rich_fields.dart';

/// The repository's `golang` directory, from wherever the test runs.
Directory golangDir() {
  final root = Platform.environment['CARDFLIP_ROOT'];
  if (root != null) return Directory('$root/golang');
  var dir = Directory.current.absolute;
  while (!File('${dir.path}/golang/go.mod').existsSync()) {
    if (dir.parent.path == dir.path) {
      throw StateError(
        'golang/go.mod not found above ${Directory.current.path}; '
        'set CARDFLIP_ROOT',
      );
    }
    dir = dir.parent;
  }
  return Directory('${dir.path}/golang');
}

class _NoPicker extends FilePickerService {
  @override
  Future<golang.StageInfo?> pick(List<String> extensions) async => null;

  @override
  Future<bool> save(String fileName, Uint8List bytes) async => false;
}

/// A wired network on the loopback subnet, without mDNS.
class _LoopbackPlatform implements SyncPlatform {
  @override
  Stream<void> get networkChanges => const Stream.empty();

  @override
  Future<NetworkState> network({required bool readWifiName}) async =>
      const NetworkState('ethernet', subnet: '127.0.0.0/8');

  @override
  bool get needsLocation => false;

  @override
  Future<bool> locationGranted() async => false;

  @override
  Future<bool> requestLocation() async => false;

  @override
  Future<String> deviceName() async => 'Cardflip test';

  @override
  Future<void> advertise(
    String name,
    int port,
    Map<String, String> txt,
  ) async {}

  @override
  Future<void> stopAdvertising() async {}

  @override
  Future<void> browse(
    void Function(FoundService service) onFound,
    void Function(String name) onLost,
  ) async {}

  @override
  Future<void> stopBrowsing() async {}
}

/// The headless peer: commands on its standard input, one JSON line per
/// answer on its standard output.
class _Peer {
  _Peer._(this._process) {
    _process.stdout
        .transform(utf8.decoder)
        .transform(const LineSplitter())
        .listen((line) {
          final m = jsonDecode(line) as Map<String, dynamic>;
          final waiting = _waiting;
          _waiting = null;
          waiting == null ? _lines.add(m) : waiting.complete(m);
        });
    _process.stderr
        .transform(utf8.decoder)
        .transform(const LineSplitter())
        .listen((line) => debugPrint(line));
  }

  final Process _process;
  final _lines = <Map<String, dynamic>>[];
  Completer<Map<String, dynamic>>? _waiting;

  /// Builds the peer and starts it in [dataDir].
  static Future<_Peer> start(Directory tmp) async {
    final binary = '${tmp.path}/syncpeer';
    final build = await Process.run('go', [
      'build',
      '-o',
      binary,
      './cmd/syncpeer',
    ], workingDirectory: golangDir().path);
    expect(build.exitCode, 0, reason: '${build.stderr}');
    final process = await Process.start(binary, [
      '-data',
      '${tmp.path}/peer',
      '-name',
      'Peer',
    ]);
    return _Peer._(process);
  }

  Future<Map<String, dynamic>> next() {
    if (_lines.isNotEmpty) return Future.value(_lines.removeAt(0));
    final waiting = _waiting = Completer();
    return waiting.future.timeout(const Duration(seconds: 30));
  }

  Future<Map<String, dynamic>> command(String cmd) {
    _process.stdin.writeln(cmd);
    return next();
  }

  Future<void> quit() async {
    _process.stdin.writeln('quit');
    await _process.exitCode.timeout(
      const Duration(seconds: 10),
      onTimeout: () {
        _process.kill();
        return -1;
      },
    );
  }
}

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('pair with a peer by address, link, merge and sync both ways', (
    tester,
  ) async {
    final dir = await Directory.systemTemp.createTemp('cardflip-sync-');
    addTearDown(() => dir.delete(recursive: true));
    tester.view.physicalSize =
        const Size(1280, 860) * tester.view.devicePixelRatio;
    addTearDown(tester.view.resetPhysicalSize);

    final peer = await _Peer.start(dir);
    addTearDown(peer.quit);
    final ready = await peer.next();
    expect(ready['ready'], isTrue);
    final peerPort = ready['port'] as int;

    // Go answers asynchronously, and syncs take a moment.
    Future<void> idle([int rounds = 3]) async {
      for (var i = 0; i < rounds; i++) {
        await Future<void>.delayed(const Duration(milliseconds: 40));
        await tester.pumpAndSettle();
      }
    }

    /// Waits until [finder] finds something, for at most 20 seconds, and
    /// lets dialogs finish opening. It does not wait for the screen to
    /// settle: progress bars run while Go pairs or syncs.
    Future<void> waitFor(Finder finder) async {
      for (var i = 0; i < 100 && finder.evaluate().isEmpty; i++) {
        await Future<void>.delayed(const Duration(milliseconds: 200));
        await tester.pump();
      }
      expect(finder, findsWidgets);
      for (var i = 0; i < 5; i++) {
        await tester.pump(const Duration(milliseconds: 100));
      }
    }

    Future<void> go(String destination) async {
      await tester.tap(
        find.descendant(
          of: find.byType(NavigationRail),
          matching: find.text(destination),
        ),
      );
      await idle();
    }

    Future<void> tapInSettings(Finder finder) async {
      await tester.scrollUntilVisible(
        finder,
        200,
        scrollable: find
            .descendant(
              of: find.byKey(const Key('settings-list')),
              matching: find.byType(Scrollable),
            )
            .first,
      );
      await idle();
      await tester.tap(finder);
      await idle();
    }

    /// Taps the deck list's sync indicator and waits for its report.
    Future<void> syncNow() async {
      await tester.tap(find.byKey(const Key('sync-indicator')));
      await waitFor(find.textContaining('Synced: '));
    }

    /// The new count of the deck list's first deck.
    Future<String?> newCount() async {
      final counts = find.byKey(
        ValueKey('counts-${(await golang.decks.tree()).first.id}'),
      );
      final texts = tester.widgetList<Text>(
        find.descendant(of: counts, matching: find.byType(Text)),
      );
      return texts.isEmpty ? null : texts.first.data;
    }

    await tester.pumpWidget(
      CardflipBootstrap(
        dataDir: () async => '${dir.path}/app',
        isWeb: false,
        picker: _NoPicker(),
        syncPlatform: _LoopbackPlatform(),
      ),
    );
    await idle();

    // Trust the network.
    await go('Settings');
    await tapInSettings(find.byKey(const Key('trust-network')));
    expect(find.text('Trust this network?'), findsOneWidget);
    await tester.tap(find.widgetWithText(FilledButton, 'Trust'));
    await idle();
    expect(
      find.byKey(const ValueKey('trusted-subnet:127.0.0.0/8')),
      findsOneWidget,
    );

    // Pair with the peer by its address; it confirms by itself.
    await tapInSettings(find.byKey(const Key('pair-device')));
    await tester.enterText(
      find.byKey(const Key('pair-address')),
      '127.0.0.1:$peerPort',
    );
    await tester.tap(find.byKey(const Key('pair-by-address')));
    await waitFor(find.text('Compare the codes'));
    expect(find.byKey(const Key('pairing-code')), findsOneWidget);
    await tester.tap(find.text('Codes match'));
    await waitFor(find.text('Paired with Peer'));
    expect((await peer.command('status'))['paired'], ['Cardflip test']);
    await tester.pageBack();
    await idle();
    expect(find.text('Peer'), findsOneWidget);

    // Link the profile with the peer's.
    await tapInSettings(find.byKey(const Key('link-profile')));
    expect(find.text('Profile on Peer'), findsOneWidget);
    await tester.tap(find.byKey(const ValueKey('remote-default')));
    await waitFor(find.text('Default on Peer'));

    // The first sync asks; merge.
    await go('Decks');
    expect(find.byKey(const Key('sync-indicator')), findsOneWidget);
    await tester.tap(find.byKey(const Key('sync-indicator')));
    await waitFor(find.text('Sync with Peer'));
    await tester.tap(find.byKey(const Key('first-sync-merge')));
    await waitFor(find.textContaining('Synced: '));

    // A note added in the app arrives at the peer.
    await go('Add');
    await enterField(tester, 0, 'gato');
    await enterField(tester, 1, 'cat');
    await tester.tap(find.byKey(const Key('editor-submit')));
    await idle();
    expect(find.text('Card added'), findsOneWidget);
    await go('Decks');
    expect(await newCount(), '1');
    await syncNow();
    expect((await peer.command('notes'))['notes'], contains('gato'));

    // A review on the peer lowers the app's due count.
    final answered = await peer.command('answer');
    expect(answered['error'], isNull);
    expect((await peer.command('status'))['reviews'], 1);
    await syncNow();
    expect(find.textContaining(', 0 sent'), findsOneWidget);
    expect(await newCount(), '0');
    expect((await golang.decks.tree()).first.counts.new_, 0);
  });
}
