package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// SearchResult includes a message's stable ID so users can inspect its revisions.
type SearchResult struct {
	Chat      string
	ID        string
	Body      string
	Timestamp time.Time
	Deleted   bool
}

// Search performs a literal substring match on locally archived text.  It
// intentionally includes revoked messages because this is a local archive.
func (d *DB) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("search term must not be empty")
	}
	if limit < 1 || limit > 100 {
		limit = 30
	}
	rows, err := d.sql.QueryContext(ctx, `
      SELECT chat_jid, message_id, body, timestamp_ns, deleted
      FROM messages WHERE instr(lower(body),lower(?)) > 0
      ORDER BY timestamp_ns DESC LIMIT ?`, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]SearchResult, 0)
	for rows.Next() {
		var r SearchResult
		var ns int64
		if err := rows.Scan(&r.Chat, &r.ID, &r.Body, &ns, &r.Deleted); err != nil {
			return nil, err
		}
		r.Timestamp = time.Unix(0, ns)
		result = append(result, r)
	}
	return result, rows.Err()
}

// Receipt levels are monotonic across updates: sent (0), delivered (1), read (2).
// Receipts can arrive ahead of archive inserts and are therefore stored independently.
func (d *DB) Receipt(ctx context.Context, chat, id string, level int) error {
	if chat == "" || id == "" || (level != 1 && level != 2) {
		return errors.New("invalid receipt")
	}
	_, err := d.sql.ExecContext(ctx, `INSERT INTO message_receipts(chat_jid,message_id,level)
       VALUES(?,?,?) ON CONFLICT(chat_jid,message_id)
       DO UPDATE SET level=max(level,excluded.level)`, chat, id, level)
	return err
}

// SaveMedia persists the protobuf metadata needed for a later on-demand media
// download. Media ciphertext/plaintext is NOT automatically downloaded.
func (d *DB) SaveMedia(ctx context.Context, chat, id, name string, payload []byte) error {
	if chat == "" || id == "" || len(payload) == 0 || len(payload) > 2<<20 {
		return errors.New("invalid media metadata")
	}
	_, err := d.sql.ExecContext(ctx, `INSERT INTO message_media(chat_jid,message_id,name,payload)
      VALUES(?,?,?,?) ON CONFLICT(chat_jid,message_id) DO NOTHING`, chat, id, name, payload)
	return err
}

var ErrMediaNotFound = errors.New("no downloadable media metadata found")

func (d *DB) Media(ctx context.Context, chat, id string) (name string, payload []byte, err error) {
	err = d.sql.QueryRowContext(ctx, `SELECT name,payload FROM message_media WHERE chat_jid=? AND message_id=?`, chat, id).Scan(&name, &payload)
	if errors.Is(err, sql.ErrNoRows) {
		err = ErrMediaNotFound
	}
	return
}

// SetReceiptLevel is deliberately not exposed for arbitrary writes: call Receipt.
func (d *DB) ReceiptLevel(ctx context.Context, chat, id string) (int, error) {
	var level int
	err := d.sql.QueryRowContext(ctx, `SELECT level FROM message_receipts WHERE chat_jid=? AND message_id=?`, chat, id).Scan(&level)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("receipt lookup: %w", err)
	}
	return level, nil
}
