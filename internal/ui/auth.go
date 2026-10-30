package ui

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sepisotoni/whatshell-go/internal/wa"
)

// The onboarding screen is a real terminal UI, not a QR image printed into a log.
// The QR is only shown on a requested login flow, and the client never stores it.
func (u *UI) setupAuthentication() {
	// A distinct WhatsCLI identity: WhatsApp's logo belongs to Meta and is not
	// used as a third-party application's logo or as an endorsement cue.
	u.welcomeButton = tview.NewButton("  Get started  →  ")
	u.welcomeButton.SetBackgroundColor(green)
	u.welcomeButton.SetLabelColor(darker)
	u.welcomeButton.SetSelectedFunc(func() { u.showAgreement() })
	u.welcomeQuit = tview.NewButton("  Exit  ")
	u.welcomeQuit.SetBackgroundColor(slate)
	u.welcomeQuit.SetLabelColor(tcell.ColorWhite)
	u.welcomeQuit.SetSelectedFunc(func() { u.app.Stop() })

	// A distinct terminal chat mark with pulsing symmetric rails.
	u.welcomeGlow = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	u.welcomeGlow.SetBackgroundColor(slate)
	u.welcomeGlow.SetText(welcomeLogoFrame(0))
	hero := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	hero.SetBackgroundColor(slate)
	hero.SetText("\n[#E9EDEF::b]Welcome to [#00E99A::b]WhatsCLI[-]\n\n" +
		"[#B5C3C8]Your conversations. Your terminal.[-]\n\n" +
		"[#245147]────────────────────────────────────────[-]")
	actions := tview.NewFlex().SetDirection(tview.FlexColumn)
	actions.AddItem(nil, 0, 1, false).
		AddItem(u.welcomeButton, 24, 0, true).
		AddItem(nil, 3, 0, false).
		AddItem(u.welcomeQuit, 11, 0, false).
		AddItem(nil, 0, 1, false)
	details := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	details.SetBackgroundColor(slate)
	details.SetText("[#607E79]INDEPENDENT PROJECT  ·  NOT AFFILIATED WITH WHATSAPP / META[-]\n" +
		"[#607E79]Designed for keyboard, mouse, and SSH terminals.[-]")
	card := tview.NewFlex().SetDirection(tview.FlexRow)
	card.AddItem(nil, 1, 0, false).
		AddItem(u.welcomeGlow, 8, 0, false).
		AddItem(hero, 0, 1, false).
		AddItem(actions, 3, 0, true).
		AddItem(details, 4, 0, false)
	card.SetBackgroundColor(slate)
	card.SetBorder(true).SetBorderColor(tcell.NewRGBColor(0, 105, 75))
	// The title lives in the centre of the welcome page.
	u.welcomeCard = card
	u.welcomeOuter = tview.NewFlex().SetDirection(tview.FlexColumn)
	u.welcomeOuter.SetBackgroundColor(darker)
	u.welcomeOuter.AddItem(nil, 0, 1, false).
		AddItem(card, 88, 0, true).
		AddItem(nil, 0, 1, false)
	welcome := tview.NewFlex().SetDirection(tview.FlexRow)
	welcome.SetBackgroundColor(darker)
	welcome.AddItem(nil, 0, 1, false).
		AddItem(u.welcomeOuter, 24, 0, true).
		AddItem(nil, 0, 1, false)
	u.welcomePage = welcome
	u.pages.AddPage("welcome", welcome, true, false)

	// Agreement is a separate page, not a tiny checkbox hidden beside the QR.
	// All buttons are actual TView controls and work with pointer and keyboard.
	u.agreementText = tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	u.agreementText.SetBackgroundColor(slate)
	u.agreementText.SetText("[#8696A0]BEFORE YOU CONNECT[-]   [#00E99A]STEP 01 / 02[-]\n" +
		"[#E9EDEF::b]Disclaimer & consent[-]\n" +
		"[#8696A0]A few important things to know before you continue.[-]\n\n" +
		"[#CAD5D7]01  Independent and unofficial.[-]\n" +
		"    WhatsCLI is not created, sponsored, certified or supported by\n" +
		"    WhatsApp or Meta. The WhatsApp name belongs to its owners.\n\n" +
		"[#CAD5D7]02  Your account, your decision.[-]\n" +
		"    Linking relies on an unofficial client and may stop working or\n" +
		"    carry account restrictions. Respect WhatsApp's terms of service.\n\n" +
		"[#CAD5D7]03  Your local data stays your responsibility.[-]\n" +
		"    Device credentials and cached chat history are stored on this\n" +
		"    computer. The archive is not encrypted by WhatsCLI; protect it.\n" +
		"    Cached copies of edits/deletions may remain on your device.\n\n" +
		"[#CAD5D7]04  Early software, with no guarantee.[-]\n" +
		"    This preview is experimental. Back up your data and use a test\n" +
		"    account first. You can stop here without connecting anything.\n\n" +
		"[#8696A0]More: https://www.whatsapp.com/legal/terms-of-service[-]")
	u.agreementCheck = tview.NewCheckbox().SetLabel(" I understand this is an unofficial app and accept the risks above")
	u.agreementCheck.SetLabelColor(tcell.ColorWhite)
	u.agreementCheck.SetChangedFunc(func(checked bool) {
		if checked {
			u.agreementNext.SetBackgroundColor(green)
			u.agreementNext.SetLabelColor(darker)
			u.agreementStatus.SetText("[#25D366]Ready to continue. Nothing is linked yet.[-]")
		} else {
			u.agreementNext.SetBackgroundColor(tcell.NewRGBColor(51, 65, 69))
			u.agreementNext.SetLabelColor(tcell.ColorGray)
			u.agreementStatus.SetText("[#8696A0]Check the box to enable Continue.[-]")
		}
	})
	u.agreementNext = tview.NewButton("  Agree & continue  →  ")
	u.agreementNext.SetBackgroundColor(tcell.NewRGBColor(51, 65, 69))
	u.agreementNext.SetLabelColor(tcell.ColorGray)
	u.agreementNext.SetSelectedFunc(func() { u.acceptAgreement() })
	u.agreementBack = tview.NewButton("  ← Back  ")
	u.agreementBack.SetBackgroundColor(slate)
	u.agreementBack.SetLabelColor(tcell.ColorWhite)
	u.agreementBack.SetSelectedFunc(func() { u.showWelcome() })
	u.agreementStatus = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	u.agreementStatus.SetBackgroundColor(slate)
	u.agreementStatus.SetText("[#8696A0]Check the box to enable Continue.[-]")
	checkLine := tview.NewFlex().SetDirection(tview.FlexColumn)
	checkLine.AddItem(nil, 2, 0, false).
		AddItem(u.agreementCheck, 0, 1, true).
		AddItem(nil, 2, 0, false)
	buttons := tview.NewFlex().SetDirection(tview.FlexColumn)
	buttons.AddItem(u.agreementBack, 15, 0, false).
		AddItem(nil, 0, 1, false).
		AddItem(u.agreementNext, 27, 0, true)
	agreementCard := tview.NewFlex().SetDirection(tview.FlexRow)
	agreementCard.AddItem(u.agreementText, 0, 1, false).
		AddItem(checkLine, 2, 0, true).
		AddItem(u.agreementStatus, 1, 0, false).
		AddItem(buttons, 3, 0, true)
	agreementCard.SetBorder(true).SetBorderColor(tcell.NewRGBColor(48, 69, 71))
	agreementCard.SetBackgroundColor(slate)
	// Heading is inside the page rather than duplicated in the border.
	u.agreementCard = agreementCard
	u.agreementOuter = tview.NewFlex().SetDirection(tview.FlexColumn)
	u.agreementOuter.SetBackgroundColor(darker)
	u.agreementOuter.AddItem(nil, 0, 1, false).
		AddItem(agreementCard, 88, 0, true).
		AddItem(nil, 0, 1, false)
	agreementPage := tview.NewFlex().SetDirection(tview.FlexRow)
	agreementPage.SetBackgroundColor(darker)
	agreementPage.AddItem(nil, 0, 1, false).
		AddItem(u.agreementOuter, 30, 0, true).
		AddItem(nil, 0, 1, false)
	u.agreementPage = agreementPage
	u.pages.AddPage("agreement", agreementPage, true, false)

	heading := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignLeft)
	heading.SetBackgroundColor(slate)
	heading.SetText("[#00E69C::b]◇ LINK A DEVICE[-]\n" +
		"[#B5C3C8]    Log in to WhatsApp[-]\n\n" +
		"[#00D78D]01[-]  On your phone, open WhatsApp.\n\n" +
		"[#00D78D]02[-]  Go to Settings or ⋮ → Linked devices.\n\n" +
		"[#00D78D]03[-]  Select Link a device and scan the QR.\n\n" +
		"[#00D78D]04[-]  Keep your terminal open until linked.\n\n" +
		"[#78948F]ⓘ The code refreshes automatically. Never share it.[-]\n" +
		"[#78948F]Phone-number pairing is not available yet.[-]")
	u.loginQR = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	u.loginQR.SetBackgroundColor(slate)
	u.loginQR.SetTextColor(tcell.ColorWhite)
	u.loginQR.SetText("[#78948F]Waiting for QR code…[-]")
	// White is confined to the code itself, not the entire right-hand panel.
	// Do not add a second glowing border or decorative Scan button.
	u.loginQR.SetBorder(false)
	u.loginQRRow = tview.NewFlex().SetDirection(tview.FlexColumn)
	u.loginQRRow.SetBackgroundColor(slate)
	u.loginQRRow.AddItem(nil, 0, 1, false).
		AddItem(u.loginQR, 47, 0, false).
		AddItem(nil, 0, 1, false)
	u.loginQRCenter = tview.NewFlex().SetDirection(tview.FlexRow)
	u.loginQRCenter.SetBackgroundColor(slate)
	u.loginQRCenter.AddItem(nil, 0, 1, false).
		AddItem(u.loginQRRow, 25, 0, false).
		AddItem(nil, 0, 1, false)
	u.loginState = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	u.loginState.SetBackgroundColor(slate)
	u.loginState.SetText("[#8696A0]Waiting for device pairing…[-]")
	u.loginInstructions = heading
	u.loginSpacer = tview.NewBox()
	u.loginColumns = tview.NewFlex().SetDirection(tview.FlexColumn)
	u.loginColumns.AddItem(heading, 0, 1, false).
		AddItem(u.loginSpacer, 2, 0, false).
		AddItem(u.loginQRCenter, 0, 1, false)
	u.loginBack = tview.NewButton("  ← Back  ")
	u.loginBack.SetBackgroundColor(slate)
	u.loginBack.SetLabelColor(tcell.ColorWhite)
	u.loginBack.SetSelectedFunc(func() { u.showWelcome() })
	u.loginExpand = tview.NewButton("  ⤢ Full-screen QR  ")
	u.loginExpand.SetBackgroundColor(green)
	u.loginExpand.SetLabelColor(darker)
	u.loginExpand.SetSelectedFunc(func() { u.showFullscreenQR() })
	actionLine := tview.NewFlex().SetDirection(tview.FlexColumn)
	actionLine.AddItem(u.loginBack, 16, 0, true).
		AddItem(nil, 0, 1, false).
		AddItem(u.loginExpand, 25, 0, false)
	loginCard := tview.NewFlex().SetDirection(tview.FlexRow)
	loginCard.AddItem(u.loginColumns, 0, 1, false).
		AddItem(u.loginState, 2, 0, false).
		AddItem(actionLine, 3, 0, true)
	loginCard.SetBackgroundColor(slate)
	loginCard.SetBorder(true).SetBorderColor(tcell.NewRGBColor(50, 61, 65))
	// The title lives inside the left column, matching the reference.
	u.loginCard = loginCard
	u.loginOuter = tview.NewFlex().SetDirection(tview.FlexColumn)
	u.loginOuter.SetBackgroundColor(darker)
	u.loginOuter.AddItem(nil, 0, 1, false).
		AddItem(loginCard, 106, 0, true).
		AddItem(nil, 0, 1, false)
	login := tview.NewFlex().SetDirection(tview.FlexRow)
	login.SetBackgroundColor(darker)
	login.AddItem(nil, 0, 1, false).
		AddItem(u.loginOuter, 37, 0, true).
		AddItem(nil, 0, 1, false)
	u.loginPage = login
	u.pages.AddPage("login", login, true, false)

	u.fullQR = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	u.fullQR.SetBackgroundColor(tcell.ColorWhite)
	u.fullQR.SetTextColor(tcell.ColorBlack)
	u.pages.AddPage("fullqr", u.fullQR, true, false)
	u.authScreen = "welcome"
	u.pages.SwitchToPage("welcome")
	u.app.SetFocus(u.welcomeButton)
}

