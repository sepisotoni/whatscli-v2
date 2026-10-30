# Security policy (development)

Whatshell is an **unofficial** WhatsApp Multi-Device client and stores highly
sensitive session keys and conversations. Do not deploy it as a shared-user
multi-tenant service in its current development state.

- Run under a dedicated, non-root Unix login. Use OpenSSH public-key authentication.
- Keep `WHATSHELL_HOME` private to that login. Session, media metadata and history
  are NOT encrypted at rest. Use disk encryption and encrypted offline backups.
- Incoming files are **never executed automatically**. Inspect documents before opening.
- `/attach` is restricted to the private `uploads/` directory and has a 16 MiB
  process-memory upload cap; SSH drag-and-drop is not supported.
- `/download` uses private, exclusive filenames and leaves downloads in `downloads/`.
  The current WhatsMeow streaming implementation is bounded by network/server behavior,
  not a strict file-size quota. Keep sufficient storage available.
- SQLite `WAL` and `FULL` synchronous durability improve reliability, not a
  proof against arbitrary power loss or malicious filesystem corruption.
- `/logout` attempts remote unlink and deletes local credentials. SSD forensic
  erasure cannot be guaranteed, and deleting session keys does not erase the chat archive.
- Edited and deleted content is intentionally preserved after reception; do not
  enable this archive on an account or machine that other people can access.
- WhatsMeow is unofficial. Protocol breakage or account restrictions are possible.

For vulnerabilities, use private GitHub vulnerability reporting when enabled.
Avoid posting message contents, QR codes, session databases, phone numbers or logs
containing secrets in public issues.
