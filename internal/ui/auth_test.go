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
	if err != nil { t.Fatal(err) }
	defer func() { _ = db.Close() }()
	u := New(db, nil, config.Paths{})
	u.PreviewOnboarding()
	if u.authScreen != "welcome" { t.Fatalf("start page = %q", u.authScreen) }
	u.width, u.height = 120, 47
	u.authLayout(u.width, u.height)
	u.showAgreement()
	u.agreementCheck.SetChecked(true)
	u.acceptAgreement()
	if u.authScreen != "login" || u.pairingStarted { t.Fatal("demo screen should never initiate a WhatsApp connection") }
	if err := u.handleQR("onboarding-regression-fake-payload"); err != nil { t.Fatal(err) }
	if !strings.Contains(u.qrText, "[black:white]") || u.qrWidth == 0 { t.Fatal("QR not rendered") }
	if !strings.Contains(u.loginQR.GetText(true), "█") { t.Fatal("pairing page did not render QR modules") }
	u.handle(wa.Event{Kind: wa.Status, Text: "Pairing: code"})
	if u.authScreen != "login" { t.Fatal("non-connected status incorrectly dismissed login") }
	u.showFullscreenQR()
	if u.authScreen != "fullqr" { t.Fatal("full-screen QR mode unavailable") }
	u.authKey(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if u.authScreen != "login" { t.Fatal("Escape should return to login") }
	u.authKey(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if u.authScreen != "agreement" { t.Fatal("Escape should return to disclaimer") }
	u.authKey(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	if u.authScreen != "welcome" { t.Fatal("Escape should return to welcome") }
	u.handle(wa.Event{Kind: wa.Connected, Text: "Connected"})
	if u.authScreen != "" { t.Fatal("confirmed connection should enter chat workspace") }
}

func TestUnpairedConsentIsRequiredBeforeLogin(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "archive.sqlite3"))
	if err != nil { t.Fatal(err) }
	defer func() { _ = db.Close() }()
	u := New(db, nil, config.Paths{})
	u.PreviewOnboarding()
	u.openLogin()
	if u.authScreen != "agreement" || u.agreed { t.Fatal("consent gate bypassed") }
	u.acceptAgreement()
	if u.authScreen != "agreement" || u.agreed { t.Fatal("unchecked disclaimer must not proceed") }
	u.agreementCheck.SetChecked(true)
	if u.agreed { t.Fatal("checking alone must not count as agreement") }
	u.acceptAgreement()
	if !u.agreed || u.authScreen != "login" || u.pairingStarted {
		t.Fatal("explicit consent should open login without networking in demo")
	}
}

func TestWelcomeAnimationAndReferenceLayout(t *testing.T) {
	first := welcomeLogoFrame(0)
	second := welcomeLogoFrame(3)
	if first == second { t.Fatal("welcome animation frame did not change") }
	for _, want := range []string{"❯_", "────────────"} {
		if !strings.Contains(first, want) { t.Fatalf("welcome mark missing %q", want) }
	}
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "archive.sqlite3"))
	if err != nil { t.Fatal(err) }
	defer func() { _ = db.Close() }()
	u := New(db, nil, config.Paths{})
	u.PreviewOnboarding()
	if u.loginQR == nil || u.loginInstructions == nil { t.Fatal("pairing page is incomplete") }
	u.width, u.height = 110, 40
	u.authLayout(u.width, u.height)
	u.showAgreement()
	u.agreementCheck.SetChecked(true)
	u.acceptAgreement()
	if err := u.handleQR("not-a-real-whatsapp-token"); err != nil { t.Fatal(err) }
	if strings.Contains(u.loginQR.GetText(true), "Scan with WhatsApp") {
		t.Fatal("pairing page must not add a scan button inside the QR area")
	}
	u.width, u.height = 68, 24
	u.authLayout(u.width, u.height)
	if !strings.Contains(u.loginQR.GetText(true), "full-screen QR") {
		t.Fatal("small-terminal QR fallback missing")
	}
}
