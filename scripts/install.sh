#!/usr/bin/env bash
# Install a verified WhatsCLI release into a user-owned bin directory.
set -euo pipefail

repo='sepisotoni/WhatsCLI'
install_dir="${WHATSCLI_INSTALL_DIR:-${HOME:?HOME is required}/.local/bin}"
channel_url="${WHATSCLI_CHANNEL_URL:-https://raw.githubusercontent.com/${repo}/main/releases/current.txt}"
release_base="${WHATSCLI_RELEASE_BASE:-https://github.com/${repo}/releases/download}"
version=''

usage() {
  cat <<'HELP'
WhatsCLI installer (Linux amd64/arm64 and macOS amd64/arm64)
Usage: bash install.sh [--version vX.Y.Z[-beta.N]] [--install-dir DIR]

Defaults to the tested prerelease channel and $HOME/.local/bin.
To upgrade, rerun the installer. Local WhatsApp sessions/history are untouched.
Override WHATSCLI_INSTALL_DIR to choose a directory already on PATH.
HELP
}

while (($#)); do
  case "$1" in
    --version) [[ $# -ge 2 ]] || { echo 'Missing --version argument' >&2; exit 2; }; version="$2"; shift 2 ;;
    --install-dir) [[ $# -ge 2 ]] || { echo 'Missing --install-dir argument' >&2; exit 2; }; install_dir="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage >&2; exit 2 ;;
  esac
done

for tool in curl mktemp uname awk install mv; do
  command -v "$tool" >/dev/null 2>&1 || { echo "Required command missing: $tool" >&2; exit 1; }
done
if command -v sha256sum >/dev/null 2>&1; then
  checksum_tool='sha256sum'
elif command -v shasum >/dev/null 2>&1; then
  checksum_tool='shasum'
else
  echo 'SHA-256 utility required (sha256sum or shasum)' >&2; exit 1
fi

if [[ -z "$version" ]]; then
  version="$(curl --fail --silent --show-error --location --proto '=https' --tlsv1.2 "$channel_url")" || {
    echo 'Could not fetch the current tested release. Pass --version vX.Y.Z-beta.N.' >&2
    exit 1
  }
  version="${version//$'\r'/}"
  version="${version//$'\n'/}"
fi
if [[ ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-(alpha|beta|rc)\.[0-9]+)?$ ]]; then
  echo "Invalid release version: $version" >&2
  exit 2
fi

case "$(uname -s)" in
  Linux) os='linux' ;;
  Darwin) os='darwin' ;;
  *) echo 'Unsupported system: only Linux and macOS are currently packaged' >&2; exit 1 ;;
esac
case "$(uname -m)" in
  x86_64|amd64) arch='amd64' ;;
  aarch64|arm64) arch='arm64' ;;
  *) echo 'Unsupported CPU: only amd64 and arm64 are currently packaged' >&2; exit 1 ;;
esac
asset="whatscli-${os}-${arch}"

workdir="$(mktemp -d)"
trap 'rm -rf -- "$workdir"' EXIT
curl --fail --silent --show-error --location --retry 3 --proto '=https' --tlsv1.2 \
  "${release_base}/${version}/${asset}" -o "${workdir}/${asset}"
curl --fail --silent --show-error --location --retry 3 --proto '=https' --tlsv1.2 \
  "${release_base}/${version}/SHA256SUMS" -o "${workdir}/SHA256SUMS"

expected="$(awk -v file="$asset" '$2 == file { print $1 }' "${workdir}/SHA256SUMS")"
if [[ ! "$expected" =~ ^[[:xdigit:]]{64}$ ]]; then
  echo "No valid SHA-256 checksum found for ${asset}" >&2
  exit 1
fi
if [[ "$checksum_tool" == 'sha256sum' ]]; then
  actual="$(sha256sum "${workdir}/${asset}" | awk '{print $1}')"
else
  actual="$(shasum -a 256 "${workdir}/${asset}" | awk '{print $1}')"
fi
if [[ "$(printf %s "$actual" | tr '[:upper:]' '[:lower:]')" != "$(printf %s "$expected" | tr '[:upper:]' '[:lower:]')" ]]; then
  echo "Checksum mismatch for ${asset}; installation aborted" >&2
  exit 1
fi

mkdir -p -- "$install_dir"
target="${install_dir%/}/whatscli"
if [[ -e "$target" || -L "$target" ]]; then
  if ! ("$target" --version 2>/dev/null | grep -q '^WhatsCLI '); then
    echo "Refusing to overwrite unrecognized executable: $target" >&2
    exit 1
  fi
fi
stage="$(mktemp "${install_dir%/}/.whatscli.XXXXXXXX")"
trap 'rm -rf -- "$workdir"; rm -f -- "${stage:-}"' EXIT
install -m 0755 "${workdir}/${asset}" "$stage"
mv -f -- "$stage" "$target"
echo "Installed ${version} to ${target} (SHA-256 verified)"
"$target" --version
case ":${PATH}:" in
  *":${install_dir%/}:"*) ;;
  *) echo "Add ${install_dir%/} to PATH, e.g. export PATH=\"${install_dir%/}:\$PATH\"" ;;
esac
