// Package sendmediagroup validates and sends Telegram media albums.
package sendmediagroup

import (
	"context"
	"fmt"
	"net/http"

	"github.com/beeyev/telegram-owl/internal/telegram/common/util"
	"github.com/beeyev/telegram-owl/internal/telegram/httpclient"
)

const telegramAPIEndpoint = "sendMediaGroup"

// Sender sends an attachment group to a Telegram chat.
type Sender interface {
	Send(ctx context.Context, opts *Options) error
}

type mediaSender struct {
	httpClient httpclient.HTTPDoer
}

// New returns a media sender backed by httpClient.
func New(httpClient httpclient.HTTPDoer) Sender {
	return mediaSender{httpClient: httpClient}
}

// Send validates opts and submits one sendMediaGroup multipart request.
// See https://core.telegram.org/bots/api#sendmediagroup.
func (s mediaSender) Send(ctx context.Context, opts *Options) error {
	payloadData, multipartFiles, err := opts.preparePayload()
	if err != nil {
		return fmt.Errorf("send media: %w", err)
	}

	formFields, err := util.StructToFormPayload(payloadData)
	if err != nil {
		return fmt.Errorf("unable to create form fields from the payload. Details: %w", err)
	}

	if err = s.httpClient.SubmitMultipart(
		ctx,
		http.MethodPost,
		telegramAPIEndpoint,
		formFields,
		multipartFiles,
	); err != nil {
		return fmt.Errorf("failed to send media: %w", err)
	}

	return nil
}
