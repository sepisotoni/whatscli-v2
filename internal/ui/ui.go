// Package ui contains the SSH-compatible terminal interface.
package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sepisotoni/whatshell-go/internal/config"
	"github.com/sepisotoni/whatshell-go/internal/store"
	"github.com/sepisotoni/whatshell-go/internal/wa"
	"github.com/skip2/go-qrcode"
)

var (
	green  = tcell.NewRGBColor(37, 211, 102)
	teal   = tcell.NewRGBColor(18, 140, 126)
	slate  = tcell.NewRGBColor(17, 27, 33)
	darker = tcell.NewRGBColor(11, 20, 26)
)

var commands = []struct{ Name, Description string }{
	{"/help", "keyboard shortcuts & commands"},
	{"/chat", "open a phone number or full chat JID"},
	{"/versions", "view cached edits: /versions MESSAGE_ID"},
	{"/search", "search cached messages"},
	{"/chats", "choose a chat on compact terminals"},
	{"/alias", "set or remove a local chat name: /alias NAME or -"},
	{"/attach", "send an image or document from uploads/"},
	{"/download", "download media using its message ID"},
	{"/backup", "create a consistent private history snapshot"},
	{"/poll", "create poll: /poll Question | Yes | No"},
	{"/logout", "unlink and remove local session tokens"},
	{"/quit", "close the terminal client"},
}

type UI struct {
	app         *tview.Application
	pages       *tview.Pages
	body        *tview.Flex
	chatList    *tview.List
	messages    *tview.TextView
	input       *tview.InputField
	palette     *tview.List
	status      *tview.TextView
	main        *tview.Flex
	archive     *store.DB
	paths       config.Paths
	client      *wa.Client
	chats       []store.Chat
	chat        string
	busy        bool
	showSidebar bool
	width       int
	height      int
}

func New(archive *store.DB, client *wa.Client, paths config.Paths) *UI {
	u := &UI{app: tview.NewApplication(), archive: archive, client: client, paths: paths}
	u.chatList = tview.NewList().ShowSecondaryText(true)
	u.chatList.SetBorder(true).SetTitle(" Chats ").SetBorderColor(teal)
	u.chatList.SetMainTextColor(tcell.ColorWhite).SetSecondaryTextColor(tcell.ColorGray)
	u.chatList.SetSelectedFunc(func(index int, _, _ string, _ rune) { u.selectChat(index) })
	u.messages = tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	u.messages.SetBorder(true).SetTitle(" Messages ").SetBorderColor(green)
	u.messages.SetBackgroundColor(slate)
	u.input = tview.NewInputField().SetLabel(" ❯ ").SetFieldWidth(0)
	u.input.SetLabelColor(green).SetFieldBackgroundColor(darker).SetFieldTextColor(tcell.ColorWhite)
	u.input.SetBorder(true).SetTitle(" Compose ")
	u.input.SetChangedFunc(func(s string) { u.updatePalette(s) })
	u.input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			u.submit()
		}
	})
	u.palette = tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true)
	u.palette.SetBackgroundColor(darker)
	u.palette.SetSelectedFunc(func(index int, _, _ string, _ rune) {
		if index >= 0 && index < u.palette.GetItemCount() {
			cmd, _ := u.palette.GetItemText(index)
			u.input.SetText(strings.Fields(cmd)[0] + " ")
			u.app.SetFocus(u.input)
		}
	})
	u.status = tview.NewTextView().SetDynamicColors(true)
	u.status.SetBackgroundColor(darker)
	u.status.SetTextColor(tcell.ColorWhite)
	u.status.SetText(" [green]●[-] Whatshell  •  connecting…  | Tab: focus  Ctrl+N: new chat  /help: commands ")
	u.body = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(u.chatList, 27, 0, false).AddItem(u.messages, 0, 1, false)
	bar := tview.NewTextView().SetDynamicColors(true).SetText(" [green::b] WHATSHELL [-]  [gray]Go • SSH-first • Local archive[-] ")
	bar.SetBackgroundColor(darker)
	u.main = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(bar, 1, 0, false).
		AddItem(u.body, 0, 1, false).
		AddItem(u.palette, 0, 0, false).
		AddItem(u.input, 3, 0, true).
		AddItem(u.status, 1, 0, false)
	u.pages = tview.NewPages().AddPage("main", u.main, true, true)
	u.app.SetRoot(u.pages, true).EnableMouse(true).EnablePaste(true).SetFocus(u.input)
	u.app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		w, h := screen.Size()
		u.width = w
		u.height = h
		side := 0
		switch {
		case w >= 82:
			side = 27
		case w >= 60:
			side = 20
		case w >= 38 && u.showSidebar:
			side = 16
		}
		u.body.ResizeItem(u.chatList, side, 0)
		return false
	})
	u.app.SetInputCapture(u.onKey)
	u.refresh()
	return u
}

