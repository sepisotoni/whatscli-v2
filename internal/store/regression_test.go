package store

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestSnapshotKeepsRevisionsReceiptsAndMedia(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	const chat, id = "archive@s.whatsapp.net", "persisted"
	at := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	if err := db.Insert(ctx, Message{Chat: chat, ID: id, Body: "before", Timestamp: at}); err != nil {
		t.Fatal(err)
	}
	if err := db.Edit(ctx, chat, id, "after", at.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := db.MarkDeleted(ctx, chat, id, at.Add(2*time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := db.Receipt(ctx, chat, id, 2); err != nil {
		t.Fatal(err)
	}
	if err := db.SaveMedia(ctx, chat, id, "image.png", []byte{3, 4, 5}); err != nil {
		t.Fatal(err)
	}
	path, err := db.Backup(ctx, filepath.Join(t.TempDir(), "snapshots"))
	if err != nil {
		t.Fatal(err)
	}
	restored, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = restored.Close() }()
	messages, err := restored.Messages(ctx, chat, 20)
	if err != nil || len(messages) != 1 {
		t.Fatalf("restore messages: %+v %v", messages, err)
	}
	m := messages[0]
	if m.Body != "after" || !m.Deleted || !m.Edited || m.Receipt != 2 {
		t.Fatalf("metadata lost in backup: %+v", m)
	}
	versions, err := restored.Versions(ctx, chat, id)
	if err != nil || len(versions) != 1 || versions[0].Body != "before" {
		t.Fatalf("revision history lost in backup: %+v %v", versions, err)
	}
	name, payload, err := restored.Media(ctx, chat, id)
	if err != nil || name != "image.png" || len(payload) != 3 {
		t.Fatalf("media pointer lost in backup: %q %v %v", name, payload, err)
	}
}

func TestConcurrentArchiveWritesRetainAllMessages(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	const workers, perWorker = 8, 25
	const chat = "concurrency@s.whatsapp.net"
	var wg sync.WaitGroup
	errors := make(chan error, workers)
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				id := fmt.Sprintf("%02d-%02d", worker, i)
				if err := db.Insert(ctx, Message{Chat: chat, ID: id, Body: id, Timestamp: time.Unix(int64(worker*perWorker+i+1), 0)}); err != nil {
					errors <- err
					return
				}
			}
		}(worker)
	}
	wg.Wait()
	close(errors)
	for err := range errors {
		t.Fatal(err)
	}
	got, err := db.Messages(ctx, chat, workers*perWorker)
	if err != nil || len(got) != workers*perWorker {
		t.Fatalf("lost concurrent messages: %d, %v", len(got), err)
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].Timestamp.After(got[i].Timestamp) {
			t.Fatalf("history not sorted at index %d", i)
		}
	}
}