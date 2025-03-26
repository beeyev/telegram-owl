package sendmediagroup

import (
	"fmt"
	"net/http"

	"github.com/beeyev/telegram-owl/internal/telegram/common/util"
	"github.com/beeyev/telegram-owl/internal/telegram/httpclient"
)

const telegramAPIEndpoint = "sendMediaGroup"

// Sender provides an interface for sending text messages to a Telegram chat.
type Sender interface {
	Send(opts *Options) error
}

type mediaSender struct {
	httpClient httpclient.HTTPDoer
}

// New returns a new instance of Sender.
func New(httpClient httpclient.HTTPDoer) Sender {
	return mediaSender{httpClient: httpClient}
}

// Send sends a group of photos, videos, documents or audios as an album
// See: https://core.telegram.org/bots/api#sendmediagroup
func (s mediaSender) Send(opts *Options) error {
	payloadData, multipartFiles, err := opts.preparePayload()
	if err != nil {
		return fmt.Errorf("send media: %w", err)
	}

	// Convert the struct payload into a form payload
	formFields, err := util.StructToFormPayload(payloadData)
	if err != nil {
		return fmt.Errorf("unable to create form fields from the payload. Details: %w", err)
	}

	// Submit the multipart/form-data request to Telegram
	if err = s.httpClient.SubmitMultipart(http.MethodPost, telegramAPIEndpoint, formFields, multipartFiles); err != nil {
		return fmt.Errorf("failed to send media: %w", err)
	}

	return nil
}