func (u *UI) onKey(key *tcell.EventKey) *tcell.EventKey {
	switch key.Key() {
	case tcell.KeyCtrlC:
		u.app.Stop()
		return nil
	case tcell.KeyCtrlL:
		u.showSidebar = !u.showSidebar
		if u.showSidebar {
			u.app.SetFocus(u.chatList)
		} else {
			u.app.SetFocus(u.input)
		}
		return nil
	case tcell.KeyCtrlF:
		u.promptSearch()
		return nil
	case tcell.KeyCtrlN:
		u.promptChat()
		return nil
	case tcell.KeyTAB:
		switch u.app.GetFocus() {
		case u.input, u.chatList, u.messages:
		default:
			return key // Preserve Tab handling in modals.
		}
		if u.app.GetFocus() == u.input && strings.HasPrefix(u.input.GetText(), "/") {
			if u.completeCommand() {
				return nil
			}
		}
		switch u.app.GetFocus() {
		case u.input:
			u.app.SetFocus(u.chatList)
		case u.chatList:
			u.app.SetFocus(u.messages)
		default:
			u.app.SetFocus(u.input)
		}
		return nil
	}
	return key
}

func (u *UI) updatePalette(text string) {
	u.palette.Clear()
	if !strings.HasPrefix(text, "/") || strings.Contains(text, " ") {
		u.main.ResizeItem(u.palette, 0, 0)
		return
	}
	for _, c := range commands {
		if strings.HasPrefix(c.Name, strings.ToLower(text)) {
			u.palette.AddItem(c.Name+"   "+c.Description, "", 0, nil)
		}
	}
	n := u.palette.GetItemCount()
	if n > 4 {
		n = 4
	}
	u.main.ResizeItem(u.palette, n, 0)
}

func (u *UI) completeCommand() bool {
	if u.palette.GetItemCount() == 0 {
		return false
	}
	name, _ := u.palette.GetItemText(0)
	u.input.SetText(strings.Fields(name)[0] + " ")
	return true
}

func (u *UI) selectChat(i int) {
	if i < 0 || i >= len(u.chats) {
		return
	}
	u.chat = u.chats[i].JID
	u.refreshMessages()
	u.app.SetFocus(u.input)
}

func (u *UI) openChat(id string) {
	id = strings.TrimSpace(id)
	if id == "" {
		u.notice("Enter a phone number or JID")
		return
	}
	if !strings.Contains(id, "@") {
		id = strings.TrimPrefix(id, "+")
		for _, r := range id {
			if r < '0' || r > '9' {
				u.notice("Phone numbers require country code and digits only")
				return
			}
		}
		id += "@s.whatsapp.net"
	}
	u.chat = id
	u.refreshMessages()
	u.app.SetFocus(u.input)
}

