import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/app/errors.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/features/browse/browse_screen.dart';
import 'package:cardflip/features/editor/field_drop_target.dart';
import 'package:cardflip/features/editor/note_editor_screen.dart';
import 'package:cardflip/features/editor/note_type_manager.dart';
import 'package:cardflip/features/editor/occlusion_editor.dart';
import 'package:cardflip/features/editor/rich_field.dart';
import 'package:cardflip/features/editor/template_preview.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';
import 'package:cardflip/widgets/card_view.dart';
import 'package:cardflip/widgets/media_cache.dart';
import 'package:cardflip/widgets/voice_recorder.dart';

import 'support/fake_go.dart';
import 'support/occlusion.dart';
import 'support/fake_media.dart';
import '../integration_test/rich_fields.dart';

Future<void> openAdd(WidgetTester tester) async {
  await tester.tap(find.text('Add').last);
  await tester.pumpAndSettle();
}

/// The error the editor shows.
String errorText(WidgetTester tester) =>
    tester.widget<Text>(find.byKey(const Key('editor-error'))).data!;

/// Scrolls to the editor's submit button and taps it.
Future<void> submitNote(WidgetTester tester) async {
  final submit = find.byKey(const Key('editor-submit'));
  await tester.scrollUntilVisible(
    submit,
    200,
    scrollable: find.byType(Scrollable).first,
  );
  await tester.tap(submit);
  await tester.pumpAndSettle();
}

