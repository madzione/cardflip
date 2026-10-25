#!/usr/bin/env bash
# Builds release packages into dist/, named cardflip-<version>-<platform>[-<arch>].<ext>,
# and writes dist/SHA256SUMS.
#
#     tool/package.sh linux apk          package for Linux and Android
#     tool/package.sh --no-check web     skip tool/check.sh before building
#
# Targets, and the host each one builds on:
#     linux               Linux     tarball of the bundle with install.sh, desktop entry and icons
#     appimage            Linux     AppImage with libmpv and the libraries it needs
#     apk                 any       one APK per ABI
#     appbundle           any       Android App Bundle for Google Play
#     web                 any       tarball of the site for a static web server
#     windows             Windows   zip of the app folder (run from Git Bash)
#     windows-installer   Windows   setup executable made with Inno Setup (run from Git Bash)
#     macos               macOS     disk image with the app
#     ios                 macOS     .ipa
#
# The version is the vX.Y.Z tag on HEAD, else the version in pubspec.yaml with
# "-dev.<commit>" added to the file names. The build number is the number of
# commits on HEAD. VERSION and BUILD_NUMBER override both.
#
# Signing:
#     android/key.properties  release key for apk and appbundle (see BUILDING.md);
#                             ALLOW_DEBUG_SIGNING=1 packages with the debug key instead
#     MACOS_SIGN_IDENTITY     "Developer ID Application: …" to sign the macOS app
#     MACOS_NOTARY_PROFILE    notarytool keychain profile to notarize the disk image
#     IOS_EXPORT_OPTIONS      ExportOptions.plist for the .ipa export
#
# Packaging tools, looked up on PATH unless set:
#     ISCC                    Inno Setup 6's ISCC.exe, for windows-installer (also
#                             found in its default install folders)
#     LINUXDEPLOY             linuxdeploy, for appimage
#     APPIMAGETOOL            appimagetool, for appimage
#     PATCHELF                patchelf, for appimage
#
# WEB_BASE_HREF sets the base href when the web app isn't served from /.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
dist="$root/dist"
export PATH="$HOME/go/bin:$PATH"

usage() { sed -n '2,/^set /{/^set /d;s/^# \{0,1\}//;p;}' "$0"; }
die() { echo "package.sh: $*" >&2; exit 1; }
warn() { echo "package.sh: warning: $*" >&2; }
step() { printf '\n== %s\n' "$*"; }

run_check=1
targets=()
for arg in "$@"; do
  case "$arg" in
    --no-check) run_check=0 ;;
    -h | --help) usage; exit 0 ;;
    linux | appimage | apk | appbundle | web | windows | windows-installer | macos | ios) targets+=("$arg") ;;
    *) die "unknown target or option: $arg (see --help)" ;;
  esac
done
if [[ ${#targets[@]} -eq 0 ]]; then
  usage >&2
  exit 1
fi

case "$(uname -s)" in
  Linux) host=linux ;;
  Darwin) host=macos ;;
  MINGW* | MSYS* | CYGWIN*) host=windows ;;
  *) host=unknown ;;
esac
case "$(uname -m)" in
  x86_64 | amd64) arch=x64 ;;
  aarch64 | arm64) arch=arm64 ;;
  *) arch="$(uname -m)" ;;
esac

# The first of the environment variable $1 and the commands $2… that exists.
find_tool() {
  local var="$1" cmd
  shift
  if [[ -n "${!var:-}" ]]; then
    command -v "${!var}" 2>/dev/null
    return
  fi
  for cmd in "$@"; do
    if command -v "$cmd" 2>/dev/null; then
      return
    fi
  done
  return 1
}

# ISCC, else Inno Setup 6 on PATH or in its machine-wide or per-user (winget)
# install folder.
find_iscc() {
  find_tool ISCC iscc ISCC.exe && return
  local dir
  for dir in "$(printenv 'ProgramFiles(x86)' || true)" "${LOCALAPPDATA:+$LOCALAPPDATA/Programs}"; do
    if [[ -n "$dir" && -f "$(cygpath -u "$dir")/Inno Setup 6/ISCC.exe" ]]; then
      echo "$(cygpath -u "$dir")/Inno Setup 6/ISCC.exe"
      return
    fi
  done
  return 1
}

