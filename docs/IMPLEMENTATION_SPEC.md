# WhatsCLI — Detailed Implementation Specification

**Companion tracker:** [TASKS.md](../TASKS.md) · **Operational test plan:** [ACCEPTANCE.md](ACCEPTANCE.md) · **Architecture:** [ARCHITECTURE.md](ARCHITECTURE.md)

This is the technical contract for a **Go-first, SSH-native, terminal WhatsApp client** using **WhatsMeow**, **tview/tcell**, and **pure-Go SQLite**. It explains exactly what each task must include and what proves it works. A compile-successful transport is not the same as a verified WhatsApp integration. Feature availability depends on the WhatsMeow version and WhatsApp's evolving protocol.

## 1. Runtime and project structure

**Linked tasks:** CORE-01, REL-01. **Priority:** P0.

- Maintain `cmd/whatshell` as the single executable, with small `internal/` packages for configuration, transport, archive, terminal UI, diagnostics and media transfer. Avoid circular dependencies and keep UI code off the WhatsApp event-handling thread.
- Go module path should eventually be aligned with `github.com/sepisotoni/WhatsCLI`. Do this as one atomic refactor across `go.mod`, imports, docs, release paths and CI; no broken transitional imports.
- Build on Linux amd64/arm64 first, with Windows/macOS releases where verified. The application must not require a running desktop environment, web browser, node process, system service or CGO SQLite.
- Provide clear `--help`, `--version` and `--doctor` commands. Operational modes must work **without** pairing WhatsApp. A non-interactive command must never open TUI nor connect remotely.
- Do not invent SSH server behavior: this is a terminal application run *inside* an existing SSH session, not an SSH daemon. Use ordinary SSH user permissions, `tmux` where helpful, and SFTP/SCP for files.

**Acceptance:** `gofmt`, `go vet`, `go test -race`, `CGO_ENABLED=0 go build`; start commands do not leak secrets, and usage matches the README. CI must enforce dependency hashes and release tags.

## 2. SQLite archive and data integrity

**Linked tasks:** DATA-01, DATA-02, DATA-03, MSG-02. **Priority:** P0.

**Current tables:** `messages`, `message_versions`, `message_receipts`, `message_media`. Key incoming records by `(chat_jid, message_id)`; timestamps and chat IDs must never be guessed. Preserve **observed** previous message bodies on edits and mark deleted messages with tombstones. Never claim recovery of unseen/view-once/disappearing content.

**Still required:**

1. Introduce `schema_migrations` or SQLite `PRAGMA user_version`; migrations must be forward-only, transactional, idempotent and covered by a test from each supported older schema.
2. Ensure late history-sync, receipts and live events can arrive in any order without accidentally resurrecting a tombstone or duplicating a message; clarify group sender vs chat identity in event parsing.
3. Paginate queries by stable `(timestamp_ns, rowid/message_id)` cursor, not only `LIMIT`. Index common chat-time and search paths. Add optional FTS5 with migration and escape the user input safely.
4. Make receipt levels monotonic and preserve them when the message itself has not arrived yet. Deduplicate retransmission. Never block the WhatsApp callback on long-running UI work.
5. Verify atomic snapshot/restore after WAL writes; exclude device/session credentials from archive backups. Add integrity checks and import/export with explicit user consent.
6. Define retention settings and the privacy limits of keeping copies of subsequently deleted messages. Optional backup encryption must have a documented key derivation/recovery and lost-key behavior.

**Tests:** multi-edit/revoke, revoked-before-original, duplicate history batches, out-of-order event ordering, concurrent writes, corrupt database failure, read-only `PRAGMA quick_check`, backup consistency and an empty archive. **Real account:** A04–A07 and A11.

## 3. WhatsApp transport and account lifecycle

**Linked tasks:** AUTH-01/02/03, MSG-01. **Priority:** P0.

- WhatsMeow Multi-Device QR pairing must show a scannable QR in the SSH terminal (no secrets in logs) and preserve login data in a private session DB. Show clear states: unpaired, QR pending, connected, reconnecting, disconnected, logged out and fatal error.
- Reconnect with jittered bounded exponential backoff, network detection and an explicit cancel path. Never start two concurrent connection attempts; handle device revocation and re-pairing safely.
- Separate transport from archive and UI through a minimal interface to support deterministic fake transport tests. Events should be processed in bounded order through worker/queue, with idempotent DB writes and safe shutdown.
- Sending must validate full international phone/JID input and handle group JIDs, per-message IDs, local pending state, remote errors and retries. Never silently resend a message when delivery is uncertain. If adding an outgoing queue, store idempotency keys and explicit retry policy.
- Add real test-account checks for initial pairing, persistence after restart, incoming and outgoing text, groups, revocation, logout, reconnect, and offline resume. Do not embed a real user account or tokens in CI.

