import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:flutter_quill/flutter_quill.dart' show ChangeSource;
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/features/editor/rich_field.dart';

// [prefix] is "field" in the note editor, "details-field" in the browser.

/// The controller of field [i].
RichFieldController fieldController(
  WidgetTester tester,
  int i, {
  String prefix = 'field',
}) => tester.widget<RichField>(find.byKey(Key('$prefix-$i'))).controller;

/// The HTML of field [i].
String fieldHtml(WidgetTester tester, int i, {String prefix = 'field'}) =>
    fieldController(tester, i, prefix: prefix).html;

/// Focuses field [i] and replaces its text with [text], through the
/// platform text input as typing does: plain text in the visual editor,
/// HTML in HTML mode.
Future<void> enterField(
  WidgetTester tester,
  int i,
  String text, {
  String prefix = 'field',
}) async {
  final field = find.byKey(Key('$prefix-$i'));
  if (fieldController(tester, i, prefix: prefix).htmlMode) {
    await tester.enterText(
      find.descendant(of: field, matching: find.byType(TextField)),
      text,
    );
    return;
  }
  final editor = find.byKey(Key('$prefix-$i-visual'));
  await tester.ensureVisible(editor);
  await tester.tap(editor);
  await tester.pump();
  // The editor's text always ends with the document's last line break.
  tester.testTextInput.updateEditingValue(
    TextEditingValue(
      text: '$text\n',
      selection: TextSelection.collapsed(offset: text.length),
    ),
  );
  await tester.pump();
}

/// Selects [start] to [end] of field [i]'s visual editor.
void selectInField(
  WidgetTester tester,
  int i,
  int start,
  int end, {
  String prefix = 'field',
}) {
  fieldController(tester, i, prefix: prefix).quill.updateSelection(
    TextSelection(baseOffset: start, extentOffset: end),
    ChangeSource.local,
  );
}

/// The tooltips of the editing buttons in order, checking that they end the
/// field toolbar.
List<String?> toolbarTools(WidgetTester tester) {
  final toolbar = tester.widget<Wrap>(find.byKey(const Key('rich-toolbar')));
  expect(toolbar.children.last.key, const Key('rich-toolbar-tools'));
  return [
    for (final b in tester.widgetList<IconButton>(
      find.descendant(
        of: find.byKey(const Key('rich-toolbar-tools')),
        matching: find.byType(IconButton),
      ),
    ))
      b.tooltip,
  ];
}

/// The field toolbar's button whose tooltip starts with [tooltip].
IconButton toolbarButton(WidgetTester tester, String tooltip) =>
    tester.widget<IconButton>(
      find.descendant(
        of: find.byKey(const Key('rich-toolbar')),
        matching: find.byWidgetPredicate(
          (w) => w is IconButton && (w.tooltip?.startsWith(tooltip) ?? false),
        ),
      ),
    );

/// A clipboard holding [htmlText], [plainText] and [imageBytes]; [install]
/// makes the fields paste from it until the test ends.
class FakeClipboard extends FieldClipboard {
  FakeClipboard({this.htmlText, this.plainText, this.imageBytes});

  String? htmlText;
  String? plainText;
  Uint8List? imageBytes;

  void install() {
    FieldClipboard.instance = this;
    addTearDown(() => FieldClipboard.instance = FieldClipboard());
  }

  @override
  Future<String?> html() async => htmlText;

  @override
  Future<String?> text() async => plainText;

  @override
  Future<Uint8List?> image() async => imageBytes;
}

/// A 1×1 PNG.
final pngPixel = Uint8List.fromList(
  base64Decode(
    'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==',
  ),
);