func (u *UI) submit() {
	if u.busy {
		return
	}
	text := strings.TrimSpace(u.input.GetText())
	if text == "" {
		return
	}
	if strings.HasPrefix(text, "/") {
		if !strings.Contains(text, " ") && u.completeCommand() && text != strings.TrimSpace(u.input.GetText()) {
			return
		}
		u.command(text)
		return
	}
	if u.chat == "" {
		u.notice("Choose a chat first (Ctrl+N or /chat PHONE)")
		return
	}
	u.busy = true
	u.notice("Sending…")
	chat := u.chat
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		_, err := u.client.Send(ctx, chat, text)
		u.app.QueueUpdateDraw(func() {
			u.busy = false
			if err != nil {
				u.notice("Send failed: " + err.Error())
				return
			}
			if u.input.GetText() == text {
				u.input.SetText("")
			}
			u.refresh()
			u.notice("✓ Sent and archived")
		})
	}()
}

func (u *UI) command(raw string) {
	parts := strings.SplitN(raw, " ", 2)
	cmd, arg := parts[0], ""
	if len(parts) == 2 {
		arg = strings.TrimSpace(parts[1])
	}
	switch cmd {
	case "/help":
		u.showHelp()
	case "/chat":
		u.openChat(arg)
	case "/chats":
		u.chooseChat()
	case "/alias":
		if u.chat == "" || arg == "" {
			u.notice("Open a chat, then /alias DISPLAY NAME (or /alias - to remove)")
			break
		}
		if arg == "-" {
			arg = ""
		}
		if err := u.archive.SetAlias(context.Background(), u.chat, arg); err != nil {
			u.notice("Alias: " + err.Error())
		} else {
			u.refresh()
			u.notice("Local chat name saved; no WhatsApp contact changed")
		}
	case "/search":
		if arg == "" {
			u.promptSearch()
		} else {
			u.search(arg)
		}
	case "/attach":
		if u.chat == "" || arg == "" {
			u.notice("Use /attach NAME in a chat; upload file via SFTP first")
			break
		}
		chat := u.chat
		u.asyncAction("Uploading attachment", func(ctx context.Context) (string, error) {
			_, err := u.client.SendFile(ctx, chat, u.paths.Uploads, arg, "")
			return "", err
		})
	case "/download":
		if u.chat == "" || arg == "" {
			u.notice("Use /download MESSAGE_ID inside a chat")
			break
		}
		chat := u.chat
		u.asyncAction("Downloading attachment", func(ctx context.Context) (string, error) {
			path, err := u.client.DownloadMedia(ctx, chat, arg, u.paths.Downloads)
			return "Downloaded to " + path, err
		})
	case "/backup":
		u.asyncAction("Backing up history", func(ctx context.Context) (string, error) {
			path, err := u.archive.Backup(ctx, u.paths.Dir+"/backups")
			return "History backup: " + path, err
		})
	case "/poll":
		if u.chat == "" {
			u.notice("Choose a chat first")
			break
		}
		parts := strings.Split(arg, "|")
		if len(parts) < 3 {
			u.notice("Use /poll Question | Option 1 | Option 2")
			break
		}
		chat := u.chat
		u.asyncAction("Sending poll", func(ctx context.Context) (string, error) {
			_, err := u.client.SendPoll(ctx, chat, parts[0], parts[1:])
			return "", err
		})
	case "/versions":
		if u.chat == "" || arg == "" {
			u.notice("Use /versions MESSAGE_ID inside a chat")
			break
		}
		versions, err := u.archive.Versions(context.Background(), u.chat, arg)
		if err != nil {
			u.notice(err.Error())
			break
		}
		if len(versions) == 0 {
			u.notice("No earlier cached versions found")
			break
		}
		var b strings.Builder
		for i, v := range versions {
			fmt.Fprintf(&b, "%d. %s: %s\n", i+1, v.ChangedAt.Format(time.RFC822), v.Body)
		}
		modal := tview.NewModal().SetText(b.String()).AddButtons([]string{"Close"}).SetDoneFunc(func(_ int, _ string) {
			u.pages.RemovePage("versions")
			u.app.SetFocus(u.input)
		})
		u.pages.AddPage("versions", modal, true, true)
	case "/logout":
		u.busy = true
		u.notice("Unlinking and wiping local session…")
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			err := u.client.Logout(ctx)
			u.app.QueueUpdateDraw(func() {
				if err != nil {
					u.notice("Local session removed; remote unlink warning: " + err.Error())
				}
				u.app.Stop()
			})
		}()
	case "/quit":
		u.app.Stop()
	default:
		u.notice("Unknown command: " + cmd)
	}
	if cmd != "/logout" {
		u.input.SetText("")
	}
}

