// An agent on Linux, with the real Go core: the user turns the MCP server
// on in Settings, an agent adds a note over HTTP with the copied
// configuration, the open browser shows the card, and the app's Undo
// removes it again (agent-access spec).
import 'dart:convert';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';

import 'package:cardflip/app/bootstrap.dart';
import 'package:cardflip/features/import/pick.dart';
import 'package:cardflip/golang/golang.dart' as golang;

class _NoPicker extends FilePickerService {
  @override
  Future<golang.StageInfo?> pick(List<String> extensions) async => null;

  @override
  Future<bool> save(String fileName, Uint8List bytes) async => false;
}

/// A minimal MCP client over Streamable HTTP, as an agent sends requests.
class _Agent {
  _Agent(this.url, this.authorization);

  final String url;
  final String authorization;
  final _http = HttpClient();
  String? _session;
  var _id = 0;

  /// Sends a JSON-RPC message and returns the response's result, or null
  /// for a notification.
  Future<Map<String, dynamic>?> send(
    String method,
    Map<String, Object?> params, {
    bool notification = false,
  }) async {
    final req = await _http.postUrl(Uri.parse(url));
    req.headers
      ..set('Authorization', authorization)
      ..set('Content-Type', 'application/json')
      ..set('Accept', 'application/json, text/event-stream');
    if (_session != null) req.headers.set('Mcp-Session-Id', _session!);
    req.write(
      jsonEncode({
        'jsonrpc': '2.0',
        if (!notification) 'id': ++_id,
        'method': method,
        'params': params,
      }),
    );
    final res = await req.close();
    _session ??= res.headers.value('mcp-session-id');
    final body = await res.transform(utf8.decoder).join();
    expect(res.statusCode, lessThan(300), reason: '$method: $body');
    if (notification) return null;
    // A JSON body, or an event stream whose data line holds it.
    final json = res.headers.contentType?.mimeType == 'text/event-stream'
        ? body.split('\n').lastWhere((l) => l.startsWith('data:')).substring(5)
        : body;
    final message = jsonDecode(json) as Map<String, dynamic>;
    expect(message['error'], isNull, reason: '$method: $json');
    return message['result'] as Map<String, dynamic>;
  }

  void close() => _http.close(force: true);
}

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('an agent adds a note that the browser shows and Undo removes', (
    tester,
  ) async {
    final dir = await Directory.systemTemp.createTemp('cardflip-mcp-');
    addTearDown(() => dir.delete(recursive: true));
    tester.view.physicalSize =
        const Size(1280, 860) * tester.view.devicePixelRatio;
    addTearDown(tester.view.resetPhysicalSize);

    Future<void> idle() async {
      for (var i = 0; i < 3; i++) {
        await Future<void>.delayed(const Duration(milliseconds: 40));
        await tester.pumpAndSettle();
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

    Future<void> scrollSettingsTo(Finder finder) async {
      await tester.scrollUntilVisible(
        finder,
        200,
        // The list's own; the URL's SelectableText has one too.
        scrollable: find
            .descendant(
              of: find.byKey(const Key('settings-list')),
              matching: find.byType(Scrollable),
            )
            .first,
      );
      await idle();
    }

    await tester.pumpWidget(
      CardflipBootstrap(
        dataDir: () async => '${dir.path}/data',
        isWeb: false,
        picker: _NoPicker(),
      ),
    );
    await idle();
    addTearDown(() => golang.mcp.setEnabled(false));

    // Turn the server on, on a port nothing else uses.
    await go('Settings');
    await scrollSettingsTo(find.byKey(const Key('mcp-switch')));
    await tester.tap(find.byKey(const Key('mcp-switch')));
    await idle();
    expect(find.text('Allow agents?'), findsOneWidget);
    await tester.tap(find.text('Allow'));
    await idle();
    final free = await ServerSocket.bind(InternetAddress.loopbackIPv4, 0);
    final port = free.port;
    await free.close();
    await scrollSettingsTo(find.byKey(const Key('mcp-port')));
    await tester.tap(find.byKey(const Key('mcp-port')));
    await idle();
    await tester.enterText(find.byType(TextField).last, '$port');
    await tester.tap(find.widgetWithText(FilledButton, 'Save'));
    await idle();
    final url = 'http://127.0.0.1:$port/mcp';
    await scrollSettingsTo(find.text(url));
    expect(find.text('Server running'), findsOneWidget);

    // The client configuration the user copies for the agent.
    String? copied;
    final messenger = tester.binding.defaultBinaryMessenger;
    messenger.setMockMethodCallHandler(SystemChannels.platform, (call) async {
      if (call.method == 'Clipboard.setData') {
        copied = (call.arguments as Map)['text'] as String;
      }
      return null;
    });
    await scrollSettingsTo(find.text('Copy client configuration'));
    await tester.tap(find.text('Copy client configuration'));
    await idle();
    messenger.setMockMethodCallHandler(SystemChannels.platform, null);
    final server =
        (jsonDecode(copied!) as Map<String, dynamic>)['mcpServers']['cardflip']
            as Map<String, dynamic>;
    expect(server['url'], url);

    // The browser shows the Default deck while the agent adds a note.
    await go('Browse');
    await tester.enterText(find.byKey(const Key('search')), 'deck:Default');
    await tester.testTextInput.receiveAction(TextInputAction.done);
    await idle();
    Finder row(String text) =>
        find.byWidgetPredicate((w) => w is Text && w.data == text);
    expect(row('hola agente'), findsNothing);

    final agent = _Agent(url, server['headers']['Authorization'] as String);
    addTearDown(agent.close);
    final init = await agent.send('initialize', {
      'protocolVersion': '2025-06-18',
      'capabilities': <String, Object?>{},
      'clientInfo': {'name': 'integration test', 'version': '1'},
    });
    expect(init!['serverInfo']['name'], 'cardflip');
    await agent.send('notifications/initialized', const {}, notification: true);
    final added = await agent.send('tools/call', {
      'name': 'add_notes',
      'arguments': {
        'notes': [
          {
            'noteType': 'Basic',
            'deck': 'Default',
            'fields': {'Front': 'hola agente', 'Back': 'hello agent'},
          },
        ],
      },
    });
    expect(added!['isError'], isNot(true), reason: '$added');
    expect(added['structuredContent']['added'], hasLength(1));

    // The change reaches the open browser without user action.
    for (var i = 0; i < 20 && row('hola agente').evaluate().isEmpty; i++) {
      await idle();
    }
    expect(row('hola agente'), findsOneWidget);

    // The app's Undo reverts the agent's step.
    await tester.sendKeyDownEvent(LogicalKeyboardKey.controlLeft);
    await tester.sendKeyEvent(LogicalKeyboardKey.keyZ);
    await tester.sendKeyUpEvent(LogicalKeyboardKey.controlLeft);
    await idle();
    expect(find.text('Undone: Add 1 note'), findsOneWidget);
    for (var i = 0; i < 20 && row('hola agente').evaluate().isNotEmpty; i++) {
      await idle();
    }
    expect(row('hola agente'), findsNothing);

    // Settings lists the agent's change.
    await go('Settings');
    await scrollSettingsTo(find.text('Add 1 note'));
    expect(find.text('Add 1 note'), findsOneWidget);
  });
}
