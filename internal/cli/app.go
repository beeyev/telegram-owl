package cli

import (
	"context"
	"fmt"

	"github.com/beeyev/telegram-owl/internal/telegram"
	"github.com/beeyev/telegram-owl/internal/telegram/common/attachment"
	"github.com/beeyev/telegram-owl/internal/version"
	"github.com/urfave/cli/v3"
)

const author = "Alexander Tebiev - https://github.com/beeyev"
const (
	maxTotalAttachments         = 10
	maxPhotoAttachmentSizeBytes = 10 * attachment.BytesPerMegabyte
	maxAttachmentSizeBytes      = 50 * attachment.BytesPerMegabyte
	maxTotalSizeBytes           = 50 * attachment.BytesPerMegabyte
)

const usageText = `Examples:
  telegram-owl --token=$TOKEN --chat=@mychannel --message "Hello"
  echo "Hi there" | telegram-owl -t $TOKEN -c 123456789 --stdin
  telegram-owl -t $TOKEN -c @group --attach file.jpg --spoiler`

func flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "token",
			Usage:    "Telegram bot token (required), environment variable:",
			Aliases:  []string{"t"},
			OnlyOnce: true,
			Sources:  cli.EnvVars("TELEGRAM_OWL_TOKEN"),
			Config:   cli.StringConfig{TrimSpace: true},
		},
		&cli.StringFlag{
			Name:     "chat",
			Usage:    "Chat ID (numeric) or channel username (@channelusername) (required). environment variable:",
			Aliases:  []string{"c"},
			OnlyOnce: true,
			Sources:  cli.EnvVars("TELEGRAM_OWL_CHAT"),
			Config:   cli.StringConfig{TrimSpace: true},
		},
		&cli.StringFlag{
			Name:     "proxy",
			Usage:    "Proxy URL for outgoing requests, environment variable:",
			OnlyOnce: true,
			Sources:  cli.EnvVars("TELEGRAM_OWL_PROXY"),
			Config:   cli.StringConfig{TrimSpace: true},
		},
		&cli.StringFlag{
			Name:     "message",
			Usage:    "Text message content. Use --stdin to read from standard input.",
			Aliases:  []string{"m"},
			OnlyOnce: true,
		},
		&cli.StringFlag{
			Name:     "format",
			Usage:    "Message format options, possible values: markdown, html",
			Aliases:  []string{"f"},
			OnlyOnce: true,
			Config:   cli.StringConfig{TrimSpace: true},
		},
		&cli.StringSliceFlag{
			Name:      "attach",
			Usage:     "File paths of attachments. Can be specified multiple times or comma-separated.",
			Aliases:   []string{"a"},
			TakesFile: true,
		},
		&cli.BoolFlag{
			Name:        "as-document",
			Usage:       "Send all attachments as documents (bypass media type detection).",
			Aliases:     []string{"d"},
			OnlyOnce:    true,
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "silent",
			Usage:       "Sends the message silently. Users will receive a notification with no sound",
			Aliases:     []string{"s"},
			OnlyOnce:    true,
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "spoiler",
			Usage:       "Cover media attachments with a spoiler animation.",
			OnlyOnce:    true,
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "protect",
			Usage:       "Protects the message content from forwarding and saving.",
			OnlyOnce:    true,
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "no-link-preview",
			Usage:       "Disable automatic link previews for messages.",
			OnlyOnce:    true,
			HideDefault: true,
		},
		&cli.StringFlag{
			Name:     "thread",
			Usage:    "Message thread ID (forum supergroup topics only), environment variable:",
			OnlyOnce: true,
			Sources:  cli.EnvVars("TELEGRAM_OWL_THREAD"),
			Config:   cli.StringConfig{TrimSpace: true},
		},
		&cli.BoolFlag{
			Name:        "stdin",
			Usage:       "Read message content from stdin. Example: echo 'Hello, world!' | telegram-owl --stdin",
			OnlyOnce:    true,
			HideDefault: true,
		},
	}
}

func NewApp(apiBotURL string) *cli.Command {
	//nolint:reassign // "reassigning variable VersionPrinter in other package cli"
	cli.VersionPrinter = func(cmd *cli.Command) {
		_, _ = fmt.Fprintf(cmd.Writer, "%s v%s\n%s\n", cmd.Name, cmd.Version, author)
	}

	return &cli.Command{
		Name:            "telegram-owl",
		Usage:           "Send messages and attachments to Telegram via the command line.",
		Description:     "A simple CLI tool to send text messages and file attachments to Telegram chats and channels.",
		Version:         version.Version(),
		HideHelpCommand: true,
		UsageText:       usageText,
		Flags:           flags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.NumFlags() == 0 {
				return cli.ShowAppHelp(cmd)
			}

			iv := &inputValues{cmd: cmd}
			if err := iv.validate(); err != nil {
				return err
			}

			telegramClient, err := telegram.NewClient(apiBotURL, cmd.String("token"), cmd.String("proxy"))
			if err != nil {
				return fmt.Errorf("failed to create Telegram client: %w", err)
			}

			attachLoader := &attachment.Loader{
				FileOpener:                  &attachment.OSFileOpener{},
				IsEverythingDocument:        cmd.Bool("as-document"),
				MaxTotalAttachments:         maxTotalAttachments,
				MaxPhotoAttachmentSizeBytes: maxPhotoAttachmentSizeBytes,
				MaxAttachmentSizeBytes:      maxAttachmentSizeBytes,
				MaxTotalSizeBytes:           maxTotalSizeBytes,
			}

			a := &action{
				client:           telegramClient,
				attachLoader:     attachLoader,
				chatID:           cmd.String("chat"),
				message:          iv.getMessage(),
				MessageFormat:    cmd.String("format"),
				attachmentsPaths: cmd.StringSlice("attach"),
				silent:           cmd.Bool("silent"),
				noLinkPreview:    cmd.Bool("no-link-preview"),
				spoiler:          cmd.Bool("spoiler"),
				protect:          cmd.Bool("protect"),
				threadID:         cmd.String("thread"),
			}

			if err = a.execute(); err != nil {
				return fmt.Errorf("failed to send message to chat ID %s: %w", a.chatID, err)
			}

			_, _ = fmt.Fprintln(cmd.Writer, "Message sent successfully. Chat ID:", a.chatID)

			return nil
		},
	}
}