// asyncAction runs blocking WhatsApp network operations outside the tview loop.
func (u *UI) asyncAction(label string, action func(context.Context) (string, error)) {
	if u.busy {
		u.notice("Another operation is in progress")
		return
	}
	u.busy = true
	u.notice(label + "…")
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		detail, err := action(ctx)
		u.app.QueueUpdateDraw(func() {
			u.busy = false
			if err != nil {
				u.notice(label + " failed: " + err.Error())
				return
			}
			u.refresh()
			if detail != "" {
				u.notice(detail)
			} else {
				u.notice(label + " complete")
			}
		})
	}()
}

func (u *UI) showHelp() {
	help := `Ctrl+N: New chat     Ctrl+F: Search offline messages
Ctrl+L: Toggle compact sidebar
Tab: Cycle focus    Mouse wheel/arrow keys: scroll
/chat PHONE         Open a person/group
/chats              Browse saved chat list
/alias NAME         Local-only chat name (/alias - clears)
/attach FILE        Send a server-side uploads/ file (SFTP first)
/download ID        Save a received attachment to downloads/
/backup             Snapshot archive (NOT login credentials)
/poll Q | A | B     Create a single-choice WhatsApp poll
/search TEXT        Search locally cached conversation bodies
/versions ID       View earlier archived edits
/logout             Unlink device and delete session secrets
/quit               Exit without unlinking`
	modal := tview.NewModal().SetText(help).AddButtons([]string{"Close"}).SetDoneFunc(func(_ int, _ string) { u.pages.RemovePage("help"); u.app.SetFocus(u.input) })
	u.pages.AddPage("help", modal, true, true)
}

func (u *UI) chooseChat() {
	list := tview.NewList().ShowSecondaryText(true)
	list.SetBorder(true).SetTitle(" Cached chats ")
	for _, c := range u.chats {
		c := c
		label := c.JID
		if c.Alias != "" {
			label = c.Alias + " (" + c.JID + ")"
		}
		list.AddItem(tview.Escape(label), tview.Escape(c.LastBody), 0, func() {
			u.pages.RemovePage("chatpicker")
			u.openChat(c.JID)
		})
	}
	list.AddItem("Close", "Esc or Enter", 0, func() { u.pages.RemovePage("chatpicker"); u.app.SetFocus(u.input) })
	list.SetInputCapture(func(k *tcell.EventKey) *tcell.EventKey {
		if k.Key() == tcell.KeyEscape {
			u.pages.RemovePage("chatpicker")
			u.app.SetFocus(u.input)
			return nil
		}
		return k
	})
	u.pages.AddPage("chatpicker", u.center(list, 80, 23), true, true)
	u.app.SetFocus(list)
}

func (u *UI) promptSearch() {
	field := tview.NewInputField().SetLabel("Find in local archive: ").SetFieldWidth(32)
	form := tview.NewForm().AddFormItem(field).
		AddButton("Search", func() { u.pages.RemovePage("searchform"); u.search(field.GetText()) }).
		AddButton("Cancel", func() { u.pages.RemovePage("searchform"); u.app.SetFocus(u.input) })
	form.SetBorder(true).SetTitle(" Search ")
	u.pages.AddPage("searchform", u.center(form, 66, 9), true, true)
	u.app.SetFocus(field)
}

