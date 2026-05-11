package sendmessage

import (
	"context"
	"fmt"
	"net/http"

	"github.com/beeyev/telegram-owl/internal/telegram/httpclient"
)

const telegramAPIEndpoint = "sendMessage"

// Sender provides an interface for sending text messages to a Telegram chat.
type Sender interface {
	Send(ctx context.Context, opts *Options) error
}

type messageSender struct {
	httpClient httpclient.HTTPDoer
}

// New returns a new instance of Sender.
func New(httpClient httpclient.HTTPDoer) Sender {
	return messageSender{httpClient: httpClient}
}

// Send sends a text message to the specified Telegram chat.
// See: https://core.telegram.org/bots/api#sendmessage
func (s messageSender) Send(ctx context.Context, opts *Options) error {
	payload, err := opts.preparePayload()
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}

	// Submit the JSON payload to the Telegram API
	if err = s.httpClient.SubmitJSON(ctx, http.MethodPost, telegramAPIEndpoint, payload); err != nil {
		return fmt.Errorf("send: failed to send message: %w", err)
	}

	return nil
}
