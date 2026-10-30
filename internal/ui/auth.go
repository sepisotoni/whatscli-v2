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

// setupAuthentication builds the three first-run pages used by an unpaired
// WhatsCLI install: Welcome -> Disclaimer -> Link device.
func (u *UI) setupAuthentication() {
	u.welcomeButton = tview.NewButton("  Get started  →  ")
	u.welcomeButton.SetBackgroundColor(green)
	u.welcomeButton.SetLabelColor(darker)
	u.welcomeButton.SetSelectedFunc(func() { u.showAgreement() })

	u.welcomeQuit = tview.NewButton("  Exit  ")
	u.welcomeQuit.SetBackgroundColor(slate)
	u.welcomeQuit.SetLabelColor(tcell.ColorWhite)
	u.welcomeQuit.SetSelectedFunc(func() { u.app.Stop() })

	// Keep this page deliberately sparse to mirror the supplied reference:
	// animated mark, title, subtitle, divider, and two actions.
	u.welcomeGlow = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	u.welcomeGlow.SetBackgroundColor(slate)
	u.welcomeGlow.SetText(welcomeLogoFrame(0))

	welcomeTitle := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	welcomeTitle.SetBackgroundColor(slate)
	welcomeTitle.SetText("[#E9EDEF::b]Welcome to [#00E99A::b]WhatsCLI[-]\n\n" +
		"[#B5C3C8]Your conversations. Your terminal.[-]\n\n" +
		"[#1F5A49]────────────────────────────────────────────[-]")

	welcomeActions := tview.NewFlex().SetDirection(tview.FlexColumn)
	welcomeActions.SetBackgroundColor(slate)
	welcomeActions.AddItem(nil, 0, 1, false).
		AddItem(u.welcomeButton, 26, 0, true).
		AddItem(nil, 3, 0, false).
		AddItem(u.welcomeQuit, 17, 0, false).
		AddItem(nil, 0, 1, false)

	welcomeCard := tview.NewFlex().SetDirection(tview.FlexRow)
	welcomeCard.SetBackgroundColor(slate)
	welcomeCard.AddItem(nil, 2, 0, false).
		AddItem(u.welcomeGlow, 7, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(welcomeTitle, 7, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(welcomeActions, 3, 0, true).
		AddItem(nil, 2, 0, false)
	welcomeCard.SetBorder(true).SetBorderColor(tcell.NewRGBColor(0, 120, 84))
	u.welcomeCard = welcomeCard

	u.welcomeOuter = tview.NewFlex().SetDirection(tview.FlexColumn)
	u.welcomeOuter.SetBackgroundColor(darker)
	u.welcomeOuter.AddItem(nil, 0, 1, false).
		AddItem(u.welcomeCard, 90, 0, true).
		AddItem(nil, 0, 1, false)

	u.welcomePage = tview.NewFlex().SetDirection(tview.FlexRow)
	u.welcomePage.SetBackgroundColor(darker)
	u.welcomePage.AddItem(nil, 0, 1, false).
		AddItem(u.welcomeOuter, 26, 0, true).
		AddItem(nil, 0, 1, false)
	u.pages.AddPage("welcome", u.welcomePage, true, false)

	// Disclaimer page: one card, four clearly separated items, one checkbox,
	// and two actions. Pairing cannot start without explicit consent.
	u.agreementText = tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	u.agreementText.SetBackgroundColor(slate)
	u.agreementText.SetText("[#718A85]BEFORE YOU CONNECT[-]                                   [#00E99A::b]STEP 01 / 02[-]\n" +
		"[#E9EDEF::b]Disclaimer & consent[-]\n" +
		"[#8FA7A2]A few important things to know about WhatsCLI before you continue.[-]\n\n" +
		"[#00E99A::b]01[-]  [#00E99A::b]Independent and unofficial[-]\n" +
		"    [#C5D0D2]WhatsCLI is an independent, unofficial client. It is not created, sponsored,[-]\n" +
		"    [#C5D0D2]or endorsed by WhatsApp or Meta. Their names and brands belong to them.[-]\n\n" +
		"[#00E99A::b]02[-]  [#00E99A::b]Your account, your responsibility[-]\n" +
		"    [#C5D0D2]Connecting is at your own risk. Unofficial clients can stop working and may[-]\n" +
		"    [#C5D0D2]carry account or service restrictions. Follow WhatsApp's terms.[-]\n\n" +
		"[#00E99A::b]03[-]  [#00E99A::b]Local data storage[-]\n" +
		"    [#C5D0D2]Linked-device credentials and cached chats are stored on this computer.[-]\n" +
		"    [#C5D0D2]WhatsCLI does not encrypt its local archive. Cached edits/deletions may remain.[-]\n\n" +
		"[#00E99A::b]04[-]  [#00E99A::b]Early software[-]\n" +
		"    [#C5D0D2]WhatsCLI is experimental. Features may change or be incomplete. Use a test[-]\n" +
		"    [#C5D0D2]account first and do not rely on it for critical communication.[-]\n\n" +
		"[#00E99A]▱ Read more  ·  WhatsApp Terms of Service ↗[-]")

	u.agreementCheck = tview.NewCheckbox().SetLabel(" I understand that this is an unofficial application and accept the risks above.")
	u.agreementCheck.SetLabelColor(tcell.ColorWhite)
	u.agreementCheck.SetChangedFunc(func(checked bool) {
		if checked {
			u.agreementNext.SetBackgroundColor(green)
			u.agreementNext.SetLabelColor(darker)
			u.agreementStatus.SetText("")
			return
		}
		u.agreementNext.SetBackgroundColor(tcell.NewRGBColor(45, 59, 62))
		u.agreementNext.SetLabelColor(tcell.ColorGray)
		u.agreementStatus.SetText("")
	})

	u.agreementNext = tview.NewButton("  Agree & continue  →  ")
	u.agreementNext.SetBackgroundColor(tcell.NewRGBColor(45, 59, 62))
	u.agreementNext.SetLabelColor(tcell.ColorGray)
	u.agreementNext.SetSelectedFunc(func() { u.acceptAgreement() })

	u.agreementBack = tview.NewButton("  ← Back  ")
	u.agreementBack.SetBackgroundColor(slate)
	u.agreementBack.SetLabelColor(tcell.ColorWhite)
	u.agreementBack.SetSelectedFunc(func() { u.showWelcome() })

	u.agreementStatus = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	u.agreementStatus.SetBackgroundColor(slate)

	agreementCheckRow := tview.NewFlex().SetDirection(tview.FlexColumn)
	agreementCheckRow.SetBackgroundColor(slate)
	agreementCheckRow.AddItem(nil, 1, 0, false).
		AddItem(u.agreementCheck, 0, 1, true).
		AddItem(nil, 1, 0, false)

	agreementButtons := tview.NewFlex().SetDirection(tview.FlexColumn)
	agreementButtons.SetBackgroundColor(slate)
	agreementButtons.AddItem(u.agreementBack, 16, 0, false).
		AddItem(nil, 0, 1, false).
		AddItem(u.agreementNext, 24, 0, true)

	u.agreementCard = tview.NewFlex().SetDirection(tview.FlexRow)
	u.agreementCard.SetBackgroundColor(slate)
	u.agreementCard.AddItem(u.agreementText, 0, 1, false).
		AddItem(agreementCheckRow, 3, 0, true).
		AddItem(u.agreementStatus, 1, 0, false).
		AddItem(agreementButtons, 3, 0, true)
	u.agreementCard.SetBorder(true).SetBorderColor(tcell.NewRGBColor(0, 120, 84))

	u.agreementOuter = tview.NewFlex().SetDirection(tview.FlexColumn)
	u.agreementOuter.SetBackgroundColor(darker)
	u.agreementOuter.AddItem(nil, 0, 1, false).
		AddItem(u.agreementCard, 92, 0, true).
		AddItem(nil, 0, 1, false)

	u.agreementPage = tview.NewFlex().SetDirection(tview.FlexRow)
	u.agreementPage.SetBackgroundColor(darker)
	u.agreementPage.AddItem(nil, 0, 1, false).
		AddItem(u.agreementOuter, 32, 0, true).
		AddItem(nil, 0, 1, false)
	u.pages.AddPage("agreement", u.agreementPage, true, false)

	// Pairing page: four instructions on the left, QR on the right. The QR has
	// no decorative border and there is intentionally no fake scan button.
	u.loginInstructions = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignLeft)
	u.loginInstructions.SetBackgroundColor(slate)
	u.loginInstructions.SetText("[#00E99A::b]◇  LINK A DEVICE[-]\n" +
		"[#C5D0D2]   Log in to WhatsApp[-]\n\n" +
		"[#00E99A::b]01[-]  [#E9EDEF]On your phone, open WhatsApp.[-]\n\n" +
		"[#00E99A::b]02[-]  [#E9EDEF]Go to Settings or ⋮ → Linked devices.[-]\n\n" +
		"[#00E99A::b]03[-]  [#E9EDEF]Select Link a device and scan the QR.[-]\n\n" +
		"[#00E99A::b]04[-]  [#E9EDEF]Keep your terminal open until linked.[-]\n\n" +
		"[#78948F]ⓘ  The code refreshes automatically. Never share it.[-]")

	u.loginQR = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	u.loginQR.SetBackgroundColor(slate)
	u.loginQR.SetTextColor(tcell.ColorWhite)
	u.loginQR.SetText("[#78948F]Preparing QR code…[-]")
	u.loginQR.SetBorder(false)

	u.loginState = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	u.loginState.SetBackgroundColor(slate)
	u.loginState.SetText("[#78948F]Waiting for device pairing…[-]")

	u.loginSpacer = tview.NewBox()
	u.loginColumns = tview.NewFlex().SetDirection(tview.FlexColumn)
	u.loginColumns.SetBackgroundColor(slate)
	u.loginColumns.AddItem(u.loginInstructions, 0, 1, false).
		AddItem(u.loginSpacer, 3, 0, false).
		AddItem(u.loginQR, 49, 0, false)

	u.loginBack = tview.NewButton("  ← Back  ")
	u.loginBack.SetBackgroundColor(slate)
	u.loginBack.SetLabelColor(tcell.ColorWhite)
	u.loginBack.SetSelectedFunc(func() { u.showAgreement() })

	u.loginExpand = tview.NewButton("  ⤢ Full-screen QR  ")
	u.loginExpand.SetBackgroundColor(green)
	u.loginExpand.SetLabelColor(darker)
	u.loginExpand.SetSelectedFunc(func() { u.showFullscreenQR() })

	loginButtons := tview.NewFlex().SetDirection(tview.FlexColumn)
	loginButtons.SetBackgroundColor(slate)
	loginButtons.AddItem(u.loginBack, 17, 0, true).
		AddItem(nil, 0, 1, false).
		AddItem(u.loginExpand, 24, 0, false)

	u.loginCard = tview.NewFlex().SetDirection(tview.FlexRow)
	u.loginCard.SetBackgroundColor(slate)
	u.loginCard.AddItem(u.loginColumns, 0, 1, false).
		AddItem(u.loginState, 2, 0, false).
		AddItem(loginButtons, 3, 0, true)
	u.loginCard.SetBorder(true).SetBorderColor(tcell.NewRGBColor(0, 120, 84))

	u.loginOuter = tview.NewFlex().SetDirection(tview.FlexColumn)
	u.loginOuter.SetBackgroundColor(darker)
	u.loginOuter.AddItem(nil, 0, 1, false).
		AddItem(u.loginCard, 100, 0, true).
		AddItem(nil, 0, 1, false)

	u.loginPage = tview.NewFlex().SetDirection(tview.FlexRow)
	u.loginPage.SetBackgroundColor(darker)
	u.loginPage.AddItem(nil, 0, 1, false).
		AddItem(u.loginOuter, 32, 0, true).
		AddItem(nil, 0, 1, false)
	u.pages.AddPage("login", u.loginPage, true, false)

	u.fullQR = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	u.fullQR.SetBackgroundColor(tcell.ColorWhite)
	u.fullQR.SetTextColor(tcell.ColorBlack)
	u.pages.AddPage("fullqr", u.fullQR, true, false)

	u.authScreen = "welcome"
	u.pages.SwitchToPage("welcome")
	u.app.SetFocus(u.welcomeButton)
}

func (u *UI) authLayout(width, height int) {
	if u.welcomeOuter == nil || u.agreementOuter == nil || u.loginOuter == nil {
		return
	}

	clampWidth := func(desired int) int {
		if width <= 6 {
			return 1
		}
		if desired > width-4 {
			return width - 4
		}
		return desired
	}
	clampHeight := func(desired int) int {
		if height <= 6 {
			return 1
		}
		if desired > height-4 {
			return height - 4
		}
		return desired
	}

	u.welcomeOuter.ResizeItem(u.welcomeCard, clampWidth(90), 0)
	u.welcomePage.ResizeItem(u.welcomeOuter, clampHeight(26), 0)
	u.agreementOuter.ResizeItem(u.agreementCard, clampWidth(92), 0)
	u.agreementPage.ResizeItem(u.agreementOuter, clampHeight(32), 0)
	u.loginOuter.ResizeItem(u.loginCard, clampWidth(100), 0)
	u.loginPage.ResizeItem(u.loginOuter, clampHeight(32), 0)

	qrColumnWidth := 49
	if u.qrWidth > 0 {
		qrColumnWidth = u.qrWidth + 4
	}
	canSplit := width >= qrColumnWidth+39 && height >= 25
	if canSplit {
		u.loginColumns.SetDirection(tview.FlexColumn)
		u.loginColumns.ResizeItem(u.loginInstructions, 0, 1)
		u.loginColumns.ResizeItem(u.loginSpacer, 3, 0)
		u.loginColumns.ResizeItem(u.loginQR, qrColumnWidth, 0)
	} else {
		u.loginColumns.SetDirection(tview.FlexColumn)
		u.loginColumns.ResizeItem(u.loginInstructions, 0, 1)
		u.loginColumns.ResizeItem(u.loginSpacer, 1, 0)
		u.loginColumns.ResizeItem(u.loginQR, 29, 0)
	}

	if u.qrText == "" {
		return
	}
	if !canSplit {
		u.loginQR.SetBackgroundColor(slate)
		u.loginQR.SetTextColor(tcell.ColorWhite)
		u.loginQR.SetText("[#78948F]QR ready — enlarge the terminal or press Q for full-screen QR.[-]")
		u.loginState.SetText("[yellow]Terminal is too small for the native QR. Press Q for full-screen QR.[-]")
		return
	}
	u.loginQR.SetBackgroundColor(tcell.ColorWhite)
	u.loginQR.SetTextColor(tcell.ColorBlack)
	u.loginQR.SetText("\n" + u.qrText)
}

func welcomeLogoFrame(step int) string {
	shades := [...]string{"#0A744E", "#0C9B64", "#00C980", "#00F0A0", "#00C980", "#0C9B64"}
	dots := [...]string{"·  ·  •  ·  ·", "·  •  ·  ·  ·", "•  ·  ·  ·  •", "·  ·  ·  •  ·", "·  ·  •  ·  ·", "·  •  ·  ·  ·"}
	accent := shades[step%len(shades)]
	rail := dots[step%len(dots)]
	return fmt.Sprintf("\n[#17473B]%s[-]     [%s]╭───────╮[-]     [#17473B]%s[-]\n"+
		"[#17473B]────────────[-]     [%s]│  ❯_   │[-]     [#17473B]────────────[-]\n"+
		"                    [%s]╰───╮   │[-]\n"+
		"                        [%s]╰───╯[-]", rail, accent, rail, accent, accent, accent)
}

func (u *UI) startWelcomeAnimation(ctx context.Context) {
	if u.welcomeGlow == nil || os.Getenv("WHATSCLI_REDUCED_MOTION") == "1" {
		return
	}
	go func() {
		ticker := time.NewTicker(260 * time.Millisecond)
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
		u.agreementStatus.SetText("[yellow]Please tick the acknowledgement before continuing.[-]")
		u.app.SetFocus(u.agreementCheck)
		return
	}
	u.agreed = true
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
		u.loginState.SetText("[#00E99A]Connecting to WhatsApp…[-]")
		u.client.Start(u.runContext, func(evt wa.Event) {
			u.app.QueueUpdateDraw(func() { u.handle(evt) })
		})
	}
}

func (u *UI) showFullscreenQR() {
	if u.qrText == "" {
		u.loginState.SetText("[yellow]Waiting for a QR from WhatsApp; keep this page open.[-]")
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
		if (event.Rune() == 'q' || event.Rune() == 'Q') && u.authScreen == "login" {
			u.showFullscreenQR()
			return nil
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
	if u.fullQR != nil {
		u.fullQR.SetText(qr)
	}
	if u.loginState != nil {
		u.loginState.SetText("[#00E99A]QR ready · WhatsApp → Linked devices → Link a device[-]")
	}
	if u.width > 0 && u.height > 0 {
		u.authLayout(u.width, u.height)
	} else if u.loginQR != nil {
		u.loginQR.SetBackgroundColor(tcell.ColorWhite)
		u.loginQR.SetTextColor(tcell.ColorBlack)
		u.loginQR.SetText("\n" + qr)
	}
	return nil
}

func (u *UI) authConnected() { u.showChats() }

func (u *UI) PreviewOnboarding() {
	if u.authScreen == "" {
		u.setupAuthentication()
	}
}
