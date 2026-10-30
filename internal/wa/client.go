// Package wa wraps the WhatsApp Multi-Device transport behind a small interface.
package wa

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sepisotoni/whatshell-go/internal/store"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

type EventKind int

const (
	Status EventKind = iota
	QR
	Message
	Failure
)

type Event struct {
	Kind EventKind
	Text string
	Chat string
}

type Client struct {
	api       *whatsmeow.Client
	device    *sqlstore.Container
	archive   *store.DB
	path      string
	emit      func(Event)
	closeOnce sync.Once
	closeErr  error
}

// New uses the pure-Go sqlite3 driver alias installed by the store package.
func New(ctx context.Context, sessionPath string, archive *store.DB) (*Client, error) {
	container, err := sqlstore.New(ctx, "sqlite3", store.DSN(sessionPath), nil)
	if err != nil {
		return nil, fmt.Errorf("open WhatsApp session: %w", err)
	}
	device, err := container.GetFirstDevice(ctx)
	if err != nil {
		_ = container.Close()
		return nil, fmt.Errorf("load WhatsApp device: %w", err)
	}
	c := &Client{
		api: whatsmeow.NewClient(device, nil), device: container,
		archive: archive, path: sessionPath,
	}
	// WhatsMeow owns reconnect/backoff; never blindly retry sent messages.
	c.api.InitialAutoReconnect = true
	c.api.AddEventHandler(c.onEvent)
	return c, nil
}

func (c *Client) notify(evt Event) {
	if c.emit != nil {
		c.emit(evt)
	}
}

func (c *Client) Start(ctx context.Context, emit func(Event)) {
	c.emit = emit
	go func() {
		if c.api.Store.ID == nil {
			codes, err := c.api.GetQRChannel(ctx)
			if err != nil {
				c.notify(Event{Failure, err.Error(), ""})
				return
			}
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					case item, ok := <-codes:
						if !ok {
							return
						}
						if item.Event == "code" {
							c.notify(Event{QR, item.Code, ""})
						} else {
							c.notify(Event{Status, "Pairing: " + item.Event, ""})
						}
					}
				}
			}()
		}
		if err := c.api.ConnectContext(ctx); err != nil {
			c.notify(Event{Failure, "connection failed: " + err.Error(), ""})
		}
	}()
}

func (c *Client) onEvent(raw any) {
	switch e := raw.(type) {
	case *events.Connected:
		c.notify(Event{Status, "Connected • local history enabled", ""})
	case *events.LoggedOut:
		c.notify(Event{Failure, "Session revoked. Restart to pair again.", ""})
	case *events.Disconnected:
		c.notify(Event{Status, "Disconnected • viewing offline cache", ""})
	case *events.Message:
		if err := c.archiveMessage(e); err != nil {
			c.notify(Event{Failure, "archive: " + err.Error(), e.Info.Chat.String()})
		} else {
			c.notify(Event{Message, "message archived", e.Info.Chat.String()})
		}
	case *events.HistorySync:
		if e.Data == nil {
			return
		}
		// WhatsMeow obtains the history blob and exposes individual web messages.
		count := 0
		for _, conv := range e.Data.GetConversations() {
			chat, err := types.ParseJID(conv.GetID())
			if err != nil {
				continue
			}
			for _, entry := range conv.GetMessages() {
				parsed, err := c.api.ParseWebMessage(chat, entry.GetMessage())
				if err != nil {
					continue
				}
				if err = c.archiveMessage(parsed); err == nil {
					count++
				}
			}
		}
		c.notify(Event{Message, fmt.Sprintf("Synced %d cached messages", count), ""})
	case *events.Receipt:
		level := 0
		switch e.Type {
		case types.ReceiptTypeRead, types.ReceiptTypeReadSelf, types.ReceiptTypePlayed:
			level = 2
		case types.ReceiptTypeDelivered, types.ReceiptTypeSender:
			level = 1
		}
		if level == 0 {
			return
		}
		for _, id := range e.MessageIDs {
			if err := c.archive.Receipt(context.Background(), e.Chat.String(), string(id), level); err != nil {
				c.notify(Event{Failure, "receipt persistence: " + err.Error(), e.Chat.String()})
				return
			}
		}
		c.notify(Event{Message, "receipts updated", e.Chat.String()})
	}
}

