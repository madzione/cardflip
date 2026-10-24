#!/usr/bin/env bash
# Installs the Linux release bundle with a desktop entry and icons, so the
# launcher, taskbar and Wayland compositors show Cardflip's name and icon.
#
#     tool/install_linux.sh               install for the current user (~/.local)
#     tool/install_linux.sh --uninstall   remove that install again
#     sudo PREFIX=/usr/local tool/install_linux.sh   install system-wide
#
# In the repository, build the bundle first with `fluffy build linux`. The
# release tarball from tool/package.sh ships this script as install.sh next to
# bundle/ and packaging/. The collection lives in
# ~/.local/share/dev.cardflip.cardflip/ and is never touched by this script.
set -euo pipefail

here="$(cd "$(dirname "$0")" && pwd)"
root="$(cd "$here/.." && pwd)"
app_id="dev.cardflip.cardflip"
prefix="${PREFIX:-$HOME/.local}"
app_dir="$prefix/lib/cardflip"
bin_link="$prefix/bin/cardflip"
desktop_file="$prefix/share/applications/$app_id.desktop"
icon_root="$prefix/share/icons/hicolor"

refresh_caches() {
  if command -v update-desktop-database >/dev/null; then
    update-desktop-database -q "$prefix/share/applications" || true
  fi
  if command -v gtk-update-icon-cache >/dev/null && [[ -f "$icon_root/index.theme" ]]; then
    gtk-update-icon-cache -q -t "$icon_root" || true
  fi
}

remove_icons() {
  local icon
  for icon in "$icon_root"/*/apps/"$app_id".{png,svg}; do
    if [[ -e "$icon" ]]; then
      rm -f "$icon"
    fi
  done
}

if [[ "${1:-}" == "--uninstall" ]]; then
  if [[ -x "$app_dir/cardflip" ]]; then
    rm -rf "$app_dir"
  fi
  if [[ -L "$bin_link" ]]; then
    rm -f "$bin_link"
  fi
  rm -f "$desktop_file"
  remove_icons
  refresh_caches
  echo "Uninstalled Cardflip from $prefix."
  exit 0
fi

case "$(uname -m)" in
  x86_64) arch=x64 ;;
  aarch64 | arm64) arch=arm64 ;;
  *) echo "Unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac
bundle="$root/build/linux/$arch/release/bundle"
packaging="$root/linux/packaging"
if [[ -x "$here/bundle/cardflip" ]]; then
  bundle="$here/bundle"
  packaging="$here/packaging"
fi
if [[ ! -x "$bundle/cardflip" ]]; then
  echo "No release bundle at $bundle; run \`fluffy build linux\` first." >&2
  exit 1
fi

# Replace an earlier install, but never delete a directory that isn't one.
if [[ -e "$app_dir" && ! -x "$app_dir/cardflip" ]]; then
  echo "$app_dir exists and isn't a Cardflip install; not overwriting it." >&2
  exit 1
fi
rm -rf "$app_dir"
mkdir -p "$app_dir" "$(dirname "$bin_link")" "$(dirname "$desktop_file")" "$icon_root"
cp -r "$bundle"/. "$app_dir"/
ln -sfn "$app_dir/cardflip" "$bin_link"

sed "s|^Exec=.*|Exec=$app_dir/cardflip|" \
  "$packaging/$app_id.desktop" >"$desktop_file"

remove_icons
cp -r "$packaging/icons/hicolor"/. "$icon_root"/

refresh_caches
echo "Installed Cardflip to $app_dir."
echo "It shows up in the application launcher as \"Cardflip\"."
