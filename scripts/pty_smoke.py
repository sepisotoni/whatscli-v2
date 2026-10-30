#!/usr/bin/env python3
"""Local-fake-data TTY smoke test; never connects to WhatsApp."""
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
    with tempfile.TemporaryDirectory(prefix="whatscli-pty-smoke-") as tmp:
        # A trap path for a real session: demo mode must never touch it.
        real_home = os.path.join(tmp, "should-not-be-created")
        env = {**os.environ, "TERM": "xterm-256color", "WHATSHELL_HOME": real_home}
        master, slave = pty.openpty()
        proc = None
        try:
            fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack("HHHH", 32, 100, 0, 0))
            proc = subprocess.Popen([binary, "--demo"], stdin=slave,
                                    stdout=slave, stderr=slave,
                                    start_new_session=True, env=env)
            os.close(slave)
            slave = -1
            output = bytearray()
            deadline = time.monotonic() + 8
            while time.monotonic() < deadline and proc.poll() is None:
                ready, _, _ = select.select([master], [], [], 0.2)
                if ready:
                    try:
                        output.extend(os.read(master, 65536))
                    except OSError:
                        break
                if b"WHATSHELL" in output or b"Whatshell" in output:
                    break
            else:
                raise RuntimeError("demo TUI did not draw before timeout")
            if b"WHATSHELL" not in output and b"Whatshell" not in output:
                raise RuntimeError("terminal output missing WhatsCLI UI header")
            for cols, rows in [(75, 24), (40, 20), (100, 32)]:
                fcntl.ioctl(master, termios.TIOCSWINSZ, struct.pack("HHHH", rows, cols, 0, 0))
                os.killpg(proc.pid, signal.SIGWINCH)
                time.sleep(0.12)
            os.write(master, b"/quit\r")
            try:
                code = proc.wait(timeout=5)
            except subprocess.TimeoutExpired:
                os.write(master, b"\x03")
                code = proc.wait(timeout=3)
            if code != 0:
                raise RuntimeError(f"demo exited with status {code}")
            if os.path.exists(real_home):
                raise RuntimeError("demo mode modified the live-session workspace")
            print("TTY demo smoke test passed: terminal draw, resize, quit, isolated data")
        finally:
            if proc is not None and proc.poll() is None:
                os.killpg(proc.pid, signal.SIGKILL)
                proc.wait(timeout=5)
            os.close(master)
            if slave != -1:
                os.close(slave)


if __name__ == "__main__":
    if len(sys.argv) != 2:
        sys.exit("Usage: pty_smoke.py /path/to/whatshell")
    main(os.path.abspath(sys.argv[1]))
