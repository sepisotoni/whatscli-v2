# Architecture and engineering decisions

## Runtime topology

```text
  OpenSSH session (PTY) / local terminal
                 |
          tview + tcell
         (single event loop)
         /             \
   SQLite archive     WhatsApp event notifications
         ^             |
         |     WhatsMeow Multi-Device client
         |             |
      incoming   network-backed E2EE transport
         |
  event callback: store first, then QueueUpdateDraw
```

Data lifecycle:
1. `config.Resolve` creates a user-private data directory.
2. `store.Open` configures `WAL`, `synchronous=FULL`, foreign keys, and idempotent tables.
3. `wa.New` opens a **separate** WhatsMeow session database with modernc's `sqlite3` compatibility alias.
4. `ui.Run` launches pairing/reconnection asynchronously, without blocking the terminal render loop.
5. The transport callback writes each incoming message/edit/revocation to the app archive **before** issuing a UI event.
6. Sending waits for WhatsApp's server acknowledgment and then stores the outgoing message.
7. `/logout` unlinks if possible, tears down the client, and deletes the local **session** files, leaving history intact.

## SQLite schema

`messages` is uniquely keyed by `(chat_jid, message_id)`; repeated message delivery is idempotent. `deleted` is a tombstone, not a hard delete. `edited` indicates a known revision. Each observed edit appends the previous body to `message_versions` transactionally. Edit events arriving before originals create an initial placeholder; an original that was not observed cannot be fabricated. The archive is designed to be auditable, not to defeat ephemeral/privacy behavior.

`message_receipts` stores monotonic delivered/read levels independently of message arrival to tolerate out-of-order events. `message_media` contains protobuf metadata for deferred encrypted media download; it is not a plaintext media cache. `/backup` uses `VACUUM INTO` on a pre-created `0600` file for a consistent point-in-time history-only snapshot.

**Next migrations:** a schema-version table, contact metadata, poll options/votes, pagination tokens, attachment hashes, FTS5 full-text indexing, and a durable outgoing queue. Backfills must not silently overwrite locally captured edits and tombstones.

## SSH compatibility

`tcell` handles terminal dimension changes and mouse tracking via terminfo/ANSI and a PTY. Do **not** manually install `SIGWINCH` handlers that compete with the terminal library. Mouse support depends on the terminal emulator and PTY; keyboard fallback is mandatory. Respect `NO_COLOR`, truecolor detection, smaller widths, SSH disconnect/reconnect and `tmux` nesting. The layout collapses the sidebar below 60 columns; `Ctrl+L` temporarily shows a compact sidebar, while `/chats` and `/chat` remain available. Modal widths are clamped to terminal dimensions. QR pairing warns when the terminal is too small.

## Security and privacy boundaries

- Do not log QR payloads, auth tokens, cryptographic key bytes, complete messages or personal IDs.
- Do not store WhatsApp credentials in Git, CI logs or diagnostics. Keep config/data within `0700` and SQLite files `0600`.
- Do not use raw shell commands for media preview. Treat remote filenames as untrusted and require canonical path validation and size limits.
- Protect session files at rest using OS disk encryption (SQLite by itself is plaintext).
- Don't expose the UI as a public unauthenticated SSH service. Use individual OS accounts or Wish with SSH public-key verification and session isolation if adding an embedded server later.
- WhatsMeow is **not** an official Meta Cloud API. A linked-device protocol can change or be restricted without notice; calls and some other desktop functions are unsupported by the backend.
- SQLite durability is not the same as backup. Provide snapshot and restore verification before claiming enterprise resilience.

## Module boundaries and future services

`internal/wa/media.go` implements the first media upload/download path and single-choice poll creation, while `internal/transfer` enforces the private SFTP-root boundaries. A future `VideoRenderer` should expose terminal capability negotiation and safe conversion, and a future `PollService` should add vote decryption and persistence without blocking event processing.

A Wish-based SSH server could eventually manage terminal sessions as Go processes, but only after per-user authentication, authorization, database isolation, rate-limits and PTY threat modeling are implemented. For the current baseline, normal OpenSSH is simpler and more secure.
