#!/usr/bin/env python3
"""Verify first-run welcome -> login navigation in a disposable TTY.

This intentionally never obtains a WhatsApp QR or accesses user credentials.
"""
import fcntl
import os
import pty
import select
import signal
import struct
import subprocess
import sys
import tempfile
import termios
import time


def main(binary: str) -> None:
    with tempfile.TemporaryDirectory(prefix="whatscli-onboarding-") as tmp:
        protected = os.path.join(tmp, "real-session-must-not-exist")
        master, slave = pty.openpty()
        proc = None
        try:
            fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack("HHHH", 44, 118, 0, 0))

            def session():
                os.setsid()
                fcntl.ioctl(slave, termios.TIOCSCTTY, 0)

            proc = subprocess.Popen(
                [binary, "--demo-onboarding"], stdin=slave, stdout=slave, stderr=slave,
                preexec_fn=session,
                env={**os.environ, "TERM": "xterm-256color", "WHATSHELL_HOME": protected},
            )
            os.close(slave)
            slave = -1

            def until(needle: bytes, timeout: float = 6.0) -> None:
                output = bytearray()
                deadline = time.monotonic() + timeout
                while time.monotonic() < deadline:
                    ready, _, _ = select.select([master], [], [], 0.05)
                    if ready:
                        try:
                            output.extend(os.read(master, 65536))
                        except OSError:
                            break
                        if needle in output:
                            return
                    if proc.poll() is not None:
                        break
                raise RuntimeError(f"missing UI text {needle!r} (exit={proc.poll()}, output={bytes(output[-800:])!r})")

            until(b"Get started")  # Styled heading uses terminal colour escapes.
            os.write(master, b"\r")  # Get started -> disclosure, not login.
            until(b"Independent and unofficial")
            # Trying Continue without acknowledgement must not open QR login.
            os.write(master, b"\t\r")
            until(b"Please read and tick")
            os.write(master, b" ")  # checkbox (focus returned by blocked Continue)
            os.write(master, b"\t\r")  # Agree & continue
            until(b"Log in to WhatsApp")
            # Change size while the login screen is visible; neither a QR token
            # nor a WhatsApp session should ever be generated in preview mode.
            fcntl.ioctl(master, termios.TIOCSWINSZ, struct.pack("HHHH", 28, 80, 0, 0))
            os.killpg(proc.pid, signal.SIGWINCH)
            deadline = time.monotonic() + 0.5
            while time.monotonic() < deadline:
                ready, _, _ = select.select([master], [], [], 0.05)
                if ready:
                    try:
                        os.read(master, 65536)
                    except OSError:
                        break
            os.write(master, b"\x03")  # Ctrl+C safely exits from onboarding
            deadline = time.monotonic() + 5
            while time.monotonic() < deadline and proc.poll() is None:
                ready, _, _ = select.select([master], [], [], 0.05)
                if ready:
                    try:
                        os.read(master, 65536)
                    except OSError:
                        break
            if proc.poll() is None:
                raise RuntimeError("Ctrl+C did not exit onboarding")
            if proc.returncode != 0:
                raise RuntimeError(f"onboarding exited {proc.returncode}")
            if os.path.exists(protected):
                raise RuntimeError("onboarding demo touched real WhatsApp session workspace")
            print("Onboarding smoke passed: welcome, click/Enter login, resize, safe exit")
        finally:
            if proc is not None and proc.poll() is None:
                os.killpg(proc.pid, signal.SIGKILL)
                proc.wait(timeout=5)
            os.close(master)
            if slave != -1:
                os.close(slave)


if __name__ == "__main__":
    if len(sys.argv) != 2:
        sys.exit("Usage: onboarding_pty_smoke.py /path/to/whatshell")
    main(os.path.abspath(sys.argv[1]))