func (u *UI) authLayout(width, height int) {
	if u.welcomeOuter == nil || u.loginOuter == nil || u.agreementOuter == nil {
		return
	}
	max := func(n, limit int) int {
		if limit < 24 {
			limit = 24
		}
		if n > limit {
			return limit
		}
		return n
	}
	u.welcomeOuter.ResizeItem(u.welcomeCard, max(88, width-4), 0)
	u.agreementOuter.ResizeItem(u.agreementCard, max(88, width-4), 0)
	heightFor := func(desired int) int {
		available := height - 5
		if available < 8 {
			return 8
		}
		if available < desired {
			return available
		}
		return desired
	}
	u.welcomePage.ResizeItem(u.welcomeOuter, heightFor(26), 0)
	u.agreementPage.ResizeItem(u.agreementOuter, heightFor(30), 0)
	u.loginPage.ResizeItem(u.loginOuter, heightFor(37), 0)
	u.loginOuter.ResizeItem(u.loginCard, max(106, width-4), 0)
	// The QR needs its original module size to remain reliably scannable.
	// Keep the instructions beside it on wider screens; on small screens use
	// the dedicated full-screen QR action rather than scaling QR modules.
	if width < 100 {
		u.loginColumns.SetDirection(tview.FlexRow)
		u.loginColumns.ResizeItem(u.loginSpacer, 0, 0)
		u.loginColumns.ResizeItem(u.loginInstructions, 12, 0)
		u.loginColumns.ResizeItem(u.loginQRCenter, 0, 1)
	} else {
		u.loginColumns.SetDirection(tview.FlexColumn)
		u.loginColumns.ResizeItem(u.loginInstructions, 0, 1)
		u.loginColumns.ResizeItem(u.loginSpacer, 2, 0)
		u.loginColumns.ResizeItem(u.loginQRCenter, 0, 1)
	}
	// Keep the scannable module size fixed, with a small white quiet zone,
	// and place the white code in the centre of the dark pairing column.
	qrWidth, qrHeight := 47, 25
	if u.qrWidth > 0 {
		qrWidth, qrHeight = u.qrWidth+4, u.qrHeight+2
	}
	u.loginQRRow.ResizeItem(u.loginQR, qrWidth, 0)
	u.loginQRCenter.ResizeItem(u.loginQRRow, qrHeight, 0)
	if u.qrWidth > 0 && (width < u.qrWidth+8 || height < u.qrHeight+12) {
		u.loginQR.SetBackgroundColor(slate)
		u.loginQR.SetTextColor(tcell.ColorWhite)
		u.loginQR.SetText("[#8696A0]QR ready — press Q for full-screen QR[-]")
		u.loginState.SetText(fmt.Sprintf("[yellow]Enlarge terminal to %dx%d or press Q for full-screen QR.[-]", u.qrWidth+8, u.qrHeight+12))
	} else if u.qrText != "" {
		u.loginQR.SetBackgroundColor(tcell.ColorWhite)
		u.loginQR.SetTextColor(tcell.ColorBlack)
		u.loginQR.SetText("\n" + u.qrText)
	}
}

