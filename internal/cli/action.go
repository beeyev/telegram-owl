package cli

import (
	"errors"
	"fmt"

	"github.com/beeyev/telegram-owl/internal/telegram"
	"github.com/beeyev/telegram-owl/internal/telegram/common/attachment"
	"github.com/beeyev/telegram-owl/internal/telegram/method/sendmediagroup"
	"github.com/beeyev/telegram-owl/internal/telegram/method/sendmessage"
)

type action struct {
	client           *telegram.Client
	attachLoader     *attachment.Loader
	chatID           string
	message          string
	attachmentsPaths []string
	silent           bool
	noLinkPreview    bool
	spoiler          bool
	protect          bool
	threadID         string
}

func (a *action) execute() error {
	if a.message == "" && len(a.attachmentsPaths) == 0 {
		return errors.New("nothing to send: provide a --message or --attach flag")
	}

	// Send message without media
	if len(a.attachmentsPaths) == 0 {
		return a.sendMessage(a.message)
	}

	// If message fits in media caption, send as single request
	if len(a.message) <= sendmediagroup.MaxCaptionLength {
		return a.sendMediaGroup(a.message)
	}

	// Otherwise split into media group and separate message
	if err := a.sendMediaGroup(""); err != nil {
		return err
	}

	return a.sendMessage(a.message)
}

func (a *action) sendMessage(message string) error {
	if message == "" {
		panic("message is required")
	}
	return a.client.SendMessage.Send(&sendmessage.Options{
		ChatID:              a.chatID,
		Text:                message,
		DisableNotification: a.silent,
		ProtectContent:      a.protect,
		MessageThreadID:     a.threadID,
		DisableLinkPreview:  a.noLinkPreview,
	})
}

func (a *action) sendMediaGroup(message string) error {
	if len(a.attachmentsPaths) == 0 {
		panic("no attachments to send")
	}

	attachments, err := a.attachLoader.LoadMultipleAttachments(a.attachmentsPaths)
	if err != nil {
		return fmt.Errorf("failed to load attachments: %w", err)
	}

	return a.client.SendMediaGroup.Send(&sendmediagroup.Options{
		ChatID:              a.chatID,
		MessageThreadID:     a.threadID,
		Caption:             message,
		HasSpoiler:          a.spoiler,
		DisableNotification: a.silent,
		ProtectContent:      a.protect,
		Attachments:         attachments,
	})
}
