import 'dart:convert';
import 'dart:io';

import 'package:flutter/widgets.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/app/errors.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/builtin_names.dart';
import 'package:cardflip/l10n/go_text.dart';
import 'package:cardflip/l10n/l10n.dart';
import 'package:cardflip/widgets/time_format.dart';

import '../tool/gen_go_messages.dart';

/// A text tagged like Go's usermsg.Text.
String tagged(String code, List<String> args, String english) =>
    '\x02${jsonEncode({'code': code, 'args': args})}\x03$english\x04';

void main() {
  group('Go texts', () {
    test('a known code is shown with its message', () {
      final s = tagged('deckNameTakenByOther', [
        'Spanish',
      ], 'The name "Spanish" is already taken by another deck.');
      expect(goText(s, english), english.goDeckNameTakenByOther('Spanish'));
    });

    test('an unknown code falls back to the English wording', () {
      final s = tagged('somethingNew', ['x'], 'Something new went wrong.');
      expect(goText(s, english), 'Something new went wrong.');
    });

    test('a malformed tag or missing argument falls back to English', () {
      expect(goText('\x02{nope\x03English.\x04', english), 'English.');
      expect(
        goText(tagged('deckNameTakenByOther', [], 'English.'), english),
        'English.',
      );
      expect(
        goText(tagged('backupIntervalRange', ['x'], 'English.'), english),
        'English.',
      );
    });

    test('untagged text and the context around tags are kept', () {
      expect(goText('plain', english), 'plain');
      expect(
        goText('import: ${tagged('csvNoRows', [], 'the file has no rows')}'),
        'import: ${english.goCsvNoRows}',
      );
    });

    test('tagged arguments are translated too', () {
      final inner = tagged('searchParenNotClosed', [], 'a parenthesis…');
      final s = tagged('searchErrorAt', [inner, '1'], 'English');
      expect(
        goText(s, english),
        english.goSearchErrorAt(english.goSearchParenNotClosed, '1'),
      );
    });

    test('Go errors are shown translated and capitalised', () {
      final e = golang.GolangException(
        'decks.Rename',
        tagged('backupFirstFailed', ['disk full'], 'the collection…'),
      );
      expect(
        errorMessage(e),
        'The collection could not be backed up first, so nothing was '
        'changed: disk full',
      );
      expect(
        errorMessage(golang.GolangException('x', 'plain English.')),
        'Plain English.',
      );
    });

    test('go_messages.dart is generated from app_en.arb', () {
      // Whitespace aside, which dart format changes.
      String code(String s) => s.replaceAll(RegExp(r'\s+'), '');
      expect(
        code(File('lib/l10n/go_messages.dart').readAsStringSync()),
        code(goMessagesSource(File('lib/l10n/app_en.arb').readAsStringSync())),
        reason: 'run: fvm dart run tool/gen_go_messages.dart',
      );
    });
  });

  group('Translations', () {
    /// The message keys of an ARB file, without its metadata.
    Set<String> keys(File f) => (jsonDecode(f.readAsStringSync()) as Map).keys
        .cast<String>()
        .where((k) => !k.startsWith('@'))
        .toSet();

    test('every ARB file has the same keys as English', () {
      final en = keys(File('lib/l10n/app_en.arb'));
      final arbs = Directory('lib/l10n')
          .listSync()
          .whereType<File>()
          .where((f) => f.path.endsWith('.arb'));
      expect(arbs.length, greaterThan(1));
      for (final arb in arbs) {
        final other = keys(arb);
        expect(en.difference(other), isEmpty, reason: '${arb.path} misses');
        expect(other.difference(en), isEmpty, reason: '${arb.path} adds');
      }
    });

    test('a Go error is shown in German', () {
      final german = lookupAppLocalizations(const Locale('de'));
      final s = tagged('deckNameTakenByOther', [
        'Spanisch',
      ], 'The name "Spanish" is already taken by another deck.');
      expect(
        goText(s, german),
        'Der Name „Spanisch“ wird bereits von einem anderen Stapel verwendet.',
      );
    });

    test('the English built-in names are Go\'s stock note types', () {
      // Go falls back to English for names that do not fit its stock note
      // types (golang/core/notetypes.go), so these must match them.
      final names = builtinNames(english);
      expect(names.deck, 'Default');
      expect(names.preset, 'Default');
      expect(
        {
          for (final t in names.noteTypes)
            t.stock: [t.name, t.fields, t.cardTypes],
        },
        {
          'Basic': [
            'Basic',
            ['Front', 'Back'],
            ['Card 1'],
          ],
          'Basic (and reversed card)': [
            'Basic (and reversed card)',
            ['Front', 'Back'],
            ['Card 1', 'Card 2'],
          ],
          'Basic (optional reversed card)': [
            'Basic (optional reversed card)',
            ['Front', 'Back', 'Add Reverse'],
            ['Card 1', 'Card 2'],
          ],
          'Basic (type in the answer)': [
            'Basic (type in the answer)',
            ['Front', 'Back'],
            ['Card 1'],
          ],
          'Cloze': [
            'Cloze',
            ['Text', 'Back Extra'],
            ['Cloze'],
          ],
          'Image Occlusion': [
            'Image Occlusion',
            ['Occlusion', 'Image', 'Header', 'Back Extra', 'Comments'],
            ['Image Occlusion'],
          ],
        },
      );
    });

    test('built-in names in German', () {
      final german = lookupAppLocalizations(const Locale('de'));
      final names = builtinNames(german);
      expect(names.deck, 'Standard');
      final basic = names.noteTypes.first;
      expect(
        [basic.name, basic.fields, basic.cardTypes],
        [
          'Einfach',
          ['Vorderseite', 'Rückseite'],
          ['Karte 1'],
        ],
      );
      expect(stockTypeName(german, 'Image Occlusion'), 'Bildverdeckung');
      expect(stockTypeName(german, 'Something new'), 'Something new');
    });

    test('interval labels use German formatting', () {
      final german = lookupAppLocalizations(const Locale('de'));
      expect(formatIntervalLabel(const Duration(days: 3), l10n: german), '3 T');
      expect(
        formatIntervalLabel(const Duration(days: 36), l10n: german),
        '1,2 Mon.',
      );
    });
  });
}
