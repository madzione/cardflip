import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/app/errors.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/features/editor/note_editor_screen.dart';

import 'support/fake_go.dart';
import '../integration_test/rich_fields.dart';

void main() {
  testWidgets('edit mode loads the note and saves changes', (tester) async {
    final go = FakeGo();
    go.handle(
      'notes.Get',
      (_) => {
        'id': 5,
        'typeId': 10,
        'fields': ['hola', 'hello'],
        'tags': ['verb'],
        'cardIds': [7],
      },
    );
    go.handle('notes.Update', (_) => null);
    final app = AppState(isWeb: false);
    await app.load();
    setSize(tester, phone);
    await tester.pumpWidget(
      AppScope(
        state: app,
        child: MaterialApp(
          scaffoldMessengerKey: messengerKey,
          home: Builder(
            builder: (context) => TextButton(
              onPressed: () => Navigator.of(context).push(
                MaterialPageRoute<void>(
                  builder: (_) => const NoteEditorScreen(noteId: 5),
                ),
              ),
              child: const Text('edit'),
            ),
          ),
        ),
      ),
    );
    await tester.tap(find.text('edit'));
    await tester.pumpAndSettle();
    expect(find.text('Edit card'), findsOneWidget);
    expect(find.byKey(const Key('type-picker')), findsNothing);
    await enterField(tester, 0, 'hola!');
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    expect(go.callsTo('notes.Update'), [
      [
        5,
        ['hola!', 'hello'],
        ['verb'],
      ],
    ]);
    expect(find.text('Edit card'), findsNothing);
    app.dispose();
  });
}
