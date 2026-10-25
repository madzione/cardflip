# Design

## Context

- `tool/package.sh` validates its targets against the host before running `tool/check.sh`, then calls one `package_<target>` function per target. Each function runs its own `fluffy build …` and stages into a `mktemp -d` work directory; `SHA256SUMS` covers everything in `dist/`.
- `package_windows` copies `build/windows/<arch>/runner/Release`, adds `msvcp140.dll`, `vcruntime140.dll` and `vcruntime140_1.dll` from `System32`, and zips with PowerShell. It runs in Git Bash, so paths go through `cygpath`.
- `package_linux` tars `build/linux/<arch>/release/bundle`, `linux/packaging/` (desktop entry `dev.cardflip.cardflip.desktop`, hicolor icons) and `tool/install_linux.sh`.
- The Flutter Linux bundle must stay together: the `cardflip` executable finds `data/` and `lib/` next to itself (RUNPATH `$ORIGIN/lib`). The plugins are copied into the bundle from the build tree and keep its RUNPATH (`<repo>/linux/flutter/ephemeral`), which doesn't exist on other machines. `libmedia_kit_video_plugin.so` links `libmpv.so.2` from the system (mpv 0.35 or newer).
- The app runs system programs on Linux (`parecord`, `ffmpeg`, `spd-say`), so the environment it is started with reaches those children.
- The collection lives in the application support directory: `%APPDATA%\dev.cardflip\cardflip` on Windows (from the company and product names in `Runner.rc`), `~/.local/share/dev.cardflip.cardflip` on Linux. Neither is inside the install folder.
- `lib/app/licenses.dart` adds libmpv and FFmpeg entries per `TargetPlatform` from `_mediaBuilds`, which has no Linux entry. Cardflip's own `LICENSE` is the GPL version 3.

## Goals / Non-Goals

**Goals:**
- The new targets reuse the existing version, naming, check and checksum logic; the existing targets' functions keep their output.
- One build per platform per run, whichever packages of that platform are requested.

**Non-Goals:**
- Code signing of the installer or the AppImage. Unsigned installers show a SmartScreen warning; signing can be added later the way macOS signing is (optional environment variables).
- AppImage update information (zsync, AppImageUpdate) and desktop integration beyond what the embedded desktop entry gives.
- Other formats (MSI, MSIX, Flatpak, .deb, .rpm), and an installer for the web or mobile targets.
- Building in a container from `package.sh`. The script builds on the host it runs on, like today; the AppImage's build host is documented.

## Decisions

### D1: Two new targets, not new behavior of `windows` and `linux`

`windows-installer` and `appimage` join the target list, the usage text and the host check (`windows-installer` on Windows, `appimage` on Linux). Folding them into `windows` and `linux` would change those packages' outputs and make every Windows zip build need Inno Setup and every Linux tarball build need the AppImage tools.

*Alternative:* flags such as `--installer`. Rejected: targets are already how `package.sh` selects outputs, and `apk`/`appbundle` set the precedent of two packages per platform.

### D2: Build and stage once per platform

`build_linux` and `build_windows` run `fluffy build …` once per invocation (guarded by a variable set on first call). Windows staging (the Release folder plus the three runtime DLLs) moves into `stage_windows`, which also runs once and returns the staged folder; `package_windows` zips it and `package_windows_installer` hands it to Inno Setup, so both contain identical files. `package_linux` keeps copying the bundle as today; `package_appimage` copies it again into its own AppDir, since it adds libraries to it.

### D3: Tools are found before anything runs

The preflight loop that checks hosts also resolves the tools of the new targets and dies naming the missing one:
- `windows-installer`: `ISCC` environment variable, else `iscc`/`ISCC.exe` on `PATH`, else `Inno Setup 6\ISCC.exe` under Program Files (x86) and under `%LOCALAPPDATA%\Programs` (where a per-user winget install puts it).
- `appimage`: `LINUXDEPLOY`, `APPIMAGETOOL` and `PATCHELF` environment variables, else `linuxdeploy` / `appimagetool` (or the upstream `-<arch>.AppImage` file names) and `patchelf` on `PATH`.

