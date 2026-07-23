// Package telegram wires the HTTP transport to the supported Bot API methods.
// It contains composition only; request validation and encoding stay in the
// individual method packages.
package telegram

import (
	"github.com/beeyev/telegram-owl/internal/telegram/httpclient"
	"github.com/beeyev/telegram-owl/internal/telegram/method/sendmediagroup"
	"github.com/beeyev/telegram-owl/internal/telegram/method/sendmessage"
	"github.com/beeyev/telegram-owl/internal/telegram/method/sendrichmessage"
)

// Client groups the Telegram operations exposed to the CLI.
type Client struct {
	SendMessage     sendmessage.Sender
	SendMediaGroup  sendmediagroup.Sender
	SendRichMessage sendrichmessage.Sender
}

// NewClient builds all method senders over one configured HTTP transport.
func NewClient(apiBotURL, token, proxyURL string) (*Client, error) {
	httpClient, err := httpclient.New(apiBotURL, token, proxyURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		SendMessage:     sendmessage.New(httpClient),
		SendMediaGroup:  sendmediagroup.New(httpClient),
		SendRichMessage: sendrichmessage.New(httpClient),
	}, nil
}
