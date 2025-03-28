package sendmessage

import (
	"errors"
	"fmt"
	"strings"
)

// MaxTextLength defines the maximum length allowed for a text message.
const MaxTextLength = 4096

type Options struct {
	ChatID              string
	MessageThreadID     string
	Text                string
	ParseMode           string
	HasSpoiler          bool
	DisableNotification bool
	ProtectContent      bool
	DisableLinkPreview  bool
}

type payload struct {
	ChatID              string              `json:"chat_id"`
	MessageThreadID     string              `json:"message_thread_id,omitempty"`
	Text                string              `json:"text"`
	ParseMode           string              `json:"parse_mode,omitempty"`
	DisableNotification bool                `json:"disable_notification,omitempty"`
	ProtectContent      bool                `json:"protect_content,omitempty"`
	LinkPreviewOptions  *linkPreviewOptions `json:"link_preview_options,omitempty"`
}

type linkPreviewOptions struct {
	IsDisabled bool `json:"is_disabled,omitempty"`
}

func (o *Options) preparePayload() (*payload, error) {
	if err := o.validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	payload := &payload{
		ChatID:              o.ChatID,
		MessageThreadID:     o.MessageThreadID,
		Text:                o.Text,
		ParseMode:           o.ParseMode,
		DisableNotification: o.DisableNotification,
		ProtectContent:      o.ProtectContent,
	}

	if o.ParseMode == "markdown" {
		payload.ParseMode = "MarkdownV2"
	}

	if o.DisableLinkPreview {
		payload.LinkPreviewOptions = &linkPreviewOptions{IsDisabled: true}
	}

	return payload, nil
}

func (o *Options) validate() error {
	var validationErrors []string

	if o.ChatID == "" {
		validationErrors = append(validationErrors, "chat ID is required")
	}
	if o.Text == "" {
		validationErrors = append(validationErrors, "message is required")
	}
	if len(o.Text) > MaxTextLength {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("message is too long: must be <= %d characters, got %d", MaxTextLength, len(o.Text)),
		)
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}
