// Package store implements the local, append-preserving chat archive.
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"github.com/sepisotoni/whatshell-go/internal/config"
	sqlite "modernc.org/sqlite"
)

// WhatsMeow's SQL dialect uses the driver name "sqlite3". Register an alias
// for the pure-Go modernc driver, without requiring CGO or libsqlite3.
func init() {
	sql.Register("sqlite3", &sqlite.Driver{})
}

func DSN(path string) string {
	path = filepath.ToSlash(path)
	if filepath.VolumeName(path) != "" {
		path = "/" + path
	}
	return (&url.URL{
		Scheme:   "file",
		Path:     path,
		RawQuery: "_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)",
	}).String()
}

type DB struct{ sql *sql.DB }

type Message struct {
	Chat      string
	ID        string
	Sender    string
	Body      string
	Timestamp time.Time
	Outgoing  bool
	Deleted   bool
	Edited    bool
	Receipt   int // 0 sent/unknown, 1 delivered, 2 read
}

type Chat struct {
	JID       string
	LastBody  string
	Timestamp time.Time
}

type Version struct {
	Body      string
	ChangedAt time.Time
}

func Open(ctx context.Context, path string) (*DB, error) {
	if err := config.ProtectFile(path); err != nil {
		return nil, fmt.Errorf("protect archive: %w", err)
	}
	db, err := sql.Open("sqlite", DSN(path))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	for _, pragma := range []string{"PRAGMA journal_mode=WAL", "PRAGMA synchronous=FULL"} {
		if _, err = db.ExecContext(ctx, pragma); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("sqlite configuration: %w", err)
		}
	}
	for _, statement := range []string{
		`CREATE TABLE IF NOT EXISTS messages (
		 chat_jid TEXT NOT NULL, message_id TEXT NOT NULL,
		 sender_jid TEXT NOT NULL DEFAULT '', body TEXT NOT NULL DEFAULT '',
		 timestamp_ns INTEGER NOT NULL, outgoing INTEGER NOT NULL DEFAULT 0,
		 deleted INTEGER NOT NULL DEFAULT 0, edited INTEGER NOT NULL DEFAULT 0,
		 PRIMARY KEY(chat_jid,message_id))`,
		`CREATE TABLE IF NOT EXISTS message_versions (
		 id INTEGER PRIMARY KEY AUTOINCREMENT, chat_jid TEXT NOT NULL,
		 message_id TEXT NOT NULL, previous_body TEXT NOT NULL,
		 changed_at_ns INTEGER NOT NULL,
		 FOREIGN KEY(chat_jid,message_id) REFERENCES messages(chat_jid,message_id))`,
		`CREATE INDEX IF NOT EXISTS idx_messages_chat_time
         ON messages(chat_jid,timestamp_ns DESC)`,
		`CREATE TABLE IF NOT EXISTS message_receipts (
          chat_jid TEXT NOT NULL, message_id TEXT NOT NULL, level INTEGER NOT NULL DEFAULT 0,
          PRIMARY KEY(chat_jid,message_id))`,
		`CREATE TABLE IF NOT EXISTS message_media (
          chat_jid TEXT NOT NULL, message_id TEXT NOT NULL, name TEXT NOT NULL DEFAULT '', payload BLOB NOT NULL,
          PRIMARY KEY(chat_jid,message_id), FOREIGN KEY(chat_jid,message_id) REFERENCES messages(chat_jid,message_id))`,
	} {
		if _, err = db.ExecContext(ctx, statement); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("sqlite migration: %w", err)
		}
	}
	return &DB{sql: db}, nil
}

func (d *DB) Close() error { return d.sql.Close() }

// Insert adds an observed message once; delayed history sync cannot overwrite edits.
func (d *DB) Insert(ctx context.Context, m Message) error {
	if m.Chat == "" || m.ID == "" {
		return errors.New("chat JID and message ID required")
	}
	_, err := d.sql.ExecContext(ctx,
		`INSERT INTO messages(chat_jid,message_id,sender_jid,body,timestamp_ns,outgoing,deleted,edited)
		 VALUES(?,?,?,?,?,?,?,?) ON CONFLICT(chat_jid,message_id) DO NOTHING`,
		m.Chat, m.ID, m.Sender, m.Body, m.Timestamp.UnixNano(), m.Outgoing, m.Deleted, m.Edited)
	return err
}

