package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePrivateDirectory(t *testing.T) {
	base := filepath.Join(t.TempDir(), "data")
	t.Setenv("WHATSHELL_HOME", base)
	paths, err := Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if paths.History != filepath.Join(base, "history.sqlite3") {
		t.Fatalf("unexpected path: %s", paths.History)
	}
	if err := ProtectFile(paths.History); err != nil {
		t.Fatal(err)
	}
	stat, err := os.Stat(paths.History)
	if err != nil {
		t.Fatal(err)
	}
	if got := stat.Mode().Perm(); got != 0o600 {
		t.Fatalf("expected private file, got %04o", got)
	}
}

func TestProtectFileRejectsSymlink(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	if err := os.WriteFile(target, []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "session.sqlite3")
	if err := os.Symlink(target, link); err != nil {
		t.Skip(err)
	}
	if err := ProtectFile(link); err == nil {
		t.Fatal("symlink accepted as database path")
	}
}