// welcomeLogoFrame animates the glow and matching accent dots on both
// sides of WhatsCLI's chat/terminal icon. It works in ordinary SSH terminals.
func welcomeLogoFrame(step int) string {
	shades := [...]string{"#08784F", "#0BAC71", "#0CDD83", "#20F2A2", "#0CDD83", "#0BAC71"}
	rail := [...]string{"●───────────", "──●─────────", "────●───────", "──────●─────", "────────●───", "──────────●─"}
	accent := shades[step%len(shades)]
	line := rail[step%len(rail)]
	return fmt.Sprintf("\n[#134338]%s[-]    [%s]╭───────╮[-]    [#134338]%s[-]\n"+
		"[#134338]%s[-]    [%s]│  ❯_   │[-]    [#134338]%s[-]\n"+
		"[#134338]%s[-]    [%s]╰──╮    │[-]    [#134338]%s[-]\n"+
		"                 [%s]╰────╯[-]", line, accent, line, line, accent, line, line, accent, line, accent)
}

// The decoration never blocks input. Ticker cancellation and queueing all
// UI mutations on the app loop avoid races when exiting or resizing.
func (u *UI) startWelcomeAnimation(ctx context.Context) {
	if u.welcomeGlow == nil || os.Getenv("WHATSCLI_REDUCED_MOTION") == "1" {
		return
	}
	go func() {
		ticker := time.NewTicker(240 * time.Millisecond)
		defer ticker.Stop()
		step := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				step++
				frame := welcomeLogoFrame(step)
				u.app.QueueUpdateDraw(func() {
					if u.authScreen == "welcome" {
						u.welcomeGlow.SetText(frame)
					}
				})
			}
		}
	}()
}

