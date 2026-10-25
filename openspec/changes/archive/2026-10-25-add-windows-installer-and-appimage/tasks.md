# Tasks

## 1. Shared script structure

- [x] 1.1 In `tool/package.sh`, add `windows-installer` and `appimage` to the argument `case`, the header comment's target table and the host check (D1), and add the tool resolution for both to the preflight loop (D3); verify with `bash -n tool/package.sh`, that `tool/package.sh --help` lists both targets, that `tool/package.sh --no-check windows-installer` on Linux dies with the host message, and that `PATH=/usr/bin:/bin tool/package.sh --no-check appimage` dies naming linuxdeploy before building
- [x] 1.2 Add `build_linux` and `build_windows` (run once per invocation) and move the Windows staging into `stage_windows` (D2), keeping `package_linux` and `package_windows` output unchanged; verify `tool/package.sh --no-check linux` runs `fluffy build linux` once and the tarball's file list (`tar -tzf`) matches the one from the current `main`

## 2. AppImage

- [x] 2.1 Add `linux/packaging/AppRun` (executable, `LIBMPV_LIBRARY_PATH` set to the bundled `libmpv.so.2`, `exec "$APPDIR/usr/lib/cardflip/cardflip" "$@"`, no `LD_LIBRARY_PATH`) per D5; verify `bash -n` and `test -x` on it, and that `tool/install_linux.sh` and the tarball still work with the extra file in `packaging/` (install to a temporary `PREFIX` and check `$PREFIX/lib/cardflip/cardflip` exists)
- [x] 2.2 Implement `package_appimage` (D5): AppDir layout, libmpv deployed through linuxdeploy into a scratch AppDir with the extra exclusions and moved into the bundle's `lib/`, the AppDir's plugin RUNPATHs set to `$ORIGIN` with patchelf, desktop entry and icons, appimagetool with `ARCH`, output `dist/$name-linux-$arch.AppImage`, and the build host's glibc version printed (D7); verify `tool/package.sh --no-check appimage` produces the file, that `--appimage-extract` shows the layout of D5, that `ldd` over `squashfs-root/usr/lib/cardflip/lib/*.so*` reports no "not found", and that no `libgtk`, `libglib`, `libva`, `libpulse` or `libpipewire` file is bundled
- [x] 2.3 Write `usr/share/doc/cardflip/bundled-libraries.txt` and copy `LICENSE` there (D6); verify the extracted file names the distribution from `/etc/os-release` and lists `libmpv.so.2` and the `libav*` libraries with package and version
- [x] 2.4 Verify `tool/package.sh --no-check linux appimage` builds once and produces both packages, and `SHA256SUMS` lists both

## 3. AppImage license entries

- [x] 3.1 In `lib/app/licenses.dart`, detect the AppImage from `APPIMAGE` and `APPDIR` through an injectable environment and add libmpv and FFmpeg entries with the GPL version 3 text and the contents of `bundled-libraries.txt` (D6); add tests to the licenses test for the spec scenarios "Media libraries in the AppImage" and "Media libraries on Linux" (no AppImage variables), and verify `fvm flutter test` on that test file passes
- [x] 3.2 Add the AppImage's bundled libraries to `THIRD_PARTY_NOTICES.md` next to the libmpv/FFmpeg paragraph; verify its links resolve (`ls` each linked path)

## 4. Windows installer

- [x] 4.1 Add `windows/packaging/cardflip.iss` per D4, with a newly generated fixed `AppId` GUID and a comment that it must never change; verify on Windows that `ISCC.exe` compiles it with the defines `package.sh` passes (a manual `iscc /D…` run), without warnings
- [x] 4.2 Implement `package_windows_installer`: `build_windows`, `stage_windows`, then `ISCC` with the defines from D4 and Windows paths via `cygpath`, output `dist/$name-windows-$arch-setup.exe`; verify on Windows in Git Bash that `tool/package.sh --no-check windows windows-installer` builds once and produces both the zip and the setup executable, and that the installer's file list matches the zip's
- [x] 4.3 Verify the installer on a Windows machine (or VM) as a standard user: default install runs without elevation into `%LOCALAPPDATA%\Programs\Cardflip`; the Start menu entry has the icon and starts the app; the desktop shortcut appears only when ticked; "Install for all users" elevates and installs into Program Files; with the app running, the installer asks to close it; installing a build with a higher `VERSION` leaves one "Installed apps" entry at the new version with the collection unchanged; uninstalling removes the folder and shortcuts and leaves `%APPDATA%\dev.cardflip\cardflip`

## 5. Documentation

- [x] 5.1 Update `BUILDING.md`: the two new rows in the packages table, the tools each needs (Inno Setup 6, `ISCC`; `linuxdeploy`, `appimagetool`, `patchelf`, `LINUXDEPLOY`/`APPIMAGETOOL`/`PATCHELF`, `APPIMAGE_EXTRACT_AND_RUN=1`), the Ubuntu 24.04 build host for the AppImage (D7), and the note that the bundled libraries' source packages must stay available for a release; verify the commands shown run as written on Linux (`tool/package.sh --help`, the AppImage build)

## 6. Integration checks

- [x] 6.1 Run `tool/check.sh` and verify every gate passes
- [x] 6.2 Run the AppImage on a host or container without mpv/libmpv installed (e.g. a fresh Ubuntu 24.04 desktop VM, or Fedora with `mpv-libs` absent): it starts with the Cardflip name and icon, plays a card's audio, opens the same collection as the tarball install, the licenses page lists libmpv and FFmpeg with the bundled versions, and recording with `parecord`/`ffmpeg` still works (no bundled libraries leak into child processes)
- [x] 6.3 Run `tool/package.sh linux` and `tool/package.sh windows` on their hosts with only the existing targets and verify `dist/` gets the same file names and contents as before the change, and no AppImage or installer
