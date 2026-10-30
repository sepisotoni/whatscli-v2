# WhatsCLI: live SSH + WhatsApp acceptance checklist

**State:** Manual acceptance **not yet run**. Go CI and automated regressions are separate from a real-account end-to-end test. Use an authorized test WhatsApp account and a private SSH server.

## Prepare

```bash
git clone https://github.com/sepisotoni/WhatsCLI.git
cd WhatsCLI
go mod tidy
CGO_ENABLED=0 go build -trimpath -o bin/whatshell ./cmd/whatshell
mkdir -p ~/.local/share/whatshell-go/uploads
chmod 700 ~/.local/share/whatshell-go/uploads
```

On another machine, start with:

```bash
ssh -tt user@server 'cd ~/WhatsCLI && ./bin/whatshell'
```

Record your terminal emulator, `TERM`, dimensions, OS, and the commit SHA. **Do not record or share** actual QR codes, chats, tokens, personal phone numbers, or `session.sqlite3`.

## Acceptance cases

| ID | Action | Expected result | Tested? |
|:--|:--|:--|:--|
| A01 | Start with no session; scan the terminal QR in WhatsApp → Linked devices | Device links; TUI shows connected | Not run |
| A02 | Quit, reopen, and reconnect over SSH | Previous session reused without scanning again | Not run |
| A03 | `/chat` a consenting test contact; send a text | Sent once; message archived locally | Not run |
| A04 | Receive two texts, disconnect network, reopen TUI | Cached texts readable offline, in order | Not run |
| A05 | Edit a test message on a linked device; check `/versions ID` | Prior cached body retained when event arrives | Not run |
| A06 | Revoke a received message after it was cached | Tombstone flagged; cached body retained if supported | Not run |
| A07 | Receive read/delivery receipts | Read state monotonically increases (never regresses) | Not run |
| A08 | SFTP upload a small JPEG and `/attach photo.jpg` | Image arrives in the target chat exactly once | Not run |
| A09 | Receive a document and `/download MESSAGE_ID` | Private file saved; not auto-executed | Not run |
| A10 | Create a single-choice `/poll Q \| A \| B` | Poll appears with both options | Not run |
| A11 | `/backup`; verify copied SQLite archive with `PRAGMA integrity_check` | Archive is consistent; device credentials excluded | Not run |
| A12 | Resize SSH window to 100×32, 75×24 and 40×20; toggle sidebar | No panic; readable navigation and input | Not run |
| A13 | Click chat, scroll, Tab through panes; type `/d` then Tab | No unexpected focus loss; `/download` completion | Not run |
| A14 | Logout, verify WhatsApp Linked devices, restart | Credentials removed, chat archive preserved | Not run |

## Automated terminal preview (not live acceptance)

```bash
go run ./cmd/whatshell --demo
```

The demo provides fictional conversations and no WhatsApp network connection. CI runs a Linux PTY smoke test against the compiled binary, exercising startup, 100×32 → 75×24 → 40×20 resize and clean exit, while confirming no real user session directory was created. **This does not complete A01–A14 or verify real SSH transport.**

## Automated checks

```bash
go test -race -count=1 ./...
go vet ./...
bash scripts/check.sh
CGO_ENABLED=0 go build ./cmd/whatshell
```

See the [CI workflow](../.github/workflows/ci.yml) for staticcheck, gosec, govulncheck, and golangci-lint. Race-enabled tests check concurrent Go access but **do not** prove zero data loss or flawless WhatsApp protocol compatibility.

## Report a bug responsibly

When an acceptance case fails, report: test ID, expected/actual behavior, commit SHA, Go version, OS, terminal and `TERM`, and sanitized error output. Never attach authentication keys, QR screenshots, unredacted message history, or confidential attachments. For security concerns, follow [SECURITY.md](../SECURITY.md).