func (u *UI) showAgreement() {
	u.authScreen = "agreement"
	u.pages.SwitchToPage("agreement")
	u.app.SetFocus(u.agreementCheck)
}

func (u *UI) acceptAgreement() {
	if !u.agreementCheck.IsChecked() {
		u.agreementStatus.SetText("[yellow]Please read and tick the agreement before continuing.[-]")
		u.app.SetFocus(u.agreementCheck)
		return
	}
	u.agreed = true // Explicit action; checking the box alone is not consent.
	u.openLogin()
}

func (u *UI) showWelcome() {
	u.authScreen = "welcome"
	u.pages.SwitchToPage("welcome")
	u.app.SetFocus(u.welcomeButton)
}

func (u *UI) openLogin() {
	if !u.agreed {
		u.showAgreement()
		return
	}
	u.authScreen = "login"
	u.pages.SwitchToPage("login")
	u.app.SetFocus(u.loginExpand)
	if u.client != nil && !u.pairingStarted && u.runContext != nil {
		u.pairingStarted = true
		u.loginState.SetText("[#25D366]Connecting to WhatsApp…[-]")
		u.client.Start(u.runContext, func(evt wa.Event) {
			u.app.QueueUpdateDraw(func() { u.handle(evt) })
		})
	}
}

func (u *UI) showFullscreenQR() {
	if u.qrText == "" {
		u.loginState.SetText("[yellow]Waiting for a QR from WhatsApp; keep this screen open.[-]")
		return
	}
	u.authScreen = "fullqr"
	u.fullQR.SetText(u.qrText)
	u.pages.SwitchToPage("fullqr")
	u.app.SetFocus(u.fullQR)
}

