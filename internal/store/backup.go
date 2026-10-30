package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Backup makes a consistent, private, hot SQLite snapshot using VACUUM INTO.
// It archives message history ONLY, never WhatsMeow's authentication keys.
// The caller must safeguard this plaintext backup with encrypted storage.
func (d *DB) Backup(ctx context.Context, dir string) (string, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("create backup folder: %w", err)
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("backup root is not a real directory: %s", dir)
	}
	if err := os.Chmod(dir, 0700); err != nil {
		return "", err
	}
	// VACUUM INTO accepts a pre-existing zero-byte output file. Creating it
	// ourselves with mode 0600 avoids a world-readable intermediate backup.
	name := "whatshell-" + time.Now().UTC().Format("20060102T150405.000000000Z") + ".sqlite3"
	path := filepath.Join(dir, name)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		return "", err
	}
	if err = file.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	if _, err = d.sql.ExecContext(ctx, `VACUUM INTO ?`, path); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("create SQLite snapshot: %w", err)
	}
	synced, err := os.Open(path)
	if err != nil {
		return "", err
	}
	syncErr := synced.Sync()
	closeErr := synced.Close()
	if syncErr != nil || closeErr != nil {
		return "", fmt.Errorf("sync backup: %w", errors.Join(syncErr, closeErr))
	}
	return path, nil
}