**Acceptance:** A01–A04, A14 pass against a consenting test account; no leaked QR/session data and no runaway reconnect loops.

## 4. Terminal experience and SSH

**Linked tasks:** UI-01/02/03. **Priority:** P0/P2.

- Display dark emerald/slate WhatsApp-inspired UI using terminal capabilities without relying on truecolor; render accessible mono/color fallback. Layout: chat sidebar, scrollable messages, footer/status, compose, command suggestions/modal views.
- Correct resizing, wrapping, mouse and keyboard focus in `xterm`, `screen`, `tmux`, narrow 40×20 terminals, 80×24, wide layouts, differing Unicode width and terminal `TERM` configurations. QR may require a bigger screen; explain a safe fallback.
- Global keys must not hijack modal form entry. Tab should complete a slash command when appropriate; otherwise move focus predictably. `Ctrl+C`, escape, backspace, mouse scroll, search, selection, and paste must not crash or leak contents.
- Escape untrusted message content and usernames before passing to `tview` dynamic colors/markup. Define behavior for emojis, combining codepoints, zero-width joiners, RTL text and long lines.
- Support configurable keybindings/themes later via per-user local config, with documented defaults. Avoid excessive re-rendering over high-latency SSH.
- Replace illustrative README mockups with real anonymized PTY screenshots once A12/A13 pass; never publish personal messages, phone numbers or live QR codes.

**Tests:** command completion, mock terminal screen tests, resize fuzzing, scrolling/focus tests and actual SSH screenshots. **Real account:** A01, A12 and A13.

## 5. Messaging and chat UX

**Linked tasks:** MSG-03/04. **Priority:** P1.

- Message rows should show sender label (including group sender), local time, outgoing/incoming state, receipt ticks, edited/deleted marker, and stable message IDs where actionable.
- Add unread counters, mark read, pinned/archived/muted chats, draft preservation, message pagination, reply threading/quote selection, search navigation and a visible send-in-progress/error state.
- Preserve backwards compatibility of slash commands: `/chat`, `/search`, `/versions`, `/attach`, `/download`, `/poll`, `/backup`, `/logout`, `/help`, `/quit`. Document new commands and test parsing including quoted filenames, spaces and Unicode.
- Presence, typing, reaction, forward and delete-for-everyone features must be implemented only when supported by the backend and privacy permissions; fail clearly when unsupported.

**Acceptance:** fast keyboard-only navigation, deterministic dummy message rendering, and real send/receive with an authorized test user.

## 6. Contacts, groups and chat management

**Linked tasks:** CONTACT-01/02. **Priority:** P1.

- Get contact-friendly names from WhatsMeow and store separately from user-defined local aliases. Do not mistake JIDs for phone numbers; handle changed contact identities and `@lid` compatibility.
- Group UI should show group title, sender names, membership/update events and unread counts; don't conflate group JID with participant JID.
- Add searchable chat selector, local alias editor, pin/mute/archive toggles, favorites and clear unread counters. Define whether each action is local-only or synchronized with WhatsApp.
- Cache presentation metadata transactionally and invalidate on contact/group updates. Provide fallback to unambiguous JID, never silently wrong identity mapping.

**Acceptance:** renamed contact, unknown contact, group sender and multiple JID forms show correct identities after restart.

## 7. Media, files, audio and video

**Linked tasks:** MEDIA-01/02/03/04. **Priority:** P1/P2.

- Existing `/attach` accepts bounded JPEG/PNG or document from private SFTP `uploads/`; `/download MESSAGE_ID` saves received regular attachments to private `downloads/`. Preserve path-traversal and symlink tests and avoid running downloaded content.
- Move large transfers to streaming/chunked I/O with progress and cancel; enforce per-type size limits, MIME detection and file-length checks. Never trust remote filename or extension. Media keys, hashes and URLs are sensitive and must not be logged.
- Support thumbnails and lightweight text/ANSI previews first. Implement send/receive audio and voice notes, with safe explicit playback integration or export. Do not auto-open unknown content.
- Video is likely a terminal/host capability rather than native full-frame TUI playback: detect Kitty/Sixel support where possible and offer low-bandwidth ANSI fallback or external player. Preserve SSH responsiveness.
- Use WhatsMeow-supported media API, encryption and retransmission semantics; test expiring URLs, malformed media, interrupted transfers and small/big files on real devices.

**Acceptance:** A08/A09, 16 MiB guard regression and SFTP escape tests; later media E2E checks and throughput budget.

