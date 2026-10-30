package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBackupUsesPrivateFileAndPreservesData(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	if err := db.Insert(ctx, Message{Chat: "friend@s.whatsapp.net", ID: "1", Body: "kept", Timestamp: time.Now()}); err != nil {
		t.Fatal(err)
	}
	path, err := db.Backup(ctx, filepath.Join(t.TempDir(), "backup"))
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("unencrypted snapshot is not private: %v", info.Mode().Perm())
	}
	restored, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	rows, err := restored.Messages(ctx, "friend@s.whatsapp.net", 5)
	if err != nil || len(rows) != 1 || rows[0].Body != "kept" {
		t.Fatalf("snapshot missing message: %+v %v", rows, err)
	}
}
