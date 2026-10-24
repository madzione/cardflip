# Building Cardflip

The UI is Flutter. Everything else lives in a Go core: the collection,
scheduling, card rendering, search, import, backup and storage. The two are
joined with [fluffy](https://github.com/madzione/fluffy), which builds the
Go code as a native library (dart:ffi) or as WebAssembly (web) and generates
the typed Dart API in `lib/golang/api.g.dart`.

## Prerequisites

- **Go 1.27** (`go version`).
- **FVM** with **Flutter 3.47.0**. The project is pinned in `.fvmrc`; run
  `fvm install` once in the repository. Use `fvm flutter …`, never a global
  `flutter`.
- **fluffy** (0.1.0 or newer) on `PATH`. The platform build hooks call
  `fluffy` during every native build, so `~/go/bin` must be on the `PATH`
  that the build sees:

  ```sh
  go install github.com/madzione/fluffy@latest
  export PATH="$HOME/go/bin:$PATH"
  ```

- Platform toolchains, only for the platforms you build:
  - **Linux**: a C compiler, CMake, Ninja, pkg-config and the GTK 3
    development files (`clang cmake ninja-build pkg-config libgtk-3-dev`),
    and libmpv for audio and video (`libmpv-dev`). See the
    [README](README.md#linux) for the packages the app needs at runtime.
  - **Android**: the Android SDK with the NDK (set `ANDROID_NDK_HOME`, or
    pass `--ndk-version` to `fluffy lib android`).
  - **macOS / iOS**: Xcode on a Mac.
  - **Windows**: Visual Studio with C++ on Windows, plus MinGW-w64 `gcc`
    for the Go library (cgo).
  - **Web**: nothing extra.

## Running

`fluffy run` generates the Dart API, builds the Go library and starts
`flutter run`:

```sh
fluffy run -d linux      # Linux desktop
fluffy run -d chrome     # web; fluffy builds web/golang.wasm first
fluffy run -d macos      # on a Mac
fluffy run -d windows    # on Windows
fluffy run -d <device>   # Android or iOS device/emulator, see `fvm flutter devices`
```

After changing Go code, run `fluffy generate` (or keep `fluffy generate
--watch` running) so the Dart API matches.

After adding, removing or updating a Go module, regenerate the list of Go
module licenses that the app's license page shows, and commit it:

```sh
cd golang && go run ./internal/licensegen   # writes assets/licenses/go_modules.json
```

It fails when a linked module has no license file.

## Release builds

```sh
fluffy build linux
fluffy build web
fluffy build apk
fluffy build macos
fluffy build windows
fluffy build ios
```

On Linux, `tool/install_linux.sh` installs the release bundle to
`~/.local/lib/cardflip` with a desktop entry and icons, which Wayland
compositors need to show the app's name and icon (`--uninstall` removes it,
`PREFIX=/usr/local` installs system-wide).

App icons for every platform are regenerated from `assets/logo/` with
`tool/generate_icons.sh`.

## Packages

`tool/package.sh` runs `tool/check.sh`, builds the given targets and puts
versioned packages and a `SHA256SUMS` into `dist/`:

```sh
tool/package.sh linux apk web     # on Linux
tool/package.sh windows           # on Windows, from Git Bash
tool/package.sh macos ios         # on a Mac
```

| Target | Package |
|---|---|
| `linux` | `cardflip-<v>-linux-<arch>.tar.gz`: the bundle, the desktop entry and icons, and `install.sh` (the same script as `tool/install_linux.sh`) |
| `apk` | `cardflip-<v>-android-<abi>.apk`, one per ABI |
| `appbundle` | `cardflip-<v>-android.aab` for Google Play |
| `web` | `cardflip-<v>-web.tar.gz`; the server must send `.wasm` as `application/wasm`. Set `WEB_BASE_HREF` when the app isn't served from `/` |
| `windows` | `cardflip-<v>-windows-<arch>.zip`, with the Visual C++ runtime DLLs |
| `macos` | `cardflip-<v>-macos.dmg` |
| `ios` | `cardflip-<v>-ios.ipa` |

The version is taken from a `vX.Y.Z` tag on the current commit. Without
one, the version in `pubspec.yaml` is used and the file names get
`-dev.<commit>`. The build number is the number of commits. `VERSION` and
`BUILD_NUMBER` override both. `--no-check` skips `tool/check.sh`.

**Android signing.** Release builds are signed with the key in
`android/key.properties` (never commit it or the keystore):

```properties
storeFile=/home/you/keys/cardflip-upload.jks
storePassword=…
keyAlias=upload
keyPassword=…
```

Create the keystore once with `keytool -genkeypair -v -keystore
cardflip-upload.jks -keyalg RSA -keysize 2048 -validity 10000 -alias upload`
and keep a backup: an app signed with a different key can't update an
installed one. Without `key.properties` Gradle falls back to the debug key,
and `tool/package.sh` refuses to package Android unless
`ALLOW_DEBUG_SIGNING=1` is set.

**macOS signing.** `MACOS_SIGN_IDENTITY` (a "Developer ID Application"
certificate) signs the app and disk image; `MACOS_NOTARY_PROFILE` (made with
`xcrun notarytool store-credentials`) also notarizes and staples it. Without
them the disk image only opens on other Macs after the user overrides
Gatekeeper.

**iOS signing.** The `.ipa` export needs a signing team set in Xcode, or an
`ExportOptions.plist` passed as `IOS_EXPORT_OPTIONS`.

## Platform status

Built and smoke-tested (release builds open to the deck list on a fresh
collection):

| Target | Command | Checked on |
|---|---|---|
| Linux | `fluffy build linux` | x86_64, KWin/Wayland (via XWayland) |
| Web | `fluffy build web` | headless Chromium 153 |
| Android | `fluffy build apk` | x86_64 emulator; the APK also contains arm64-v8a and armeabi-v7a |

Not built here, because they need their own host:

- **Windows** (`fluffy build windows`): a Windows machine with Visual Studio
  (Desktop development with C++) and MinGW-w64 `gcc` on `PATH` for the Go
  library. Cross-building the Go library from Linux is possible with
  `fluffy lib windows` and `x86_64-w64-mingw32-gcc`, but the Flutter runner
  itself must be built on Windows.
- **macOS** (`fluffy build macos`) and **iOS** (`fluffy build ios`): a Mac
  with Xcode and its command line tools; for iOS also a signing identity
  (or `--no-codesign` for a build check). fluffy builds the Go library as an
  Xcode build phase.

media_kit's native libraries make the app larger: the release APK is about
119 MB.

## Checks and tests

```sh
tool/check.sh
```

runs every gate: `go vet` for native and `js/wasm`, `go test`, `go run
./internal/licensegen -check` (the Go license list matches the linked
modules), `fluffy generate --check` (the committed bindings are up to
date), `fvm flutter analyze` and `fvm flutter test` (widget tests against a
fake Go side).

Integration tests use the real Go library:

```sh
fvm flutter test integration_test -d linux
```

The same flow runs in Chrome against the Go core compiled to WebAssembly,
with the collection in IndexedDB (needs `chromedriver` matching your
Chrome or Chromium):

```sh
tool/web_integration.sh            # headless
HEADED=1 tool/web_integration.sh   # with a visible browser window
```

The browser performance check (50,000 cards: time to first rows and frame
timings while scrolling) runs as a profile build:

```sh
fvm flutter drive --profile --driver test_driver/integration_test.dart \
  --target integration_test/perf/browse_perf.dart -d linux
```

It is kept out of `flutter test integration_test`, because the Flutter tool
cannot launch a desktop app twice in one `flutter test` run.

Never run `dart format` over `lib/golang/`: `api.g.dart` and `src/` belong to
fluffy, and `fluffy generate --check` fails when they change.

## Anki test fixtures

The `.apkg` files in `golang/importer/apkg/testdata/` are made with the
official `anki` Python package by a committed script:

```sh
pip install anki
tool/make_apkg_fixtures.py          # regenerate the fixtures
tool/make_apkg_fixtures.py --check  # check that a fresh build has identical content
```

See `golang/importer/apkg/testdata/README.md` for what each fixture contains.

## Storage format

The collection is a JSON snapshot plus an append-only journal (one entry per
change, synced to disk), monthly review log segments, a deletion log of
every deleted note and media files. A crash loses at most the change being
written. On native platforms it lives in the application support directory
under `collection/`; on the web in IndexedDB.

Backups are zip files with `manifest.json`, `collection.json`,
`revlog.jsonl`, `deleted.jsonl` and `media/`.

The MCP settings (`mcp.json`) and the device's sync identity, trusted
networks and paired devices (`sync.json`) are in the data directory; each
profile's sync links are in `sync-links.json` in its directory.

## Project layout

```
golang/              Go module: the core and the bridge services (svc_*.go)
  core/              collection model, ops, journal, queues, change events
  scheduler/         FSRS (go-fsrs), learning steps, fuzz, day boundary
  render/            templates, cloze, sanitizing, CSS inlining
  search/            browser query language
  storage/           files / IndexedDB backends, generations, journal
  importer/csv       CSV/TSV import
  importer/apkg      Anki packages (with an in-repo read-only SQLite reader)
  backup/            .cardflip archives
lib/
  app/               bootstrap, theme, adaptive scaffold, errors
  data/              change notifications and shared state
  features/          decks, study, browse, editor, import, settings
  widgets/           CardView and helpers
  golang/            generated Dart API (do not edit)
test/                widget tests (fake Go side)
integration_test/    tests against the real Go library
tool/                check.sh, fixture generator, icon and install scripts
```
