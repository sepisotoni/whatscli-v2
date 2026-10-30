# Change log

## Unreleased — Animated welcome and consent (2026-10-30)

- Added a terminal-native animated WhatsCLI welcome, separate disclaimer page, consent checkbox and explicit Agree & continue action before QR pairing.
- Disclosure covers unofficial status, possible service/account restrictions, local plaintext history/session risks and no warranties.
- Keyboard/mouse navigation, reduced-motion option, responsive panels, and agreement-gate Go/PTY regression tests.
- WhatsApp's official branding is not used as WhatsCLI's app logo, in keeping with Meta's independent-client branding restrictions.

## Earlier first-run onboarding and pairing UX (2026-10-30)

- Added a full-screen welcome experience, a dedicated WhatsApp QR pairing screen, keyboard/mouse action controls and a QR-fullscreen fallback for small terminals.
- Keep the QR visible across pairing status updates; open the chat workspace only on a confirmed Connected event.
- Added `--demo-onboarding` with fictional local data, unit tests and terminal pseudo-TTY smoke coverage.

## v0.2.1-beta.1 — Experimental prerelease (2026-10-30)

- Tagged release binaries for Linux/macOS/Windows with embedded `--version` and exact commit.
- User-local verified `curl` installer, safe uninstall retaining chats/credentials, checksum verification and installer tests.
- GitHub release workflow with checksums and a separate opt-in default install channel.
- SQLite schema v2 upgrades, local aliases, offline `--doctor`, isolated `--demo` and PTY resize tests.
- **Not yet verified:** actual WhatsApp QR/device linking, live messaging or an authorized SSH session. Not a stable release.

## Earlier v0.2 development (2026-10-30)

- **WhatsApp:** ingest Multi-Device history sync events, archive incoming media metadata for later downloads, and persist delivered/read receipts in a separate monotonic table.
- **Media:** added bounded (16 MiB) SFTP-root upload of JPEG/PNG images and generic documents, plus on-demand, collision-resistant attachment downloads for supported message types.
- **Polls:** create single-choice WhatsApp polls using WhatsMeow's poll builder. Voting and results are still pending.
- **Archive:** literal case-insensitive search across messages, an exclusive private `/backup` snapshot, including previously cached deleted content. Improved receipt idempotence and added regression tests.
- **Terminal:** Ctrl+F search, Ctrl+L compact sidebar toggle, /chats picker, /attach, /download and /poll. Updated command palette and inline read receipts.
- **Security:** explicit private upload/download directories, filename sanitization, symlink and path-traversal defenses, exclusive download creation, no executable media auto-opening, and file-permission tests.
- **Delivery:** tagged GitHub release workflow producing CGO-free Linux, macOS and Windows binaries plus SHA-256 checksums.

**Verification limits:** Go 1.27 dependencies could not be fetched in this environment. Syntax formatting and independent standard-library tests passed, but full Go compilation, protocol integration and CI remain unverified.

## v0.1 — Original foundation

QR pairing, session persistence, incoming and outgoing text, SQLite message versions/tombstones, SSH-friendly TUI and static analysis configuration.
