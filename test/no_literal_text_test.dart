import 'dart:io';

import 'package:flutter_test/flutter_test.dart';

/// Literal texts that are meant to stay as they are, by file: names that
/// are not translated.
const allowed = {
  'lib/app/app.dart': ['Cardflip'],
  'lib/app/bootstrap.dart': ['Cardflip'],
  'lib/features/settings/settings_screen.dart': ['Cardflip'],
};

/// A widget text or a text argument, followed by a string literal.
final _literal = RegExp(
  r'''(?:\bText\(|\b(?:tooltip|labelText|hintText|helperText|title|message|label):)\s*(?:const\s+)?r?('(?:[^'\\\n]|\\.)*'|"(?:[^"\\\n]|\\.)*")''',
);

/// Interpolations, which are code rather than text. An interpolation that
/// contains quotes ends the match early, so an unclosed one runs to the end.
final _interpolation = RegExp(r'\$\{[^}]*\}?|\$[A-Za-z_]\w*');

void main() {
  test('interface texts come from the translation files', () {
    final found = <String>[];
    final files = Directory('lib')
        .listSync(recursive: true)
        .whereType<File>()
        .where(
          (f) =>
              f.path.endsWith('.dart') &&
              !f.path.startsWith('lib/l10n/') &&
              !f.path.startsWith('lib/golang/'),
        );
    for (final file in files) {
      final source = file.readAsStringSync();
      for (final m in _literal.allMatches(source)) {
        final quoted = m.group(1)!;
        final text = quoted.substring(1, quoted.length - 1);
        if (!RegExp('[A-Za-z]').hasMatch(text.replaceAll(_interpolation, ''))) {
          continue;
        }
        if (allowed[file.path]?.contains(text) ?? false) continue;
        final line = '\n'.allMatches(source.substring(0, m.start)).length + 1;
        found.add('${file.path}:$line: $quoted');
      }
    }
    expect(found, isEmpty, reason: 'Move these texts to lib/l10n/app_en.arb');
  });

  test('the check finds literal texts', () {
    String? hit(String source) => _literal.firstMatch(source)?.group(1);
    expect(hit("Text('Add card')"), "'Add card'");
    expect(hit('tooltip:\n    "Undo"'), '"Undo"');
    expect(hit("title: const Text('x')"), "'x'");
    expect(hit('Text(l.addCard)'), isNull);
    expect("\${n.toString().padLeft(2, '".replaceAll(_interpolation, ''), '');
  });
}
