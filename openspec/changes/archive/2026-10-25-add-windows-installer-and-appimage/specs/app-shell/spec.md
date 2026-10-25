# Spec Delta

## MODIFIED Requirements

### Requirement: Licenses page contents
The licenses page SHALL list Cardflip itself with the GNU GPL version 3 text; the code taken from other projects (fsrs-rs, the Anki Image Occlusion template, fluffy) with their licenses; every Dart and Flutter package, including the Flutter engine; every Go module linked into the app, and the Go standard library; and, on the platforms and in the packages where the app bundles them, libmpv and FFmpeg.

#### Scenario: Cardflip and ported code
- **WHEN** the user opens the licenses page
- **THEN** it lists Cardflip once, as its package `cardflip`, with the GPL version 3, "fsrs-rs" with the BSD 3-Clause license, "Anki" with the GNU AGPL version 3 and "fluffy" with the MIT license

#### Scenario: Go modules
- **WHEN** the user opens the licenses page
- **THEN** it lists `github.com/klauspost/compress`, `github.com/open-spaced-repetition/go-fsrs/v3` and the Go standard library, each with its license text

#### Scenario: Web build
- **WHEN** the user opens the licenses page in the web build
- **THEN** it lists the Go modules compiled into the web build's Go code, the same as the native apps

#### Scenario: Media libraries on Windows
- **WHEN** the user opens the licenses page in the Windows app
- **THEN** it lists libmpv and FFmpeg with the licenses of the bundled builds

#### Scenario: Media libraries on Linux
- **WHEN** the user opens the licenses page in the Linux app, which uses the system's libmpv
- **THEN** libmpv and FFmpeg are not listed

#### Scenario: Media libraries in the AppImage
- **WHEN** the user opens the licenses page in the Linux app started from the AppImage, which bundles libmpv
- **THEN** it lists libmpv and FFmpeg with the GNU GPL version 3 text, the versions bundled and where their source is published
