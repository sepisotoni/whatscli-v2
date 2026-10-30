package wa

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sepisotoni/whatshell-go/internal/store"
	"github.com/sepisotoni/whatshell-go/internal/transfer"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// SendFile uploads a bounded server-local SFTP file as an image or document.
// It never interprets local filenames as commands or arbitrary URLs.
func (c *Client) SendFile(ctx context.Context, jidString, root, filename, caption string) (store.Message, error) {
	jid, err := types.ParseJID(jidString)
	if err != nil {
		return store.Message{}, fmt.Errorf("recipient: %w", err)
	}
	data, name, err := transfer.ReadUpload(root, filename)
	if err != nil {
		return store.Message{}, err
	}
	mediaType := whatsmeow.MediaDocument
	mimeType := mime.TypeByExtension(strings.ToLower(filepath.Ext(name)))
	detected := http.DetectContentType(data)
	if mimeType == "" {
		mimeType = detected
	}
	var content *waE2E.Message
	if detected == "image/jpeg" || detected == "image/png" {
		mimeType = detected
		mediaType = whatsmeow.MediaImage
	}
	uploaded, err := c.api.Upload(ctx, data, mediaType)
	if err != nil {
		return store.Message{}, fmt.Errorf("upload: %w", err)
	}
	var preview string
	if mediaType == whatsmeow.MediaImage {
		content = &waE2E.Message{ImageMessage: &waE2E.ImageMessage{
			Mimetype: proto.String(mimeType), Caption: proto.String(caption),
			URL: &uploaded.URL, DirectPath: &uploaded.DirectPath, MediaKey: uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256, FileSHA256: uploaded.FileSHA256, FileLength: &uploaded.FileLength,
		}}
		preview = "[image] " + name
	} else {
		content = &waE2E.Message{DocumentMessage: &waE2E.DocumentMessage{
			FileName: proto.String(name), Mimetype: proto.String(mimeType), Caption: proto.String(caption),
			URL: &uploaded.URL, DirectPath: &uploaded.DirectPath, MediaKey: uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256, FileSHA256: uploaded.FileSHA256, FileLength: &uploaded.FileLength,
		}}
		preview = "[document] " + name
	}
	resp, err := c.api.SendMessage(ctx, jid, content)
	if err != nil {
		return store.Message{}, fmt.Errorf("send uploaded file: %w", err)
	}
	msg := store.Message{Chat: jid.String(), ID: string(resp.ID), Sender: "self", Body: preview, Outgoing: true, Timestamp: resp.Timestamp}
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}
	if err := c.archive.Insert(ctx, msg); err != nil {
		return msg, fmt.Errorf("sent but not locally archived: %w", err)
	}
	return msg, nil
}

// DownloadMedia downloads a previously received attachment on user request.
// A document is not executed, and the local file is not automatically opened.
func (c *Client) DownloadMedia(ctx context.Context, chat, id, root string) (string, error) {
	name, payload, err := c.archive.Media(ctx, chat, id)
	if err != nil {
		return "", err
	}
	var message waE2E.Message
	if err := proto.Unmarshal(payload, &message); err != nil {
		return "", fmt.Errorf("stored media metadata: %w", err)
	}
	var part whatsmeow.DownloadableMessage
	switch {
	case message.GetImageMessage() != nil:
		part = message.GetImageMessage()
	case message.GetDocumentMessage() != nil:
		part = message.GetDocumentMessage()
	case message.GetVideoMessage() != nil:
		part = message.GetVideoMessage()
	case message.GetAudioMessage() != nil:
		part = message.GetAudioMessage()
	default:
		return "", errors.New("message has no downloadable regular attachment")
	}
	return transfer.SaveDownload(root, name, func(file *os.File) error {
		return c.api.DownloadToFile(ctx, part, file)
	})
}

// SendPoll creates a single-choice poll. Vote collection/results are a later phase.
func (c *Client) SendPoll(ctx context.Context, jidString, question string, options []string) (store.Message, error) {
	jid, err := types.ParseJID(jidString)
	if err != nil {
		return store.Message{}, err
	}
	question = strings.TrimSpace(question)
	if question == "" || len(options) < 2 || len(options) > 12 {
		return store.Message{}, errors.New("poll requires a question and 2-12 options")
	}
	clean := make([]string, 0, len(options))
	seen := map[string]bool{}
	for _, option := range options {
		option = strings.TrimSpace(option)
		if option == "" || seen[strings.ToLower(option)] {
			return store.Message{}, errors.New("poll options must be unique and nonempty")
		}
		seen[strings.ToLower(option)] = true
		clean = append(clean, option)
	}
	resp, err := c.api.SendMessage(ctx, jid, c.api.BuildPollCreation(question, clean, 1))
	if err != nil {
		return store.Message{}, err
	}
	m := store.Message{Chat: jid.String(), ID: string(resp.ID), Sender: "self", Body: "[poll] " + question, Timestamp: resp.Timestamp, Outgoing: true}
	if m.Timestamp.IsZero() {
		m.Timestamp = time.Now()
	}
	if err := c.archive.Insert(ctx, m); err != nil {
		return m, fmt.Errorf("poll sent but archive failed: %w", err)
	}
	return m, nil
}
