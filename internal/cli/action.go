package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/beeyev/telegram-owl/internal/telegram"
	"github.com/beeyev/telegram-owl/internal/telegram/common/attachment"
	"github.com/beeyev/telegram-owl/internal/telegram/method/sendmediagroup"
	"github.com/beeyev/telegram-owl/internal/telegram/method/sendmessage"
)

type action struct {
	ctx              context.Context
	client           *telegram.Client
	attachLoader     *attachment.Loader
	warningWriter    io.Writer
	chatID           string
	message          string
	MessageFormat    string
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

	// Text-only sends use sendMessage because Telegram media groups require at
	// least one attachment.
	if len(a.attachmentsPaths) == 0 {
		return a.sendMessage(a.message)
	}

	// Telegram limits a media-group caption to 1,024 Unicode code points. Keep
	// short text attached to the media so the operation remains one request.
	if utf8.RuneCountInString(a.message) <= sendmediagroup.MaxCaptionLength {
		return a.sendMediaGroup(a.message)
	}

	// Longer text cannot be a caption. Send the attachments first, then the
	// complete text as a separate message. Do not send the text if upload fails.
	if err := a.sendMediaGroup(""); err != nil {
		return err
	}

	return a.sendMessage(a.message)
}

func (a *action) sendMessage(message string) error {
	if message == "" {
		return errors.New("message is required")
	}

	return a.client.SendMessage.Send(a.ctx, &sendmessage.Options{
		ChatID:              a.chatID,
		Text:                message,
		ParseMode:           a.MessageFormat,
		DisableNotification: a.silent,
		ProtectContent:      a.protect,
		MessageThreadID:     a.threadID,
		DisableLinkPreview:  a.noLinkPreview,
	})
}

func (a *action) sendMediaGroup(message string) error {
	if len(a.attachmentsPaths) == 0 {
		return errors.New("no attachments to send")
	}

	attachments, err := a.attachLoader.LoadMultipleAttachments(a.attachmentsPaths)
	if err != nil {
		return fmt.Errorf("failed to load attachments: %w", err)
	}

	// The loader transfers ownership of open files to this action. Keep them
	// open through the synchronous upload, then close each file exactly once.
	// The HTTP adapter hides io.Closer from Resty so Resty cannot close them.
	sendErr := a.client.SendMediaGroup.Send(a.ctx, &sendmediagroup.Options{
		ChatID:              a.chatID,
		MessageThreadID:     a.threadID,
		Caption:             message,
		ParseMode:           a.MessageFormat,
		HasSpoiler:          a.spoiler,
		DisableNotification: a.silent,
		ProtectContent:      a.protect,
		Attachments:         attachments,
	})
	if sendErr != nil {
		sendErr = fmt.Errorf("send attachments: %w", sendErr)
	}

	closeErr := attachments.Close()
	if closeErr == nil {
		return sendErr
	}
	closeErr = fmt.Errorf("close attachments: %w", closeErr)

	if sendErr != nil {
		// Preserve both failures. The upload error explains the failed operation;
		// the cleanup error identifies leaked or partially closed resources.
		return errors.Join(sendErr, closeErr)
	}

	// The upload already succeeded. Report cleanup as a warning so a long
	// message can still continue with its second, text-only request.
	if a.warningWriter != nil {
		_, _ = fmt.Fprintf(
			a.warningWriter,
			"warning: attachments sent, but cleanup failed: %v\n",
			closeErr,
		)
	}

	return nil
}
