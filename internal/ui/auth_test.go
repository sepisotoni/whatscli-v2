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
	u.showAgreement()
	u.agreementCheck.SetChecked(true)
	u.acceptAgreement()
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
	if u.authScreen != "agreement" {
		t.Fatal("Escape should return to the agreement")
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

func TestUnpairedConsentIsRequiredBeforeLogin(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "archive.sqlite3"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	u := New(db, nil, config.Paths{})
	u.PreviewOnboarding()
	u.openLogin() // Calling the method directly must not bypass consent.
	if u.authScreen != "agreement" || u.agreed {
		t.Fatal("consent gate bypassed")
	}
	u.acceptAgreement()
	if u.authScreen != "agreement" || u.agreed {
		t.Fatal("unchecked disclosure must not proceed")
	}
	u.agreementCheck.SetChecked(true)
	if u.agreed {
		t.Fatal("checking alone must not count as agreement")
	}
	u.acceptAgreement()
	if !u.agreed || u.authScreen != "login" || u.pairingStarted {
		t.Fatal("explicit consent should open login but not contact WhatsApp in demo")
	}
	u.authKey(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if u.authScreen != "agreement" {
		t.Fatal("Back from login should show the disclosure")
	}
	u.authKey(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if u.authScreen != "welcome" {
		t.Fatal("Back from the disclosure should show welcome")
	}
	for _, dims := range [][2]int{{118, 44}, {80, 28}, {58, 19}} {
		u.authLayout(dims[0], dims[1])
	}
}

func TestWelcomeLogoAnimationAndBorderlessQR(t *testing.T) {
	a := welcomeLogoFrame(0)
	b := welcomeLogoFrame(2)
	if a == b || !strings.Contains(a, "❯_") || !strings.Contains(a, "╭") {
		t.Fatal("welcome logo and symmetric glow rails must animate")
	}
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "archive.sqlite3"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	view := New(db, nil, config.Paths{})
	view.PreviewOnboarding()
	if view.loginQRRow == nil || view.loginQRCenter == nil {
		t.Fatal("QR must be centred within a borderless dark panel")
	}
	view.width, view.height = 118, 44
	view.authLayout(view.width, view.height)
	view.showAgreement()
	view.agreementCheck.SetChecked(true)
	view.acceptAgreement()
	if err := view.handleQR("not-a-real-whatsapp-token"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(view.loginQR.GetText(true), "█") {
		t.Fatal("QR modules missing after layout")
	}
}