# Fail before building anything when a target can't be packaged here.
for target in "${targets[@]}"; do
  case "$target" in
    linux | windows | macos)
      [[ "$host" == "$target" ]] || die "$target packages only build on $target, not on $(uname -s)." ;;
    ios)
      [[ "$host" == macos ]] || die "ios packages only build on macOS, not on $(uname -s)." ;;
    appimage)
      [[ "$host" == linux ]] || die "AppImages only build on Linux, not on $(uname -s)."
      case "$arch" in
        x64) appimage_arch=x86_64 ;;
        arm64) appimage_arch=aarch64 ;;
        *) die "AppImages don't build on $arch." ;;
      esac
      linuxdeploy="$(find_tool LINUXDEPLOY linuxdeploy "linuxdeploy-$appimage_arch.AppImage")" ||
        die "linuxdeploy not found; put it on PATH or set LINUXDEPLOY (see BUILDING.md)."
      appimagetool="$(find_tool APPIMAGETOOL appimagetool "appimagetool-$appimage_arch.AppImage")" ||
        die "appimagetool not found; put it on PATH or set APPIMAGETOOL (see BUILDING.md)."
      patchelf="$(find_tool PATCHELF patchelf)" ||
        die "patchelf not found; put it on PATH or set PATCHELF (see BUILDING.md)." ;;
    windows-installer)
      [[ "$host" == windows ]] || die "windows-installer packages only build on Windows, not on $(uname -s)."
      iscc="$(find_iscc)" ||
        die "Inno Setup 6 (ISCC.exe) not found; install it (winget install JRSoftware.InnoSetup) or set ISCC." ;;
    apk | appbundle)
      if [[ ! -f "$root/android/key.properties" ]]; then
        if [[ "${ALLOW_DEBUG_SIGNING:-}" == 1 ]]; then
          warn "android/key.properties is missing; $target is signed with the debug key and can't update a release install."
        else
          die "android/key.properties is missing, so $target would be signed with the debug key. See BUILDING.md, or set ALLOW_DEBUG_SIGNING=1 for a test build."
        fi
      fi ;;
  esac
done

cd "$root"
pubspec_version="$(sed -n 's/^version: *//p' pubspec.yaml)"
suffix=""
if [[ -n "${VERSION:-}" ]]; then
  version="$VERSION"
elif tag="$(git describe --tags --exact-match --match 'v[0-9]*' 2>/dev/null)" &&
  [[ "$tag" =~ ^v([0-9]+\.[0-9]+\.[0-9]+)$ ]]; then
  version="${BASH_REMATCH[1]}"
else
  version="${pubspec_version%%+*}"
  suffix="-dev"
  if commit="$(git rev-parse --short HEAD 2>/dev/null)"; then
    suffix="-dev.$commit"
  fi
