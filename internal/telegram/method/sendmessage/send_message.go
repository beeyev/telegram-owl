// Package sendmessage validates and sends Telegram sendMessage requests.
package sendmessage

import (
	"context"
	"fmt"
	"net/http"

	"github.com/beeyev/telegram-owl/internal/telegram/httpclient"
)

const telegramAPIEndpoint = "sendMessage"

// Sender sends text messages to a Telegram chat.
type Sender interface {
	Send(ctx context.Context, opts *Options) error
}

type messageSender struct {
	httpClient httpclient.HTTPDoer
}

// New returns a sender backed by httpClient.
func New(httpClient httpclient.HTTPDoer) Sender {
	return messageSender{httpClient: httpClient}
}

// Send validates opts and submits one sendMessage request.
// See https://core.telegram.org/bots/api#sendmessage.
func (s messageSender) Send(ctx context.Context, opts *Options) error {
	payload, err := opts.preparePayload()
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}

	if err = s.httpClient.SubmitJSON(ctx, http.MethodPost, telegramAPIEndpoint, payload); err != nil {
		return fmt.Errorf("send: failed to send message: %w", err)
	}

	return nil
}
