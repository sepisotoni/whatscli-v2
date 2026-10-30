# WhatsCLI — Project Tasks & Progress

**Repository:** https://github.com/sepisotoni/WhatsCLI  
**Last planning review:** 2026-10-30  
**Stage:** v0.2 development → v0.3  
**Estimated overall completion: 42%** (weighted scope estimate; not a verified measure of functionality).  
**Evidence:** The Go CI has passed, including race tests, build, lint, static analysis and security scans. Real WhatsApp account + SSH acceptance cases [A01–A14](docs/ACCEPTANCE.md) have **not** been completed.

## How this plan works

- Every task has a stable **ID**, priority (**P0** must work before beta; **P1** useful for beta; **P2** advanced/optional), a checkbox, and a link to its exact section in the [detailed implementation specification](docs/IMPLEMENTATION_SPEC.md).
- `[x]` means implemented and supported by source/automated tests **only**. It does **not** certify real WhatsApp compatibility. `[ ]` means unfinished or awaiting acceptance.
- A feature is done only after implementation, tests, security review where relevant, documentation, and [real-device acceptance](docs/ACCEPTANCE.md) when applicable.
- Percentages are **planning estimates**, not automatic counts of checked boxes. Recalculate after meaningful milestones, not after cosmetic commits.

## Weighted completion score

| Workstream | Total weight | Earned | Evidence / biggest gap |
|---|---:|---:|---|
| Foundation & project tooling | 8 | 7 | Go app, CI, GitHub releases configured; no production tag |
| SQLite archive & sync reliability | 14 | 8 | Edit versions, tombstones, search, backup; migrations, recovery and full history proof outstanding |
| Auth, connectivity & messaging | 14 | 4 | WhatsMeow implementation compiles; account-level verification missing |
| Terminal UI, SSH & accessibility | 14 | 6 | Tview UI, commands, mouse, QR; real PTY and responsive acceptance missing |
| Contacts, groups & chat controls | 8 | 0 | Contact names, group member UX and pin/mute controls missing |
| Files, audio, images & video | 12 | 3 | JPEG/document upload and download source; streaming and real-media testing missing |
| Polls, statuses & advanced features | 10 | 2 | Poll creation source; voting, results and statuses missing |
| Security, privacy & operations | 10 | 6 | Private paths, basic safety, backup and source security; encrypted backup, restore and audits missing |
| Tests, diagnostics & release quality | 10 | 6 | CI, unit/race tests and acceptance plan; doctor added in this milestone; real integration and tagged release missing |
| **Total** | **100** | **42** | **Early development, not production-ready** |

## Milestone A — reliable first-run beta (P0)

