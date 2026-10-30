# WhatsCLI release process and version control

- **Version format:** `vMAJOR.MINOR.PATCH[-beta.N]` (SemVer prerelease). Example first development release: `v0.2.1-beta.1`.
- **Branches:** `main` is the integration branch; production-ready status is *not* implied by a tag.
- **One-way versioning:** never mutate an existing tag or silently replace published release assets; publish `-beta.2` for fixes.
- **Build provenance:** `whatscli --version` embeds the exact tag and Git commit SHA via Go `-ldflags -X`.

## Publish a new release

1. Run `go test -race -count=1 ./...`, `go vet ./...`, shell script tests and the terminal demo smoke test; CI must be green. Before major bumps, back up and test schema migrations.
2. Edit `releases/REQUEST` on `main` to the desired new SemVer tag. Every change to this file triggers the `Versioned release` workflow. Alternatively use *Actions → Versioned release → Run workflow* and pass a tag.
3. The release workflow validates the version, checks that it is not reusing an existing tag, runs tests, builds binaries for Linux/macOS/Windows with CGO disabled, embeds version/commit, generates a `SHA256SUMS`, and creates the GitHub Release and matching version tag. Prerelease suffixes automatically mark the release **Pre-release**.
4. Inspect the Actions run, the GitHub Release page, downloadable assets, and checksum verification. **Only after success**, update `releases/current.txt` to the new tag; that advances the one-line install channel. Verify installation and uninstallation on Ubuntu and the `--version`, `--doctor` and `--demo` commands.
5. Document the result in `docs/CHANGELOG.md` and link the release to the associated GitHub commit. Keep manual WhatsApp account acceptance separate from compilation/CI success.

## Versioning policy

- `-beta.N`: unverified/experimental account integration; may break compatibility. Use a test account and back up history.
- `v0.x.y`: pre-1.0 APIs may change, migration logic should preserve already observed local history.
- `v1.0.0`: reserved for the full release criteria in `TASKS.md` and `docs/ACCEPTANCE.md`, including real SSH + WhatsApp account pairing/message/media acceptance.

No credentials, QR codes, chat archives or private messages should enter a public release or GitHub Actions logs.
