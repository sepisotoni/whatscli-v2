package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// SetAlias adds or clears a *local-only* display name. It does not modify any
// WhatsApp contact. No aliases are inferred from phone numbers or linked IDs.
func (d *DB) SetAlias(ctx context.Context, jid, name string) error {
	jid, name = strings.TrimSpace(jid), strings.TrimSpace(name)
	if jid == "" {
		return errors.New("a chat JID is required")
	}
	if !strings.Contains(jid, "@") || strings.ContainsAny(jid, " \t\r\n") {
		return errors.New("invalid chat JID")
	}
	if !utf8.ValidString(name) || utf8.RuneCountInString(name) > 96 {
		return errors.New("alias must contain at most 96 valid UTF-8 characters")
	}
	for _, c := range name {
		if unicode.IsControl(c) {
			return errors.New("alias cannot contain control characters")
		}
	}
	if name == "" {
		_, err := d.sql.ExecContext(ctx, `DELETE FROM chat_aliases WHERE jid=?`, jid)
		return err
	}
	_, err := d.sql.ExecContext(ctx, `INSERT INTO chat_aliases(jid,alias,updated_ns)
        VALUES(?,?,?) ON CONFLICT(jid) DO UPDATE SET alias=excluded.alias,
        updated_ns=excluded.updated_ns`, jid, name, time.Now().UTC().UnixNano())
	return err
}

// Alias returns an empty name when no local alias exists. A missing alias is
// not an error and must not trigger an online contact lookup.
func (d *DB) Alias(ctx context.Context, jid string) (string, error) {
	var name string
	err := d.sql.QueryRowContext(ctx, `SELECT alias FROM chat_aliases WHERE jid=?`, jid).Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return name, err
}