void main() {
  richFieldTests();

  testWidgets('a note type and deck merged away by sync are replaced, '
      'keeping the text', (tester) async {
    final go = FakeGo();
    go.handle('notes.Add', (_) => 99);
    await pumpApp(tester);
    await openAdd(tester);
    await enterField(tester, 0, 'gato');
    await enterField(tester, 1, 'cat');

    // The first sync merged this device's Basic and Default into the other
    // device's.
    go.noteTypes = [
      noteTypeJson(5, 'Basic', ['Front', 'Back']),
      ...go.noteTypes.skip(1),
    ];
    go.tree = [deckNode(2, 'Default')];
    go.emitChange(kind: 'sync', noteTypes: true, decks: true);
    await tester.pumpAndSettle();
    expect(fieldHtml(tester, 0), contains('gato'));

    await submitNote(tester);
    final note = go.callsTo('notes.Add').single.single as Map<String, dynamic>;
    expect(note['typeId'], 5);
    expect(note['deckId'], 2);
    expect(note['fields'], [contains('gato'), contains('cat')]);
    expect(find.text('Card added'), findsOneWidget);
  });

  group('Image Occlusion', () {
    testWidgets('create from an image (spec)', (tester) async {
      final go = FakeGo();
      go.handle('notes.Add', (_) => 99);
      final image = await occlusionImage(tester);
      final picker = FakePicker()..media = [(name: 'heart.png', bytes: image)];
      await pumpApp(tester, picker: picker);
      await openAdd(tester);
      // The collection has no Image Occlusion note type yet.
      await tester.tap(find.byKey(const Key('type-picker')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Image Occlusion').last);
      await tester.pumpAndSettle();
      expect(go.callsTo('notes.ImageOcclusionType'), hasLength(1));
      // The mask editor stands for Occlusion and Image.
      expect(find.byKey(const Key('occlusion-section')), findsOneWidget);
      expect(find.byKey(const Key('field-0')), findsNothing);
      expect(find.byKey(const Key('field-1')), findsNothing);
      expect(find.byKey(const Key('field-2')), findsOneWidget);

      await submitNote(tester);
      expect(errorText(tester), 'Choose an image to draw on.');

      await tester.ensureVisible(find.byKey(const Key('occlusion-pick-image')));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('occlusion-pick-image')));
      await tester.pumpAndSettle();
      await waitForCanvas(tester);
      expect(go.media['heart.png'], image);

      await submitNote(tester);
      expect(errorText(tester), 'Draw a shape over each part to ask.');

      await tester.ensureVisible(find.byKey(const Key('occlusion-tool-rect')));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('occlusion-tool-rect')));
      await showCanvas(tester);
      await tester.pump();
      for (final (x, y) in [(.1, .1), (.6, .5)]) {
        await tester.dragFrom(onCanvas(tester, x, y), const Offset(60, 30));
        await tester.pump();
      }
      await enterField(tester, 2, 'Heart');
      await submitNote(tester);

      final note =
          go.callsTo('notes.Add').single.single as Map<String, dynamic>;
      final [occlusion as String, img, header, _, _] =
          note['fields'] as List<dynamic>;
      expect(note['typeId'], 13);
      final shapes = parseOcclusionField(occlusion);
      expect(shapes.map((s) => (s.kind, s.ord, s.occludeInactive)), [
        ('rect', 1, true),
        ('rect', 2, true),
      ]);
      expect(shapes.first.left, moreOrLessEquals(.1, epsilon: 1e-3));
      expect(img, '<img src="heart.png">');
      expect(header, 'Heart');
      // Ready for the next note, with a new image.
      expect(find.byKey(const Key('occlusion-canvas')), findsNothing);
      expect(find.text('Choose image', skipOffstage: false), findsOneWidget);
    });

    testWidgets('masks of a translated note type with other fields', (
      tester,
    ) async {
      // German field names, and the image before the shapes.
      final type = {
        ...occlusionTypeJson(),
        'name': 'Bildverdeckung',
        'fields': [
          'Bild',
          'Verdeckung',
          'Kopfzeile',
          'Rückseite Extra',
          'Kommentare',
        ],
        'occlusionField': 1,
        'imageField': 0,
      };
      final go = FakeGo()
        ..noteTypes = [
          noteTypeJson(10, 'Basic', ['Front', 'Back']),
          type,
        ];
      go.media['heart.png'] = await occlusionImage(tester);
      const field =
          '{{c1::image-occlusion:rect:left=.05:top=.05:width=.2:height=.2}}<br>'
          '{{c2::image-occlusion:rect:left=.6:top=.6:width=.3:height=.25}}';
      go.handle(
        'notes.Get',
        (_) => {
          'id': 5,
          'typeId': 13,
          'fields': ['<img src="heart.png">', field, 'Heart', '', ''],
          'tags': <String>[],
          'cardIds': [7, 8],
          'cardOrds': [0, 1],
        },
      );
      go.handle('notes.Update', (_) => null);
      final app = await pumpEditor(tester, const NoteEditorScreen(noteId: 5));
      await waitForCanvas(tester);
      await showCanvas(tester);
      await tester.tapAt(onCanvas(tester, .1, .1));
      await tester.pump();
      await tester.sendKeyEvent(LogicalKeyboardKey.delete);
      await tester.pump();
      await submitNote(tester);
      expect(go.callsTo('notes.Update').single[1], [
        '<img src="heart.png">',
        '{{c2::image-occlusion:rect:left=.6:top=.6:width=.3:height=.25}}',
        'Heart',
        '',
        '',
      ]);
      app.dispose();
    });

    testWidgets('edit masks (spec)', (tester) async {
      final go = FakeGo()
        ..noteTypes = [
          noteTypeJson(10, 'Basic', ['Front', 'Back']),
          occlusionTypeJson(),
        ];
      go.media['heart.png'] = await occlusionImage(tester);
      const field =
          '{{c1::image-occlusion:rect:left=.05:top=.05:width=.2:height=.2}}<br>'
          '{{c2::image-occlusion:rect:left=.6:top=.6:width=.3:height=.25}}';
      go.handle(
        'notes.Get',
        (_) => {
          'id': 5,
          'typeId': 13,
          'fields': [field, '<img src="heart.png">', 'Heart', '', ''],
          'tags': <String>[],
          'cardIds': [7, 8],
          'cardOrds': [0, 1],
        },
      );
      go.handle('notes.Update', (_) => null);
      final app = await pumpEditor(tester, const NoteEditorScreen(noteId: 5));
      await waitForCanvas(tester);
      await showCanvas(tester);
      // Delete shape 1.
      await tester.tapAt(onCanvas(tester, .1, .1));
      await tester.pump();
      expect(
        find.byKey(const Key('occlusion-delete')).hitTestable(),
        findsNothing,
      );
      // The canvas has the focus: Delete removes the selected shape.
      await tester.sendKeyEvent(LogicalKeyboardKey.delete);
      await tester.pump();
      await submitNote(tester);
      // Shape 2 keeps its number, so card 2 keeps its scheduling; Go keeps
      // card 1 as an empty card.
      expect(go.callsTo('notes.Update').single[1], [
        '{{c2::image-occlusion:rect:left=.6:top=.6:width=.3:height=.25}}',
        '<img src="heart.png">',
        'Heart',
        '',
        '',
      ]);
      expect(find.byType(NoteEditorScreen), findsNothing);
      app.dispose();
    });
  });

  group('attach media', () {
    testWidgets('an image goes into the focused Back field (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      go.handle('notes.Add', (_) => 99);
      final picker = FakePicker()..media = [(name: 'cat.jpg', bytes: pngPixel)];
      await pumpApp(tester, picker: picker);
      await openAdd(tester);
      await enterField(tester, 1, 'Cat ');
      await tester.tap(find.byKey(const Key('attach-media')));
      await tester.pumpAndSettle();

      expect(picker.mediaPicks, 1);
      final [name, _] = go.callsTo('media.Add').single;
      expect(name, 'cat.jpg');
      expect(go.media['cat.jpg'], pngPixel);
      expect(fieldHtml(tester, 1), 'Cat <img src="cat.jpg">');
      expect(fieldHtml(tester, 0), isEmpty);
      // The field shows the stored image.
      await tester.runAsync(
        () => Future<void>.delayed(const Duration(milliseconds: 200)),
      );
      await tester.pumpAndSettle();
      expect(
        find.descendant(
          of: find.byKey(const Key('field-1')),
          matching: find.byType(Image),
        ),
        findsOneWidget,
      );

      await enterField(tester, 0, 'cat');
      await tester.tap(find.byKey(const Key('editor-submit')));
      await tester.pumpAndSettle();
      final note =
          go.callsTo('notes.Add').single.single as Map<String, dynamic>;
      expect(note['fields'], ['cat', 'Cat <img src="cat.jpg">']);
    });

    testWidgets('audio goes into the Front field as [sound:] (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      final picker = FakePicker()
        ..media = [
          (name: 'hola.mp3', bytes: Uint8List.fromList([1, 2, 3])),
        ];
      await pumpApp(tester, picker: picker);
      await openAdd(tester);
      await tester.tap(find.byKey(const Key('attach-media')));
      await tester.pumpAndSettle();
      expect(fieldHtml(tester, 0), '[sound:hola.mp3]');
      final first = tester.widget<RichField>(find.byKey(const Key('field-0')));
      expect(first.focusNode.hasFocus, isTrue);

      // The name Go stores the file under is inserted: a different file
      // named hola.mp3 is renamed.
      picker.media = [
        (name: 'hola.mp3', bytes: Uint8List.fromList([4])),
      ];
      await tester.tap(find.byKey(const Key('attach-media')));
      await tester.pumpAndSettle();
      expect(fieldHtml(tester, 0), '[sound:hola.mp3][sound:hola-1.mp3]');
      expect(go.media.keys, ['hola.mp3', 'hola-1.mp3']);
    });

    testWidgets('HTML mode inserts markup; a cancelled picker adds nothing', (
      tester,
    ) async {
      final go = FakeGo();
      final picker = FakePicker();
      await pumpApp(tester, picker: picker);
      await openAdd(tester);
      await tester.tap(find.byKey(const ValueKey('field-1-html')));
      await tester.pumpAndSettle();
      await enterField(tester, 1, '<b>x</b>');
      await tester.tap(find.byKey(const Key('attach-media')));
      await tester.pumpAndSettle();
      expect(go.callsTo('media.Add'), isEmpty);
      expect(fieldHtml(tester, 1), '<b>x</b>');

      picker.media = [
        (name: 'a&b.png', bytes: pngPixel),
        (name: 'clip.mp4', bytes: Uint8List.fromList([9])),
      ];
      await tester.tap(find.byKey(const Key('attach-media')));
      await tester.pumpAndSettle();
      expect(
        fieldHtml(tester, 1),
        '<b>x</b><img src="a&amp;b.png">[sound:clip.mp4]',
      );
    });

    testWidgets('a failure is shown and the input kept', (tester) async {
      final go = FakeGo();
      go.handle('media.Add', (_) => throw 'The file could not be saved: full');
      final picker = FakePicker()..media = [(name: 'cat.jpg', bytes: pngPixel)];
      await pumpApp(tester, picker: picker);
      await openAdd(tester);
      await enterField(tester, 0, 'kept');
      await tester.tap(find.byKey(const Key('attach-media')));
      await tester.pumpAndSettle();
      expect(find.text('The file could not be saved: full'), findsOneWidget);
      expect(fieldHtml(tester, 0), 'kept');
    });
  });

  group('editing buttons in the field toolbar', () {
    List<String?> appBarActions(WidgetTester tester) => [
      for (final a
          in tester
              .widget<AppBar>(find.widgetWithText(AppBar, 'Add card'))
              .actions!)
        (a as IconButton).tooltip,
    ];

    testWidgets('Basic: Math, Attach and Record audio end it (spec)', (
      tester,
    ) async {
      FakeGo();
      await pumpApp(tester);
      await openAdd(tester);
      expect(toolbarTools(tester), [
        'Math',
        'Attach picture, audio or video',
        'Record audio',
      ]);
      expect(appBarActions(tester), [
        'History: cards added today',
        'Note types',
      ]);
    });

    testWidgets('Cloze: Cloze deletion comes first (spec)', (tester) async {
      FakeGo();
      await pumpApp(tester);
      await openAdd(tester);
      await tester.tap(find.byKey(const Key('type-picker')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Cloze').last);
      await tester.pumpAndSettle();
      expect(toolbarTools(tester), [
        'Cloze deletion',
        'Math',
        'Attach picture, audio or video',
        'Record audio',
      ]);
      expect(appBarActions(tester), [
        'History: cards added today',
        'Note types',
      ]);
    });

    testWidgets('they stay enabled in HTML source (spec)', (tester) async {
      FakeGo();
      final picker = FakePicker()..media = [(name: 'cat.jpg', bytes: pngPixel)];
      await pumpApp(tester, picker: picker);
      await openAdd(tester);
      await tester.tap(find.byKey(const ValueKey('field-1-html')));
      await tester.pumpAndSettle();
      await enterField(tester, 1, '');
      expect(toolbarButton(tester, 'Bold').onPressed, isNull);
      expect(toolbarButton(tester, 'Math').onPressed, isNotNull);
      expect(toolbarButton(tester, 'Attach').onPressed, isNotNull);

      await tester.tap(find.byTooltip('Math'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('MathJax inline'));
      await tester.pumpAndSettle();
      final source = fieldController(tester, 1).source;
      expect(source.text, r'\(\)');
      // The cursor goes after the math.
      source.selection = const TextSelection.collapsed(offset: 4);
      await tester.tap(find.byKey(const Key('attach-media')));
      await tester.pumpAndSettle();
      expect(source.text, r'\(\)<img src="cat.jpg">');
      expect(fieldHtml(tester, 0), isEmpty);
    });
  });

  testWidgets(
    'rapid entry: fields clear, deck, type and tags stay, focus returns',
    (tester) async {
      final go = FakeGo();
      go.tree = [deckNode(1, 'Default'), deckNode(2, 'Spanish')];
      go.lastDeck = 2;
      go.handle('notes.Add', (_) => 99);
      await pumpApp(tester);
      await openAdd(tester);

      // The title and the button say "Add card"; the template is a note type.
      expect(find.text('Add card'), findsNWidgets(2));
      expect(find.text('Note type'), findsOneWidget);
      expect(find.text('Basic'), findsOneWidget);
      expect(find.text('Spanish'), findsOneWidget);
      await enterField(tester, 0, 'hola');
      await enterField(tester, 1, 'hello');
      await tester.enterText(
        find.byKey(const Key('tags-input')),
        'verb irregular ',
      );
      await tester.pump();
      expect(find.widgetWithText(InputChip, 'verb'), findsOneWidget);
      expect(find.widgetWithText(InputChip, 'irregular'), findsOneWidget);

      await tester.tap(find.byKey(const Key('editor-submit')));
      await tester.pumpAndSettle();
      final note =
          go.callsTo('notes.Add').single.single as Map<String, dynamic>;
      expect(note['typeId'], 10);
      expect(note['deckId'], 2);
      expect(note['fields'], ['hola', 'hello']);
      expect(note['tags'], ['verb', 'irregular']);

      expect(fieldHtml(tester, 0), isEmpty);
      expect(fieldHtml(tester, 1), isEmpty);
      expect(find.widgetWithText(InputChip, 'verb'), findsOneWidget);
      expect(find.text('Basic'), findsOneWidget);
      expect(find.text('Spanish'), findsOneWidget);
      final first = tester.widget<RichField>(find.byKey(const Key('field-0')));
      expect(first.focusNode.hasFocus, isTrue);
      expect(find.text('Card added'), findsOneWidget);
    },
  );

  testWidgets('pinned field kept after an add (spec)', (tester) async {
    final go = FakeGo();
    go.noteTypes = [
      noteTypeJson(10, 'Basic', ['Front', 'Back', 'Source']),
    ];
    go.handle('notes.Add', (_) => 99);
    await pumpApp(tester);
    await openAdd(tester);
    final pin = find.byKey(const Key('field-2-pin'));
    expect(tester.widget<IconButton>(pin).isSelected, isFalse);
    await tester.tap(pin);
    await tester.pumpAndSettle();
    expect(go.callsTo('notes.SetFieldSticky').single, [10, 2, true]);
    expect(tester.widget<IconButton>(pin).isSelected, isTrue);

    await enterField(tester, 0, 'hola');
    await enterField(tester, 1, 'hello');
    await enterField(tester, 2, 'Book 1');
    await tester.tap(find.byKey(const Key('editor-submit')));
    await tester.pumpAndSettle();
    final note = go.callsTo('notes.Add').single.single as Map<String, dynamic>;
    expect(note['fields'], ['hola', 'hello', 'Book 1']);
    expect(fieldHtml(tester, 0), isEmpty);
    expect(fieldHtml(tester, 1), isEmpty);
    expect(fieldHtml(tester, 2), 'Book 1');
  });

  group('per-field presentation', () {
    /// The text the visual editor of field [i] renders.
    RichText fieldText(WidgetTester tester, int i) => tester.widget<RichText>(
      find
          .descendant(
            of: find.byKey(Key('field-$i-visual')),
            matching: find.byType(RichText),
          )
          .first,
    );

    testWidgets('a right-to-left field is laid out from the right (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      go.noteTypes = [
        noteTypeJson(
          10,
          'Arabic',
          ['English', 'Arabic'],
          options: {
            1: {'rtl': true, 'font': 'Amiri', 'size': 28},
          },
        ),
      ];
      final app = await pumpEditor(tester, const NoteEditorScreen());
      await enterField(tester, 1, 'كتاب');
      await enterField(tester, 0, 'book');
      final arabic = fieldText(tester, 1);
      expect(arabic.textDirection, TextDirection.rtl);
      expect(arabic.textAlign, TextAlign.start);
      // The text is aligned right: it ends at the editor's right edge.
      final box = tester.getRect(find.byKey(const Key('field-1-visual')));
      final text = tester.getRect(find.byWidget(arabic));
      final rendered = tester.renderObject<RenderParagraph>(
        find.byWidget(arabic),
      );
      final width = rendered.getMaxIntrinsicWidth(double.infinity);
      expect(text.right, closeTo(box.right, 1));
      expect(text.right - width, greaterThan(box.center.dx));
      // The editing font and size apply to that field only.
      expect(arabic.text.style?.fontFamily, 'Amiri');
      expect(arabic.text.style?.fontSize, 28);
      final english = fieldText(tester, 0);
      expect(english.textDirection, TextDirection.ltr);
      expect(english.text.style?.fontFamily, isNot('Amiri'));
      expect(english.text.style?.fontSize, 16);
      app.dispose();
    });

    testWidgets('a field collapsed by default expands when tapped (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      go.noteTypes = [
        noteTypeJson(
          10,
          'Basic',
          ['Front', 'Back', 'Notes'],
          options: {
            2: {'collapsed': true},
          },
        ),
      ];
      go.handle(
        'notes.Get',
        (_) => {
          'id': 5,
          'typeId': 10,
          'fields': ['go', 'went', 'An <b>irregular</b>\n verb'],
          'tags': <String>[],
          'cardIds': [7],
          'cardOrds': [0],
        },
      );
      go.handle('notes.Update', (_) => null);
      final app = await pumpEditor(tester, const NoteEditorScreen(noteId: 5));
      expect(find.byKey(const Key('field-2-visual')), findsNothing);
      expect(find.byKey(const Key('field-1-visual')), findsOneWidget);
      // Collapsed, the field shows the start of its text.
      final preview = find.byKey(const Key('field-2-collapsed'));
      expect(
        find.descendant(of: preview, matching: find.text('An irregular verb')),
        findsOneWidget,
      );
      await tester.ensureVisible(preview);
      await tester.tap(preview);
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('field-2-visual')), findsOneWidget);
      expect(preview, findsNothing);
      // It can be collapsed again; its content is kept.
      await tester.tap(find.byKey(const Key('field-2-collapse')));
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('field-2-visual')), findsNothing);
      await submitNote(tester);
      expect(
        (go.callsTo('notes.Update').single[1] as List)[2],
        contains('irregular'),
      );
      app.dispose();
    });
  });

  testWidgets('pins are remembered when the add screen is reopened', (
    tester,
  ) async {
    final go = FakeGo();
    go.noteTypes = [
      noteTypeJson(10, 'Basic', ['Front', 'Back', 'Source'], sticky: {2}),
    ];
    go.handle('notes.Add', (_) => 99);
    final app = await pumpEditor(tester, const NoteEditorScreen());
    bool pinned(int i) =>
        tester.widget<IconButton>(find.byKey(Key('field-$i-pin'))).isSelected!;
    expect([pinned(0), pinned(1), pinned(2)], [false, false, true]);
    await tester.tap(find.byKey(const Key('field-0-pin')));
    await tester.pumpAndSettle();
    expect(go.callsTo('notes.SetFieldSticky').single, [10, 0, true]);

    await tester.pageBack();
    await tester.pumpAndSettle();
    await tester.tap(find.text('open'));
    await tester.pumpAndSettle();
    expect([pinned(0), pinned(1), pinned(2)], [true, false, true]);
    // Unpinned fields clear after an add again.
    await tester.tap(find.byKey(const Key('field-2-pin')));
    await tester.pumpAndSettle();
    expect(go.callsTo('notes.SetFieldSticky').last, [10, 2, false]);
    await enterField(tester, 0, 'Chapter 3');
    await enterField(tester, 1, 'hello');
    await enterField(tester, 2, 'Book 2');
    await pressKeys(tester, [
      LogicalKeyboardKey.controlLeft,
      LogicalKeyboardKey.enter,
    ]);
    expect(go.callsTo('notes.Add'), hasLength(1));
    expect(fieldHtml(tester, 0), 'Chapter 3');
    expect(fieldHtml(tester, 1), isEmpty);
    expect(fieldHtml(tester, 2), isEmpty);
    app.dispose();
  });

  testWidgets('Ctrl+Enter in the Back field adds the card (spec)', (
    tester,
  ) async {
    final go = FakeGo();
    go.handle('notes.Add', (_) => 99);
    await pumpApp(tester);
    await openAdd(tester);
    await enterField(tester, 0, 'hola');
    await enterField(tester, 1, 'hello');
    await pressKeys(tester, [
      LogicalKeyboardKey.controlLeft,
      LogicalKeyboardKey.enter,
    ]);
    final note = go.callsTo('notes.Add').single.single as Map<String, dynamic>;
    expect(note['fields'], ['hola', 'hello']);
    expect(fieldHtml(tester, 0), isEmpty);
    expect(fieldHtml(tester, 1), isEmpty);
    expect(find.text('Card added'), findsOneWidget);
  });

  testWidgets('Ctrl+Enter in the HTML editor and tags adds the card', (
    tester,
  ) async {
    final go = FakeGo();
    go.handle('notes.Add', (_) => 99);
    await pumpApp(tester);
    await openAdd(tester);
    await enterField(tester, 0, 'hola');
    await tester.tap(find.byKey(const Key('field-1-html')));
    await tester.pumpAndSettle();
    await enterField(tester, 1, '<b>hello</b>');
    await pressKeys(tester, [
      LogicalKeyboardKey.controlLeft,
      LogicalKeyboardKey.enter,
    ]);
    expect((go.callsTo('notes.Add').single.single as Map)['fields'], [
      'hola',
      '<b>hello</b>',
    ]);
    // Typed tags are committed with the note.
    await enterField(tester, 0, 'adiós');
    await tester.enterText(find.byKey(const Key('tags-input')), 'verb');
    await pressKeys(tester, [
      LogicalKeyboardKey.controlLeft,
      LogicalKeyboardKey.numpadEnter,
    ]);
    expect((go.callsTo('notes.Add').last.single as Map)['tags'], ['verb']);
  });

  testWidgets('Cmd+Enter adds the card on macOS', (tester) async {
    debugDefaultTargetPlatformOverride = TargetPlatform.macOS;
    try {
      final go = FakeGo();
      go.handle('notes.Add', (_) => 99);
      await pumpApp(tester);
      await openAdd(tester);
      await enterField(tester, 0, 'hola');
      await pressKeys(tester, [
        LogicalKeyboardKey.metaLeft,
        LogicalKeyboardKey.enter,
      ]);
      expect(go.callsTo('notes.Add'), hasLength(1));
    } finally {
      debugDefaultTargetPlatformOverride = null;
    }
  });

  testWidgets('Ctrl+Enter saves in the edit screen, which has no pins', (
    tester,
  ) async {
    final go = FakeGo();
    go.handle(
      'notes.Get',
      (_) => {
        'id': 5,
        'typeId': 10,
        'fields': ['comer', 'to eat'],
        'tags': <String>[],
        'cardIds': [7],
      },
    );
    go.handle('notes.Update', (_) => null);
    final app = await pumpEditor(tester, const NoteEditorScreen(noteId: 5));
    expect(find.byKey(const Key('field-0-pin')), findsNothing);
    await enterField(tester, 1, 'to eat (something)');
    await pressKeys(tester, [
      LogicalKeyboardKey.controlLeft,
      LogicalKeyboardKey.enter,
    ]);
    expect(go.callsTo('notes.Update').single[1], [
      'comer',
      'to eat (something)',
    ]);
    // Saved: the editor closed.
    expect(find.byType(NoteEditorScreen), findsNothing);
    app.dispose();
  });

  testWidgets('History opens Browse with added:1 (spec)', (tester) async {
    final go = FakeGo();
    var next = 100;
    go.handle('notes.Add', (_) => next++);
    // Go lists the cards added today for added:1.
    go.handle(
      'browse.Search',
      (args) =>
          args[0] == 'added:1' ? [for (var i = 100; i < next; i++) i] : <int>[],
    );
    go.handle(
      'browse.Rows',
      (args) => [
        for (final id in args[0] as List)
          {
            'cardId': id,
            'noteId': id,
            'question': 'word $id',
            'deck': 'Default',
            'noteType': 'Basic',
            'cardType': 'Card 1',
            'answer': '',
            'sortField': '',
            'tags': '',
            'reviews': 0,
            'lapses': 0,
            'state': 'new',
            'due': null,
            'suspended': false,
            'flag': 0,
            'marked': false,
          },
      ],
    );
    await pumpApp(tester);
    await openAdd(tester);
    for (var i = 0; i < 5; i++) {
      await enterField(tester, 0, 'word $i');
      await enterField(tester, 1, 'meaning $i');
      await tester.tap(find.byKey(const Key('editor-submit')));
      await tester.pumpAndSettle();
    }
    expect(go.callsTo('notes.Add'), hasLength(5));

    await tester.tap(find.byTooltip('History: cards added today'));
    await tester.pumpAndSettle();
    expect(find.byType(BrowseScreen), findsOneWidget);
    expect(go.callsTo('browse.Search').last[0], 'added:1');
    expect(find.widgetWithText(TextField, 'added:1'), findsOneWidget);
    for (var id = 100; id < 105; id++) {
      expect(find.byKey(Key('row-$id')), findsOneWidget);
    }

    // Back in the add screen, which kept its state.
    await tester.pageBack();
    await tester.pumpAndSettle();
    expect(find.byType(NoteEditorScreen), findsOneWidget);
  });

  testWidgets('the edit screen has no History action', (tester) async {
    final go = FakeGo();
    go.handle(
      'notes.Get',
      (args) => {
        'id': 5,
        'typeId': 10,
        'fields': ['comer', 'to eat'],
        'tags': <String>[],
        'cardIds': [7],
      },
    );
    final app = await pumpEditor(tester, const NoteEditorScreen(noteId: 5));
    expect(find.byKey(const Key('editor-history')), findsNothing);
    app.dispose();
  });

  group('duplicate warning', () {
    final warning = find.byKey(const Key('duplicate-warning'));
    bool marked(WidgetTester tester) =>
        tester.widget<RichField>(find.byKey(const Key('field-0'))).marked;

    Future<void> typeFirst(WidgetTester tester, String text) async {
      await enterField(tester, 0, text);
      await tester.pump(NoteEditorScreen.duplicateDebounce);
      await tester.pumpAndSettle();
    }

    testWidgets('duplicate word (spec)', (tester) async {
      final go = FakeGo();
      go.existingNotes = [(id: 5, typeId: 10, first: '<b>perro</b>')];
      go.handle('browse.Search', (_) => [7, 8]);
      go.handle(
        'notes.Get',
        (args) => {
          'id': args[0],
          'typeId': 10,
          'fields': ['perro', 'dog'],
          'tags': <String>[],
          'cardIds': [7],
        },
      );
      go.handle(
        'browse.Rows',
        (args) => [
          for (final id in args[0] as List)
            {
              'cardId': id,
              'noteId': 5,
              'question': 'perro',
              'deck': 'Default',
              'noteType': 'Basic',
              'cardType': 'Card 1',
              'answer': '',
              'sortField': '',
              'tags': '',
              'reviews': 0,
              'lapses': 0,
              'state': 'new',
              'due': null,
              'suspended': false,
              'flag': 0,
              'marked': false,
            },
        ],
      );
      await pumpApp(tester);
      await openAdd(tester);
      expect(warning, findsNothing);

      await enterField(tester, 0, 'Perro ');
      // Not before the debounce.
      await tester.pump(const Duration(milliseconds: 100));
      expect(warning, findsNothing);
      await tester.pump(NoteEditorScreen.duplicateDebounce);
      await tester.pumpAndSettle();
      expect(go.callsTo('notes.Duplicates').last[0], 10);
      expect(go.callsTo('notes.Duplicates').last[2], 0);
      expect(warning, findsOneWidget);
      expect(
        find.text(
          'This is a duplicate: another note has the same Front field.',
        ),
        findsOneWidget,
      );
      expect(marked(tester), isTrue);

      // "Show duplicates" opens Browse with the note's cards only.
      await tester.tap(find.byKey(const Key('show-duplicates')));
      await tester.pumpAndSettle();
      expect(find.byType(BrowseScreen), findsOneWidget);
      expect(go.callsTo('notes.Get').last, [5]);
      expect(find.byKey(const Key('row-7')), findsOneWidget);
      expect(find.byKey(const Key('row-8')), findsNothing);
      expect(find.text('Only 1 note'), findsOneWidget);
      // Removing the filter shows every card again.
      await tester.tap(find.byTooltip('Show all cards'));
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('row-8')), findsOneWidget);

      // Back in the editor, a different word clears the warning.
      await tester.pageBack();
      await tester.pumpAndSettle();
      await typeFirst(tester, 'perros');
      expect(warning, findsNothing);
      expect(marked(tester), isFalse);
    });

    testWidgets('different note type (spec)', (tester) async {
      final go = FakeGo();
      go.existingNotes = [(id: 5, typeId: 12, first: 'perro')];
      await pumpApp(tester);
      await openAdd(tester);
      await typeFirst(tester, 'perro');
      expect(go.callsTo('notes.Duplicates').last.take(2), [10, 'perro']);
      expect(warning, findsNothing);
      expect(marked(tester), isFalse);
    });

    testWidgets('add anyway (spec)', (tester) async {
      final go = FakeGo();
      go.existingNotes = [(id: 5, typeId: 10, first: 'perro')];
      go.handle('notes.Add', (_) => 99);
      await pumpApp(tester);
      await openAdd(tester);
      await typeFirst(tester, 'perro');
      await enterField(tester, 1, 'dog');
      expect(warning, findsOneWidget);
      await tester.tap(find.byKey(const Key('editor-submit')));
      await tester.pumpAndSettle();
      final note =
          go.callsTo('notes.Add').single.single as Map<String, dynamic>;
      expect(note['fields'], ['perro', 'dog']);
      expect(find.text('Card added'), findsOneWidget);
      // The fields cleared, so the warning went with them.
      await tester.pump(NoteEditorScreen.duplicateDebounce);
      await tester.pumpAndSettle();
      expect(warning, findsNothing);
    });

    testWidgets('the edit screen ignores the edited note', (tester) async {
      final go = FakeGo();
      go.existingNotes = [(id: 5, typeId: 10, first: 'comer')];
      go.handle(
        'notes.Get',
        (_) => {
          'id': 5,
          'typeId': 10,
          'fields': ['comer', 'to eat'],
          'tags': <String>[],
          'cardIds': [7],
        },
      );
      final app = await pumpEditor(tester, const NoteEditorScreen(noteId: 5));
      await tester.pump(NoteEditorScreen.duplicateDebounce);
      await tester.pumpAndSettle();
      expect(go.callsTo('notes.Duplicates').last, [10, 'comer', 5]);
      expect(warning, findsNothing);

      // Another note with the same word is reported.
      go.existingNotes.add((id: 6, typeId: 10, first: 'Comer'));
      await typeFirst(tester, 'comer ');
      expect(go.callsTo('notes.Duplicates').last[2], 5);
      expect(warning, findsOneWidget);
      app.dispose();
    });
  });

  testWidgets('a copy starts with its note type, deck, fields and tags', (
    tester,
  ) async {
    final go = FakeGo();
    go.tree = [deckNode(1, 'Default'), deckNode(2, 'Spanish')];
    go.handle('notes.Add', (_) => 99);
    final app = await pumpEditor(
      tester,
      const NoteEditorScreen(
        initial: NoteDraft(
          typeId: 11,
          deckId: 2,
          fields: ['hola', 'hello'],
          tags: ['greeting'],
        ),
      ),
    );
    expect(find.text('Add card'), findsWidgets);
    expect(find.text('Basic (and reversed card)'), findsOneWidget);
    expect(find.text('Spanish'), findsOneWidget);
    expect(fieldHtml(tester, 0), 'hola');
    expect(fieldHtml(tester, 1), 'hello');
    expect(find.widgetWithText(InputChip, 'greeting'), findsOneWidget);

    await enterField(tester, 0, 'buenos días');
    await tester.tap(find.byKey(const Key('editor-submit')));
    await tester.pumpAndSettle();
    final note = go.callsTo('notes.Add').single.single as Map<String, dynamic>;
    expect(note['typeId'], 11);
    expect(note['deckId'], 2);
    expect(note['fields'], ['buenos días', 'hello']);
    expect(note['tags'], ['greeting']);
    app.dispose();
  });

  group('deck for new cards', () {
    FakeGo decks() {
      final go = FakeGo();
      go.tree = [
        deckNode(1, 'Default'),
        deckNode(2, 'Spanish'),
        deckNode(3, 'Biology'),
      ];
      go.lastDeck = 2;
      go.deckByType = {12: 3, 10: 2};
      return go;
    }

    Future<void> pickType(WidgetTester tester, String name) async {
      await tester.tap(find.byKey(const Key('type-picker')));
      await tester.pumpAndSettle();
      await tester.tap(find.text(name).last);
      await tester.pumpAndSettle();
    }

    int? shownDeck(WidgetTester tester) => tester
        .widget<DropdownButtonFormField<int>>(
          find.byKey(const Key('deck-picker')),
        )
        .initialValue;

    testWidgets('the deck follows the note type (spec)', (tester) async {
      final go = decks();
      go.settings = {...go.settings, 'deckByNoteType': true};
      await pumpApp(tester);
      await openAdd(tester);
      expect(shownDeck(tester), 2);
      await pickType(tester, 'Cloze');
      expect(go.callsTo('notes.Defaults').last, [12]);
      expect(shownDeck(tester), 3);
      expect(find.text('Biology'), findsOneWidget);
      // A note type without a remembered deck gets the last used one.
      await pickType(tester, 'Basic (and reversed card)');
      expect(shownDeck(tester), 2);
    });

    testWidgets('"Current deck" keeps the deck when the type changes', (
      tester,
    ) async {
      final go = decks();
      await pumpApp(tester);
      await openAdd(tester);
      final asked = go.callsTo('notes.Defaults').length;
      await pickType(tester, 'Cloze');
      expect(go.callsTo('notes.Defaults'), hasLength(asked));
      expect(shownDeck(tester), 2);
      expect(find.text('Spanish'), findsOneWidget);
    });
  });

  testWidgets('empty note: the error is shown and the input kept', (
    tester,
  ) async {
    final go = FakeGo();
    go.handle(
      'notes.Add',
      (_) => throw 'This would produce no cards. Fill in the Front field.',
    );
    await pumpApp(tester);
    await openAdd(tester);
    await enterField(tester, 1, 'only a back');
    await tester.tap(find.byKey(const Key('editor-submit')));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('editor-error')), findsOneWidget);
    expect(
      find.text('This would produce no cards. Fill in the Front field.'),
      findsOneWidget,
    );
    expect(fieldHtml(tester, 1), 'only a back');
  });

  testWidgets(
    'save failure: a message explains it, the editor keeps the input',
    (tester) async {
      final go = FakeGo();
      go.handle(
        'notes.Add',
        (_) => throw 'could not save the change: disk full',
      );
      await pumpApp(tester);
      await openAdd(tester);
      await enterField(tester, 0, 'hola');
      await enterField(tester, 1, 'hello');
      await tester.tap(find.byKey(const Key('editor-submit')));
      await tester.pumpAndSettle();
      expect(find.text('Could not save the change: disk full'), findsOneWidget);
      expect(fieldHtml(tester, 0), 'hola');
      expect(fieldHtml(tester, 1), 'hello');
      // Fixed: the second try goes through.
      go.handle('notes.Add', (_) => 1);
      // Above the snackbar of the failure.
      await Scrollable.ensureVisible(
        tester.element(find.byKey(const Key('editor-submit'))),
        alignment: 0.5,
      );
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('editor-submit')));
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('editor-error')), findsNothing);
      expect(fieldHtml(tester, 0), isEmpty);
    },
  );

  testWidgets('changing the note type keeps fields with the same name', (
    tester,
  ) async {
    FakeGo();
    await pumpApp(tester);
    await openAdd(tester);
    await enterField(tester, 0, 'hola');
    await tester.tap(find.byKey(const Key('type-picker')));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Basic (and reversed card)').last);
    await tester.pumpAndSettle();
    expect(fieldHtml(tester, 0), 'hola');
    await tester.tap(find.byKey(const Key('type-picker')));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Cloze').last);
    await tester.pumpAndSettle();
    expect(find.widgetWithText(RichField, 'Text'), findsOneWidget);
    expect(find.widgetWithText(RichField, 'Back Extra'), findsOneWidget);
    // The cloze button wraps text in the next deletion.
    await enterField(tester, 0, 'Paris');
    await tester.tap(find.byTooltip('Cloze deletion'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Next cloze number'));
    await tester.pumpAndSettle();
    expect(fieldHtml(tester, 0), 'Paris{{c1::}}');
  });

  group('cloze insertion', () {
    const text = '{{c1::Paris}} is in France';

    Future<void> openCloze(WidgetTester tester) async {
      FakeGo();
      await pumpApp(tester);
      await openAdd(tester);
      await tester.tap(find.byKey(const Key('type-picker')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Cloze').last);
      await tester.pumpAndSettle();
    }

    Future<void> nextNumber(WidgetTester tester) => pressKeys(tester, [
      LogicalKeyboardKey.controlLeft,
      LogicalKeyboardKey.shiftLeft,
      LogicalKeyboardKey.keyC,
    ]);

    Future<void> sameNumber(WidgetTester tester) => pressKeys(tester, [
      LogicalKeyboardKey.controlLeft,
      LogicalKeyboardKey.altLeft,
      LogicalKeyboardKey.shiftLeft,
      LogicalKeyboardKey.keyC,
    ]);

    Future<void> chooseFromMenu(WidgetTester tester, String entry) async {
      await tester.tap(find.byTooltip('Cloze deletion'));
      await tester.pumpAndSettle();
      await tester.tap(find.text(entry));
      await tester.pumpAndSettle();
    }

    testWidgets('Ctrl+Shift+C uses the next number (spec)', (tester) async {
      await openCloze(tester);
      await enterField(tester, 0, text);
      selectInField(tester, 0, 20, 26);
      await nextNumber(tester);
      expect(fieldHtml(tester, 0), '{{c1::Paris}} is in {{c2::France}}');
    });

    testWidgets('Ctrl+Alt+Shift+C uses the same number (spec)', (tester) async {
      await openCloze(tester);
      await enterField(tester, 0, text);
      selectInField(tester, 0, 20, 26);
      await sameNumber(tester);
      expect(fieldHtml(tester, 0), '{{c1::Paris}} is in {{c1::France}}');
    });

    testWidgets('numbers come from all fields (spec)', (tester) async {
      await openCloze(tester);
      await enterField(tester, 1, '{{c3::x}}');
      await enterField(tester, 0, 'Paris');
      selectInField(tester, 0, 0, 5);
      await nextNumber(tester);
      expect(fieldHtml(tester, 0), '{{c4::Paris}}');
      await enterField(tester, 0, 'Rome');
      selectInField(tester, 0, 0, 4);
      await sameNumber(tester);
      expect(fieldHtml(tester, 0), '{{c3::Rome}}');
    });

    testWidgets('multi-number deletions count with their highest number', (
      tester,
    ) async {
      await openCloze(tester);
      await enterField(tester, 0, '{{c1, 4::Paris}} {{c0::x}} France');
      selectInField(tester, 0, 27, 33);
      await nextNumber(tester);
      expect(fieldHtml(tester, 0), '{{c1, 4::Paris}} {{c0::x}} {{c5::France}}');
    });

    testWidgets('the same number is 1 without clozes', (tester) async {
      await openCloze(tester);
      await enterField(tester, 0, 'Paris');
      selectInField(tester, 0, 0, 5);
      await sameNumber(tester);
      expect(fieldHtml(tester, 0), '{{c1::Paris}}');
    });

    testWidgets('an empty selection puts the cursor inside', (tester) async {
      await openCloze(tester);
      await enterField(tester, 0, 'Paris ');
      await nextNumber(tester);
      final c = fieldController(tester, 0);
      expect(c.html, 'Paris {{c1::}}');
      expect(c.quill.selection, const TextSelection.collapsed(offset: 12));
    });

    testWidgets('the toolbar menu offers both numbers', (tester) async {
      await openCloze(tester);
      await enterField(tester, 0, text);
      selectInField(tester, 0, 20, 26);
      await tester.tap(find.byTooltip('Cloze deletion'));
      await tester.pumpAndSettle();
      expect(find.text('Ctrl+Shift+C'), findsOneWidget);
      expect(find.text('Ctrl+Alt+Shift+C'), findsOneWidget);
      await tester.tap(find.text('Same cloze number'));
      await tester.pumpAndSettle();
      expect(fieldHtml(tester, 0), '{{c1::Paris}} is in {{c1::France}}');

      await enterField(tester, 1, 'Seine');
      selectInField(tester, 1, 0, 5);
      await chooseFromMenu(tester, 'Next cloze number');
      expect(fieldHtml(tester, 1), '{{c2::Seine}}');
    }, variant: TargetPlatformVariant.only(TargetPlatform.linux));

    testWidgets('Basic note types ignore the shortcuts', (tester) async {
      FakeGo();
      await pumpApp(tester);
      await openAdd(tester);
      await enterField(tester, 0, 'Paris');
      selectInField(tester, 0, 0, 5);
      await nextNumber(tester);
      expect(fieldHtml(tester, 0), 'Paris');
    });
  });

  testWidgets('the math menu wraps the selection (spec)', (tester) async {
    FakeGo();
    await pumpApp(tester);
    await openAdd(tester);
    await enterField(tester, 0, 'Area x^2 m');
    selectInField(tester, 0, 5, 8);
    await tester.tap(find.byTooltip('Math'));
    await tester.pumpAndSettle();
    expect(find.text('Ctrl+M, M'), findsOneWidget);
    await tester.tap(find.text('MathJax inline'));
    await tester.pumpAndSettle();
    expect(fieldHtml(tester, 0), r'Area \(x^2\) m');

    // Chemistry goes inside inline math.
    await enterField(tester, 1, 'H2O');
    selectInField(tester, 1, 0, 3);
    await tester.tap(find.byTooltip('Math'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('MathJax chemistry'));
    await tester.pumpAndSettle();
    expect(fieldHtml(tester, 1), r'\(\ce{H2O}\)');
  }, variant: TargetPlatformVariant.only(TargetPlatform.linux));

  testWidgets('the math menu shows no shortcuts on Android (spec)', (
    tester,
  ) async {
    FakeGo();
    await pumpApp(tester);
    await openAdd(tester);
    await enterField(tester, 0, 'x');
    await tester.tap(find.byTooltip('Math'));
    await tester.pumpAndSettle();
    expect(find.text('MathJax inline'), findsOneWidget);
    expect(find.text('Ctrl+M, M'), findsNothing);
    await tester.tapAt(Offset.zero);
    await tester.pumpAndSettle();
    expect(find.byTooltip('Ctrl+Enter'), findsNothing);
  }, variant: TargetPlatformVariant.only(TargetPlatform.android));

  testWidgets('narrow desktop window (spec)', (tester) async {
    FakeGo();
    await pumpApp(tester, size: const Size(340, 700));
    await openAdd(tester);
    await enterField(tester, 0, 'x');
    expect(find.byTooltip('Ctrl+Enter'), findsOneWidget);
    await tester.tap(find.byTooltip('Math'));
    await tester.pumpAndSettle();
    expect(find.text('MathJax inline'), findsOneWidget);
    expect(find.text('Ctrl+M, M'), findsOneWidget);
  }, variant: TargetPlatformVariant.only(TargetPlatform.linux));

  testWidgets('Ctrl+M then E inserts display math around the cursor (spec)', (
    tester,
  ) async {
    FakeGo();
    await pumpApp(tester);
    await openAdd(tester);
    await enterField(tester, 0, '');
    await tester.sendKeyDownEvent(LogicalKeyboardKey.controlLeft);
    await tester.sendKeyEvent(LogicalKeyboardKey.keyM);
    await tester.sendKeyUpEvent(LogicalKeyboardKey.controlLeft);
    await tester.sendKeyEvent(LogicalKeyboardKey.keyE);
    await tester.pump();
    final c = fieldController(tester, 0);
    expect(c.html, r'\[\]');
    expect(c.quill.selection, const TextSelection.collapsed(offset: 2));

    // Typing goes between the delimiters; other chords work the same way.
    await tester.sendKeyDownEvent(LogicalKeyboardKey.controlLeft);
    await tester.sendKeyEvent(LogicalKeyboardKey.keyT);
    await tester.sendKeyEvent(LogicalKeyboardKey.keyM);
    await tester.sendKeyUpEvent(LogicalKeyboardKey.controlLeft);
    await tester.pump();
    expect(c.html, r'\[[$$][/$$]\]');
    expect(c.quill.selection, const TextSelection.collapsed(offset: 6));
  });

  testWidgets('remove field asks for confirmation naming the notes', (
    tester,
  ) async {
    final go = FakeGo();
    go.noteTypes = [
      noteTypeJson(10, 'Basic', ['Front', 'Back'], noteCount: 4),
    ];
    go.handle('notes.Type', (_) => go.noteTypes.first);
    go.handle('notes.RemoveField', (args) {
      go.noteTypes = [
        noteTypeJson(10, 'Basic', ['Front'], noteCount: 4),
      ];
      return null;
    });
    setSize(tester, phone);
    await tester.pumpWidget(
      const MaterialApp(home: NoteTypeScreen(typeId: 10)),
    );
    await tester.pumpAndSettle();
    await tester.tap(find.byTooltip('Remove field').last);
    await tester.pumpAndSettle();
    expect(find.text('Remove field “Back”?'), findsOneWidget);
    expect(find.textContaining('4 notes'), findsOneWidget);
    await tester.tap(find.text('Remove'));
    await tester.pumpAndSettle();
    expect(go.callsTo('notes.RemoveField'), [
      [10, 1],
    ]);
    expect(find.text('Back'), findsNothing);
  });

  group('field options', () {
    Future<FakeGo> pumpBasic(WidgetTester tester) async {
      final go = FakeGo();
      go.noteTypes = [
        noteTypeJson(10, 'Basic', ['Front', 'Back', 'Notes']),
      ];
      go.handle('notes.Type', (_) => go.noteTypes.first);
      setSize(tester, phone);
      await tester.pumpWidget(
        const MaterialApp(home: NoteTypeScreen(typeId: 10)),
      );
      await tester.pumpAndSettle();
      return go;
    }

    double top(WidgetTester tester, String text) =>
        tester.getTopLeft(find.text(text)).dy;

    Future<void> openOptions(WidgetTester tester, int field) async {
      await tester.tap(find.byTooltip('Field options').at(field));
      await tester.pumpAndSettle();
    }

    testWidgets('"Use HTML editor by default" is set per field', (
      tester,
    ) async {
      final go = await pumpBasic(tester);
      await openOptions(tester, 1);
      expect(find.text('Options for “Back”'), findsOneWidget);
      await tester.ensureVisible(find.byKey(const Key('html-by-default')));
      await tester.tap(find.byKey(const Key('html-by-default')));
      await tester.tap(find.text('Save'));
      await tester.pumpAndSettle();
      final [typeId, index, options as Map] = go
          .callsTo('notes.SetFieldOptions')
          .single;
      expect([typeId, index], [10, 1]);
      expect(options['htmlByDefault'], isTrue);
      expect(options['rtl'], isFalse);
      // Reloaded from Go: listed for Back only.
      expect(find.byKey(const Key('field-summary-1')), findsOneWidget);
      expect(
        tester.widget<Text>(find.byKey(const Key('field-summary-1'))).data,
        'HTML editor',
      );
      expect(go.callsTo('notes.SetSortField'), isEmpty);
    });

    testWidgets('a field is moved with its handle (spec)', (tester) async {
      final go = await pumpBasic(tester);
      expect(top(tester, 'Front'), lessThan(top(tester, 'Back')));
      // The first field sorts by default.
      expect(
        tester.widget<Text>(find.byKey(const Key('field-summary-0'))).data,
        'Sort field',
      );

      final handle = find.byKey(const Key('move-field-1'));
      final gesture = await tester.startGesture(tester.getCenter(handle));
      for (var i = 0; i < 10; i++) {
        await gesture.moveBy(const Offset(0, -12));
        await tester.pump(const Duration(milliseconds: 16));
      }
      await gesture.up();
      await tester.pumpAndSettle();

      expect(go.callsTo('notes.MoveField'), [
        [10, 1, 0],
      ]);
      // Back is listed first; Front is still the sort field.
      expect(top(tester, 'Back'), lessThan(top(tester, 'Front')));
      expect(go.noteTypes.first['sortField'], 1);
      expect(
        tester.widget<Text>(find.byKey(const Key('field-summary-1'))).data,
        'Sort field',
      );
    });

    testWidgets('the sort field and the editing options are set (spec)', (
      tester,
    ) async {
      final go = await pumpBasic(tester);
      // The sort field cannot be switched off, only moved to another one.
      await openOptions(tester, 0);
      expect(
        tester
            .widget<SwitchListTile>(find.byKey(const Key('field-sort')))
            .onChanged,
        isNull,
      );
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();

      await openOptions(tester, 1);
      await tester.tap(find.byKey(const Key('field-sort')));
      await tester.enterText(find.byKey(const Key('field-font')), 'Amiri');
      await tester.enterText(find.byKey(const Key('field-size')), '0');
      await tester.ensureVisible(find.byKey(const Key('field-rtl')));
      await tester.tap(find.byKey(const Key('field-rtl')));
      await tester.ensureVisible(find.byKey(const Key('field-exclude')));
      await tester.tap(find.byKey(const Key('field-exclude')));
      await tester.tap(find.text('Save'));
      await tester.pumpAndSettle();
      // An invalid size keeps the dialog open.
      expect(find.byKey(const Key('field-options-error')), findsOneWidget);
      expect(go.callsTo('notes.SetFieldOptions'), isEmpty);
      await tester.enterText(find.byKey(const Key('field-size')), '28');
      await tester.tap(find.text('Save'));
      await tester.pumpAndSettle();

      expect(go.callsTo('notes.SetSortField'), [
        [10, 1],
      ]);
      final options = go.callsTo('notes.SetFieldOptions').single[2] as Map;
      expect(options, {
        'htmlByDefault': false,
        'sticky': false,
        'font': 'Amiri',
        'size': 28,
        'rtl': true,
        'collapsed': false,
        'excludeFromSearch': true,
      });
      expect(find.byKey(const Key('field-summary-0')), findsNothing);
      expect(
        tester.widget<Text>(find.byKey(const Key('field-summary-1'))).data,
        'Sort field · Amiri 28 · Right to left · Not in unqualified searches',
      );
    });
  });

  group('note type list', () {
    Future<void> pumpList(WidgetTester tester) async {
      setSize(tester, phone);
      await tester.pumpWidget(const MaterialApp(home: NoteTypeListScreen()));
      await tester.pumpAndSettle();
    }

    Future<void> openMenu(WidgetTester tester, int row) async {
      await tester.tap(find.byTooltip('Note type actions').at(row));
      await tester.pumpAndSettle();
    }

    String nameText(WidgetTester tester) => tester
        .widget<TextField>(find.byKey(const Key('note-type-name')))
        .controller!
        .text;

    testWidgets('add from a stock type under a new name (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      await pumpList(tester);
      await tester.tap(find.byTooltip('Add note type'));
      await tester.pumpAndSettle();

      for (final s in go.stockTypes) {
        expect(find.widgetWithText(RadioListTile<String>, s), findsOneWidget);
      }
      // Basic is taken, so a free name is proposed; it follows the stock
      // type until the user changes it.
      expect(nameText(tester), 'Basic 2');
      await tester.tap(find.text('Basic (optional reversed card)'));
      await tester.pumpAndSettle();
      expect(nameText(tester), 'Basic (optional reversed card)');

      // Name taken: the error is shown in the dialog, which stays open.
      await tester.enterText(find.byKey(const Key('note-type-name')), 'basic');
      await tester.tap(find.widgetWithText(FilledButton, 'Add'));
      await tester.pumpAndSettle();
      expect(find.textContaining('"Basic" already exists'), findsOneWidget);

      await tester.enterText(
        find.byKey(const Key('note-type-name')),
        'Vocabulary',
      );
      await tester.tap(find.widgetWithText(FilledButton, 'Add'));
      await tester.pumpAndSettle();
      expect(go.callsTo('notes.AddType').last, [
        'Basic (optional reversed card)',
        'Vocabulary',
      ]);
      expect(find.byType(AlertDialog), findsNothing);
      expect(find.text('Vocabulary'), findsOneWidget);
    });

    testWidgets('stock types are listed and named in German', (tester) async {
      final go = FakeGo();
      setSize(tester, phone);
      await tester.pumpWidget(
        const MaterialApp(
          locale: Locale('de'),
          supportedLocales: AppLocalizations.supportedLocales,
          localizationsDelegates: AppLocalizations.localizationsDelegates,
          home: NoteTypeListScreen(),
        ),
      );
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Notiztyp hinzufügen'));
      await tester.pumpAndSettle();
      for (final s in [
        'Einfach',
        'Einfach (mit umgekehrter Karte)',
        'Lückentext',
      ]) {
        expect(find.widgetWithText(RadioListTile<String>, s), findsOneWidget);
      }
      expect(find.widgetWithText(RadioListTile<String>, 'Cloze'), findsNothing);
      // The English Basic does not take the German name.
      expect(nameText(tester), 'Einfach');
      await tester.tap(find.text('Lückentext'));
      await tester.pumpAndSettle();
      expect(nameText(tester), 'Lückentext');
      await tester.tap(find.widgetWithText(FilledButton, 'Hinzufügen'));
      await tester.pumpAndSettle();
      // The stock type is still identified by its English name.
      expect(go.callsTo('notes.AddType').last, ['Cloze', 'Lückentext']);
    });

    testWidgets('clone proposes a free copy name', (tester) async {
      final go = FakeGo();
      go.noteTypes = [
        ...go.noteTypes,
        noteTypeJson(13, 'Basic copy', ['Front', 'Back']),
      ];
      await pumpList(tester);
      await openMenu(tester, 0);
      await tester.tap(find.text('Clone'));
      await tester.pumpAndSettle();
      expect(find.text('Clone “Basic”'), findsOneWidget);
      expect(nameText(tester), 'Basic copy 2');

      await tester.tap(find.widgetWithText(FilledButton, 'Clone'));
      await tester.pumpAndSettle();
      expect(go.callsTo('notes.CloneType'), [
        [10, 'Basic copy 2'],
      ]);
      expect(find.text('Basic copy 2'), findsOneWidget);
    });

    testWidgets('delete confirms with the notes and cards it deletes (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      go.noteTypes = [
        noteTypeJson(10, 'Basic', ['Front', 'Back'], noteCount: 10),
        noteTypeJson(12, 'Cloze', ['Text', 'Back Extra'], cloze: true),
      ];
      go.handle('notes.TypeDeletionInfo', (_) => {'notes': 10, 'cards': 20});
      await pumpList(tester);
      await openMenu(tester, 0);
      await tester.tap(find.text('Delete'));
      await tester.pumpAndSettle();
      expect(find.text('Delete “Basic”?'), findsOneWidget);
      expect(find.textContaining('10 notes and 20 cards'), findsOneWidget);

      // Cancel keeps it.
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(go.callsTo('notes.DeleteType'), isEmpty);

      await openMenu(tester, 0);
      await tester.tap(find.text('Delete'));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(FilledButton, 'Delete'));
      await tester.pumpAndSettle();
      expect(go.callsTo('notes.DeleteType'), [
        [10],
      ]);
      expect(find.text('Basic'), findsNothing);
      expect(find.text('Cloze'), findsOneWidget);
    });

    testWidgets('the last note type cannot be deleted (spec)', (tester) async {
      final go = FakeGo();
      go.noteTypes = [go.noteTypes.first];
      await pumpList(tester);
      await openMenu(tester, 0);
      final delete = find.ancestor(
        of: find.text('Delete'),
        matching: find.byWidgetPredicate((w) => w is PopupMenuItem),
      );
      expect(tester.widget<PopupMenuItem<dynamic>>(delete).enabled, isFalse);
      await tester.tap(find.text('Delete'), warnIfMissed: false);
      await tester.pumpAndSettle();
      expect(go.callsTo('notes.TypeDeletionInfo'), isEmpty);
      expect(go.callsTo('notes.DeleteType'), isEmpty);
    });
  });

  group('template preview', () {
    golang.NoteType basic(FakeGo go) =>
        golang.NoteType.fromJson(go.noteTypes.first);

    Future<void> pumpTemplateEditor(
      WidgetTester tester,
      golang.NoteType type, {
      Size size = desktop,
    }) async {
      setSize(tester, size);
      await tester.pumpWidget(
        MaterialApp(
          home: TemplateEditorScreen(
            noteType: type,
            index: 0,
            template: type.templates.first,
          ),
        ),
      );
      await tester.pumpAndSettle();
    }

    CardView side(WidgetTester tester, String side) =>
        tester.widget<CardView>(find.byKey(Key('preview-$side')));

    testWidgets('typing updates the preview after the debounce (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      go.previewNotes = {
        10: [
          const PreviewNote(1, {'Front': 'hola', 'Back': 'hello'}),
        ],
      };
      await pumpTemplateEditor(tester, basic(go));
      // Side by side on a wide screen, showing the note.
      expect(find.text('Preview'), findsNothing);
      expect(side(tester, 'front').html, contains('hola'));
      expect(go.callsTo('notes.PreviewTemplate').last[3], 1);

      await tester.enterText(
        find.byKey(const Key('front-template')),
        '<b>{{Front}}</b>',
      );
      // The debounce starts with the frame showing the new text.
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 250));
      expect(side(tester, 'front').html, isNot(contains('<b>')));
      await tester.pump(const Duration(milliseconds: 50));
      await tester.pumpAndSettle();
      expect(side(tester, 'front').html, contains('<b>hola</b>'));
      expect(
        (go.callsTo('notes.PreviewTemplate').last[1] as Map)['qfmt'],
        '<b>{{Front}}</b>',
      );
      // Nothing is saved.
      expect(go.callsTo('notes.UpdateTemplate'), isEmpty);
    });

    testWidgets('placeholder text without notes, in a tab on phones (spec)', (
      tester,
    ) async {
      final go = FakeGo();
      await pumpTemplateEditor(tester, basic(go), size: phone);
      expect(find.byKey(const Key('preview-front')), findsNothing);
      await tester.tap(find.text('Preview'));
      await tester.pumpAndSettle();
      expect(find.text('No notes yet'), findsOneWidget);
      expect(side(tester, 'front').html, contains('(Front)'));
      expect(go.callsTo('notes.PreviewTemplate').last[3], 0);
    });

    testWidgets('cloze number picker and dark toggle', (tester) async {
      final go = FakeGo();
      go.previewNotes = {
        12: [
          const PreviewNote(
            5,
            {'Text': '{{c1::uno}} {{c2::dos}}', 'Back Extra': ''},
            clozeNumbers: [1, 2],
          ),
        ],
      };
      final cloze = golang.NoteType.fromJson(go.noteTypes[2]);
      await pumpTemplateEditor(tester, cloze);
      expect(side(tester, 'front').html, contains('card1'));

      await tester.tap(find.byKey(const Key('preview-cloze')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Cloze 2').last);
      await tester.pumpAndSettle();
      expect(go.callsTo('notes.PreviewTemplate').last[4], 2);
      expect(side(tester, 'front').html, contains('card2'));

      await tester.tap(find.byTooltip('Preview dark theme'));
      await tester.pumpAndSettle();
      expect(go.callsTo('notes.PreviewTemplate').last[5], true);
      expect(side(tester, 'front').html, contains('nightMode'));
      expect(find.byTooltip('Preview light theme'), findsOneWidget);
    });

    testWidgets('a CSS change updates the preview', (tester) async {
      final go = FakeGo();
      setSize(tester, desktop);
      await tester.pumpWidget(
        MaterialApp(home: CssEditorScreen(noteType: basic(go))),
      );
      await tester.pumpAndSettle();
      expect(side(tester, 'front').html, contains('(Front)'));
      expect(side(tester, 'front').color, isNull);

      await tester.enterText(
        find.byKey(const Key('css-input')),
        '.card { color: red }',
      );
      await tester.pump();
      await tester.pump(TemplatePreview.debounce);
      await tester.pumpAndSettle();
      expect(side(tester, 'front').color, 'red');
      expect(go.callsTo('notes.SetCSS'), isEmpty);
    });

    TextButton saveButton(WidgetTester tester) =>
        tester.widget<TextButton>(find.widgetWithText(TextButton, 'Save'));

    testWidgets('a template problem shows while typing (spec)', (tester) async {
      final go = FakeGo();
      await pumpTemplateEditor(tester, basic(go));
      expect(find.byKey(const Key('template-problem')), findsNothing);
      expect(saveButton(tester).onPressed, isNotNull);

      await tester.enterText(
        find.byKey(const Key('front-template')),
        '{{Frnt}}',
      );
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 250));
      expect(find.byKey(const Key('template-problem')), findsNothing);
      await tester.pump(const Duration(milliseconds: 50));
      await tester.pumpAndSettle();
      expect(
        find.text(
          'The front template uses {{Frnt}}, but there is no field "Frnt".',
        ),
        findsOneWidget,
      );
      final check = go.callsTo('notes.CheckTemplate').last;
      expect(check[1], 0);
      expect((check[2] as Map)['qfmt'], '{{Frnt}}');
      // Save is disabled, nothing is saved.
      expect(saveButton(tester).onPressed, isNull);
      await tester.tap(find.text('Save'), warnIfMissed: false);
      await tester.pumpAndSettle();
      expect(go.callsTo('notes.UpdateTemplate'), isEmpty);

      // Fixing the template enables Save again.
      await tester.enterText(
        find.byKey(const Key('front-template')),
        '{{Front}}',
      );
      await tester.pump();
      await tester.pump(TemplateEditorScreen.checkDebounce);
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('template-problem')), findsNothing);
      expect(saveButton(tester).onPressed, isNotNull);
    });

    testWidgets('a broken card type disables Save when opened', (tester) async {
      final go = FakeGo();
      final json = {
        ...go.noteTypes.first,
        'templates': [
          {'name': 'Card 1', 'qfmt': '{{Front}}', 'afmt': '{{#Bak}}{{/Bak}}'},
        ],
      };
      go.noteTypes = [json, ...go.noteTypes.skip(1)];
      await pumpTemplateEditor(tester, golang.NoteType.fromJson(json));
      expect(
        find.text(
          'The back template uses {{#Bak}}, but there is no field "Bak".',
        ),
        findsOneWidget,
      );
      expect(saveButton(tester).onPressed, isNull);
    });

    testWidgets('saving templates keeps the card type options', (tester) async {
      final go = FakeGo();
      final json = {
        ...go.noteTypes.first,
        'templates': [
          {
            'name': 'Card 1',
            'qfmt': '{{Front}}',
            'afmt': '{{Back}}',
            'deckOverride': 5,
            'browserQfmt': '{{Front}}!',
          },
        ],
      };
      go.noteTypes = [json, ...go.noteTypes.skip(1)];
      await pumpTemplateEditor(tester, golang.NoteType.fromJson(json));
      await tester.enterText(
        find.byKey(const Key('front-template')),
        'Q: {{Front}}',
      );
      await tester.pump(TemplateEditorScreen.checkDebounce);
      await tester.pumpAndSettle();
      await tester.tap(find.text('Save'));
      await tester.pumpAndSettle();
      final saved = go.callsTo('notes.UpdateTemplate').single[2] as Map;
      expect(saved['qfmt'], 'Q: {{Front}}');
      expect(saved['deckOverride'], 5);
      expect(saved['browserQfmt'], '{{Front}}!');
    });
  });

  group('card type options', () {
    FakeGo vocab() {
      final go = FakeGo();
      go.noteTypes = [
        noteTypeJson(
          10,
          'Vocab',
          ['Front', 'Back'],
          templates: ['Recognition', 'Production'],
        ),
      ];
      go.handle('notes.Type', (_) => go.noteTypes.first);
      go.tree = [
        deckNode(1, 'Default'),
        deckNode(2, 'Spanish', hasChildren: true),
        deckNode(3, 'Spanish::Production', parentId: 2),
        deckNode(4, 'Cram', filtered: true),
      ];
      return go;
    }

    Future<void> pumpType(WidgetTester tester) async {
      setSize(tester, phone);
      await tester.pumpWidget(
        const MaterialApp(home: NoteTypeScreen(typeId: 10)),
      );
      await tester.pumpAndSettle();
    }

    double top(WidgetTester tester, String text) =>
        tester.getTopLeft(find.text(text)).dy;

    testWidgets('a card type is moved with its handle (spec)', (tester) async {
      final go = vocab();
      await pumpType(tester);
      expect(top(tester, 'Recognition'), lessThan(top(tester, 'Production')));

      final handle = find.byKey(const Key('move-template-1'));
      final gesture = await tester.startGesture(tester.getCenter(handle));
      for (var i = 0; i < 10; i++) {
        await gesture.moveBy(const Offset(0, -12));
        await tester.pump(const Duration(milliseconds: 16));
      }
      await gesture.up();
      await tester.pumpAndSettle();

      expect(go.callsTo('notes.MoveTemplate'), [
        [10, 1, 0],
      ]);
      expect(top(tester, 'Production'), lessThan(top(tester, 'Recognition')));
    });

    testWidgets('cloze note types have no handle', (tester) async {
      final go = vocab();
      go.noteTypes = [
        noteTypeJson(10, 'Cloze', ['Text'], cloze: true),
      ];
      await pumpType(tester);
      expect(find.byKey(const Key('move-template-0')), findsNothing);
      expect(find.byTooltip('Card type options'), findsOneWidget);
    });

    testWidgets('a deck override and browser templates are set (spec)', (
      tester,
    ) async {
      final go = vocab();
      await pumpType(tester);
      await tester.tap(find.byTooltip('Card type options').last);
      await tester.pumpAndSettle();
      expect(find.text('Options for “Production”'), findsOneWidget);

      await tester.tap(find.byKey(const Key('deck-override')));
      await tester.pumpAndSettle();
      // Filtered decks cannot get new cards.
      expect(find.text('Cram'), findsNothing);
      await tester.tap(find.text('Spanish::Production').last);
      await tester.pumpAndSettle();
      await tester.enterText(
        find.byKey(const Key('browser-front-template')),
        '{{Back}}',
      );
      await tester.tap(find.text('Save'));
      await tester.pumpAndSettle();

      final [typeId, index, template as Map] = go
          .callsTo('notes.UpdateTemplate')
          .single;
      expect([typeId, index], [10, 1]);
      expect(template['name'], 'Production');
      expect(template['qfmt'], '{{Front}}');
      expect(template['deckOverride'], 3);
      expect(template['browserQfmt'], '{{Back}}');
      expect(template.containsKey('browserAfmt'), isFalse);
      expect(find.text('Options for “Production”'), findsNothing);

      // Reopened, the dialog shows the saved options; None removes the
      // override.
      await tester.tap(find.byTooltip('Card type options').last);
      await tester.pumpAndSettle();
      expect(find.text('Spanish::Production'), findsOneWidget);
      expect(
        tester
            .widget<TextField>(find.byKey(const Key('browser-front-template')))
            .controller!
            .text,
        '{{Back}}',
      );
      await tester.tap(find.byKey(const Key('deck-override')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('None (the deck chosen when adding)').last);
      await tester.pumpAndSettle();
      await tester.tap(find.text('Save'));
      await tester.pumpAndSettle();
      final cleared = go.callsTo('notes.UpdateTemplate').last[2] as Map;
      expect(cleared.containsKey('deckOverride'), isFalse);
      expect(cleared['browserQfmt'], '{{Back}}');
    });

    testWidgets('an error keeps the dialog open', (tester) async {
      final go = vocab();
      go.handle('notes.UpdateTemplate', (_) {
        throw 'That deck no longer exists.';
      });
      await pumpType(tester);
      await tester.tap(find.byTooltip('Card type options').first);
      await tester.pumpAndSettle();
      await tester.tap(find.text('Save'));
      await tester.pumpAndSettle();
      expect(find.text('That deck no longer exists.'), findsOneWidget);
      expect(find.text('Options for “Recognition”'), findsOneWidget);
    });
  });
}

