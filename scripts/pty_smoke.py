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
            # TCell opens /dev/tty. Make the PTY a controlling terminal
            # after setsid() rather than merely redirecting standard streams.
            def take_controlling_tty():
                os.setsid()
                fcntl.ioctl(slave, termios.TIOCSCTTY, 0)

            proc = subprocess.Popen([binary, "--demo"], stdin=slave,
                                    stdout=slave, stderr=slave,
                                    preexec_fn=take_controlling_tty, env=env)
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
                raise RuntimeError(f"terminal UI header missing (exit={proc.poll()}, output={bytes(output[-600:])!r})")
            def pump_until_exit(seconds):
                deadline = time.monotonic() + seconds
                while time.monotonic() < deadline:
                    if proc.poll() is not None:
                        return proc.returncode
                    ready, _, _ = select.select([master], [], [], 0.05)
                    if ready:
                        try:
                            os.read(master, 65536)
                        except OSError:
                            if proc.poll() is not None:
                                return proc.returncode
                return None

            for cols, rows in [(75, 24), (40, 20), (100, 32)]:
                fcntl.ioctl(master, termios.TIOCSWINSZ, struct.pack("HHHH", rows, cols, 0, 0))
                os.killpg(proc.pid, signal.SIGWINCH)
                # Drain redraw traffic so the app cannot block on its own output.
                ready, _, _ = select.select([master], [], [], 0.12)
                if ready:
                    try:
                        os.read(master, 65536)
                    except OSError:
                        pass
            os.write(master, b"/quit\r")
            code = pump_until_exit(5)
            if code is None:
                os.write(master, b"\x03")
                code = pump_until_exit(3)
            if code is None:
                raise RuntimeError("demo did not quit after /quit and Ctrl+C")
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
