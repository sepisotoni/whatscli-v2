package transfer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadDoesNotChmodSymlinkTarget(t *testing.T) {
	base := t.TempDir()
	target := filepath.Join(base, "target")
	if err := os.Mkdir(target, 0755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(base, "downloads")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := SaveDownload(link, "bad.txt", func(*os.File) error { return nil }); err == nil {
		t.Fatal("symlink download directory unexpectedly accepted")
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0755 {
		t.Fatalf("symlink check changed unrelated directory permissions to %04o", info.Mode().Perm())
	}
}

func TestUploadRejectsOversizeWithoutReadingIntoMemory(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "large.bin")
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Truncate(path, MaxUpload+1); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReadUpload(root, "large.bin"); err == nil {
		t.Fatal("oversized file was accepted")
	}
}