func (c *Client) archiveMessage(e *events.Message) error {
	if e == nil || e.Message == nil {
		return nil
	}
	chat := e.Info.Chat.String()
	if p := e.Message.GetProtocolMessage(); p != nil {
		key := p.GetKey()
		if key == nil || key.GetID() == "" {
			return nil
		}
		switch p.GetType() {
		case waE2E.ProtocolMessage_REVOKE:
			return c.archive.MarkDeleted(context.Background(), chat, key.GetID(), e.Info.Timestamp)
		case waE2E.ProtocolMessage_MESSAGE_EDIT:
			return c.archive.Edit(context.Background(), chat, key.GetID(), messageText(p.GetEditedMessage()), e.Info.Timestamp)
		}
		return nil
	}
	m := store.Message{Chat: chat, ID: string(e.Info.ID), Sender: e.Info.Sender.String(),
		Body: messageText(e.Message), Timestamp: e.Info.Timestamp, Outgoing: e.Info.IsFromMe}
	if m.ID == "" {
		return nil
	}
	if err := c.archive.Insert(context.Background(), m); err != nil {
		return err
	}
	// Only normal attachments, not wrapped view-once or ephemeral payloads.
	if e.Message.GetImageMessage() != nil || e.Message.GetDocumentMessage() != nil ||
		e.Message.GetAudioMessage() != nil || e.Message.GetVideoMessage() != nil {
		payload, err := proto.Marshal(e.Message)
		if err != nil {
			return err
		}
		filename := strings.TrimSpace(e.Message.GetDocumentMessage().GetFileName())
		if filename == "" {
			switch {
			case e.Message.GetImageMessage() != nil:
				filename = m.ID + ".jpg"
			case e.Message.GetVideoMessage() != nil:
				filename = m.ID + ".mp4"
			case e.Message.GetAudioMessage() != nil:
				filename = m.ID + ".ogg"
			default:
				filename = m.ID + ".bin"
			}
		}
		if err := c.archive.SaveMedia(context.Background(), chat, m.ID, filename, payload); err != nil {
			return err
		}
	}
	return nil
}

func messageText(m *waE2E.Message) string {
	if m == nil {
		return "[empty message]"
	}
	if s := m.GetConversation(); s != "" {
		return s
	}
	if s := m.GetExtendedTextMessage().GetText(); s != "" {
		return s
	}
	if m.GetImageMessage() != nil {
		return "[image] " + m.GetImageMessage().GetCaption()
	}
	if m.GetVideoMessage() != nil {
		return "[video] " + m.GetVideoMessage().GetCaption()
	}
	if m.GetAudioMessage() != nil {
		return "[audio / voice note]"
	}
	if m.GetDocumentMessage() != nil {
		return "[document] " + m.GetDocumentMessage().GetFileName()
	}
	if m.GetPollCreationMessage() != nil {
		return "[poll] " + m.GetPollCreationMessage().GetName()
	}
	return "[unsupported rich message]"
}

func (c *Client) Send(ctx context.Context, jidString, body string) (store.Message, error) {
	if body == "" {
		return store.Message{}, errors.New("cannot send an empty message")
	}
	jid, err := types.ParseJID(jidString)
	if err != nil {
		return store.Message{}, fmt.Errorf("invalid recipient: %w", err)
	}
	resp, err := c.api.SendMessage(ctx, jid, &waE2E.Message{Conversation: proto.String(body)})
	if err != nil {
		return store.Message{}, err
	}
	m := store.Message{Chat: jid.String(), ID: string(resp.ID), Sender: "self", Body: body, Timestamp: resp.Timestamp, Outgoing: true}
	if m.Timestamp.IsZero() {
		m.Timestamp = time.Now()
	}
	if err := c.archive.Insert(ctx, m); err != nil {
		return m, fmt.Errorf("sent to WhatsApp but failed local archive: %w", err)
	}
	return m, nil
}

// Logout requests remote unlink and removes the local session database, not history.
// If remote unlink fails, local credentials are still removed; warn the caller.
func (c *Client) Logout(ctx context.Context) error {
	remoteErr := c.api.Logout(ctx)
	c.api.Disconnect()
	closeErr := c.Close()
	var fileErr error
	for _, name := range []string{c.path, c.path + "-wal", c.path + "-shm"} {
		if err := os.Remove(name); err != nil && !os.IsNotExist(err) {
			fileErr = errors.Join(fileErr, err)
		}
	}
	return errors.Join(remoteErr, closeErr, fileErr)
}

func (c *Client) Close() error {
	c.closeOnce.Do(func() {
		c.api.Disconnect()
		c.closeErr = c.device.Close()
	})
	return c.closeErr
}
