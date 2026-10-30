# Installing, upgrading and uninstalling WhatsCLI

> **Experimental prerelease.** Official WhatsApp account/SSH integration tests have not passed yet. Use a test account, and keep QR codes and device tokens private.

The install script fetches prebuilt releases, checks each binary against the accompanying SHA-256 manifest, and installs **without root** into `~/.local/bin/whatscli`. Go is not needed to install a release binary. Ubuntu Linux x86_64/ARM64 and macOS Intel/Apple Silicon are supported by the installer; Windows binary assets are available from the Releases page.

## One-line install (Ubuntu/macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/sepisotoni/WhatsCLI/main/scripts/install.sh | bash
```

For a **specific, reproducible version**, pin the tag:

```bash
curl -fsSL https://raw.githubusercontent.com/sepisotoni/WhatsCLI/main/scripts/install.sh | bash -s -- --version v0.2.1-beta.1
```

To select an installation directory:

```bash
curl -fsSL https://raw.githubusercontent.com/sepisotoni/WhatsCLI/main/scripts/install.sh | bash -s -- --install-dir "$HOME/bin"
```

If `~/.local/bin` is not on `PATH`, add `export PATH="$HOME/.local/bin:$PATH"` to `~/.bashrc` or `~/.zshrc` and reopen the shell (or run `source ~/.bashrc`). **Do not use `sudo`** for the default user install.

To preview the shell installer before executing it, download it:

```bash
curl -fsSLo install-whatscli.sh https://raw.githubusercontent.com/sepisotoni/WhatsCLI/main/scripts/install.sh
less install-whatscli.sh
bash install-whatscli.sh
```

## First run, test, and upgrade

```bash
whatscli --version
whatscli --doctor    # offline permissions + SQLite integrity diagnostics
whatscli --demo      # fake chats in a temporary database (no WhatsApp connection)
whatscli             # starts live linked-device pairing
```

To upgrade or repair the binary, **rerun the same installer**. It atomically replaces the binary after SHA-256 verification, without deleting `~/.local/share/whatshell-go/` or configured `WHATSHELL_HOME` data. Running a different binary version can migrate the archive; make a backup first when upgrading across development versions.

## One-line uninstall

```bash
curl -fsSL https://raw.githubusercontent.com/sepisotoni/WhatsCLI/main/scripts/uninstall.sh | bash
```

If you installed with a custom path, pass the same `--install-dir` option. The uninstaller checks that the executable identifies itself as WhatsCLI before removing it. **It intentionally does not remove chat history, your linked-device session, backups, or media.**

If you wish to unlink from WhatsApp first, run `whatscli` and use `/logout`, then verify on the phone in WhatsApp → Linked devices. Uninstalling alone does not remotely unlink a device.

## Supported binaries

| Platform | Release asset | Installer support |
|---|---|---|
| Linux x86_64 | `whatscli-linux-amd64` | Yes |
| Linux ARM64 | `whatscli-linux-arm64` | Yes |
| macOS Intel | `whatscli-darwin-amd64` | Yes |
| macOS Apple Silicon | `whatscli-darwin-arm64` | Yes |
| Windows x86_64 | `whatscli-windows-amd64.exe` | Manual download |

Every release also contains `SHA256SUMS`. When pinned to a tag, the installer always installs that exact version and checks the bytes. The default channel is the `releases/current.txt` version in this repository and only changes after a release is verified.

## Troubleshooting

- `curl: (22) 404`: release not yet published, incorrect tag, or unsupported target. Check the [releases page](https://github.com/sepisotoni/WhatsCLI/releases).
- `Checksum mismatch`: binary did not match the expected manifest. The installer aborts and leaves the previous executable intact.
- `whatscli: command not found`: add `~/.local/bin` to PATH or run `~/.local/bin/whatscli` directly.
- Permission denied: use your normal login account, do not pipe to `sudo bash`. Check ownership of the selected `--install-dir`.
- QR does not render: start `whatscli --demo` and check your terminal's `TERM` and available width; then test the live client with a consenting test account.

Do not report or upload unredacted QR codes, credentials, or `session.sqlite3` files.