- [x] **CORE-01**: Set up the Go project, dependencies, static binary build, CI, and release workflow. [Spec §1](docs/IMPLEMENTATION_SPEC.md#1-runtime-and-project-structure)
- [x] **DATA-01**: Persist observed messages, versions and deletion tombstones in private local SQLite storage. [Spec §2](docs/IMPLEMENTATION_SPEC.md#2-sqlite-archive-and-data-integrity)
- [x] **DATA-02**: Add offline chat search, read receipts, bounded queries and consistent local snapshot creation. [Spec §2](docs/IMPLEMENTATION_SPEC.md#2-sqlite-archive-and-data-integrity)
- [x] **UI-01**: Create keyboard/mouse TUI with chat list, compose, slash commands and QR renderer. [Spec §4](docs/IMPLEMENTATION_SPEC.md#4-terminal-experience-and-ssh)
- [x] **QA-01**: Enforce gofmt, vet, race tests, lint, static analysis, gosec and vulnerability checks in CI. [Spec §10](docs/IMPLEMENTATION_SPEC.md#10-testing-quality-gates-and-acceptance)
- [x] **OPS-01**: Add offline `--doctor` diagnostics with read-only history integrity check. [Spec §9](docs/IMPLEMENTATION_SPEC.md#9-operations-cli-and-observability)
- [ ] **AUTH-01**: Verify initial QR/pairing and linked-device persistence with a consenting test account (**A01, A02**). [Spec §3](docs/IMPLEMENTATION_SPEC.md#3-whatsapp-transport-and-account-lifecycle)
- [ ] **MSG-01**: Test one-to-one and group send/receive; fix message IDs, sender attribution, errors and duplication (**A03, A04**). [Spec §3](docs/IMPLEMENTATION_SPEC.md#3-whatsapp-transport-and-account-lifecycle)
- [ ] **MSG-02**: Verify history sync, edits, revocations, late arrivals, and delivered/read state (**A05–A07**). [Spec §2](docs/IMPLEMENTATION_SPEC.md#2-sqlite-archive-and-data-integrity)
- [ ] **UI-02**: Real SSH/tmux/PTY run across minimum dimensions, mouse, QR, resizing and Ctrl+C (**A12, A13**). [Spec §4](docs/IMPLEMENTATION_SPEC.md#4-terminal-experience-and-ssh)
- [ ] **AUTH-02**: Verify logout revokes the correct session while preserving history (**A14**). [Spec §3](docs/IMPLEMENTATION_SPEC.md#3-whatsapp-transport-and-account-lifecycle)
- [ ] **DATA-03**: Version schema migrations and test upgrades from both an empty database and previous schema. [Spec §2](docs/IMPLEMENTATION_SPEC.md#2-sqlite-archive-and-data-integrity)
- [ ] **REL-01**: Commit reproducible dependency sums, create first tested prerelease, and verify binary artifacts on Linux/macOS/Windows. [Spec §11](docs/IMPLEMENTATION_SPEC.md#11-builds-packaging-and-release)

## Milestone B — useful everyday terminal messaging (P1)

- [ ] **AUTH-03**: Add reconnect/backoff, offline outgoing queue, send retry policy and clear connection state. [Spec §3](docs/IMPLEMENTATION_SPEC.md#3-whatsapp-transport-and-account-lifecycle)
- [ ] **MSG-03**: Implement proper message pagination, sent/read indicators, typing/presence when permitted, and manual mark-read. [Spec §5](docs/IMPLEMENTATION_SPEC.md#5-messaging-and-chat-ux)
- [ ] **MSG-04**: Add reply/quote, emoji-safe rendering, forwarding if supported, and clear unsuccessful-send handling. [Spec §5](docs/IMPLEMENTATION_SPEC.md#5-messaging-and-chat-ux)
- [ ] **CONTACT-01**: Fetch/cache human-readable contacts and group titles; local aliases and reliable identity handling. [Spec §6](docs/IMPLEMENTATION_SPEC.md#6-contacts-groups-and-chat-management)
- [ ] **CONTACT-02**: Add contact search, pinned/muted/unread chat controls, and sensible group sender labels. [Spec §6](docs/IMPLEMENTATION_SPEC.md#6-contacts-groups-and-chat-management)
- [ ] **MEDIA-01**: Verify real `/attach` and `/download` behavior and media errors (**A08, A09**). [Spec §7](docs/IMPLEMENTATION_SPEC.md#7-media-files-audio-and-video)
- [ ] **MEDIA-02**: Add streaming uploads/downloads, configurable limits, progress, cancellation, and safe thumbnail detection. [Spec §7](docs/IMPLEMENTATION_SPEC.md#7-media-files-audio-and-video)
- [ ] **MEDIA-03**: Support audio/voice-note sending and terminal-friendly download/playback hooks. [Spec §7](docs/IMPLEMENTATION_SPEC.md#7-media-files-audio-and-video)
- [ ] **POLL-01**: Verify poll creation (**A10**) and implement vote selection, encrypted updates, and result history. [Spec §8](docs/IMPLEMENTATION_SPEC.md#8-polls-statuses-and-advanced-protocol-features)
- [ ] **SEC-01**: Test full encrypted backup/restore, secret handling, secure temporary files and recovery after crashes. [Spec §9](docs/IMPLEMENTATION_SPEC.md#9-operations-cli-and-observability)
- [ ] **QA-02**: Build deterministic fake transport tests for reconnect, receipts, media failures and event ordering. [Spec §10](docs/IMPLEMENTATION_SPEC.md#10-testing-quality-gates-and-acceptance)

## Milestone C — polished feature parity / v1.0 goals (P2)

- [ ] **MEDIA-04**: Add video download playback delegation and optional Kitty/Sixel/ANSI still/frame previews with capability detection. [Spec §7](docs/IMPLEMENTATION_SPEC.md#7-media-files-audio-and-video)
- [ ] **STATUS-01**: View and send permitted WhatsApp statuses and handle replies; verify backend API support first. [Spec §8](docs/IMPLEMENTATION_SPEC.md#8-polls-statuses-and-advanced-protocol-features)
- [ ] **SEC-02**: Define multi-account and encryption-at-rest architecture, user-approved retention/export/deletion, and threat model. [Spec §9](docs/IMPLEMENTATION_SPEC.md#9-operations-cli-and-observability)
- [ ] **UI-03**: Add user-configurable themes, rich keybindings, text accessibility, contrast and terminal screenshot tests. [Spec §4](docs/IMPLEMENTATION_SPEC.md#4-terminal-experience-and-ssh)
- [ ] **QA-03**: Run memory/CPU profiles, fuzz tests, soak testing, dependency review and complete real-world acceptance checklist. [Spec §10](docs/IMPLEMENTATION_SPEC.md#10-testing-quality-gates-and-acceptance)
- [ ] **REL-02**: Ship changelog, signed/checksummed reproducible binaries, installation instructions, security statement and v1.0 gate. [Spec §11](docs/IMPLEMENTATION_SPEC.md#11-builds-packaging-and-release)

## Current sprint — 2026-10-30

1. ✅ Publish this plan and the [implementation specification](docs/IMPLEMENTATION_SPEC.md).
2. ✅ Add a real offline `--doctor` path; test success, missing session, insecure files and corrupted history.
3. ✅ Run CI on the change; correct issues before marking it green.
4. ⏭️ Next: A01–A04 real account + SSH test; then work on automatic reconnect and contact names.

## Definition of done and blockers

**Done:** Source merged, relevant Go tests passing, CI green, docs updated, and acceptance case successfully run when the task involves the actual WhatsApp account or SSH rendering. **Blocked (external):** A01–A14 need a consenting test WhatsApp account, an interactive SSH terminal, and a linked device. No amount of unit or GitHub Actions success can substitute for those tests.

**Scope controls:** Do not claim that a cached archive can recover messages never observed, bypass view-once/ephemeral restrictions, or guarantee zero data loss. This is an unofficial third-party WhatsApp client; follow the platform's terms.