func (u *UI) search(query string) {
	results, err := u.archive.Search(context.Background(), query, 60)
	if err != nil {
		u.notice(err.Error())
		return
	}
	if len(results) == 0 {
		u.notice("No cached messages match: " + query)
		return
	}
	list := tview.NewList().ShowSecondaryText(true)
	list.SetBorder(true).SetTitle(" Search: " + tview.Escape(query) + " (Enter to open) ")
	for _, r := range results {
		r := r
		preview := strings.ReplaceAll(r.Body, "\n", " ")
		if len([]rune(preview)) > 65 {
			preview = string([]rune(preview)[:65]) + "…"
		}
		if r.Deleted {
			preview += " [deleted]"
		}
		list.AddItem(tview.Escape(r.Chat), tview.Escape(preview), 0, func() {
			u.pages.RemovePage("searchresults")
			u.openChat(r.Chat)
			u.notice("Found message ID: " + r.ID + " • use /versions " + r.ID)
		})
	}
	list.AddItem("Close", "Esc or Enter", 0, func() { u.pages.RemovePage("searchresults"); u.app.SetFocus(u.input) })
	list.SetInputCapture(func(k *tcell.EventKey) *tcell.EventKey {
		if k.Key() == tcell.KeyEscape {
			u.pages.RemovePage("searchresults")
			u.app.SetFocus(u.input)
			return nil
		}
		return k
	})
	u.pages.AddPage("searchresults", u.center(list, 88, 22), true, true)
	u.app.SetFocus(list)
}

func (u *UI) promptChat() {
	field := tview.NewInputField().SetLabel("Number (+country code) or JID: ").SetFieldWidth(36)
	form := tview.NewForm().AddFormItem(field).
		AddButton("Open", func() { u.pages.RemovePage("newchat"); u.openChat(field.GetText()) }).
		AddButton("Cancel", func() { u.pages.RemovePage("newchat"); u.app.SetFocus(u.input) })
	form.SetBorder(true).SetTitle(" New chat ")
	u.pages.AddPage("newchat", u.center(form, 68, 10), true, true)
	u.app.SetFocus(field)
}

func (u *UI) center(view tview.Primitive, width, height int) tview.Primitive {
	if u.width > 8 && width > u.width-2 {
		width = u.width - 2
	}
	if u.height > 8 && height > u.height-2 {
		height = u.height - 2
	}
	return tview.NewFlex().AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).AddItem(view, height, 0, true).
			AddItem(nil, 0, 1, false), width, 0, true).
		AddItem(nil, 0, 1, false)
}

func (u *UI) notice(s string) {
	u.status.SetText(" [green]●[-] " + tview.Escape(s))
}

func (u *UI) refresh() {
	list, err := u.archive.Chats(context.Background())
	if err != nil {
		u.notice("Archive read failed: " + err.Error())
		return
	}
	u.chats = list
	u.chatList.Clear()
	selected := -1
	for i, c := range list {
		if c.JID == u.chat {
			selected = i
		}
		preview := c.LastBody
		if len([]rune(preview)) > 35 {
			preview = string([]rune(preview)[:35]) + "…"
		}
		label := c.Alias
		if label == "" {
			label = strings.Split(c.JID, "@")[0]
		}
		u.chatList.AddItem(tview.Escape(label), tview.Escape(preview), 0, nil)
	}
	if selected >= 0 {
		u.chatList.SetCurrentItem(selected)
	}
	u.refreshMessages()
}

