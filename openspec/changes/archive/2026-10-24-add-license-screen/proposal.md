# Proposal

## Why

Cardflip is now licensed under GPL-3.0-or-later, and every build that goes out (release archives, APKs, app stores, the web build) has to carry the license texts of what it contains: its own license, the code it took from fsrs-rs, Anki and fluffy, the Dart packages, the Go modules and, where bundled, libmpv and FFmpeg. The app has no place that shows any of them today. Flutter already collects the Dart package licenses, but nothing shows them, and it knows nothing about the Go modules compiled into the Go library and `golang.wasm`.

## What Changes

- Settings' About section gets an "Open-source licenses" entry that opens a license page listing every component with its license text, and a "Source code" entry that opens the project's repository.
- The About entry names Cardflip's license (GPL-3.0-or-later).
- The license page lists, next to the Dart and Flutter packages Flutter already collects:
  - Cardflip itself (GPL-3.0-or-later);
  - the code in the repository that comes from other projects: fsrs-rs (BSD-3-Clause), the Anki Image Occlusion template (AGPL-3.0-or-later) and fluffy's message layer (MIT);
  - every Go module linked into the app, plus the Go standard library and `wasm_exec.js`;
  - libmpv and FFmpeg on the platforms where the app bundles them.
- A generator writes the Go module licenses from the module graph into a file that ships with the app. A new check gate fails when that file is out of date, for example after a dependency is added or updated, or when a linked module has no license file.
- The license texts of the projects in `THIRD_PARTY_NOTICES.md` move into `LICENSES/`, so the page and the notices use the same files.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `app-shell`: adds a requirement for the open-source licenses page and the source code link in Settings' About section.

## Impact

- `lib/features/settings/settings_screen.dart`: the About section's new entries.
- New `lib/app/licenses.dart`: registers Cardflip's own, the ported-code and the Go module licenses with Flutter's `LicenseRegistry` at startup.
- `lib/l10n/app_en.arb`, `lib/l10n/app_de.arb`: strings for the new entries.
- `pubspec.yaml`: the license files become assets.
- New Go generator under `golang/internal/` and its generated `assets/licenses/go_modules.json`; `tool/check.sh` gains a gate that checks it is current.
- `LICENSES/`, `THIRD_PARTY_NOTICES.md`: the texts are split into files that the notices link to.
- `test/settings_test.dart`: tests for the entries and the registered licenses.
- No change to data, sync, backups or the Go API.
