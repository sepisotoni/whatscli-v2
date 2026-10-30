package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func openTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open(context.Background(), filepath.Join(t.TempDir(), "test.sqlite3"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	return db
}

func TestEditHistoryAndTombstone(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	at := time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC)
	original := Message{Chat: "123@s.whatsapp.net", ID: "a1", Body: "first", Timestamp: at}
	if err := db.Insert(ctx, original); err != nil {
		t.Fatal(err)
	}
	if err := db.Edit(ctx, original.Chat, original.ID, "second", at.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := db.Edit(ctx, original.Chat, original.ID, "third", at.Add(2*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := db.MarkDeleted(ctx, original.Chat, original.ID, at.Add(3*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := db.Insert(ctx, original); err != nil {
		t.Fatal(err)
	} // delayed duplicate must not roll back edits
	messages, err := db.Messages(ctx, original.Chat, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 1 || messages[0].Body != "third" || !messages[0].Deleted || !messages[0].Edited {
		t.Fatalf("edit or tombstone was lost: %+v", messages)
	}
	versions, err := db.Versions(ctx, original.Chat, original.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 2 || versions[0].Body != "first" || versions[1].Body != "second" {
		t.Fatalf("expected append-only previous versions, got %+v", versions)
	}
}

func TestUnknownRevocationRemainsTombstone(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	chat := "123@s.whatsapp.net"
	at := time.Now()
	if err := db.MarkDeleted(ctx, chat, "unknown", at); err != nil {
		t.Fatal(err)
	}
	if err := db.Insert(ctx, Message{Chat: chat, ID: "unknown", Body: "late", Timestamp: at}); err != nil {
		t.Fatal(err)
	}
	got, err := db.Messages(ctx, chat, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || !got[0].Deleted || got[0].Body != "" {
		t.Fatalf("unseen original content should not be fabricated: %+v", got)
	}
}

func TestSeparateChatsAllowSameMessageID(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	for _, chat := range []string{"a@g.us", "b@g.us"} {
		if err := db.Insert(ctx, Message{Chat: chat, ID: "same", Body: chat, Timestamp: time.Now()}); err != nil {
			t.Fatal(err)
		}
	}
	chats, err := db.Chats(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(chats) != 2 {
		t.Fatalf("expected 2 chats, got %+v", chats)
	}
}
