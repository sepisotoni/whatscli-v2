# Change log

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