### D4: Inno Setup script in `windows/packaging/cardflip.iss`

The script lives next to the platform code like `linux/packaging/`, and `package.sh` passes everything variable with `/D`: the version, a numeric `<version>.<build number>` for the file version, the architecture, the staged source folder and the output folder and base name (`cardflip-<v><suffix>-windows-<arch>-setup`). Key settings:
- A fixed `AppId` GUID, generated once and never changed; it is what makes a newer installer upgrade the earlier install and keep one "Installed apps" entry.
- `PrivilegesRequired=lowest` with `PrivilegesRequiredOverridesAllowed=dialog commandline`, and `DefaultDirName={autopf}\Cardflip`: per user by default in `%LOCALAPPDATA%\Programs\Cardflip`, all users in Program Files after elevation.
- `ArchitecturesAllowed` / `ArchitecturesInstallIn64BitMode` set from the architecture (`x64compatible` or `arm64`).
- `[Icons]` in `{autoprograms}` and an unchecked `desktopicon` task for `{autodesktop}`; `[Run]` with `postinstall skipifsilent` to start Cardflip.
- `CloseApplications=yes` so the Restart Manager asks to close a running Cardflip; `RestartApplications=no`.
- `[InstallDelete]` of `{app}\data` so an upgrade doesn't keep assets the new version no longer has. Nothing deletes outside `{app}`, so the collection in `%APPDATA%` survives uninstall and upgrade without special handling.
- `SetupIconFile` and `UninstallDisplayIcon` from `windows/runner/resources/app_icon.ico`, `WizardStyle=modern`, English and German wizard languages (the app's two interface languages), `lzma2` solid compression.

*Alternative:* NSIS. Rejected with the user: Inno Setup's per-user/all-users switch, Restart Manager support and upgrade-by-`AppId` are built in, where NSIS needs them scripted.

### D5: AppImage built around the unchanged Flutter bundle

AppDir layout:

```
AppRun                                  linux/packaging/AppRun
dev.cardflip.cardflip.desktop           linux/packaging, Exec=cardflip
dev.cardflip.cardflip.png, .DirIcon     256x256 icon
usr/lib/cardflip/                       the release bundle (cardflip, data/, lib/)
usr/lib/cardflip/lib/libmpv.so.2 …      libmpv and its bundled dependencies
usr/share/icons/hicolor/…               linux/packaging/icons
usr/share/doc/cardflip/                 LICENSE, bundled-libraries.txt
```

libmpv and its dependencies go into the bundle's `lib/`, next to the plugin that links it. linuxdeploy is used only as the dependency resolver: it deploys `libmpv.so.2` (resolved on the host) into a scratch AppDir, applying its exclude list, and sets `$ORIGIN` RUNPATHs; the resulting `usr/lib/*.so*` are moved into the bundle's `lib/`. The plugins in the AppDir's copy of the bundle get the RUNPATH `$ORIGIN` with `patchelf` in place of the build-tree path they carry, so `libmpv.so.2` resolves next to them; with the deployed libraries all in one directory, every lookup resolves inside the AppImage. Only the AppImage's copies are patched: the build output and the tarball stay as they are.

`linux/packaging/AppRun` is only for the AppImage; `package_linux` leaves it out of the tarball's `packaging/`, so the tarball's contents don't change.

`AppRun` sets `LIBMPV_LIBRARY_PATH` to the bundled `libmpv.so.2` and `exec`s `"$APPDIR/usr/lib/cardflip/cardflip" "$@"`. media_kit's Dart side opens libmpv by name and tries `libmpv.so` first; where mpv is installed that is the system's file, a second libmpv beside the bundled one the video plugin links, and mpv aborts (`m_config_cache_from_shadow: group_index >= 0`) when the two share a player. media_kit opens `LIBMPV_LIBRARY_PATH` before any name, which reuses the loaded copy; nothing else reads the variable, so it is harmless in the programs the app starts. `AppRun` sets no `LD_LIBRARY_PATH`: that would leak into `ffmpeg`, `parecord` and `spd-say`, which the app starts, and make the system's `ffmpeg` load the bundled FFmpeg libraries.

Beyond linuxdeploy's exclude list (glibc, GL, X11, fontconfig, …), libraries that must match the host's drivers and daemons are excluded too: `libva*`, `libvdpau*`, `libvulkan*`, `libdrm*`, `libgbm*`, `libwayland-*`, `libxkbcommon*`, `libpulse*`, `libpipewire*`, `libasound*`, `libjack*`. GTK and GLib are excluded as well: only libmpv's tree is deployed, but it can reach GLib through libraries such as HarfBuzz. appimagetool then builds the file with `ARCH` set (`x86_64`/`aarch64`); the output name keeps the repository's `x64`/`arm64` convention.

The app's data location is unchanged (the bundle is the same binary with the same application ID), so the AppImage and the tarball install share `~/.local/share/dev.cardflip.cardflip`.

*Alternative:* linuxdeploy on the whole bundle with its GTK plugin. Rejected: it moves the Flutter libraries out of the bundle layout and bundles GTK, which the spec leaves to the host.

### D6: License entries for the AppImage's libmpv and FFmpeg

`package.sh` writes `usr/share/doc/cardflip/bundled-libraries.txt`: the build host's distribution (from `/etc/os-release`) and, for every library it moved into the bundle, the file name and the owning package and version (`dpkg-query -S`, `rpm -qf` or `pacman -Qo`, whichever exists; "unknown" otherwise), plus the note that the source is the distribution's source package of that version.

`licenses.dart` treats the Linux app as running from an AppImage when the `APPIMAGE` and `APPDIR` environment variables are set (the AppImage runtime sets both). It then adds libmpv and FFmpeg entries with Cardflip's GPL version 3 text and the contents of `$APPDIR/usr/share/doc/cardflip/bundled-libraries.txt`. Distribution builds of mpv and FFmpeg are GPL version 2 or later, or LGPL, depending on the distribution; both can be passed on under the GPL version 3, so that text is correct whichever build was bundled, which an LGPL-only text would not be. The environment lookup is injectable so tests cover tarball and AppImage without either.

`THIRD_PARTY_NOTICES.md` gets a paragraph on the AppImage's bundled libraries next to the existing libmpv/FFmpeg paragraph.

### D7: Build host for the AppImage

The AppImage needs glibc at least as new as the build host's, and libmpv 0.35 or newer for `libmpv.so.2`. `BUILDING.md` recommends building it on Ubuntu 24.04 (or in such a container), the oldest common base with `libmpv2`, with `APPIMAGE_EXTRACT_AND_RUN=1` where FUSE isn't available for the tools themselves. `package.sh` prints the glibc version of the build host after building so the release notes can state the minimum.

## Risks / Trade-offs

- [The AppImage pulls a large, distribution-specific dependency tree from the host's libmpv (FFmpeg, libass, codecs)] → It is built on a documented base distribution; the verification task runs it on a host without mpv and checks `ldd` on the bundled libraries for "not found" and for excluded families.
- [An excluded host library is missing on a minimal system (no PulseAudio/PipeWire client libraries)] → The spec targets desktop systems, which have them; `BUILDING.md` lists the runtime expectation.
- [GPL source obligations for the bundled distribution libraries] → The in-app entry and `bundled-libraries.txt` name the exact package versions; the release process in `BUILDING.md` notes that those source packages must stay available (the distribution's archive) or be published with the release.
- [The installer is unsigned, so SmartScreen warns on download] → Same situation as the zip today; signing is a separate change.
- [An `AppId` change would break upgrades] → The GUID is a fixed constant in the script with a comment that it must never change.
- [Media playback in the AppImage depends on the plugin resolving libmpv through `$ORIGIN`] → Covered by the playback check on a host without mpv; media_kit's own `dlopen` of the same soname reuses the already-loaded library.

## Migration Plan

Additive. Existing targets, their file names and their users are unaffected; developers who don't request the new targets need no new tools. Rolling back means removing the two targets and the new files.
