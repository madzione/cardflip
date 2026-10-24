import 'dart:io';

import 'package:flutter_test/flutter_test.dart';

/// The calls that open a dialog or menu, and the argument that sets its
/// animation.
const _styles = {
  'showDialog': 'animationStyle',
  'showMenu': 'popUpAnimationStyle',
  'PopupMenuButton': 'popUpAnimationStyle',
};

final _call = RegExp(
  r'\b(showDialog|showMenu|PopupMenuButton)\s*(?:<[^()]*?>)?\(',
);

/// The arguments of the call whose opening parenthesis is at [open].
String _arguments(String source, int open) {
  var depth = 0;
  for (var i = open; i < source.length; i++) {
    switch (source[i]) {
      case '(' || '[' || '{':
        depth++;
      case ')' || ']' || '}':
        depth--;
        if (depth == 0) return source.substring(open, i + 1);
    }
  }
  return source.substring(open);
}

void main() {
  // Reduce motion (app-shell spec): dialogs and menus open at once.
  test('every dialog and menu follows reduce motion', () {
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
      for (final m in _call.allMatches(source)) {
        final style = _styles[m.group(1)]!;
        if (_arguments(source, m.end - 1).contains('$style: popupAnimation(')) {
          continue;
        }
        final line = '\n'.allMatches(source.substring(0, m.start)).length + 1;
        found.add('${file.path}:$line: ${m.group(1)}');
      }
    }
    expect(
      found,
      isEmpty,
      reason:
          'Pass the popupAnimation(context) style (lib/app/appearance.dart)',
    );
  });
}
