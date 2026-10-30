package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sepisotoni/whatshell-go/internal/wa"
)

// The onboarding screen is a real terminal UI, not a QR image printed into a log.
// The QR is only shown on a requested login flow, and the client never stores it.
func (u *UI) setupAuthentication() {
	u.welcomeButton = tview.NewButton("  Log in to WhatsApp  →  ")
	u.welcomeButton.SetBackgroundColor(green)
	u.welcomeButton.SetLabelColor(darker)
	u.welcomeButton.SetSelectedFunc(func() { u.openLogin() })
	u.welcomeQuit = tview.NewButton("  Exit  ")
	u.welcomeQuit.SetBackgroundColor(slate)
	u.welcomeQuit.SetLabelColor(tcell.ColorWhite)
	u.welcomeQuit.SetSelectedFunc(func() { u.app.Stop() })

	brand := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	brand.SetBackgroundColor(darker)
	brand.SetText("\n[#25D366::b]◉  WhatsCLI[-]  [#8696A0] /  TERMINAL MESSENGER[-]")

	hero := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	hero.SetBackgroundColor(slate)
	hero.SetText("\n\n[#25D366::b]◉[-]\n\n[white::b]Your conversations. Your terminal.[-]\n\n[#B5C3C8]Your chats in one beautifully simple workspace.\nSecurely link WhatsApp on your phone, then stay in control\nfrom your keyboard or mouse — even over SSH.[-]\n\n[#8696A0]No browser required  •  Local message archive  •  Offline access[-]")

	actions := tview.NewFlex().SetDirection(tview.FlexColumn)
	actions.AddItem(nil, 0, 1, false).
		AddItem(u.welcomeButton, 30, 0, true).
		AddItem(nil, 2, 0, false).
		AddItem(u.welcomeQuit, 12, 0, false).
		AddItem(nil, 0, 1, false)
	details := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	details.SetBackgroundColor(slate)
	details.SetText("[#8696A0]NEW DEVICE?  Scan a QR code to link your account.\nDEMO MODE?  Run whatscli --demo to explore safely without signing in.[-]")

	card := tview.NewFlex().SetDirection(tview.FlexRow)
	card.AddItem(hero, 0, 1, false).
		AddItem(actions, 3, 0, true).
		AddItem(details, 3, 0, false)
	card.SetBackgroundColor(slate)
	card.SetBorder(true).SetBorderColor(tcell.NewRGBColor(50, 61, 65))
	card.SetTitle("  Welcome to WhatsCLI  ").SetTitleColor(green)

	u.welcomeCard = card
	u.welcomeOuter = tview.NewFlex().SetDirection(tview.FlexColumn)
	u.welcomeOuter.SetBackgroundColor(darker)
	u.welcomeOuter.AddItem(nil, 0, 1, false).
		AddItem(card, 84, 0, true).
		AddItem(nil, 0, 1, false)
	welcome := tview.NewFlex().SetDirection(tview.FlexRow)
	welcome.SetBackgroundColor(darker)
	welcome.AddItem(brand, 3, 0, false).
		AddItem(nil, 0, 1, false).
		AddItem(u.welcomeOuter, 23, 0, true).
		AddItem(nil, 0, 1, false)
	u.pages.AddPage("welcome", welcome, true, false)

	heading := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignLeft)
	heading.SetBackgroundColor(slate)
	heading.SetText("[#25D366::b]LINK A DEVICE[-]\n\n[white::b]Log in to WhatsApp[-]\n\n[#B5C3C8]1  On your phone, open WhatsApp.\n\n2  Go to Settings or ⋮ Menu → Linked devices.\n\n3  Select Link a device and scan the QR.\n\n4  Keep your terminal open until linked.[-]\n\n[#8696A0]The code refreshes automatically. Never share it.\nPhone-number pairing is not supported yet.[-]")
	u.loginQR = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	u.loginQR.SetBackgroundColor(tcell.ColorWhite)
	u.loginQR.SetTextColor(tcell.ColorBlack)
	u.loginQR.SetText("\n\n\n\n\n    Preparing secure QR code…    ")
	u.loginQR.SetBorder(true).SetBorderColor(tcell.NewRGBColor(50, 61, 65))
	u.loginState = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	u.loginState.SetBackgroundColor(slate)
	u.loginState.SetText("[#8696A0]Waiting for device pairing…[-]")
	u.loginInstructions = heading
	u.loginSpacer = tview.NewBox()
	u.loginColumns = tview.NewFlex().SetDirection(tview.FlexColumn)
	u.loginColumns.AddItem(heading, 0, 1, false).
		AddItem(u.loginSpacer, 2, 0, false).
		AddItem(u.loginQR, 0, 1, false)
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
	loginCard.SetTitle("  WhatsCLI  /  Link device  ").SetTitleColor(green)
	u.loginCard = loginCard
	u.loginOuter = tview.NewFlex().SetDirection(tview.FlexColumn)
	u.loginOuter.SetBackgroundColor(darker)
	u.loginOuter.AddItem(nil, 0, 1, false).
		AddItem(loginCard, 106, 0, true).
		AddItem(nil, 0, 1, false)
	login := tview.NewFlex().SetDirection(tview.FlexRow)
	login.SetBackgroundColor(darker)
	login.AddItem(brand, 3, 0, false).
		AddItem(nil, 0, 1, false).
		AddItem(u.loginOuter, 37, 0, true).
		AddItem(nil, 0, 1, false)
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
	if u.welcomeOuter == nil || u.loginOuter == nil {
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
	u.welcomeOuter.ResizeItem(u.welcomeCard, max(84, width-4), 0)
	u.loginOuter.ResizeItem(u.loginCard, max(106, width-4), 0)
	// The QR needs its original module size to remain reliably scannable.
	// Keep the instructions beside it on wider screens; on small screens use
	// the dedicated full-screen QR action rather than scaling QR modules.
	if width < 100 {
		u.loginColumns.SetDirection(tview.FlexRow)
		u.loginColumns.ResizeItem(u.loginSpacer, 0, 0)
		u.loginColumns.ResizeItem(u.loginInstructions, 12, 0)
		u.loginColumns.ResizeItem(u.loginQR, 0, 1)
	} else {
		u.loginColumns.SetDirection(tview.FlexColumn)
		u.loginColumns.ResizeItem(u.loginInstructions, 0, 1)
		u.loginColumns.ResizeItem(u.loginSpacer, 2, 0)
		u.loginColumns.ResizeItem(u.loginQR, 0, 1)
	}
	if u.qrWidth > 0 && (width < u.qrWidth+4 || height < u.qrHeight+12) {
		u.loginQR.SetText("\n\n\n[#17262B]QR ready. Press Q or select Full-screen QR to scan.[-]")
		u.loginState.SetText(fmt.Sprintf("[yellow]Terminal too small for QR (%dx%d); enlarge or press Q for full screen.[-]", u.qrWidth+4, u.qrHeight+12))
	} else if u.qrText != "" {
		u.loginQR.SetText(u.qrText)
	}
}

func (u *UI) showWelcome() {
	u.authScreen = "welcome"
	u.pages.SwitchToPage("welcome")
	u.app.SetFocus(u.welcomeButton)
}

func (u *UI) openLogin() {
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
		u.loginQR.SetText(qr)
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