func (u *UI) refreshMessages() {
	if u.chat == "" {
		u.messages.SetText("\n  [green]Welcome to Whatshell[-]\n\n  Linked devices sync live text messages.\n  Open a chat with Ctrl+N or /chat NUMBER.\n  Locally cached conversations remain readable while offline.")
		return
	}
	msg, err := u.archive.Messages(context.Background(), u.chat, 200)
	if err != nil {
		u.notice("Messages read failed: " + err.Error())
		return
	}
	var b strings.Builder
	fmt.Fprintf(&b, " [green]🔒[-] %s — cached messages\n\n", tview.Escape(u.chat))
	for _, m := range msg {
		who := "[teal]◀"
		if m.Outgoing {
			who = "[green]▶"
		}
		body := tview.Escape(strings.NewReplacer("\n", " ⏎ ", "\r", " ").Replace(m.Body))
		if m.Deleted {
			body += " [gray](deleted on WhatsApp)[-]"
		}
		if m.Outgoing {
			switch m.Receipt {
			case 2:
				body += " [blue]✓✓ read[-]"
			case 1:
				body += " [gray]✓✓ delivered[-]"
			default:
				body += " [gray]✓ sent[-]"
			}
		}
		if strings.HasPrefix(m.Body, "[image]") || strings.HasPrefix(m.Body, "[document]") ||
			strings.HasPrefix(m.Body, "[video]") || strings.HasPrefix(m.Body, "[audio") {
			body += " [gray]id=" + tview.Escape(m.ID) + "[-]"
		}
		if m.Edited {
			body += " [yellow](edited)[-] [gray]id=" + tview.Escape(m.ID) + "[-]"
		}
		fmt.Fprintf(&b, " %s[-] [gray]%s[-] %s\n", who, m.Timestamp.Format("02 Jan 15:04"), body)
	}
	if len(msg) == 0 {
		b.WriteString("  No messages cached here yet.\n")
	}
	title := u.chat
	if name, err := u.archive.Alias(context.Background(), u.chat); err == nil && name != "" {
		title = name + " (" + u.chat + ")"
	}
	u.messages.SetTitle(" " + tview.Escape(title) + " ")
	u.messages.SetText(b.String())
	u.messages.ScrollToEnd()
}

func renderQR(text string) (string, int, int, error) {
	qr, err := qrcode.New(text, qrcode.Medium)
	if err != nil {
		return "", 0, 0, err
	}
	bits := qr.Bitmap()
	var b strings.Builder
	b.WriteString("[black:white]")
	for y := 0; y < len(bits); y += 2 {
		for x := 0; x < len(bits[y]); x++ {
			top := bits[y][x]
			bottom := false
			if y+1 < len(bits) {
				bottom = bits[y+1][x]
			}
			switch {
			case top && bottom:
				b.WriteRune('█')
			case top:
				b.WriteRune('▀')
			case bottom:
				b.WriteRune('▄')
			default:
				b.WriteRune(' ')
			}
		}
		b.WriteString("\n")
	}
	return b.String(), len(bits[0]), (len(bits) + 1) / 2, nil
}

func (u *UI) handle(evt wa.Event) {
	switch evt.Kind {
	case wa.QR:
		qr, qrWidth, qrHeight, err := renderQR(evt.Text)
		if err != nil {
			u.notice("QR rendering failed: " + err.Error())
			return
		}
		view := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
		view.SetBorder(true).SetTitle(" Scan QR using WhatsApp > Linked Devices ")
		view.SetBackgroundColor(tcell.ColorWhite)
		view.SetText(qr)
		u.pages.RemovePage("qr")
		u.pages.AddPage("qr", u.center(view, qrWidth+4, qrHeight+4), true, true)
		if u.width < qrWidth+4 || u.height < qrHeight+4 {
			u.notice(fmt.Sprintf("Enlarge terminal to at least %dx%d to scan QR", qrWidth+4, qrHeight+4))
		} else {
			u.notice("Scan the pairing code on your phone")
		}
	case wa.Status:
		u.pages.RemovePage("qr")
		u.notice(evt.Text)
	case wa.Message:
		u.refresh()
		if evt.Chat == "" {
			u.notice(evt.Text)
		} else if evt.Chat != u.chat {
			u.notice("● New message in " + evt.Chat)
		}
	case wa.Failure:
		u.notice("Error: " + evt.Text)
	}
}

func (u *UI) Run(ctx context.Context) error {
	u.client.Start(ctx, func(evt wa.Event) {
		u.app.QueueUpdateDraw(func() { u.handle(evt) })
	})
	return u.app.Run()
}