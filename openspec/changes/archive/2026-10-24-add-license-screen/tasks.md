# Tasks

## 1. License files

- [x] 1.1 Move the fsrs-rs BSD and fluffy MIT texts out of `THIRD_PARTY_NOTICES.md` into `LICENSES/BSD-3-Clause-fsrs-rs.txt` and `LICENSES/MIT-fluffy.txt` and link them from the notices (D3); verify the notices' links resolve (`ls` each linked path) and that the copyright lines are unchanged (`grep` "Open Spaced Repetition" and "madzione" in `LICENSES/`)
- [x] 1.2 Determine the license of the libmpv/FFmpeg builds media-kit bundles for Windows and Android from each build repository's configuration (darwin is the LGPL "default" build), add `LICENSES/LGPL-2.1.txt` (and `LICENSES/GPL-2.0.txt` if one is GPL) from gnu.org, and record the result in `THIRD_PARTY_NOTICES.md`; verify by comparing the files' md5 with gnu.org's copies

## 2. Go license list

- [x] 2.1 Write `golang/internal/licensegen` (D2): the union of `go list -deps -json .` over fluffy's six targets, the license files of each module root and linked package directory, a "Go" entry from `GOROOT/LICENSE`, written sorted to `assets/licenses/go_modules.json`, and `-check` mode; generate the file and verify it lists `github.com/klauspost/compress` (with its subpackage licenses), `github.com/open-spaced-repetition/go-fsrs/v3` and "Go", and no test-only module such as `modernc.org/sqlite` or `github.com/google/go-cmp`
- [x] 2.2 Add Go tests for the generator's collection and comparison on a fixture module tree (missing license file fails naming the module; a changed entry makes `-check` fail naming the file; a subpackage license is included); verify `go test ./internal/licensegen` passes
- [x] 2.3 Add `go run ./internal/licensegen -check` to `tool/check.sh` after `go test`, and document regenerating the list after Go dependency changes in `BUILDING.md` next to `fluffy generate`; verify the gate passes, then fails after editing a version in the JSON by hand (restore afterwards)

## 3. Registering licenses

- [x] 3.1 Declare `LICENSE`, `LICENSES/` and `assets/licenses/` as assets in `pubspec.yaml`; verify `fvm flutter analyze` is clean and the files appear in `build/flutter_assets` after a build
- [x] 3.2 Add `lib/app/licenses.dart` with `registerLicenses()` (fsrs-rs, Anki, fluffy, the Go modules, and libmpv/FFmpeg on Windows, macOS, iOS and Android only, per D3 and D4) and the repository URL constant, and call it from `main()`; add a test reading `LicenseRegistry.licenses` that checks the spec scenarios "Cardflip and ported code", "Go modules", "Media libraries on Windows" and "Media libraries on Linux" (via `debugDefaultTargetPlatformOverride`); verify `fvm flutter test` on that test passes

## 4. About section

- [x] 4.1 Add `openSourceLicenses`, `sourceCode` and `licenseLegalese` to `app_en.arb` and `app_de.arb`, add the license line to `aboutCardflip` in both, and regenerate with `fvm flutter gen-l10n`; verify `test/l10n_test.dart` passes
- [x] 4.2 Add the "Open-source licenses" entry (calling `showLicensePage` with name, version, icon and legalese) and the "Source code" entry (via an overridable `openUrl`) to the About section in `lib/features/settings/settings_screen.dart`; add tests to `test/settings_test.dart` for "License named", "Source code opens the repository", "Open the page", "Read a license" (`flutter_quill` shows MIT text) and "German interface"; verify `fvm flutter test test/settings_test.dart test/no_literal_text_test.dart` passes

## 5. Integration checks

- [x] 5.1 Run `tool/check.sh` and verify every gate passes, including the new license list gate
- [x] 5.2 Run the app on Linux (`fluffy run -d linux`) and in the web build (`fluffy run -d chrome`), open Settings → Open-source licenses, and verify Cardflip, fsrs-rs, Anki, fluffy, `github.com/klauspost/compress` and Go are listed with their texts, libmpv/FFmpeg are absent, and "Source code" opens the repository URL
