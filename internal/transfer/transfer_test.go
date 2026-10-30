package transfer

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUploadRejectsEscapesAndSymlinks(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "uploads")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "ok.txt"), []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(base, "secret")
	if err := os.WriteFile(outside, []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "bad")); err != nil {
		t.Skip(err)
	}
	for _, name := range []string{"../secret", "bad", outside} {
		if _, _, err := ReadUpload(root, name); err == nil {
			t.Fatalf("escape accepted: %q", name)
		}
	}
	data, name, err := ReadUpload(root, "ok.txt")
	if err != nil || name != "ok.txt" || string(data) != "hello" {
		t.Fatalf("upload: %s %q %v", data, name, err)
	}
}
func TestDownloadUniqueAndAtomic(t *testing.T) {
	root := filepath.Join(t.TempDir(), "downloads")
	for i := 0; i < 2; i++ {
		p, err := SaveDownload(root, "../../evil\\oops.txt", func(w *os.File) error { _, e := w.WriteString("file"); return e })
		if err != nil || !strings.HasPrefix(p, root+string(filepath.Separator)) {
			t.Fatalf("invalid target: %s %v", p, err)
		}
	}
	_, err := SaveDownload(root, "failure.txt", func(w *os.File) error { _, _ = w.Write([]byte("part")); return errors.New("network failure") })
	if err == nil {
		t.Fatal("expected failure")
	}
	if _, err := os.Stat(filepath.Join(root, "failure.txt")); !os.IsNotExist(err) {
		t.Fatal("partial file not removed")
	}
}