func (u *UI) showChats() {
	u.authScreen = ""
	u.pages.SwitchToPage("main")
	u.app.SetFocus(u.input)
	u.refresh()
}

func (u *UI) authKey(event *tcell.EventKey) *tcell.EventKey {
	if u.authScreen == "" {
		return event
	}
	switch event.Key() {
	case tcell.KeyEscape:
		switch u.authScreen {
		case "fullqr":
			u.authScreen = "login"
			u.pages.SwitchToPage("login")
			u.app.SetFocus(u.loginExpand)
		case "login":
			u.showAgreement()
		case "agreement":
			u.showWelcome()
		case "welcome":
			u.app.Stop()
		}
		return nil
	case tcell.KeyCtrlC:
		u.app.Stop()
		return nil
	case tcell.KeyTAB, tcell.KeyBacktab:
		switch u.authScreen {
		case "welcome":
			if u.app.GetFocus() == u.welcomeButton {
				u.app.SetFocus(u.welcomeQuit)
			} else {
				u.app.SetFocus(u.welcomeButton)
			}
		case "agreement":
			focused := u.app.GetFocus()
			if event.Key() == tcell.KeyBacktab {
				switch focused {
				case u.agreementBack:
					u.app.SetFocus(u.agreementNext)
				case u.agreementNext:
					u.app.SetFocus(u.agreementCheck)
				default:
					u.app.SetFocus(u.agreementBack)
				}
			} else {
				switch focused {
				case u.agreementCheck:
					u.app.SetFocus(u.agreementNext)
				case u.agreementNext:
					u.app.SetFocus(u.agreementBack)
				default:
					u.app.SetFocus(u.agreementCheck)
				}
			}
		case "login":
			if u.app.GetFocus() == u.loginExpand {
				u.app.SetFocus(u.loginBack)
			} else {
				u.app.SetFocus(u.loginExpand)
			}
		}
		return nil
	case tcell.KeyRune:
		if event.Rune() == 'q' || event.Rune() == 'Q' {
			if u.authScreen == "login" {
				u.showFullscreenQR()
				return nil
			}
		}
	}
	return event
}

func (u *UI) handleQR(payload string) error {
	qr, width, height, err := renderQR(payload)
	if err != nil {
		return err
	}
	u.qrText = qr
	u.qrWidth = width
	u.qrHeight = height
	if u.loginQR != nil {
		u.loginQR.SetBackgroundColor(tcell.ColorWhite)
		u.loginQR.SetTextColor(tcell.ColorBlack)
		u.loginQR.SetText("\n" + qr)
		u.authLayout(u.width, u.height)
	}
	if u.fullQR != nil {
		u.fullQR.SetText(qr)
	}
	if u.loginState != nil {
		u.loginState.SetText("[#25D366]QR ready · Open WhatsApp → Linked devices → Link a device[-]")
	}
	return nil
}

// authConnected is deliberately triggered only by a confirmed Connected event.
func (u *UI) authConnected() { u.showChats() }

// PreviewOnboarding makes the login experience testable using fictional demo
// state without creating WhatsApp credentials or requesting a real QR token.
func (u *UI) PreviewOnboarding() {
	if u.authScreen == "" {
		u.setupAuthentication()
	}
}
