// Package config resolves user-controlled, private application data paths.
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type Paths struct {
	Dir       string
	History   string
	Session   string
	Uploads   string
	Downloads string
}

func Resolve() (Paths, error) {
	base := os.Getenv("WHATSHELL_HOME")
	if base == "" {
		base = os.Getenv("XDG_DATA_HOME")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return Paths{}, fmt.Errorf("home directory: %w", err)
			}
			base = filepath.Join(home, ".local", "share")
		}
		base = filepath.Join(base, "whatshell-go")
	}
	base, err := filepath.Abs(base)
	if err != nil {
		return Paths{}, fmt.Errorf("resolve data directory: %w", err)
	}
	if err := os.MkdirAll(base, 0o700); err != nil {
		return Paths{}, fmt.Errorf("create data directory: %w", err)
	}
	baseInfo, err := os.Lstat(base)
	if err != nil {
		return Paths{}, err
	}
	if !baseInfo.IsDir() {
		return Paths{}, fmt.Errorf("data root is not a real directory: %s", base)
	}
	if err := os.Chmod(base, 0o700); err != nil {
		return Paths{}, fmt.Errorf("protect data directory: %w", err)
	}
	for _, folder := range []string{"uploads", "downloads"} {
		path := filepath.Join(base, folder)
		if err := os.MkdirAll(path, 0o700); err != nil {
			return Paths{}, fmt.Errorf("create %s: %w", folder, err)
		}
		// Never accept a symlink pretending to be a private folder.
		info, err := os.Lstat(path)
		if err != nil {
			return Paths{}, err
		}
		if !info.IsDir() {
			return Paths{}, fmt.Errorf("%s is not a directory", path)
		}
		if err := os.Chmod(path, 0o700); err != nil {
			return Paths{}, err
		}
	}
	return Paths{Dir: base, History: filepath.Join(base, "history.sqlite3"), Session: filepath.Join(base, "session.sqlite3"), Uploads: filepath.Join(base, "uploads"), Downloads: filepath.Join(base, "downloads")}, nil
}

// ProtectFile ensures a new DB is private before the SQLite driver opens it.
func ProtectFile(path string) error {
	info, err := os.Lstat(path)
	if err == nil && !info.Mode().IsRegular() {
		return fmt.Errorf("database path is not a regular file: %s", path)
	}
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}
