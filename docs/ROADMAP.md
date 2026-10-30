# Whatshell development roadmap

## v0.1 — Foundation (written, not fully integration-tested)
- QR linking, session storage, SSH-friendly tview, outgoing text and incoming text.
- SQLite history with edit versions and deletion tombstones.
- Slash palette, chat selection, initial CI configuration.

## v0.2 — Offline reliability and rich-content transport (source implemented)
- Import WhatsMeow HistorySync events; persist receipt levels and media pointers.
- Text search, compact chat picker/sidebar, improved modal commands.
- SFTP-root JPEG/PNG image and document uploads; on-demand incoming media downloads.
- Poll **creation only**; no vote collection/results yet.
- GitHub tag-based Linux/macOS/Windows build and SHA-256 release checksums.
- Add unit tests for archive query methods and path security.
- **Gate:** compile and run the full toolchain in CI, link test account, and execute a real SSH/tmux acceptance test.

## v0.2.1 — Schema upgrades and local aliases\n- Transactional schema v2 upgrade from pre-versioned archives, preserving cached messages and previous versions.\n- Private local-only `/alias NAME` and `/alias -` for persistent chat display names.\n- WhatsMeow initial reconnect for linked accounts; no blind message retransmission.\n- Unit tests for migrations, downgrade refusal and aliases.\n\n## v0.3 — Message UX and protocol parity
- Contacts, group names and avatars; delivery status per group member.
- Poll voting, live result decryption, persistent poll metadata.
- Reliable reconnect, offline outgoing queue, media progress and cancellation.
- Mark read, typing/presence, chats pinned/muted/unread, message pagination.
- Image previews, safe external player controller, voice note upload/playback.
- Schema versioning, encrypted backups, restore tests, and configurable retention.

## v0.4 — Status and media rendering
- WhatsApp statuses (where supported by backend), text/media posting and replies.
- Video frame renderer negotiation: Kitty/Sixel optional, low-bandwidth ANSI fallback.
- Terminal/tmux notification hooks and user-defined quiet hours.
- TUI screenshot tests and resize fuzzing under different TERM profiles.

## v1.0 — Release quality
- Account integration tests, portable deterministic builds, memory/CPU profiling,
  fuzz tests, secure runtime audits, error-budget and recovery testing.

No phase is considered complete until its code has compiled and passed
interactive tests in a real SSH PTY against an authorized WhatsApp account.
