import 'package:flutter/foundation.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/app/licenses.dart';

/// The registered license texts by package name, with only the licenses
/// registerLicenses adds.
Future<Map<String, String>> registered() async {
  LicenseRegistry.reset();
  registerLicenses();
  final texts = <String, String>{};
  await for (final entry in LicenseRegistry.licenses) {
    final text = entry.paragraphs.map((p) => p.text).join('\n');
    for (final package in entry.packages) {
      texts[package] = text;
    }
  }
  return texts;
}

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  tearDown(LicenseRegistry.reset);

  test('the ported code is listed with its licenses', () async {
    final texts = await registered();
    // Flutter lists Cardflip itself, as the cardflip package.
    expect(texts.keys, isNot(contains('Cardflip')));
    expect(texts['fsrs-rs'], contains('BSD 3-Clause License'));
    expect(texts['fsrs-rs'], contains('Open Spaced Repetition'));
    expect(texts['Anki'], contains('GNU AFFERO GENERAL PUBLIC LICENSE'));
    expect(texts['Anki'], contains('Ankitects Pty Ltd'));
    expect(texts['fluffy'], contains('MIT License'));
    expect(texts['fluffy'], contains('madzione'));
  });

  test('the Go modules and the Go standard library are listed', () async {
    final texts = await registered();
    expect(texts['github.com/klauspost/compress'], contains('Klaus Post'));
    expect(texts['github.com/open-spaced-repetition/go-fsrs/v3'], isNotEmpty);
    expect(texts['Go'], contains('The Go Authors'));
    expect(texts.keys, isNot(contains('modernc.org/sqlite')));
  });

  test('libmpv and FFmpeg are listed on Windows', () async {
    debugDefaultTargetPlatformOverride = TargetPlatform.windows;
    addTearDown(() => debugDefaultTargetPlatformOverride = null);
    final texts = await registered();
    expect(texts['libmpv'], contains('libmpv-win32-video-build'));
    expect(texts['libmpv'], contains('Version 2.1, February 1999'));
    expect(texts['FFmpeg'], contains('libmpv-win32-video-build'));
    expect(texts['FFmpeg'], contains('GNU LESSER GENERAL PUBLIC LICENSE'));
    expect(texts['FFmpeg'], contains('GNU GENERAL PUBLIC LICENSE'));
  });

  test('libmpv and FFmpeg are not listed on Linux', () async {
    debugDefaultTargetPlatformOverride = TargetPlatform.linux;
    addTearDown(() => debugDefaultTargetPlatformOverride = null);
    final texts = await registered();
    expect(texts.keys, isNot(contains('libmpv')));
    expect(texts.keys, isNot(contains('FFmpeg')));
  });
}
