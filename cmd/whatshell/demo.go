package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sepisotoni/whatshell-go/internal/config"
	"github.com/sepisotoni/whatshell-go/internal/store"
	"github.com/sepisotoni/whatshell-go/internal/ui"
)

// runDemo has no WhatsMeow client, never reads the user's actual session, and
// keeps fixtures in a private disposable directory. This mode can be used to
// review UI layout over SSH without pairing a WhatsApp account.
func runDemo(ctx context.Context) error {
	root, err := os.MkdirTemp("", "whatscli-demo-")
	if err != nil {
		return fmt.Errorf("create demo directory: %w", err)
	}
	// #nosec G703 -- root is allocated by MkdirTemp and never supplied by users.
	defer func() { _ = os.RemoveAll(root) }()
	paths := config.Paths{Dir: root, History: filepath.Join(root, "history.sqlite3"), Uploads: filepath.Join(root, "uploads"), Downloads: filepath.Join(root, "downloads")}
	db, err := store.Open(ctx, paths.History)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	if err := seedDemo(ctx, db); err != nil {
		return err
	}
	if err := ui.New(db, nil, paths).Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func seedDemo(ctx context.Context, db *store.DB) error {
	at := time.Now().UTC().Add(-50 * time.Minute)
	samples := []struct {
		jid, alias string
		bodies     []string
	}{
		{"sample-1@demo.invalid", "Design Team", []string{"UI draft ready 🟢", "The terminal layout looks good.", "Review chat aliases next."}},
		{"sample-2@demo.invalid", "Project Updates", []string{"This is fake, offline demo data.", "Type /help to see shortcuts.", "Type /alias NAME to rename this chat locally."}},
		{"sample-3@demo.invalid", "Test Conversation", []string{"Hello from WhatsCLI demo!", "Message versions and receipts work offline.", "No real WhatsApp messages are sent here."}},
	}
	for i, s := range samples {
		if err := db.SetAlias(ctx, s.jid, s.alias); err != nil {
			return err
		}
		for j, body := range s.bodies {
			msg := store.Message{Chat: s.jid, ID: fmt.Sprintf("demo-%d-%d", i, j), Sender: s.jid, Body: body,
				Timestamp: at.Add(time.Duration(i*7+j*3) * time.Minute), Outgoing: j%2 == 1}
			if err := db.Insert(ctx, msg); err != nil {
				return err
			}
		}
	}
	if err := db.Edit(ctx, samples[0].jid, "demo-0-0", "UI draft approved 🟢", at.Add(5*time.Minute)); err != nil {
		return err
	}
	if err := db.MarkDeleted(ctx, samples[1].jid, "demo-1-0", at.Add(15*time.Minute)); err != nil {
		return err
	}
	return db.Receipt(ctx, samples[2].jid, "demo-2-1", 2)
}
