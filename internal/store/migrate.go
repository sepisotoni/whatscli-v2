package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// schemaVersion is the on-disk archive format, independent of WhatsMeow's
// device/session database. Do not change an existing version after release.
const schemaVersion = 2

// migrate upgrades an empty or legacy v0/v1 SQLite archive transactionally.
// v0 includes the pre-versioned v0.2 database. Existing messages, deleted
// flags, old versions, receipt levels, and media pointers are never dropped.
func migrate(ctx context.Context, db *sql.DB) error {
	var version int
	if err := db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	if version < 0 || version > schemaVersion {
		return fmt.Errorf("unsupported archive schema version %d (supported: 0-%d); refusing downgrade", version, schemaVersion)
	}
	if version == schemaVersion {
		return nil
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if version < 1 {
		for _, query := range []string{
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
			`CREATE INDEX IF NOT EXISTS idx_messages_chat_time ON messages(chat_jid,timestamp_ns DESC)`,
			`CREATE TABLE IF NOT EXISTS message_receipts (
                chat_jid TEXT NOT NULL, message_id TEXT NOT NULL, level INTEGER NOT NULL DEFAULT 0,
                PRIMARY KEY(chat_jid,message_id))`,
			`CREATE TABLE IF NOT EXISTS message_media (
                chat_jid TEXT NOT NULL, message_id TEXT NOT NULL,
                name TEXT NOT NULL DEFAULT '', payload BLOB NOT NULL,
                PRIMARY KEY(chat_jid,message_id),
                FOREIGN KEY(chat_jid,message_id) REFERENCES messages(chat_jid,message_id))`,
		} {
			if _, err := tx.ExecContext(ctx, query); err != nil {
				return fmt.Errorf("install baseline schema: %w", err)
			}
		}
	}
	if version < 2 {
		if _, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS chat_aliases (
            jid TEXT PRIMARY KEY, alias TEXT NOT NULL,
            updated_ns INTEGER NOT NULL)`); err != nil {
			return fmt.Errorf("install local chat aliases: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, `PRAGMA user_version = 2`); err != nil {
		return errors.Join(errors.New("set schema version"), err)
	}
	return tx.Commit()
}
