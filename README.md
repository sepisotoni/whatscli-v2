# Whatshell Go — SSH-first WhatsApp TUI

A **Go-only**, self-hosted, terminal-native WhatsApp **starter client**, with an emerald-and-slate interface, SQLite offline archive, and WhatsApp Multi-Device transport via [WhatsMeow](https://github.com/tulir/whatsmeow).

> **Project state: v0.2 development (not desktop parity).** The code implements the first functional vertical slice, but the full dependency graph has **not been compiled or tested against a real WhatsApp account** in the authoring environment. This is an **unofficial** third-party client; use at your own discretion and respect WhatsApp's terms. It is not affiliated with Meta or WhatsApp.

## Status at a glance

| Feature | Status |
|---|---|
| SSH terminal drawing, colors, keyboard, mouse, resizing | Implemented using tview/tcell; still needs on-device testing |
| QR pairing, persistent Multi-Device session | Transport integrated; needs account-level verification |
| Send/receive text | Transport integrated; needs account-level verification |
| SQLite chat archive and offline browsing | Implemented; unit tests included |
| Store previously cached bodies after delete events | Implemented for protocol revocations received while linked |
| Preserve edits as message versions | Implemented for protocol edits received while linked |
| Live command suggestions and `/chat`, `/versions`, `/logout` | Implemented |
| Initial history sync import / contact names | Source implemented / contact names planned |
| Read/delivered receipts / offline text search | Source implemented; account testing pending |
| Send JPEG/PNG images and documents via SFTP `/attach` | Source implemented; WhatsApp testing pending |
| On-demand incoming image/video/audio/document `/download` | Source implemented; WhatsApp testing pending |
| Voice note sending / safe built-in playback | Planned |
| Poll creation / voting / live results | Creation implemented in source; voting/results planned |
| Status stories/replies/posts | Planned |
| Terminal video/Sixel/Kitty rendering | Planned and terminal-dependent |
| Local notifications / desktop notifications | In-app notification footer; native hooks planned |
| Plaintext history `/backup` / encryption-at-rest / multiple accounts | Snapshot implemented / encryption and multi-account planned |

Local archiving cannot recover deleted content that **was never received**. It also does not bypass disappearing messages, view-once protections, or WhatsApp retention limits. Full desktop feature parity and a guarantee of zero data loss are not realistic claims for this initial release.

## Requirements

- Go **1.27.3** (CI toolchain target), with internet access to fetch modules on first build.
- Linux, macOS, or Windows terminal (SSH sessions are the primary target).
- An existing WhatsApp account for device linking.
- A real UTF-8 terminal and a working `TERM` value (for example `xterm-256color` or `screen-256color`).
- Terminal size **at least 90×30** preferred for pairing QR code and side-by-side layout.

Everything in the app is Go; SQLite uses `modernc.org/sqlite` (no C compiler for ordinary builds). The code registers `sqlite3` as a driver alias to keep WhatsMeow's SQLite dialect compatible with the pure-Go driver.

## Quick start

```bash
git clone https://github.com/YOUR_USERNAME/whatshell-go.git
cd whatshell-go
go mod tidy     # also creates go.sum; commit it before opening a PR
go run ./cmd/whatshell
```

If you downloaded the starter ZIP, extract it and run the last two commands instead. The repository URL is illustrative until the project has been published.

On first login, scan the on-screen QR using WhatsApp on your phone → **Linked devices** → **Link a device**. The terminal may need to be widened to show the QR cleanly. Re-launching uses the saved device session without scanning again unless the session expired/revoked.

```bash
# Static build (the app itself needs no CGO):
CGO_ENABLED=0 go build -trimpath -o bin/whatshell ./cmd/whatshell

# Use the binary from a normal OpenSSH shell:
ssh -t myuser@myserver '/home/myuser/whatshell-go/bin/whatshell'

# Safer for terminal allocation if needed:
ssh -tt myuser@myserver 'TERM=xterm-256color /home/myuser/whatshell-go/bin/whatshell'
```

**SSH security:** Prefer public-key login, keep the server patched, and **do not** run Whatshell as root or expose its files to other users. Use separate OS accounts for separate users. Preserve `TERM` from the SSH client rather than forcing `xterm-256color` if your client supports another terminfo type.

## Controls

| Shortcut / command | Meaning |
|---|---|
| Mouse click chat | Open chat |
| Mouse wheel, ↑/↓, PgUp/PgDn | Scroll focused pane |
| `Tab` | Cycle visible panes; autocomplete first slash suggestion |
| `/` and type | Real-time filtered command list |
| `Enter` in composer | Send text or run complete command |
| `Ctrl+N` | New chat; enter international phone number with country code |
| `Ctrl+F` | Search locally archived text |
| `Ctrl+L` | Toggle sidebar on narrow SSH windows |
| `/chats` | Pick a cached chat in a modal |
| `/search words` | Case-insensitive literal search across cached messages |
| `/attach filename.jpg` | Send an image/document uploaded by SFTP into `uploads/` (16 MiB cap) |
| `/download MESSAGE_ID` | Download incoming media into `downloads/` (no auto-open) |
| `/poll Question \| Yes \| No` | Create single-choice WhatsApp poll |
| `/backup` | Private SQLite history snapshot (does not include session keys) |
| `/chat 27123456789` | Open a chat by phone number |
| `/chat 123@g.us` | Open a group by its full JID |
| `/versions MESSAGE_ID` | Show previously cached text versions for a specific message |
| `/help` | Show controls |
| `/logout` | Request remote unlink, then delete the **session** DB and its WAL files; keep history |
| `/quit` / `Ctrl+C` | Exit without unlinking |

Tip: message IDs are shown beside media messages to support `/download ID`. Sender and chat IDs are intentionally shown in the initial version; WhatsApp contact names and a contact picker are planned.

## Data & SQLite

Default directory: `~/.local/share/whatshell-go/`, or `$XDG_DATA_HOME/whatshell-go`, with override `$WHATSHELL_HOME`. Directory is `0700` and created database files are `0600`.

- `history.sqlite3`: app-owned `messages`, `message_versions`, `message_receipts`, `message_media` and indexes.
- `session.sqlite3`: WhatsMeow's persistent device/session/cryptographic keys.
- `uploads/`: private SFTP drop-off directory; `downloads/`: private on-demand attachment download directory; `backups/`: plaintext history snapshots created by `/backup`.
- `-wal`/`-shm`: SQLite sidecars that can exist while the app runs.

**Important:** SQLite content is not encrypted at rest. Filesystem permissions are not disk encryption. Use full-disk encryption, protect backups, and do not share session database files. Deleting tokens using `/logout` removes local session files, but **cannot guarantee forensic secure erasure on SSDs**. If remote unlink fails, local data is still wiped, and the remote unlink warning is displayed briefly before exit; check WhatsApp's linked devices list on your phone.

```bash
# Back up a consistent snapshot (requires sqlite3 CLI):
sqlite3 "$HOME/.local/share/whatshell-go/history.sqlite3" \
  ".backup '$HOME/whatshell-history-backup.sqlite3'"

# Inspect tables (do not publish the output):
sqlite3 "$HOME/.local/share/whatshell-go/history.sqlite3" '.tables'
```

Run `/backup` for an atomic SQLite history snapshot in the private `backups/` directory; **encrypt the backup separately before moving it off the machine**. In production, use an encrypted, versioned backup destination, and test recovery periodically. WAL + `synchronous=FULL` improves resilience to crashes but does **not** guarantee zero loss under arbitrary hardware failure or third-party protocol changes.

## Transfer files over SSH

Conventional SSH **does not transport GUI drag-and-drop into a TUI**. Use `sftp` or `scp` to place files in a server-side upload directory. **Image/document sending is implemented in source as `/attach NAME`**, with a 16 MiB file cap. File uploads and downloads still require live account-level testing.

```bash
ssh myuser@myserver 'mkdir -p ~/.local/share/whatshell-go/uploads && chmod 700 ~/.local/share/whatshell-go/uploads'
sftp myuser@myserver
# sftp> put ./photo.jpg .local/share/whatshell-go/uploads/photo.jpg
```

File controls currently validate path roots, reject symlink escapes, enforce a 16 MiB upload limit, and avoid arbitrary command execution. Resumable uploads and preview controls are pending. Video rendering should be an optional adapter only after capability negotiation (plain ANSI fallback; Kitty/Sixel when supported). Never assume `DISPLAY` over raw SSH.

## Tests and code quality

```bash
go mod tidy
go test -race -count=1 ./...
go vet ./...
bash scripts/check.sh
go install honnef.co/go/tools/cmd/staticcheck@v0.8.1
staticcheck ./...
```

GitHub Actions runs **on every push and pull request**: gofmt, custom checks, go vet, race tests, coverage, CGO-free build, golangci-lint v2, staticcheck, gosec and govulncheck. Dependabot checks Go modules and Actions weekly. Race detection requires a supported compiler/CGO at **test time** even though release builds can use `CGO_ENABLED=0`.

Note: linters find likely mistakes; they do **not** autonomously rewrite or optimize all code, prove memory-leak freedom, or guarantee security. Add benchmarks, load tests, profiling and target-specific acceptance tests before production use.

## Publish to a new repository

The ZIP is already a local Git repository structure. After running `go mod tidy`, use GitHub CLI with an authenticated account:

```bash
git init -b main
git add .
git commit -m "feat: initial SSH-first WhatsApp TUI foundation"
gh auth login              # if not already authenticated
gh repo create whatshell-go --private --source=. --remote=origin --push
```

The module path is initially `github.com/sepisotoni/whatshell-go`. If you create the repository in another account or under a new name, replace the path in `go.mod` and all Go import paths before building.

## Design overview

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md), [docs/ROADMAP.md](docs/ROADMAP.md), [docs/CHANGELOG.md](docs/CHANGELOG.md) and [SECURITY.md](SECURITY.md). The implementation separates tview UI, WhatsMeow integration, app-owned SQLite archive, and configuration. Protocol callbacks write to storage first, then update the UI on the tview event loop. Session secrets are kept separate from historical conversations.

## Tagged releases

The `release.yml` workflow builds CGO-free Linux amd64/arm64, macOS amd64/arm64, and Windows amd64 artifacts when a `v*` tag is pushed. It publishes a `SHA256SUMS` manifest. Create tags only after CI and a linked-account smoke test pass.
