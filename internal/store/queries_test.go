package store

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSearchFindsRevokedAndLiteral(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	chat := "a@s.whatsapp.net"
	for _, m := range []Message{
		{Chat: chat, ID: "first", Body: "Hi 100% sure", Timestamp: time.Now()},
		{Chat: chat, ID: "second", Body: "Hi hello", Timestamp: time.Now().Add(time.Second)},
	} {
		if err := db.Insert(ctx, m); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.MarkDeleted(ctx, chat, "first", time.Now()); err != nil {
		t.Fatal(err)
	}
	hits, err := db.Search(ctx, "100%", 50)
	if err != nil || len(hits) != 1 || hits[0].ID != "first" || !hits[0].Deleted {
		t.Fatalf("got %#v, %v", hits, err)
	}
	hits, err = db.Search(ctx, "' OR 1=1", 10)
	if err != nil || len(hits) != 0 {
		t.Fatalf("query not literal: %v %v", hits, err)
	}
}
func TestReceiptsMonotonicAndEarly(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	chat := "a@s.whatsapp.net"
	for _, level := range []int{2, 1, 2} {
		if err := db.Receipt(ctx, chat, "id", level); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Insert(ctx, Message{Chat: chat, ID: "id", Timestamp: time.Now(), Outgoing: true}); err != nil {
		t.Fatal(err)
	}
	rows, err := db.Messages(ctx, chat, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Receipt != 2 {
		t.Fatalf("receipt regressed: %+v", rows)
	}
}
func TestMediaMetadata(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	chat := "a@s.whatsapp.net"
	if _, _, err := db.Media(ctx, chat, "missing"); !errors.Is(err, ErrMediaNotFound) {
		t.Fatal(err)
	}
	if err := db.Insert(ctx, Message{Chat: chat, ID: "id", Timestamp: time.Now()}); err != nil {
		t.Fatal(err)
	}
	if err := db.SaveMedia(ctx, chat, "id", "test.png", []byte{1, 2}); err != nil {
		t.Fatal(err)
	}
	name, raw, err := db.Media(ctx, chat, "id")
	if err != nil || name != "test.png" || len(raw) != 2 {
		t.Fatalf("media: %s %v %v", name, raw, err)
	}
}
