package store

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func schemaAt(t *testing.T, d *DB) int {
	t.Helper()
	var n int
	if err := d.sql.QueryRow(`PRAGMA user_version`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestMigrateNewDatabaseAndAliasPersistence(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "fresh.sqlite3")
	db, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if got := schemaAt(t, db); got != schemaVersion {
		t.Fatalf("schema = %d", got)
	}
	chat := "123@s.whatsapp.net"
	if err := db.Insert(ctx, Message{Chat: chat, ID: "m1", Body: "hi", Timestamp: time.Now()}); err != nil {
		t.Fatal(err)
	}
	if err := db.SetAlias(ctx, chat, "Family Chat 🟢"); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	chats, err := db.Chats(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(chats) != 1 || chats[0].Alias != "Family Chat 🟢" {
		t.Fatalf("alias after reopen: %+v", chats)
	}
	if err := db.SetAlias(ctx, chat, ""); err != nil {
		t.Fatal(err)
	}
	alias, err := db.Alias(ctx, chat)
	if err != nil || alias != "" {
		t.Fatalf("clear alias %q %v", alias, err)
	}
	for _, invalid := range []string{"bad\nname", strings.Repeat("x", 97)} {
		if err := db.SetAlias(ctx, chat, invalid); err == nil {
			t.Fatalf("accepted invalid alias: %q", invalid)
		}
	}
}

func TestUpgradePreVersionedArchivePreservesMessagesAndVersions(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "legacy.sqlite3")
	legacy, err := sql.Open("sqlite", DSN(path))
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`CREATE TABLE messages(chat_jid TEXT NOT NULL,message_id TEXT NOT NULL,sender_jid TEXT NOT NULL DEFAULT '',body TEXT NOT NULL DEFAULT '',timestamp_ns INTEGER NOT NULL,outgoing INTEGER NOT NULL DEFAULT 0,deleted INTEGER NOT NULL DEFAULT 0,edited INTEGER NOT NULL DEFAULT 0,PRIMARY KEY(chat_jid,message_id))`,
		`CREATE TABLE message_versions(id INTEGER PRIMARY KEY AUTOINCREMENT,chat_jid TEXT NOT NULL,message_id TEXT NOT NULL,previous_body TEXT NOT NULL,changed_at_ns INTEGER NOT NULL,FOREIGN KEY(chat_jid,message_id) REFERENCES messages(chat_jid,message_id))`,
		`INSERT INTO messages(chat_jid,message_id,body,timestamp_ns,deleted,edited) VALUES('a@g.us','m','edited body',123,1,1)`,
		`INSERT INTO message_versions(chat_jid,message_id,previous_body,changed_at_ns) VALUES('a@g.us','m','original body',123)`,
	} {
		if _, err := legacy.ExecContext(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}
	if err := legacy.Close(); err != nil {
		t.Fatal(err)
	}
	db, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if got := schemaAt(t, db); got != schemaVersion {
		t.Fatalf("schema %d", got)
	}
	messages, err := db.Messages(ctx, "a@g.us", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 1 || !messages[0].Deleted || messages[0].Body != "edited body" {
		t.Fatalf("lost cached message %+v", messages)
	}
	versions, err := db.Versions(ctx, "a@g.us", "m")
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 1 || versions[0].Body != "original body" {
		t.Fatalf("lost cached versions %+v", versions)
	}
	if err := db.Receipt(ctx, "a@g.us", "m", 2); err != nil {
		t.Fatal(err)
	}
	if err := db.SetAlias(ctx, "a@g.us", "Local Group"); err != nil {
		t.Fatal(err)
	}
}

func TestMigrateRejectsFutureSchemaWithoutDowngrading(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "future.sqlite3")
	db, err := sql.Open("sqlite", DSN(path))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.ExecContext(ctx, "PRAGMA user_version=999"); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(ctx, path); err == nil || !strings.Contains(err.Error(), "unsupported archive schema") {
		t.Fatalf("expected downgrade refusal, got %v", err)
	}
	db, err = sql.Open("sqlite", DSN(path))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	var version int
	if err = db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != 999 {
		t.Fatalf("future version overwritten: %d", version)
	}
}