// Edit preserves the previous body, including when a message is currently deleted.
// If the original was not seen, the first observed edit is used as a placeholder.
func (d *DB) Edit(ctx context.Context, chat, id, body string, at time.Time) error {
	tx, err := d.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var old string
	err = tx.QueryRowContext(ctx, `SELECT body FROM messages WHERE chat_jid=? AND message_id=?`, chat, id).Scan(&old)
	if errors.Is(err, sql.ErrNoRows) {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO messages(chat_jid,message_id,body,timestamp_ns,edited) VALUES(?,?,?,?,1)`,
			chat, id, body, at.UnixNano())
	} else if err == nil && old != body {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO message_versions(chat_jid,message_id,previous_body,changed_at_ns) VALUES(?,?,?,?)`,
			chat, id, old, at.UnixNano())
		if err == nil {
			_, err = tx.ExecContext(ctx, `UPDATE messages SET body=?,edited=1 WHERE chat_jid=? AND message_id=?`, body, chat, id)
		}
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

// MarkDeleted retains the already-cached body. Unseen revocations become tombstones.
func (d *DB) MarkDeleted(ctx context.Context, chat, id string, at time.Time) error {
	_, err := d.sql.ExecContext(ctx,
		`INSERT INTO messages(chat_jid,message_id,timestamp_ns,deleted) VALUES(?,?,?,1)
		 ON CONFLICT(chat_jid,message_id) DO UPDATE SET deleted=1`, chat, id, at.UnixNano())
	return err
}

func (d *DB) Chats(ctx context.Context) ([]Chat, error) {
	rows, err := d.sql.QueryContext(ctx,
		`SELECT m.chat_jid,m.body,m.timestamp_ns FROM messages m
		 WHERE m.rowid=(SELECT rowid FROM messages x WHERE x.chat_jid=m.chat_jid
		 ORDER BY x.timestamp_ns DESC,x.rowid DESC LIMIT 1)
		 ORDER BY m.timestamp_ns DESC LIMIT 200`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var chats []Chat
	for rows.Next() {
		var c Chat
		var ns int64
		if err := rows.Scan(&c.JID, &c.LastBody, &ns); err != nil {
			return nil, err
		}
		c.Timestamp = time.Unix(0, ns)
		chats = append(chats, c)
	}
	return chats, rows.Err()
}

func (d *DB) Messages(ctx context.Context, chat string, limit int) ([]Message, error) {
	if limit <= 0 || limit > 2000 {
		limit = 150
	}
	rows, err := d.sql.QueryContext(ctx,
		`SELECT m.chat_jid,m.message_id,m.sender_jid,m.body,m.timestamp_ns,
        m.outgoing,m.deleted,m.edited,coalesce(r.level,0)
        FROM (SELECT * FROM messages WHERE chat_jid=? ORDER BY timestamp_ns DESC LIMIT ?) m
        LEFT JOIN message_receipts r ON r.chat_jid=m.chat_jid AND r.message_id=m.message_id
        ORDER BY m.timestamp_ns ASC`, chat, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var messages []Message
	for rows.Next() {
		var m Message
		var ns int64
		if err := rows.Scan(&m.Chat, &m.ID, &m.Sender, &m.Body, &ns, &m.Outgoing, &m.Deleted, &m.Edited, &m.Receipt); err != nil {
			return nil, err
		}
		m.Timestamp = time.Unix(0, ns)
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

func (d *DB) Versions(ctx context.Context, chat, id string) ([]Version, error) {
	rows, err := d.sql.QueryContext(ctx,
		`SELECT previous_body,changed_at_ns FROM message_versions
		 WHERE chat_jid=? AND message_id=? ORDER BY id ASC`, chat, id)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var versions []Version
	for rows.Next() {
		var v Version
		var ns int64
		if err := rows.Scan(&v.Body, &ns); err != nil {
			return nil, err
		}
		v.ChangedAt = time.Unix(0, ns)
		versions = append(versions, v)
	}
	return versions, rows.Err()
}
