package sendrichmessage

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// FormatMarkdown selects Telegram Rich Markdown.
	FormatMarkdown = "rich-markdown"
	// FormatHTML selects Telegram Rich HTML.
	FormatHTML = "rich-html"
)

// IsFormat reports whether format selects a rich message representation.
func IsFormat(format string) bool {
	return format == FormatMarkdown || format == FormatHTML
}

// Options contains the sendRichMessage parameters supported by the CLI.
type Options struct {
	ChatID              string
	MessageThreadID     string
	Text                string
	Format              string
	DisableNotification bool
	ProtectContent      bool
}

type payload struct {
	ChatID              string           `json:"chat_id"`
	MessageThreadID     string           `json:"message_thread_id,omitempty"`
	RichMessage         inputRichMessage `json:"rich_message"`
	DisableNotification bool             `json:"disable_notification,omitempty"`
	ProtectContent      bool             `json:"protect_content,omitempty"`
}

type inputRichMessage struct {
	Markdown string `json:"markdown,omitempty"`
	HTML     string `json:"html,omitempty"`
}

func (o *Options) preparePayload() (*payload, error) {
	if err := o.validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	richMessage := inputRichMessage{}
	switch o.Format {
	case FormatMarkdown:
		richMessage.Markdown = o.Text
	case FormatHTML:
		richMessage.HTML = o.Text
	}

	return &payload{
		ChatID:              o.ChatID,
		MessageThreadID:     o.MessageThreadID,
		RichMessage:         richMessage,
		DisableNotification: o.DisableNotification,
		ProtectContent:      o.ProtectContent,
	}, nil
}

func (o *Options) validate() error {
	var validationErrors []string

	if o.ChatID == "" {
		validationErrors = append(validationErrors, "chat ID is required")
	}
	if o.Text == "" {
		validationErrors = append(validationErrors, "message is required")
	}
	if !IsFormat(o.Format) {
		validationErrors = append(validationErrors, "format must be rich-markdown or rich-html")
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}
