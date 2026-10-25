# release-packaging Specification

## Purpose

Defines the release packages `tool/package.sh` builds for each platform: their names, contents, the host each builds on, and how users install, upgrade and remove them.

## Requirements

### Requirement: Existing packages unchanged
Adding the Windows installer and the AppImage SHALL NOT change the existing targets. `linux`, `apk`, `appbundle`, `web`, `windows`, `macos` and `ios` SHALL keep their target names, file names and contents, and a run that names only existing targets SHALL produce exactly the packages it produced before.

#### Scenario: Windows zip alone
- **WHEN** a developer runs `tool/package.sh windows` on Windows
- **THEN** `dist/` gets `cardflip-<v>-windows-<arch>.zip` with the app folder and the Visual C++ runtime DLLs, and no installer

#### Scenario: Linux tarball alone
- **WHEN** a developer runs `tool/package.sh linux` on Linux
- **THEN** `dist/` gets `cardflip-<v>-linux-<arch>.tar.gz` with `bundle/`, `packaging/` and `install.sh`, and no AppImage

### Requirement: Windows installer package
The `windows-installer` target SHALL build, on a Windows host, a single installer `cardflip-<version><suffix>-windows-<arch>-setup.exe` in `dist/`, versioned like the other packages. It SHALL install the same app files as the Windows zip, including the Visual C++ runtime DLLs.

#### Scenario: Installer built
- **WHEN** a developer runs `tool/package.sh windows-installer` on an x64 Windows machine without a version tag
- **THEN** `dist/` gets `cardflip-<pubspec version>-dev.<commit>-windows-x64-setup.exe` and `SHA256SUMS` lists it

#### Scenario: Runs without the redistributable
- **WHEN** a user installs Cardflip with the installer on a Windows machine without the Visual C++ redistributable
- **THEN** Cardflip starts

### Requirement: Installer install scope
The installer SHALL install for the current user without administrator rights by default, under the user's programs folder. It SHALL offer installing for all users instead, which asks for administrator rights.

#### Scenario: Standard user
- **WHEN** a user without administrator rights runs the installer with the defaults
- **THEN** Cardflip is installed for that user without an elevation prompt

#### Scenario: All users
- **WHEN** a user chooses to install for all users
- **THEN** Windows asks for administrator rights and Cardflip is installed under Program Files for every user

### Requirement: Installer shortcuts
The installer SHALL add a "Cardflip" entry with the app icon to the Start menu, SHALL offer a desktop shortcut (off by default), and SHALL offer to start Cardflip when it finishes.

#### Scenario: Start menu
- **WHEN** the installation finishes
- **THEN** the Start menu has a "Cardflip" entry with the Cardflip icon that starts the app

#### Scenario: Desktop shortcut
- **WHEN** the user ticks the desktop shortcut option during installation
- **THEN** a "Cardflip" shortcut is placed on the desktop

### Requirement: Installer uninstall keeps the collection
The installer SHALL register Cardflip in Windows' installed apps list with its name, version and icon, and uninstalling SHALL remove the installed app files and shortcuts. Uninstalling SHALL NOT delete the user's collection, settings, profiles or backups.

#### Scenario: Uninstall
- **WHEN** the user uninstalls Cardflip from "Installed apps"
- **THEN** the program folder and shortcuts are removed and the collection is still there when Cardflip is installed again

### Requirement: Installer upgrades in place
Running a newer installer over an existing installer-made install SHALL replace the installed version in the same location, keep the user's data and shortcuts, and leave a single Cardflip entry in the installed apps list. When Cardflip is running, the installer SHALL ask to close it before replacing files.

#### Scenario: Upgrade
- **WHEN** Cardflip 1.0.0 was installed with the installer and the user runs the 1.1.0 installer
- **THEN** the installed apps list shows one Cardflip entry at version 1.1.0 and the collection is unchanged

#### Scenario: App running
- **WHEN** Cardflip is running and the user starts the installer
- **THEN** the installer asks to close Cardflip before installing, instead of failing on files in use

### Requirement: AppImage package
The `appimage` target SHALL build, on a Linux host, a single executable `cardflip-<version><suffix>-linux-<arch>.AppImage` in `dist/`, versioned like the other packages. It SHALL run without installation and without root, and SHALL embed the desktop entry and icon so AppImage integration tools show Cardflip's name and icon.

#### Scenario: AppImage built
- **WHEN** a developer runs `tool/package.sh appimage` on an x86_64 Linux machine
- **THEN** `dist/` gets `cardflip-<v>-linux-x64.AppImage` and `SHA256SUMS` lists it

#### Scenario: Run directly
- **WHEN** a user marks the AppImage executable and runs it
- **THEN** Cardflip starts and its window shows the Cardflip name and icon

### Requirement: AppImage bundles media libraries
The AppImage SHALL contain libmpv and the libraries it needs that a desktop Linux system does not provide, so audio and video play on a system without mpv or libmpv installed. It SHALL use the host's glibc, GTK and graphics drivers.

#### Scenario: No mpv on the host
- **WHEN** the AppImage runs on a supported distribution where mpv and libmpv are not installed and the user studies a card with audio
- **THEN** the audio plays

### Requirement: AppImage shares the Linux collection
The AppImage SHALL keep the collection in the same location as the Linux tarball install, so a user who switches between them keeps one collection.

#### Scenario: Switch from the tarball
- **WHEN** a user who studied with the tarball install starts the AppImage
- **THEN** the AppImage opens the same decks and review history

### Requirement: Shared builds per run
When a run names an addition together with its existing sibling (`linux` with `appimage`, `windows` with `windows-installer`), the app SHALL be built once and both packages SHALL contain the same build.

#### Scenario: Both Windows packages
- **WHEN** a developer runs `tool/package.sh windows windows-installer`
- **THEN** `dist/` gets both the zip and the setup executable from one Windows build

### Requirement: Packaging tools checked first
Before running the checks or building anything, `tool/package.sh` SHALL stop with a message naming the missing tool when a requested target's packaging tool is unavailable, and SHALL stop when a target is requested on a host it doesn't build on.

#### Scenario: Inno Setup missing
- **WHEN** a developer runs `tool/package.sh windows-installer` on Windows without Inno Setup installed
- **THEN** the script exits with an error naming Inno Setup before building

#### Scenario: Wrong host
- **WHEN** a developer runs `tool/package.sh appimage` on macOS
- **THEN** the script exits with an error that AppImages build only on Linux
