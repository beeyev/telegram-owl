package telegram

import (
	"github.com/beeyev/telegram-owl/internal/telegram/httpclient"
	"github.com/beeyev/telegram-owl/internal/telegram/method/sendmediagroup"
	"github.com/beeyev/telegram-owl/internal/telegram/method/sendmessage"
)

type Client struct {
	SendMessage    sendmessage.Sender
	SendMediaGroup sendmediagroup.Sender
}

func NewClient(apiBotURL, token string) (*Client, error) {
	httpClient, err := httpclient.New(apiBotURL, token)
	if err != nil {
		return nil, err
	}

	return &Client{
		SendMessage:    sendmessage.New(httpClient),
		SendMediaGroup: sendmediagroup.New(httpClient),
	}, nil
}
