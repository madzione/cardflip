# Proposal

## Why

`tool/package.sh` produces a zip for Windows and a tarball with `install.sh` for Linux. Both require the user to unpack files and, on Windows, to place the folder and create shortcuts by hand; on Linux, media playback breaks wherever mpv isn't installed. A Windows installer and an AppImage give users the usual "download, run" experience on both desktops without removing the existing packages that scripts, power users and distro packagers already rely on.

## What Changes

- New `package.sh` target `windows-installer` (Windows host, Git Bash): an Inno Setup installer `cardflip-<v>-windows-<arch>-setup.exe`. It installs per user without admin rights by default (all users optional), adds a Start menu entry, an optional desktop shortcut and an uninstaller in "Apps & features", upgrades an earlier install in place, and never removes the user's collection.
- New `package.sh` target `appimage` (Linux host): `cardflip-<v>-linux-<arch>.AppImage` with the release bundle, the desktop entry and icons, and libmpv with its non-system dependencies, so audio and video work without mpv installed. GTK, glibc and the graphics stack still come from the host.
- Both are additions: the `windows` zip and `linux` tarball targets keep their names, contents and file names. When an addition and its existing sibling are requested together (`linux appimage`, `windows windows-installer`), the app is built once.
- The new targets fail before any build when their host or packaging tool (Inno Setup's `ISCC.exe`; `linuxdeploy`, `appimagetool` and `patchelf`) is missing, like the existing host checks.
- The licenses page lists libmpv and FFmpeg when the Linux app runs from the AppImage, because that package bundles them; the tarball install still uses the system's libmpv and doesn't list them.
- `BUILDING.md` documents the new targets, their tools and the build host to use for the AppImage.

## Capabilities

### New Capabilities
- `release-packaging`: the release packages `tool/package.sh` builds for each platform, their file names and contents, and how they install, upgrade and uninstall. This change adds the Windows installer and AppImage requirements, plus the requirement that the existing packages stay unchanged.

### Modified Capabilities
- `app-shell`: "Licenses page contents" lists libmpv and FFmpeg in the Linux AppImage, which bundles them, while the Linux app on the system's libmpv still doesn't list them.

## Impact

- `tool/package.sh`: two new targets, preflight checks for their tools, and shared build/staging helpers so the Linux and Windows builds run once per invocation.
- New files: `windows/packaging/cardflip.iss` (Inno Setup script) and `linux/packaging/AppRun` (AppImage entry point).
- `lib/app/licenses.dart`: detect the AppImage (`APPIMAGE` environment variable) and add the libmpv/FFmpeg entries for it, with tests.
- `THIRD_PARTY_NOTICES.md` and `BUILDING.md`: the AppImage's bundled libraries and the new packaging tools.
- New build-host tools, needed only for the new targets: Inno Setup 6 on Windows; `linuxdeploy`, `appimagetool` and `patchelf` on Linux. Nothing changes for the existing targets.
