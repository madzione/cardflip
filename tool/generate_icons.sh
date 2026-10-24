#!/usr/bin/env bash
# Regenerates the app icons for every platform from the PNGs in assets/logo/.
# Needs ImageMagick (`magick`) for the web maskable icons and the Linux icon.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
logo="$root/assets/logo"
cd "$root"

# Android, iOS, macOS, Windows and web (see flutter_launcher_icons.yaml).
fvm dart run flutter_launcher_icons

# The generator makes the web maskable icons from the rounded logo, whose
# transparent corners would show once a launcher applies its own mask. Use
# the full-bleed square instead.
for size in 192 512; do
  magick "$logo/cardflip_icon_square.png" -resize "${size}x${size}" \
    "web/icons/Icon-maskable-$size.png"
done

# The generator writes a single 256 px image into the .ico; include the small
# sizes too so the taskbar and title bar don't show a blurry downscale.
magick "$logo/cardflip_logo_1024.png" \
  -define icon:auto-resize=256,128,64,48,32,24,16 \
  windows/runner/resources/app_icon.ico

# Linux: installed next to the executable and set as the window icon by
# linux/runner/my_application.cc.
magick "$logo/cardflip_logo_1024.png" -resize 256x256 \
  linux/runner/resources/app_icon.png

# Linux: hicolor theme icons for the .desktop entry, which is where Wayland
# compositors take the icon from (tool/install_linux.sh installs them).
hicolor="linux/packaging/icons/hicolor"
for size in 16 22 24 32 48 64 128 256 512; do
  mkdir -p "$hicolor/${size}x${size}/apps"
  magick "$logo/cardflip_logo_1024.png" -resize "${size}x${size}" \
    "$hicolor/${size}x${size}/apps/dev.cardflip.cardflip.png"
done
mkdir -p "$hicolor/scalable/apps"
cp "$logo/cardflip_logo.svg" "$hicolor/scalable/apps/dev.cardflip.cardflip.svg"
