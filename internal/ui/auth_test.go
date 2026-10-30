package ui

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/sepisotoni/whatshell-go/internal/config"
	"github.com/sepisotoni/whatshell-go/internal/store"
	"github.com/sepisotoni/whatshell-go/internal/wa"
)

func TestOnboardingNavigationAndQRStatus(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "archive.sqlite3"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	u := New(db, nil, config.Paths{})
	u.PreviewOnboarding()
	if u.authScreen != "welcome" {
		t.Fatalf("start page = %q", u.authScreen)
	}
	u.authLayout(120, 47)
	u.openLogin()
	if u.authScreen != "login" || u.pairingStarted {
		t.Fatal("demo screen should never initiate a WhatsApp connection")
	}
	if err := u.handleQR("onboarding-regression-fake-payload"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(u.qrText, "[black:white]") || u.qrWidth == 0 {
		t.Fatal("QR not rendered")
	}
	u.handle(wa.Event{Kind: wa.Status, Text: "Pairing: code"})
	if u.authScreen != "login" {
		t.Fatal("non-connected pairing status incorrectly dismissed login")
	}
	u.showFullscreenQR()
	if u.authScreen != "fullqr" {
		t.Fatal("full-screen QR mode unavailable")
	}
	u.authKey(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if u.authScreen != "login" {
		t.Fatal("Escape should return to login, not welcome")
	}
	u.authKey(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if u.authScreen != "welcome" {
		t.Fatal("Escape should return to welcome")
	}
	u.handle(wa.Event{Kind: wa.Connected, Text: "Connected"})
	if u.authScreen != "" {
		t.Fatal("confirmed connection should enter chat workspace")
	}
}
