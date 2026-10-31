# BUGS.md — WhatsCLI v2 audit backlog

> Targeted static inspection, 2026-10-31. **Risks are not necessarily reproduced bugs.** No fresh integration/race tests were run and this is not an exhaustive audit.

| ID | Priority | Confidence | File(s) | Finding / next action |
|---|---|---|---|---|
| WCV2-001 | High | Concurrency risk | `internal/wa/client.go` (`Start`, `notify`) | The event callback is assigned while background event/connection workers can read it, and multiple `Start` calls could launch overlapping workers. Enforce single-start semantics, synchronize callback lifecycle, track shutdown. Verify with `go test -race ./...` and targeted tests. No race reproduced yet. |
| WCV2-002 | Medium | Filesystem race risk | `internal/config/config.go` (`ProtectFile`) | `Lstat` followed by `OpenFile` and path-based `Chmod` creates a check/use gap; symlink/path replacement may affect the intended DB. Open with no-follow, verify via descriptor, secure SQLite sidecars. No exploitation demonstrated. |
| WCV2-003 | Medium | Shutdown/lifecycle risk | `internal/ui/ui.go` (send, logout, async actions) | Detached goroutines use `context.Background()` plus timeouts; UI shutdown may not cancel/wait for them before deferred client/DB closures. Use shared lifecycle context, cancel and wait on UI exit, and test concurrent quit/send/logout. |
| WCV2-004 | Medium | Behavioral risk / expected by design | `internal/wa/client.go` (`Logout`) | Local session files are removed even when remote WhatsApp unlink fails; user may remain linked remotely. Clearly distinguish remote unlink from local wipe and give manual unlink guidance. Document intended behavior. |
| WCV2-005 | Medium | Delivery consistency | `internal/wa/client.go` (`Send`), `internal/wa/media.go` (`SendFile`, `SendPoll`) | Send succeeds before history insertion. If local storage fails, retries could duplicate messages. Consider durable pending state, stable IDs, and explicit delivery-unknown status. Test DB failure after network success. |
| WCV2-006 | Low | Maintainability | `internal/ui/ui.go`, `internal/ui/auth.go` | Large UI modules; consider splitting into tested components after correctness and lifecycle work. Not a confirmed bug. |

## Existing positive controls
- SQLite WAL and full synchronization settings; archive and schema migration transactions.
- SFTP upload/download boundaries and MIME detection.
- Diagnostic `--doctor` mode.
- Welcome animation has context-based cancellation.

## Recommended fix order
1. Lifecycle and synchronization: WCV2-001, WCV2-003.
2. File/session safety: WCV2-002.
3. Sending and unlinking semantics: WCV2-005, WCV2-004.
4. Refactoring: WCV2-006.

No application code was modified by this audit backlog.