/// Shows [screen] on top of a home route, as the browser opens the editor.
/// The test disposes the returned state at its end.
Future<AppState> pumpEditor(WidgetTester tester, Widget screen) async {
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
            onPressed: () =>
                Navigator.of(context)
                    .push(MaterialPageRoute<void>(builder: (_) => screen)),
            child: const Text('open'),
          ),
        ),
      ),
    ),
  );
  await tester.tap(find.text('open'));
  await tester.pumpAndSettle();
  return app;
}

void richFieldTests() {
  testWidgets('Ctrl+B makes the selection bold (spec)', (tester) async {
    final go = FakeGo();
    go.handle('notes.Add', (_) => 1);
    await pumpApp(tester);
    await openAdd(tester);
    await enterField(tester, 0, 'perro');
    selectInField(tester, 0, 0, 5);
    await pressKeys(tester, [
      LogicalKeyboardKey.controlLeft,
      LogicalKeyboardKey.keyB,
    ]);
    await tester.pump();
    await enterField(tester, 1, 'dog');
    await tester.tap(find.byKey(const Key('editor-submit')));
    await tester.pumpAndSettle();
    final note = go.callsTo('notes.Add').single.single as Map<String, dynamic>;
    expect(note['fields'], ['<b>perro</b>', 'dog']);
  });

  testWidgets('the toolbar formats the field being edited', (tester) async {
    FakeGo();
    await pumpApp(tester);
    await openAdd(tester);
    await enterField(tester, 1, 'H2O');
    selectInField(tester, 1, 1, 2);
    await tester.tap(find.byTooltip('Subscript'));
    await tester.pump();
    selectInField(tester, 1, 0, 3);
    await tester.tap(find.byTooltip('Text colour'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Red'));
    await tester.pumpAndSettle();
    expect(
      fieldHtml(tester, 1),
      '<span style="color: #ff0000">H<sub>2</sub>O</span>',
    );
    // Ctrl+R removes the formatting again.
    await pressKeys(tester, [
      LogicalKeyboardKey.controlLeft,
      LogicalKeyboardKey.keyR,
    ]);
    await tester.pump();
    expect(fieldHtml(tester, 1), 'H2O');
    await tester.tap(find.byTooltip('Bulleted list'));
    await tester.pump();
    expect(fieldHtml(tester, 1), '<ul><li>H2O</li></ul>');
  });

  testWidgets('tooltip on a tablet (spec)', (tester) async {
    FakeGo();
    await pumpApp(tester, size: const Size(800, 1280));
    await openAdd(tester);
    await enterField(tester, 0, 'perro');
    await tester.longPress(find.byTooltip('Bold'));
    await tester.pumpAndSettle();
    expect(
      find.descendant(of: find.byType(Overlay), matching: find.text('Bold')),
      findsOneWidget,
    );
    expect(find.text('Bold (Ctrl+B)'), findsNothing);
    expect(find.byTooltip('HTML editor'), findsWidgets);
  }, variant: TargetPlatformVariant.only(TargetPlatform.android));

  testWidgets('toolbar tooltips show the shortcuts on Linux', (tester) async {
    FakeGo();
    await pumpApp(tester);
    await openAdd(tester);
    await enterField(tester, 0, 'perro');
    expect(find.byTooltip('Bold (Ctrl+B)'), findsOneWidget);
    expect(find.byTooltip('Superscript (Ctrl+Shift+=)'), findsOneWidget);
    expect(find.byTooltip('HTML editor (Ctrl+Shift+X)'), findsWidgets);
  }, variant: TargetPlatformVariant.only(TargetPlatform.linux));

  testWidgets('edit source: HTML typed in HTML mode shows formatted (spec)', (
    tester,
  ) async {
    FakeGo();
    await pumpApp(tester);
    await openAdd(tester);
    await enterField(tester, 1, 'x');
    await tester.tap(find.byKey(const Key('field-1-html')));
    await tester.pumpAndSettle();
    expect(fieldController(tester, 1).htmlMode, isTrue);
    expect(find.byKey(const Key('field-1-source')), findsOneWidget);
    await enterField(tester, 1, '<i>x</i>');
    // Ctrl+Shift+X toggles back.
    await pressKeys(tester, [
      LogicalKeyboardKey.controlLeft,
      LogicalKeyboardKey.shiftLeft,
      LogicalKeyboardKey.keyX,
    ]);
    await tester.pumpAndSettle();
    final c = fieldController(tester, 1);
    expect(c.htmlMode, isFalse);
    expect(c.quill.document.toPlainText(), 'x\n');
    expect(c.quill.document.collectStyle(0, 1).attributes['italic'], isNotNull);
    expect(c.html, '<i>x</i>');
  });

  testWidgets('imported table: opens as HTML and is saved unchanged (spec)', (
    tester,
  ) async {
    const table =
        '<table class="conj"><tr><td>yo</td><td>como</td></tr></table>';
    final go = FakeGo();
    go.handle(
      'notes.Get',
      (_) => {
        'id': 5,
        'typeId': 10,
        'fields': ['comer', table],
        'tags': <String>[],
        'cardIds': [7],
      },
    );
    go.handle('notes.Update', (_) => null);
    final app = await pumpEditor(tester, const NoteEditorScreen(noteId: 5));
    expect(fieldController(tester, 0).htmlMode, isFalse);
    expect(fieldController(tester, 1).htmlMode, isTrue);
    expect(find.text(lossyNote), findsOneWidget);
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    expect(go.callsTo('notes.Update').single[1], ['comer', table]);
    app.dispose();
  });

  testWidgets('switching a lossy field to visual asks first', (tester) async {
    final go = FakeGo();
    go.handle(
      'notes.Get',
      (_) => {
        'id': 5,
        'typeId': 10,
        'fields': ['<span class="hint">perro</span>', 'dog'],
        'tags': <String>[],
        'cardIds': [7],
      },
    );
    final app = await pumpEditor(tester, const NoteEditorScreen(noteId: 5));
    await tester.tap(find.byKey(const Key('field-0-html')));
    await tester.pumpAndSettle();
    expect(find.text('Switch to the visual editor?'), findsOneWidget);
    await tester.tap(find.text('Cancel'));
    await tester.pumpAndSettle();
    expect(fieldController(tester, 0).htmlMode, isTrue);
    await tester.tap(find.byKey(const Key('field-0-html')));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('confirm-visual')));
    await tester.pumpAndSettle();
    final c = fieldController(tester, 0);
    expect(c.htmlMode, isFalse);
    // Unedited, the field still saves exactly as it was.
    expect(c.html, '<span class="hint">perro</span>');
    expect(find.text(lossyNote), findsNothing);
    await tester.pump(const Duration(seconds: 1));
    app.dispose();
  });

  testWidgets('dragging an image handle sets its width (spec)', (tester) async {
    final requested = <String>[];
    final media = MediaCache.instance;
    MediaCache.instance = MediaCache(
      load: (name) async {
        requested.add(name);
        return pngPixel;
      },
    );
    addTearDown(() => MediaCache.instance = media);
    final go = FakeGo();
    go.handle(
      'notes.Get',
      (_) => {
        'id': 5,
        'typeId': 10,
        'fields': ['perro<br><img src="dog.png" width="200">', 'dog'],
        'tags': <String>[],
        'cardIds': [7],
      },
    );
    go.handle('notes.Update', (_) => null);
    final app = await pumpEditor(tester, const NoteEditorScreen(noteId: 5));
    expect(requested, ['dog.png']);
    // Images decode outside the test's fake clock.
    await tester.runAsync(
      () => Future<void>.delayed(const Duration(milliseconds: 200)),
    );
    await tester.pumpAndSettle();
    final image = find.descendant(
      of: find.byKey(const Key('field-0')),
      matching: find.byType(Image),
    );
    expect(tester.getSize(image), const Size(200, 200));
    await tester.drag(
      find.byKey(const ValueKey('resize-dog.png')),
      const Offset(-100, 0),
      touchSlopX: 0,
    );
    await tester.pumpAndSettle();
    // Half the width, same aspect ratio.
    expect(tester.getSize(image), const Size(100, 100));
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    expect(go.callsTo('notes.Update').single[1], [
      'perro<br><img src="dog.png" width="100">',
      'dog',
    ]);
    app.dispose();
  });

  group('record audio', () {
    testWidgets('a recording goes into the focused field (spec)', (
      tester,
    ) async {
      final recorder = FakeRecorder();
      useFakeRecorder(recorder);
      final go = FakeGo();
      await pumpApp(tester);
      await openAdd(tester);
      await enterField(tester, 1, 'hola ');
      await tester.tap(find.byKey(const Key('record-audio')));
      await tester.pumpAndSettle();
      expect(recorder.recording, isTrue);
      expect(find.text('Recording…'), findsOneWidget);
      await tester.pump(const Duration(seconds: 3));
      expect(find.text('0:03'), findsOneWidget);

      await tester.tap(find.byKey(const Key('record-insert')));
      await tester.pumpAndSettle();
      expect(recorder.recording, isFalse);
      final [name, bytes] = go.callsTo('media.Add').single;
      expect(name, '.m4a');
      expect(bytes, 'mem:voice-0'.codeUnits);
      expect(fieldHtml(tester, 1), 'hola [sound:media-1.m4a]');
      // The temporary recording is deleted once it is in the media.
      expect(recorder.discarded, ['mem:voice-0']);
    });

    testWidgets('editing an existing note has the media buttons', (
      tester,
    ) async {
      final recorder = FakeRecorder();
      useFakeRecorder(recorder);
      final go = FakeGo();
      go.handle(
        'notes.Get',
        (_) => {
          'id': 5,
          'typeId': 10,
          'fields': ['hola', 'hello'],
          'tags': <String>[],
          'cardIds': [7],
        },
      );
      go.handle('notes.Update', (_) => null);
      final app = await pumpEditor(tester, const NoteEditorScreen(noteId: 5));
      expect(find.byKey(const Key('attach-media')), findsOneWidget);
      await tester.tap(find.byKey(const Key('record-audio')));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('record-insert')));
      await tester.pumpAndSettle();
      expect(fieldHtml(tester, 0), '[sound:media-1.m4a]hola');
      await tester.tap(find.text('Save'));
      await tester.pumpAndSettle();
      expect(go.callsTo('notes.Update').single[1], [
        '[sound:media-1.m4a]hola',
        'hello',
      ]);
      app.dispose();
    });

    testWidgets('cancelling adds nothing', (tester) async {
      final recorder = FakeRecorder();
      useFakeRecorder(recorder);
      final go = FakeGo();
      await pumpApp(tester);
      await openAdd(tester);
      await tester.tap(find.byKey(const Key('record-audio')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(go.callsTo('media.Add'), isEmpty);
      expect(fieldHtml(tester, 0), isEmpty);
      expect(recorder.discarded, ['mem:voice-0']);
    });

    testWidgets('refused microphone access is explained, input kept (spec)', (
      tester,
    ) async {
      useFakeRecorder(FakeRecorder()..refuse = true);
      final go = FakeGo();
      await pumpApp(tester);
      await openAdd(tester);
      await enterField(tester, 0, 'kept');
      await tester.tap(find.byKey(const Key('record-audio')));
      await tester.pumpAndSettle();
      expect(find.text('Recording…'), findsNothing);
      expect(
        tester.widget<Text>(find.byKey(const Key('editor-error'))).data,
        const MicrophoneRefused().toString(),
      );
      expect(fieldHtml(tester, 0), 'kept');
      expect(go.callsTo('media.Add'), isEmpty);
    });
  });

  group('drop', () {
    testWidgets('media dropped on a field go into it; other files not', (
      tester,
    ) async {
      final go = FakeGo();
      await pumpApp(tester, size: desktop);
      await openAdd(tester);
      final target = tester.widget<FieldDropTarget>(
        find.byKey(const Key('drop-1')),
      );
      target.onDrop([
        (name: 'cat.JPG', bytes: pngPixel),
        (name: 'notes.txt', bytes: Uint8List.fromList([1])),
        (name: 'hola.mp3', bytes: Uint8List.fromList([2])),
      ]);
      await tester.pumpAndSettle();
      expect(go.media.keys, ['cat.JPG', 'hola.mp3']);
      expect(fieldHtml(tester, 1), '<img src="cat.JPG">[sound:hola.mp3]');
      expect(fieldHtml(tester, 0), isEmpty);
      expect(
        find.text('Only pictures, audio and video can be added to a field.'),
        findsOneWidget,
      );
      final back = tester.widget<RichField>(find.byKey(const Key('field-1')));
      expect(back.focusNode.hasFocus, isTrue);
    });
  });

  testWidgets('"Align left" sets left alignment, not none', (tester) async {
    FakeGo();
    await pumpApp(tester, size: desktop);
    await openAdd(tester);
    await enterField(tester, 0, 'perro');
    Future<void> align(String choice) async {
      await tester.tap(find.byTooltip('Alignment'));
      await tester.pumpAndSettle();
      await tester.tap(find.text(choice));
      await tester.pumpAndSettle();
    }

    await align('Center');
    expect(fieldHtml(tester, 0), '<div style="text-align: center">perro</div>');
    // The card's CSS centres unaligned text, so left must be explicit.
    await align('Align left');
    expect(fieldHtml(tester, 0), '<div style="text-align: left">perro</div>');
  });

  group('paste', () {
    FakeClipboard webPage() => FakeClipboard(
      htmlText:
          '<meta charset="utf-8"><style>b { color: red }</style>'
          '<b>perro</b> <a href="https://example.com">dog</a> '
          '<span style="font-size: 30px; color: rgb(0, 128, 0)">big</span>'
          '<img src="https://example.com/x.png">',
      plainText: 'perro dog big',
    )..install();

    Future<void> paste(WidgetTester tester, {bool plain = false}) async {
      await pressKeys(tester, [
        LogicalKeyboardKey.controlLeft,
        if (plain) LogicalKeyboardKey.shiftLeft,
        LogicalKeyboardKey.keyV,
      ]);
      await tester.pumpAndSettle();
    }

    testWidgets('keeps the toolbar formatting and drops the rest', (
      tester,
    ) async {
      FakeGo();
      webPage();
      await pumpApp(tester);
      await openAdd(tester);
      await enterField(tester, 0, 'a ');
      await paste(tester);
      expect(
        fieldHtml(tester, 0),
        'a <b>perro</b> dog <span style="color: rgb(0, 128, 0)">big</span>',
      );
    });

    testWidgets('Ctrl+Shift+V pastes plain text (spec)', (tester) async {
      FakeGo();
      webPage();
      await pumpApp(tester);
      await openAdd(tester);
      await enterField(tester, 0, '');
      await paste(tester, plain: true);
      expect(fieldHtml(tester, 0), 'perro dog big');
    });

    testWidgets('a screenshot is stored as a JPEG (spec)', (tester) async {
      final go = FakeGo();
      FakeClipboard(imageBytes: pngPixel).install();
      await pumpApp(tester);
      await openAdd(tester);
      await enterField(tester, 0, 'a ');
      await paste(tester);
      final [bytes, png] = go.callsTo('media.AddPastedImage').single;
      expect(bytes, pngPixel);
      expect(png, isFalse);
      expect(go.media.keys, ['paste-1.jpg']);
      expect(fieldHtml(tester, 0), 'a <img src="paste-1.jpg">');
    });

    testWidgets('"Paste images as PNG" stores a PNG (spec)', (tester) async {
      final go = FakeGo();
      go.settings = {...go.settings, 'pasteImagesAsPng': true};
      FakeClipboard(imageBytes: pngPixel).install();
      await pumpApp(tester);
      await openAdd(tester);
      await enterField(tester, 1, '');
      await paste(tester, plain: true);
      expect(go.callsTo('media.AddPastedImage').single[1], isTrue);
      expect(fieldHtml(tester, 1), '<img src="paste-1.png">');
    });

    testWidgets('text on the clipboard wins over an image', (tester) async {
      final go = FakeGo();
      // Office apps put a picture of the copied text next to it.
      FakeClipboard(plainText: 'words', imageBytes: pngPixel).install();
      await pumpApp(tester);
      await openAdd(tester);
      await enterField(tester, 0, '');
      await paste(tester);
      expect(fieldHtml(tester, 0), 'words');
      expect(go.callsTo('media.AddPastedImage'), isEmpty);
    });

    testWidgets('an image pasted into the HTML source becomes <img>', (
      tester,
    ) async {
      final go = FakeGo();
      FakeClipboard(imageBytes: pngPixel).install();
      await pumpApp(tester);
      await openAdd(tester);
      await tester.tap(find.byKey(const ValueKey('field-1-html')));
      await tester.pumpAndSettle();
      await enterField(tester, 1, 'x');
      await paste(tester);
      expect(go.media.keys, ['paste-1.jpg']);
      expect(fieldHtml(tester, 1), 'x<img src="paste-1.jpg">');
    });

    testWidgets('without HTML on the clipboard, the text is pasted', (
      tester,
    ) async {
      FakeGo();
      FakeClipboard(plainText: 'line 1\r\nline 2').install();
      await pumpApp(tester);
      await openAdd(tester);
      await enterField(tester, 1, '');
      await paste(tester);
      expect(fieldHtml(tester, 1), 'line 1<br>line 2');
    });

    testWidgets('"Strip formatting when pasting" inverts the shortcuts '
        '(spec)', (tester) async {
      final go = FakeGo();
      go.settings = {...go.settings, 'stripPasteFormatting': true};
      webPage();
      await pumpApp(tester);
      await openAdd(tester);
      await enterField(tester, 0, '');
      await paste(tester);
      expect(fieldHtml(tester, 0), 'perro dog big');
      await enterField(tester, 1, '');
      await paste(tester, plain: true);
      expect(fieldHtml(tester, 1), contains('<b>perro</b>'));
    });

    testWidgets('in HTML mode pasting inserts the text', (tester) async {
      FakeGo();
      webPage();
      await pumpApp(tester);
      await openAdd(tester);
      await tester.tap(find.byKey(const Key('field-0-html')));
      await tester.pumpAndSettle();
      await paste(tester, plain: true);
      expect(fieldHtml(tester, 0), 'perro dog big');
    });
  });

  testWidgets('a field with "HTML editor by default" opens as HTML', (
    tester,
  ) async {
    final go = FakeGo();
    go.noteTypes = [
      noteTypeJson(10, 'Basic', ['Front', 'Back'], htmlByDefault: {1}),
    ];
    await pumpApp(tester);
    await openAdd(tester);
    expect(fieldController(tester, 0).htmlMode, isFalse);
    expect(fieldController(tester, 1).htmlMode, isTrue);
    // No note: nothing would be lost.
    expect(find.text(lossyNote), findsNothing);
  });
}
