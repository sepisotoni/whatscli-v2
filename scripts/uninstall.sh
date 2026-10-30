#!/usr/bin/env bash
# Remove only the WhatsCLI executable; preserve user's chat history/session.
set -euo pipefail
install_dir="${WHATSCLI_INSTALL_DIR:-${HOME:?HOME is required}/.local/bin}"
usage() {
  cat <<'HELP'
WhatsCLI uninstaller
Usage: bash uninstall.sh [--install-dir DIR]
Removes only the verified whatscli executable. Does NOT erase chat history,
WhatsApp linked-device credentials, backups, uploads or downloads.
HELP
}
while (($#)); do
  case "$1" in
    --install-dir) [[ $# -ge 2 ]] || { echo 'Missing --install-dir argument' >&2; exit 2; }; install_dir="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage >&2; exit 2 ;;
  esac
done
target="${install_dir%/}/whatscli"
if [[ ! -e "$target" && ! -L "$target" ]]; then
  echo "WhatsCLI is not installed at ${target}"
  exit 0
fi
if [[ -L "$target" || ! -f "$target" ]]; then
  echo "Refusing to remove a symlink or non-regular file: ${target}" >&2
  exit 1
fi
if ! ("$target" --version 2>/dev/null | grep -q '^WhatsCLI '); then
  echo "Refusing to remove an unrecognized program: ${target}" >&2
  exit 1
fi
rm -- "$target"
echo "Uninstalled WhatsCLI from ${target}"
echo 'Your WhatsApp login/session, archived messages and backups are preserved.'
