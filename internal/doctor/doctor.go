// Package doctor performs offline health checks without connecting to WhatsApp.
package doctor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/sepisotoni/whatshell-go/internal/config"
	"github.com/sepisotoni/whatshell-go/internal/store"
)

// Check is one privacy-safe local diagnostic finding.
type Check struct {
	Name   string
	Status string // PASS, WARN or FAIL
	Detail string
}

// Inspect never connects to WhatsApp and does not create database files.
// The calling executable may create private app folders via config.Resolve.
func Inspect(ctx context.Context, paths config.Paths) ([]Check, error) {
	var checks []Check
	var failures []error
	fail := func(name, detail string, err error) {
		checks = append(checks, Check{name, "FAIL", detail})
		failures = append(failures, fmt.Errorf("%s: %w", name, err))
	}
	for _, entry := range []struct{ name, path string }{
		{"application directory", paths.Dir},
		{"upload directory", paths.Uploads},
		{"download directory", paths.Downloads},
	} {
		info, err := os.Lstat(entry.path)
		if err != nil {
			fail(entry.name, "unavailable", err)
			continue
		}
		if !info.IsDir() {
			fail(entry.name, "not a real directory (symlinks rejected)", errors.New("invalid directory"))
			continue
		}
		if runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
			fail(entry.name, "readable or writable by other users", errors.New("insecure permissions"))
			continue
		}
		checks = append(checks, Check{entry.name, "PASS", "private directory"})
	}
	for _, entry := range []struct {
		name, path string
		archive    bool
	}{
		{"history database", paths.History, true},
		{"device session", paths.Session, false},
	} {
		info, err := os.Lstat(entry.path)
		if os.IsNotExist(err) {
			checks = append(checks, Check{entry.name, "WARN", "not created yet"})
			continue
		}
		if err != nil {
			fail(entry.name, "cannot inspect file", err)
			continue
		}
		if !info.Mode().IsRegular() {
			fail(entry.name, "not a regular file (symlinks rejected)", errors.New("invalid database file"))
			continue
		}
		if runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
			fail(entry.name, "file is accessible to other users", errors.New("insecure permissions"))
			continue
		}
		if !entry.archive {
			checks = append(checks, Check{entry.name, "PASS", "private session file; credentials not inspected"})
			continue
		}
		if err := checkHistory(ctx, entry.path); err != nil {
			fail(entry.name, "SQLite quick_check failed", err)
			continue
		}
		checks = append(checks, Check{entry.name, "PASS", "read-only SQLite integrity check passed"})
	}
	return checks, errors.Join(failures...)
}

func checkHistory(ctx context.Context, path string) error {
	// SQLite URI mode=ro prevents creating a new archive during diagnostics.
	conn, err := sql.Open("sqlite", store.DSN(path)+"&mode=ro")
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	var result string
	if err := conn.QueryRowContext(ctx, "PRAGMA quick_check").Scan(&result); err != nil {
		return err
	}
	if !strings.EqualFold(result, "ok") {
		return fmt.Errorf("unexpected SQLite integrity result: %s", result)
	}
	return nil
}

// Run prints a concise report with no message contents, phone numbers or keys.
// It returns an error if any local safety or integrity check fails.
func Run(ctx context.Context, paths config.Paths, out io.Writer) error {
	checks, failures := Inspect(ctx, paths)
	if _, err := fmt.Fprintln(out, "WhatsCLI offline diagnostics (no WhatsApp connection)"); err != nil {
		return err
	}
	for _, c := range checks {
		if _, err := fmt.Fprintf(out, "[%-4s] %s: %s\n", c.Status, c.Name, c.Detail); err != nil {
			return err
		}
	}
	if failures != nil {
		return fmt.Errorf("local diagnostics failed: %w", failures)
	}
	_, err := fmt.Fprintln(out, "Local checks completed. Real QR pairing and messaging are NOT tested.")
	return err
}
