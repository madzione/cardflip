# Spec Delta

## ADDED Requirements

### Requirement: License and source in About
Settings' About section and the licenses page SHALL show Cardflip's copyright notice and state that Cardflip is free software under the GNU GPL version 3 or later and comes with no warranty, and the About section SHALL offer "Source code", which opens the project's public repository in the browser.

#### Scenario: License named
- **WHEN** the user opens Settings
- **THEN** the About section shows the version, the copyright notice, that Cardflip is licensed under the GNU GPL version 3 or later and that it comes with no warranty

#### Scenario: Source code opens the repository
- **WHEN** the user taps "Source code" in the About section
- **THEN** the project's repository opens in the system browser, or in a new tab in the web build

### Requirement: Open-source licenses page
Settings' About section SHALL offer "Open-source licenses", which opens a page that lists every component the running app contains, by name, and shows each one's license text when it is selected. The page SHALL work offline, on every platform including the web build, and in the interface language.

#### Scenario: Open the page
- **WHEN** the user taps "Open-source licenses" in the About section
- **THEN** a page opens with Cardflip's name, version and license at the top and a list of components below

#### Scenario: Read a license
- **WHEN** the user selects `flutter_quill` on the licenses page
- **THEN** its MIT license text is shown

#### Scenario: Offline
- **WHEN** the device has no network connection and the user opens the licenses page
- **THEN** all components and their license texts are shown

#### Scenario: German interface
- **WHEN** the interface language is German and the user opens Settings
- **THEN** the entry reads "Open-Source-Lizenzen" and the page's own labels are German

### Requirement: Licenses page contents
The licenses page SHALL list Cardflip itself with the GNU GPL version 3 text; the code taken from other projects (fsrs-rs, the Anki Image Occlusion template, fluffy) with their licenses; every Dart and Flutter package, including the Flutter engine; every Go module linked into the app, and the Go standard library; and, on the platforms where the app bundles them, libmpv and FFmpeg.

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

### Requirement: Licenses list kept current
The list of Go module licenses SHALL match the modules the app links on every platform. The verification gates SHALL fail when a module is added, removed or updated without the list being regenerated, or when a linked module has no license file.

#### Scenario: Dependency added
- **WHEN** a developer adds a Go module that the app uses and runs the verification gates without regenerating the list
- **THEN** the gates fail and name the license list as out of date

#### Scenario: Module without license
- **WHEN** a linked Go module has no license file
- **THEN** generating the list fails and names the module
