package main

import (
	"context"
	"github.com/sepisotoni/whatshell-go/internal/store"
	"path/filepath"
	"testing"
)

func TestSeedDemoIsFictionalAndIdempotent(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "demo.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if err := seedDemo(ctx, db); err != nil {
		t.Fatal(err)
	}
	if err := seedDemo(ctx, db); err != nil {
		t.Fatal(err)
	}
	chats, err := db.Chats(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(chats) != 3 {
		t.Fatalf("expected 3 fictional chats, got %d", len(chats))
	}
	for _, c := range chats {
		if c.Alias == "" {
			t.Fatalf("missing demo alias for %s", c.JID)
		}
		messages, err := db.Messages(ctx, c.JID, 100)
		if err != nil {
			t.Fatal(err)
		}
		if len(messages) != 3 {
			t.Fatalf("expected deduped 3 messages, got %d", len(messages))
		}
	}
	prev, err := db.Versions(ctx, "sample-1@demo.invalid", "demo-0-0")
	if err != nil || len(prev) != 1 {
		t.Fatalf("expected one prior edit: %d %v", len(prev), err)
	}
	last, err := db.ReceiptLevel(ctx, "sample-3@demo.invalid", "demo-2-1")
	if err != nil || last != 2 {
		t.Fatalf("missing read receipt: %d %v", last, err)
	}
}
