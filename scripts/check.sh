#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
echo '[check] gofmt'
mapfile -t unformatted < <(gofmt -l ./cmd ./internal)
if ((${#unformatted[@]})); then
    printf 'Unformatted: %s\n' "${unformatted[@]}"
    exit 1
fi
echo '[check] go vet'
go vet ./...
echo '[check] untracked secrets and database files'
if git ls-files | grep -E '(^|/)(\.env|session\.sqlite3|history\.sqlite3|.*\.pem)$'; then
    echo 'ERROR: private data file tracked in git' >&2
    exit 1
fi
echo '[check] finished'
