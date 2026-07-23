package sendmediagroup

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	attach "github.com/beeyev/telegram-owl/internal/telegram/common/attachment"
	"github.com/beeyev/telegram-owl/internal/telegram/common/parsemode"
	"github.com/beeyev/telegram-owl/internal/telegram/httpclient"
)

// MaxCaptionLength defines the maximum length allowed for an InputMedia caption.
const MaxCaptionLength = 1024

// Options contains the user-visible sendMediaGroup parameters supported by the
// CLI. Attachments must remain open until Sender.Send returns.
type Options struct {
	ChatID              string
	MessageThreadID     string
	Caption             string
	ParseMode           string
	HasSpoiler          bool
	DisableNotification bool
	ProtectContent      bool
	Attachments         attach.Attachments
}

type payload struct {
	ChatID              string `json:"chat_id"`
	MessageThreadID     string `json:"message_thread_id,omitempty"`
	Media               string `json:"media"`
	DisableNotification bool   `json:"disable_notification,omitempty"`
	ProtectContent      bool   `json:"protect_content,omitempty"`
}

// media is one InputMedia entry in Telegram's JSON-encoded media form field.
type media struct {
	Type       string `json:"type"`
	Media      string `json:"media"`
	Caption    string `json:"caption,omitempty"`
	ParseMode  string `json:"parse_mode,omitempty"`
	HasSpoiler bool   `json:"has_spoiler,omitempty"`
}

func (o *Options) preparePayload() (*payload, []httpclient.MultipartFile, error) {
	if err := o.validate(); err != nil {
		return nil, nil, fmt.Errorf("validation failed: %w", err)
	}

	medias, multipartFiles := o.prepareMedia()

	mediaJSON, err := json.Marshal(medias)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert media data to JSON. Details: %w", err)
	}

	payloadData := &payload{
		ChatID:              o.ChatID,
		MessageThreadID:     o.MessageThreadID,
		Media:               string(mediaJSON),
		DisableNotification: o.DisableNotification,
		ProtectContent:      o.ProtectContent,
	}

	return payloadData, multipartFiles, nil
}

func (o *Options) prepareMedia() ([]media, []httpclient.MultipartFile) {
	medias := make([]media, 0, len(o.Attachments))
	multipartFiles := make([]httpclient.MultipartFile, 0, len(o.Attachments))

	for i, attachment := range o.Attachments {
		formFieldName := fmt.Sprintf("file%d", i)

		medias = append(medias, media{
			Type:       attachment.AType.String(),
			Media:      "attach://" + formFieldName,
			HasSpoiler: o.HasSpoiler,
		})

		multipartFiles = append(multipartFiles, httpclient.MultipartFile{
			FieldName:  formFieldName,
			FileName:   attachment.FileName,
			FileReader: attachment.File,
		})
	}

	// prepareMedia is called only after validate, so medias always has a last
	// element. Telegram renders the last item's caption with the whole album.
	lastMedia := &medias[len(medias)-1]
	lastMedia.Caption = o.Caption
	if o.Caption != "" {
		lastMedia.ParseMode = parsemode.Normalize(o.ParseMode)
	}

	return medias, multipartFiles
}

func (o *Options) validate() error {
	var validationErrors []string

	if o.ChatID == "" {
		validationErrors = append(validationErrors, "chat ID is required")
	}
	if len(o.Attachments) == 0 {
		validationErrors = append(validationErrors, "at least one attachment required")
	}
	// Telegram applies the limit after parsing entities. Without duplicating
	// Telegram's parser, only plain-text length can be validated accurately.
	captionLen := utf8.RuneCountInString(o.Caption)
	if o.ParseMode == "" && captionLen > MaxCaptionLength {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("message is too long: must be <= %d characters, got %d", MaxCaptionLength, captionLen),
		)
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}