fi
[[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || die "version \"$version\" isn't X.Y.Z."
if [[ -n "${BUILD_NUMBER:-}" ]]; then
  build_number="$BUILD_NUMBER"
elif ! build_number="$(git rev-list --count HEAD 2>/dev/null)"; then
  # No commits yet.
  build_number=1
  if [[ "$pubspec_version" == *+* ]]; then
    build_number="${pubspec_version#*+}"
  fi
fi
name="cardflip-$version$suffix"
build_flags=(--release --build-name "$version" --build-number "$build_number")

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
mkdir -p "$dist"

# The Linux and Windows builds run once per invocation, however many packages
# of the platform are requested.
linux_built=""
build_linux() {
  if [[ -z "$linux_built" ]]; then
    fluffy build linux "${build_flags[@]}"
    linux_built=1
  fi
}

package_linux() {
  build_linux
  local stage="$work/$name-linux-$arch"
  mkdir -p "$stage"
  cp -r "$root/build/linux/$arch/release/bundle" "$stage/bundle"
  cp -r "$root/linux/packaging" "$stage/packaging"
  rm "$stage/packaging/AppRun" # only for the AppImage
  cp "$root/tool/install_linux.sh" "$stage/install.sh"
  tar -C "$work" -czf "$dist/$name-linux-$arch.tar.gz" "$name-linux-$arch"
}

# Libraries libmpv needs that must match the host's drivers and sound
# daemons, beyond those linuxdeploy never bundles; and GTK and GLib, which
# the app takes from the host.
appimage_excludes=(
  'libva*' 'libvdpau*' 'libvulkan*' 'libdrm*' 'libgbm*' 'libwayland-*' 'libxkbcommon*'
  'libpulse*' 'libpipewire*' 'libasound*' 'libjack*'
  'libgtk-*' 'libgdk-*' 'libglib-2.0*' 'libgobject-2.0*' 'libgio-2.0*' 'libgmodule-2.0*'
)

# The distribution package and version that installed the file $1, or
# "unknown".
owning_package() {
  local file="$1" pkg f
  [[ -n "$file" ]] || { echo unknown; return; }
  if command -v dpkg-query >/dev/null; then
    # dpkg records /lib paths on merged-/usr systems that ldd reports as /usr/lib.
    for f in "$file" "${file#/usr}"; do
      if pkg="$(dpkg-query -S "$f" 2>/dev/null)"; then
        pkg="${pkg%%:*}"
        echo "$pkg $(dpkg-query -W -f '${Version}' "$pkg")"
        return
      fi
    done
  elif command -v rpm >/dev/null && pkg="$(rpm -qf --qf '%{NAME} %{VERSION}-%{RELEASE}\n' "$file" 2>/dev/null)"; then
    echo "$pkg"
    return
  elif command -v pacman >/dev/null && pkg="$(pacman -Qqo "$file" 2>/dev/null)"; then
    pacman -Q "$pkg"
    return
  fi
  echo unknown
}

package_appimage() {
  build_linux
  local appdir="$work/$name-linux-$arch.AppDir"
  local app="$appdir/usr/lib/cardflip"
  local packaging="$root/linux/packaging"
  local app_id=dev.cardflip.cardflip
  mkdir -p "$appdir/usr/lib" "$appdir/usr/share/icons" "$appdir/usr/share/doc/cardflip"
  cp -r "$root/build/linux/$arch/release/bundle" "$app"
  cp "$packaging/AppRun" "$appdir/AppRun"
  cp "$packaging/$app_id.desktop" "$appdir/"
  cp "$packaging/icons/hicolor/256x256/apps/$app_id.png" "$appdir/$app_id.png"
  ln -s "$app_id.png" "$appdir/.DirIcon"
  cp -r "$packaging/icons/hicolor" "$appdir/usr/share/icons/"

  # The plugins carry the build tree's RUNPATH; $ORIGIN makes them find
  # libmpv and the other bundled libraries next to them.
  local plugin
  for plugin in "$app"/lib/lib*_plugin.so; do
    "$patchelf" --set-rpath '$ORIGIN' "$plugin"
  done

  # libmpv and the libraries it needs that desktop systems don't provide,
  # resolved on this host and moved next to the plugin that links it.
  local libmpv
  libmpv="$(ldd "$app/lib/libmedia_kit_video_plugin.so" | awk '$1 == "libmpv.so.2" { print $3 }')"
  [[ -f "$libmpv" ]] || die "libmpv.so.2 not found; install libmpv (mpv 0.35 or newer)."
  local deps="$work/appimage-deps" exclude_flags=() pattern
  for pattern in "${appimage_excludes[@]}"; do
    exclude_flags+=(--exclude-library "$pattern")
  done
  "$linuxdeploy" --appdir "$deps" --library "$libmpv" "${exclude_flags[@]}"
  local -A host_paths=([libmpv.so.2]="$libmpv")
  local soname path
  while read -r soname path; do
    host_paths[$soname]="$path"
  done < <(ldd "$libmpv" | awk '$2 == "=>" && $3 ~ /^\// { print $1, $3 }')

  # Which package each bundled library came from, for its license entry.
  local list="$appdir/usr/share/doc/cardflip/bundled-libraries.txt" lib file
  {
    echo "Libraries bundled in this AppImage, copied from the distribution packages"
    echo "of the system it was built on. The source of each is the distribution's"
    echo "source package of the version listed."
    echo
    echo "Distribution: $(. /etc/os-release && echo "${PRETTY_NAME:-$NAME}")"
    echo
  } >"$list"
  for lib in "$deps"/usr/lib/*.so*; do
    file="$(basename "$lib")"
    [[ ! -e "$app/lib/$file" ]] || die "$file is in the Flutter bundle already."
    mv "$lib" "$app/lib/"
    echo "$file  $(owning_package "${host_paths[$file]:-}")" >>"$list"
  done
  cp "$root/LICENSE" "$appdir/usr/share/doc/cardflip/"

  ARCH="$appimage_arch" "$appimagetool" "$appdir" "$dist/$name-linux-$arch.AppImage"
  echo "Built against $(getconf GNU_LIBC_VERSION); the AppImage needs that glibc or newer."
}

package_apk() {
  fluffy build apk --split-per-abi "${build_flags[@]}"
  local abi
  for abi in armeabi-v7a arm64-v8a x86_64; do
    cp "$root/build/app/outputs/flutter-apk/app-$abi-release.apk" "$dist/$name-android-$abi.apk"
  done
}

package_appbundle() {
  fluffy build appbundle "${build_flags[@]}"
  cp "$root/build/app/outputs/bundle/release/app-release.aab" "$dist/$name-android.aab"
}

package_web() {
  local base_flags=()
  if [[ -n "${WEB_BASE_HREF:-}" ]]; then
    base_flags=(--base-href "$WEB_BASE_HREF")
  fi
  fluffy build web "${build_flags[@]}" ${base_flags[@]+"${base_flags[@]}"}
  cp -r "$root/build/web" "$work/$name-web"
  tar -C "$work" -czf "$dist/$name-web.tar.gz" "$name-web"
}

windows_built=""
build_windows() {
  if [[ -z "$windows_built" ]]; then
    fluffy build windows "${build_flags[@]}"
    windows_built=1
  fi
}

# Stages the app folder once into $windows_stage, for the zip and the
# installer alike.
windows_stage=""
stage_windows() {
  [[ -z "$windows_stage" ]] || return 0
  local stage="$work/$name-windows-$arch"
  cp -r "$root/build/windows/$arch/runner/Release" "$stage"
  # The Visual C++ runtime, so the app starts on machines without the
  # redistributable installed.
  local system32 dll
  system32="$(cygpath -u "$WINDIR")/System32"
  for dll in msvcp140.dll vcruntime140.dll vcruntime140_1.dll; do
    if [[ -f "$system32/$dll" ]]; then
      cp "$system32/$dll" "$stage/"
    else
      warn "$dll not found in $system32; the app needs the Visual C++ redistributable to run."
    fi
  done
  windows_stage="$stage"
}

package_windows() {
  build_windows
  stage_windows
  powershell.exe -NoProfile -Command "Compress-Archive -Force -Path '$(cygpath -w "$windows_stage")' -DestinationPath '$(cygpath -w "$dist/$name-windows-$arch.zip")'"
}

package_windows_installer() {
  build_windows
  stage_windows
  # Options with "-" rather than "/", which Git Bash would rewrite as paths.
  "$iscc" -Qp \
    "-DAppVersion=$version" \
    "-DFileVersion=$version.$build_number" \
    "-DArch=$arch" \
    "-DSourceDir=$(cygpath -w "$windows_stage")" \
    "-DOutputDir=$(cygpath -w "$dist")" \
    "-DOutputBaseFilename=$name-windows-$arch-setup" \
    "$(cygpath -w "$root/windows/packaging/cardflip.iss")"
}

package_macos() {
  fluffy build macos "${build_flags[@]}"
  local stage="$work/$name-macos"
  local app="$stage/cardflip.app"
  local dmg="$dist/$name-macos.dmg"
  mkdir -p "$stage"
  ditto "$root/build/macos/Build/Products/Release/cardflip.app" "$app"
  if [[ -n "${MACOS_SIGN_IDENTITY:-}" ]]; then
    # Nested code first and without the app's entitlements, then the app.
    find "$app/Contents/Frameworks" -depth \( -name '*.framework' -o -name '*.dylib' \) -print0 |
      while IFS= read -r -d '' item; do
        codesign --force --options runtime --timestamp --sign "$MACOS_SIGN_IDENTITY" "$item" || exit 1
      done
    codesign --force --options runtime --timestamp \
      --entitlements "$root/macos/Runner/Release.entitlements" \
      --sign "$MACOS_SIGN_IDENTITY" "$app"
  else
    warn "MACOS_SIGN_IDENTITY is not set; Gatekeeper blocks the app on other Macs."
  fi
  ln -s /Applications "$stage/Applications"
  rm -f "$dmg"
  hdiutil create -volname Cardflip -srcfolder "$stage" -format UDZO "$dmg"
  if [[ -n "${MACOS_SIGN_IDENTITY:-}" ]]; then
    codesign --timestamp --sign "$MACOS_SIGN_IDENTITY" "$dmg"
  fi
  if [[ -n "${MACOS_NOTARY_PROFILE:-}" ]]; then
    [[ -n "${MACOS_SIGN_IDENTITY:-}" ]] || die "notarizing needs MACOS_SIGN_IDENTITY."
    xcrun notarytool submit "$dmg" --keychain-profile "$MACOS_NOTARY_PROFILE" --wait
    xcrun stapler staple "$dmg"
  fi
}

package_ios() {
  local export_flags=()
  if [[ -n "${IOS_EXPORT_OPTIONS:-}" ]]; then
    export_flags=(--export-options-plist "$IOS_EXPORT_OPTIONS")
  fi
  rm -rf "$root/build/ios/ipa"
  fluffy build ipa "${build_flags[@]}" ${export_flags[@]+"${export_flags[@]}"}
  set -- "$root"/build/ios/ipa/*.ipa
  [[ -f "$1" ]] || die "no .ipa was exported; check the signing team in Xcode or IOS_EXPORT_OPTIONS."
  cp "$1" "$dist/$name-ios.ipa"
}

if [[ "$run_check" == 1 ]]; then
  "$root/tool/check.sh"
fi

for target in "${targets[@]}"; do
  step "package $target ($version$suffix, build $build_number)"
  "package_${target//-/_}"
done

step "checksums"
cd "$dist"
rm -f SHA256SUMS
if command -v sha256sum >/dev/null; then
  sha256sum -- * >"$work/SHA256SUMS"
else
  shasum -a 256 -- * >"$work/SHA256SUMS"
fi
mv "$work/SHA256SUMS" SHA256SUMS
cat SHA256SUMS
