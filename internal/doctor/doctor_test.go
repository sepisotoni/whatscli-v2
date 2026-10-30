package doctor

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/sepisotoni/whatshell-go/internal/config"
	"github.com/sepisotoni/whatshell-go/internal/store"
)

func setup(t *testing.T) config.Paths {
	t.Helper()
	t.Setenv("WHATSHELL_HOME", filepath.Join(t.TempDir(), "private"))
	p, err := config.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestEmptyDirectoryDoesNotCreateDatabases(t *testing.T) {
	p := setup(t)
	var output bytes.Buffer
	if err := Run(context.Background(), p, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "WARN") || !strings.Contains(output.String(), "NOT tested") {
		t.Fatalf("missing diagnostic context: %q", output.String())
	}
	for _, path := range []string{p.History, p.Session} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("doctor unexpectedly created %s: %v", path, err)
		}
	}
}

func TestHealthyAndCorruptArchive(t *testing.T) {
	p := setup(t)
	db, err := store.Open(context.Background(), p.History)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := Run(context.Background(), p, &bytes.Buffer{}); err != nil {
		t.Fatalf("healthy database rejected: %v", err)
	}
	if err := os.WriteFile(p.History, []byte("not a SQLite database"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Run(context.Background(), p, &bytes.Buffer{}); err == nil {
		t.Fatal("corrupted history was incorrectly accepted")
	}
}

func TestInsecureSessionPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission bits not applicable")
	}
	p := setup(t)
	if err := os.WriteFile(p.Session, []byte("test fixture only"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Run(context.Background(), p, &bytes.Buffer{}); err == nil {
		t.Fatal("world-readable session was accepted")
	}
}

func TestSymlinkUploadRootRejected(t *testing.T) {
	p := setup(t)
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.Mkdir(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(p.Uploads); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, p.Uploads); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := Run(context.Background(), p, &bytes.Buffer{}); err == nil {
		t.Fatal("symlinked upload directory was accepted")
	}
}
