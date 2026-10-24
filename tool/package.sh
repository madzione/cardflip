#!/usr/bin/env bash
# Builds release packages into dist/, named cardflip-<version>-<platform>[-<arch>].<ext>,
# and writes dist/SHA256SUMS.
#
#     tool/package.sh linux apk          package for Linux and Android
#     tool/package.sh --no-check web     skip tool/check.sh before building
#
# Targets, and the host each one builds on:
#     linux       Linux     tarball of the bundle with install.sh, desktop entry and icons
#     apk         any       one APK per ABI
#     appbundle   any       Android App Bundle for Google Play
#     web         any       tarball of the site for a static web server
#     windows     Windows   zip of the app folder (run from Git Bash)
#     macos       macOS     disk image with the app
#     ios         macOS     .ipa
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
    linux | apk | appbundle | web | windows | macos | ios) targets+=("$arg") ;;
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

# Fail before building anything when a target can't be packaged here.
for target in "${targets[@]}"; do
  case "$target" in
    linux | windows | macos)
      [[ "$host" == "$target" ]] || die "$target packages only build on $target, not on $(uname -s)." ;;
    ios)
      [[ "$host" == macos ]] || die "ios packages only build on macOS, not on $(uname -s)." ;;
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

package_linux() {
  fluffy build linux "${build_flags[@]}"
  local stage="$work/$name-linux-$arch"
  mkdir -p "$stage"
  cp -r "$root/build/linux/$arch/release/bundle" "$stage/bundle"
  cp -r "$root/linux/packaging" "$stage/packaging"
  cp "$root/tool/install_linux.sh" "$stage/install.sh"
  tar -C "$work" -czf "$dist/$name-linux-$arch.tar.gz" "$name-linux-$arch"
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

package_windows() {
  fluffy build windows "${build_flags[@]}"
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
      warn "$dll not found in $system32; the zip needs the Visual C++ redistributable to run."
    fi
  done
  powershell.exe -NoProfile -Command "Compress-Archive -Force -Path '$(cygpath -w "$stage")' -DestinationPath '$(cygpath -w "$dist/$name-windows-$arch.zip")'"
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
  "package_$target"
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
