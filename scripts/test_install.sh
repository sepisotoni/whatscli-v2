#!/usr/bin/env bash
# Deterministic offline installer checks, no internet or WhatsApp account.
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work="$(mktemp -d)"
trap 'rm -rf -- "$work"' EXIT
mkdir -p "$work/fixtures" "$work/fakebin"
cat > "$work/fixtures/whatscli-linux-amd64" <<'BIN'
#!/usr/bin/env sh
if [ "$1" = '--version' ]; then echo 'WhatsCLI v0.2.1-beta.1 (test)'; exit 0; fi
exit 1
BIN
chmod +x "$work/fixtures/whatscli-linux-amd64"
(cd "$work/fixtures" && sha256sum whatscli-linux-amd64 > SHA256SUMS)
cat > "$work/fakebin/curl" <<'CURL'
#!/usr/bin/env bash
set -euo pipefail
out=''; url=''
while (($#)); do
  case "$1" in
    -o) out="$2"; shift 2 ;;
    *) url="$1"; shift ;;
  esac
done
case "$url" in
  */whatscli-linux-amd64) cp "$WHATSCLI_TEST_FIXTURES/whatscli-linux-amd64" "$out" ;;
  */SHA256SUMS) cp "$WHATSCLI_TEST_FIXTURES/SHA256SUMS" "$out" ;;
  *) echo "unexpected url $url" >&2; exit 1 ;;
esac
CURL
chmod +x "$work/fakebin/curl"
# Tests target Ubuntu CI (linux-amd64), and deliberately leave all user data alone.
export WHATSCLI_TEST_FIXTURES="$work/fixtures"
export PATH="$work/fakebin:$PATH"
mkdir -p "$work/bin"
HOME="$work/home" bash "$root/scripts/install.sh" --version v0.2.1-beta.1 --install-dir "$work/bin"
"$work/bin/whatscli" --version | grep -q 'v0.2.1-beta.1'
# A failed checksum must not replace a working existing installation.
printf '0%.0s' {1..64} | awk '{print $0 "  whatscli-linux-amd64"}' > "$work/fixtures/SHA256SUMS"
if bash "$root/scripts/install.sh" --version v0.2.1-beta.1 --install-dir "$work/bin" > "$work/failed.log" 2>&1; then
  echo 'installer accepted wrong checksum' >&2; exit 1
fi
"$work/bin/whatscli" --version | grep -q 'v0.2.1-beta.1'
# Only remove the executable, not the archive.
mkdir -p "$work/home/.local/share/whatshell-go"
printf 'keep history' > "$work/home/.local/share/whatshell-go/history.sqlite3"
HOME="$work/home" bash "$root/scripts/uninstall.sh" --install-dir "$work/bin"
test ! -e "$work/bin/whatscli"
test -f "$work/home/.local/share/whatshell-go/history.sqlite3"
echo 'Installer/uninstaller offline tests passed'
