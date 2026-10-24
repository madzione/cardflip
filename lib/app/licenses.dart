import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';

/// The project's public repository, which "Source code" opens.
const repositoryUrl = 'https://github.com/madzione/cardflip';

/// Adds the licenses Flutter doesn't collect from the Dart packages to the
/// license page: the code taken from other projects, the Go modules
/// compiled into the Go library and, where the app bundles them, libmpv and
/// FFmpeg. Cardflip's own license is collected by Flutter, as that of the
/// `cardflip` package. Called once, before the app starts.
void registerLicenses() => LicenseRegistry.addLicense(_licenses);

/// Code in the repository that comes from other projects, with the part of
/// Cardflip it covers (see THIRD_PARTY_NOTICES.md).
const _portedCode = [
  (
    name: 'fsrs-rs',
    file: 'LICENSES/BSD-3-Clause-fsrs-rs.txt',
    intro:
        'The FSRS optimizer and simulator in Cardflip are ported from '
        'fsrs-rs (https://github.com/open-spaced-repetition/fsrs-rs): the '
        'model, its parameter bounds, the training procedure and the review '
        'simulation.',
  ),
  (
    name: 'Anki',
    file: 'LICENSES/AGPL-3.0.txt',
    intro:
        "The Image Occlusion note type's card template and the shape colour "
        'variables of its styling are reproduced unchanged from Anki '
        '(https://github.com/ankitects/anki). Copyright: Ankitects Pty Ltd '
        'and contributors. Licensed under the GNU Affero General Public '
        'License, version 3 or later.',
  ),
  (
    name: 'fluffy',
    file: 'LICENSES/MIT-fluffy.txt',
    intro:
        "Cardflip's message layer between Dart and Go is written by fluffy "
        '(https://github.com/madzione/fluffy).',
  ),
];

/// The repositories with the source and build scripts of the libmpv and
/// FFmpeg builds media_kit bundles, by platform. Linux uses the system's
/// libmpv, the web the browser's media element.
const _mediaBuilds = {
  TargetPlatform.windows:
      'https://github.com/media-kit/libmpv-win32-video-build',
  TargetPlatform.android:
      'https://github.com/media-kit/libmpv-android-video-build',
  TargetPlatform.macOS: 'https://github.com/media-kit/libmpv-darwin-build',
  TargetPlatform.iOS: 'https://github.com/media-kit/libmpv-darwin-build',
};

/// Not cached: the texts are read only while the license page is open.
Future<String> _asset(String key) => rootBundle.loadString(key, cache: false);

Stream<LicenseEntry> _licenses() async* {
  for (final code in _portedCode) {
    yield LicenseEntryWithLineBreaks([
      code.name,
    ], '${code.intro}\n\n${await _asset(code.file)}');
  }

  final modules = jsonDecode(await _asset('assets/licenses/go_modules.json'));
  for (final m in (modules as List).cast<Map<String, dynamic>>()) {
    yield LicenseEntryWithLineBreaks([
      m['module'] as String,
    ], m['license'] as String);
  }

  final build = kIsWeb ? null : _mediaBuilds[defaultTargetPlatform];
  if (build != null) {
    final gpl = await _asset('LICENSE');
    yield LicenseEntryWithLineBreaks(
      const ['libmpv'],
      'The libmpv bundled with Cardflip on this platform is built without '
      'GPL code and is licensed under the GNU Lesser General Public '
      'License, version 2.1 or later. Its source and build scripts: '
      '$build\n\n${await _asset('LICENSES/LGPL-2.1.txt')}',
    );
    yield LicenseEntryWithLineBreaks(
      const ['FFmpeg'],
      'The FFmpeg bundled with Cardflip on this platform is built without '
      'GPL and nonfree code and is licensed under the GNU Lesser General '
      'Public License, version 3 or later, which follows together with '
      'the GNU General Public License, version 3, that it builds on. Its '
      'source and build scripts: $build\n\n'
      '${await _asset('LICENSES/LGPL-3.0.txt')}\n\n$gpl',
    );
  }
}
