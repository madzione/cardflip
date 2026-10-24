# Design

## Context

- Settings' About section (`lib/features/settings/settings_screen.dart`) is a single `ListTile` with the version from `AppStatus`. Nothing in the app calls `showLicensePage` or `LicenseRegistry`.
- At build time Flutter writes the `LICENSE` file of every Dart package (and the engine's third-party licenses) into the `NOTICES` asset, which `LicenseRegistry` reads. The `fluffy_golang` path package (`golang/`) has no `LICENSE`, so it adds nothing.
- fluffy builds `golang/` (package `main`) with plain `go build -trimpath` and no build tags: `CGO_ENABLED=1` for linux, windows, darwin, ios and android, `GOOS=js GOARCH=wasm CGO_ENABLED=0` for the web. The Go modules it links are those of `go list -deps .` for each of these targets, and none of their licenses reach `NOTICES`.
- `media_kit_libs_video` downloads libmpv and FFmpeg builds for Windows, macOS, iOS and Android and bundles them. Its own `LICENSE` is MIT and says nothing about them. Linux uses the system's libmpv, and the web uses the browser's media element.
- `tool/check.sh` runs the verification gates. Generated files are committed and checked for staleness (`fluffy generate --check`).
- `test/no_literal_text_test.dart` allows the literal `Cardflip` in `settings_screen.dart`; every other UI string comes from the ARB files.

## Goals / Non-Goals

**Goals:**
- One place for each license text: the files in `LICENSE` and `LICENSES/` are used both by `THIRD_PARTY_NOTICES.md` and by the app.
- A Go license list that cannot go stale without a gate failing.

**Non-Goals:**
- A custom-designed license page. Flutter's `LicensePage` is used as it is.
- License texts for the platform toolchains' own runtime pieces (the Android, iOS and Windows SDKs and their C runtimes). Their licenses don't require a notice in the app.
- Checking the Dart package licenses. Flutter collects them already.

## Decisions

### D1: Flutter's `LicensePage`, fed through `LicenseRegistry`

"Open-source licenses" calls `showLicensePage` with the application name, version, icon and a short localized legalese with the copyright notice ("Copyright © 2026 madzione"), the license and that there is no warranty, the Appropriate Legal Notices of GPL-3.0 section 0. Everything the page doesn't already get from `NOTICES` is added with `LicenseRegistry.addLicense` in a new `lib/app/licenses.dart`, called from `main()` before `runApp`. The page is localized by `MaterialLocalizations`, which the app already provides for English and German, and it reads only bundled assets, so it works offline and on the web.

*Alternative:* a hand-written page reading `NOTICES` and the extra files. This duplicates what Flutter already has (package list, grouping, search on wide screens) for no gain.

### D2: The Go license list is generated, committed and checked

A small program, `golang/internal/licensegen` (package `main`, so it isn't linked into the app), does three things:
- It runs `go list -deps -json .` in `golang/` for each target fluffy builds (linux/amd64, windows/amd64, darwin/arm64, ios/arm64, android/arm64 with `CGO_ENABLED=1`; js/wasm with `CGO_ENABLED=0`) and takes the union of the non-standard-library modules, leaving out the main module.
- For each module, it collects the `LICENSE*`, `LICENCE*`, `COPYING*` and `NOTICE*` files in the module root and in the directory of every linked package of that module. This catches vendored code with its own license, such as `klauspost/compress`'s subpackages.
- It adds one "Go" entry from `$(go env GOROOT)/LICENSE`, which covers the standard library and `wasm_exec.js`.

It writes `assets/licenses/go_modules.json`, a list of `{module, version, license}` sorted by module path. With `-check` it writes nothing and exits non-zero, naming the file, when the result differs from the committed file. It fails in both modes when a module has no license file, naming the module. `tool/check.sh` runs `go run ./internal/licensegen -check` right after `go test`.

The version is part of each entry, so every dependency update regenerates the file and its license text is checked again.

*Alternatives:*
- `github.com/google/go-licenses`: a new tool to install, and it classifies licenses, which isn't needed here; it also handles cross-target unions poorly.
- Embedding the list in the Go library with `go:embed` and serving it over the fluffy API: the list would always match the build. But it costs an API call and a few dozen KB in every Go binary, and the check gate keeps the asset just as accurate.
- One list per target: the targets differ by at most a few `golang.org/x/sys` files inside the same module, so one union list is shown everywhere.

### D3: Cardflip's own and the ported-code licenses come from `LICENSE` and `LICENSES/`

The texts now inlined in `THIRD_PARTY_NOTICES.md` move into files, and the notices link to them:

- `LICENSE`: GPL-3.0. Flutter collects it into `NOTICES` as the license of the app's own package, `cardflip`, so `licenses.dart` doesn't register it again.
- `LICENSES/BSD-3-Clause-fsrs-rs.txt`, `LICENSES/MIT-fluffy.txt`, `LICENSES/AGPL-3.0.txt`.
- `LICENSES/LGPL-2.1.txt` (libmpv) and `LICENSES/LGPL-3.0.txt` (FFmpeg, built with `--enable-version3`); no bundled build is GPL (see D4).

`pubspec.yaml` lists `LICENSE`, `LICENSES/` and `assets/licenses/` as assets. `lib/app/licenses.dart` has a fixed table of entry name → file. The fsrs-rs, Anki and fluffy entries each start with one line saying which part of Cardflip they cover, taken from the notices.

### D4: libmpv and FFmpeg only where they are bundled

The libmpv and FFmpeg entries are added only when the app is not on the web and the platform is Windows, macOS, iOS or Android. Each entry names the build repository with its source and build scripts, as `THIRD_PARTY_NOTICES.md` does. That link is what makes the corresponding source available for these binaries.

media-kit's darwin builds are "default" (LGPL) builds. For the Windows and Android builds, the license is read from each build's configuration during implementation, and the matching license text is shipped. Either way it is compatible with GPL-3.0.

### D5: About section entries

Below the existing version tile:

- **Open-source licenses**: opens the page (D1).
- **Source code**: opens the repository URL, a constant in `lib/app/licenses.dart`. It uses `launchUrl` with `LaunchMode.externalApplication`, through an overridable `openUrl` like `DeckOverviewScreen.openUrl`, so tests can catch the call.

The version tile's subtitle (`aboutCardflip`) gains the license line. New ARB keys go into `app_en.arb` and `app_de.arb`: `openSourceLicenses` ("Open-source licenses" / "Open-Source-Lizenzen"), `sourceCode` ("Source code" / "Quellcode") and `licenseLegalese`.

## Risks / Trade-offs

- [The license list depends on the module cache; `go list` needs the modules downloaded] → The gate runs after `go test`, which downloads them. The generator runs `go mod download` itself when a module directory is missing.
- [A new fluffy target or build flag (e.g. tags) would change the linked set without the generator knowing] → The target list sits in one place in the generator, with a comment pointing to fluffy's build code. A fluffy update that adds a target is the moment to update it.
- [A Go toolchain update changes `GOROOT/LICENSE`] → It hasn't changed in years. If it does, the gate fails and regenerating fixes it, which is the desired result.
- [`NOTICES` might miss something Flutter doesn't collect, such as native code that plugins pull in through CocoaPods or Gradle] → Out of scope here, except libmpv and FFmpeg (D4). Worth a review before the first app store release.
- [The asset adds about 100 KB] → Negligible next to the Go library and media libraries.

## Open Questions

- The repository URL for "Source code" isn't published yet. The plan assumes `https://github.com/madzione/cardflip`, defined as one constant so it's easy to change before release.