## 8. Polls, statuses and advanced protocol features

**Linked tasks:** POLL-01, STATUS-01. **Priority:** P1/P2.

- Poll creation supports a question and 2–12 unique nonempty options as a single-choice WhatsApp poll. Verify actual delivery and option text. Add vote submission, vote decryption/aggregation, changed votes, author visibility and persistence when supported.
- Statuses require a separate feasibility check against exact WhatsMeow API and platform rules: list/read statuses, download attached media, post supported types, replies and expiry. Do not display statuses as permanent chat history unless retention policy is explicit.
- Later enhancements: reactions, stickers, locations, contacts and other rich message types with clear unsupported placeholders.

**Acceptance:** A10, live vote from a separate account, multiple-option source-specific cases, status send/read on authorized test account.

## 9. Operations, CLI and observability

**Linked tasks:** OPS-01, SEC-01/02. **Priority:** P0/P1/P2.

- Implement `--doctor` as an **offline** check, with explicit exit codes and no remote calls. Validate root/uploads/downloads directories are not symlinks and are private, check existing database files are regular and mode 0600 (Unix), and perform **read-only** `PRAGMA quick_check` on an existing history archive. Missing credentials/empty history are informative, **not** errors. Never print message bodies, auth tokens, raw QR codes, or sensitive paths without reason.
- `--doctor` may create the default private application data directories as `config.Resolve` does, but must **not create** history or session databases or send network requests. Handle permission failures and corruption clearly.
- Define safe logging levels, redaction, state transitions, graceful cancellation, resource use, backup verification, restore dry runs, crash reports without secrets and privacy-first defaults.
- Session credentials must remain 0600 on Unix. Encryption-at-rest and multiple account directories need an explicit threat model and passphrase/key-store design; don't add a fake toggle that does not encrypt.

**Tests:** doctor with empty workspace, healthy/corrupt archive, symlink folders, insecure file modes, and no accidental session creation. Run in CI and document output.

## 10. Testing, quality gates and acceptance

**Linked tasks:** QA-01/02/03. **Priority:** P0/P1/P2.

- Keep GitHub Actions gates for Go formatting, vet, `go test -race ./...`, coverage, `CGO_ENABLED=0 go build`, golangci-lint, Staticcheck, gosec, and govulncheck. Avoid a green badge by disabling a failing gate; document focused security suppressions with rationale.
- Add deterministic fake-transport event fixtures, table-driven parsing tests, property/fuzz tests for message and path handling, concurrency races and database migration compatibility.
- Distinguish: **unit tests** (no phone), **GitHub CI** (no phone), **manual live WhatsApp account** (A01–A14), **SSH PTY rendering**, and **long-running reliability**. Never claim live integration passed because CI is green.
- Keep [ACCEPTANCE.md](ACCEPTANCE.md) as the canonical manual checklist. Record commit, Go version, operating system, terminal and `TERM` on each attempt with redacted logs.
- Definition of feature completion: coded, automated tests, docs, security review as needed, real-device acceptance if relevant, and no open release-blocking bugs.

## 11. Builds, packaging and release

**Linked tasks:** REL-01/02. **Priority:** P0/P2.

- Pin and commit `go.sum`, avoid running `go mod tidy` as a surprise step in CI that silently updates dependencies, and require `go mod verify`.
- Build Linux amd64/arm64 and supported macOS/Windows targets without CGO; verify releases include LICENSE, README, SECURITY, checksums and version metadata, and publish release assets only for reviewed tags.
- Use semver `v0.x` prereleases until real SSH and WhatsApp checks pass; do not call `v1.0` while A01–A14 remain unverified. Add automated changelog, reproducibility checks and optionally SBOM/build attestations.
- Maintain concise install/update/uninstall guides and backup/restore warning, especially that a deleted local session credential can require a new WhatsApp QR link.

**Release gates:** All quality jobs green on tagged SHA, manual test evidence, no secret artifacts, documented supported OS/terminal versions and a rollback procedure.

## Priority sequence

1. **P0:** doctor diagnostics, schema migrations, live QR/SSH acceptance, robust offline/history behavior, initial tagged prerelease.
2. **P1:** reconnect with safe retry, contacts/groups, chat UX, reliable media, polls/voting, encrypted backups and fake-transport tests.
3. **P2:** statuses/video, voice playback enhancements, multi-account support, theme customization and v1.0 hardening.

**Progress reporting format after every substantive sprint:** commit SHA, tasks changed (IDs), CI link and exact results, manual tests performed/not performed, newly discovered bugs, and updated weighted score in TASKS.md. If no real-device test occurred, say so plainly.