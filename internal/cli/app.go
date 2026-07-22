package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/beeyev/telegram-owl/internal/telegram"
	"github.com/beeyev/telegram-owl/internal/telegram/common/attachment"
	"github.com/beeyev/telegram-owl/internal/version"
)

const (
	author                      = "Alexander Tebiev - https://github.com/beeyev"
	maxTotalAttachments         = 10
	maxPhotoAttachmentSizeBytes = 10 * attachment.BytesPerMegabyte
	maxAttachmentSizeBytes      = 50 * attachment.BytesPerMegabyte
	maxTotalSizeBytes           = 50 * attachment.BytesPerMegabyte
)

const usageText = `Examples:
  telegram-owl --token=$TOKEN --chat=@mychannel --message "Hello"
  echo "Hi there" | telegram-owl -t $TOKEN -c 123456789 --stdin
  telegram-owl -t $TOKEN -c @group --attach file.jpg --spoiler`

func versionFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:        "version",
		Usage:       "print the version",
		Aliases:     []string{"v"},
		OnlyOnce:    true,
		HideDefault: true,
	}
}

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
		&cli.BoolFlag{
			Name:        "verbose",
			Usage:       "Print success messages.",
			OnlyOnce:    true,
			HideDefault: true,
		},
		versionFlag(),
	}
}

func NewApp(apiBotURL string) *cli.Command {
	return &cli.Command{
		Name:            "telegram-owl",
		Usage:           "Send messages and attachments to Telegram via the command line.",
		Description:     "A simple CLI tool to send text messages and file attachments to Telegram chats and channels.",
		Version:         version.Version(),
		HideHelpCommand: true,
		UsageText:       usageText,
		Flags:           flags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Bool("version") {
				return printVersion(cmd)
			}

			if cmd.NumFlags() == 0 {
				return cli.ShowAppHelp(cmd)
			}

			iv := &inputValues{cmd: cmd}
			if err := iv.validate(); err != nil {
				return err
			}

			message, err := iv.getMessage()
			if err != nil {
				return err
			}

			telegramClient, err := telegram.NewClient(apiBotURL, cmd.String("token"), cmd.String("proxy"))
			if err != nil {
				return fmt.Errorf("create telegram client: %w", err)
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
				ctx:              ctx,
				client:           telegramClient,
				attachLoader:     attachLoader,
				warningWriter:    cmd.ErrWriter,
				chatID:           cmd.String("chat"),
				message:          message,
				MessageFormat:    cmd.String("format"),
				attachmentsPaths: cmd.StringSlice("attach"),
				silent:           cmd.Bool("silent"),
				noLinkPreview:    cmd.Bool("no-link-preview"),
				spoiler:          cmd.Bool("spoiler"),
				protect:          cmd.Bool("protect"),
				threadID:         cmd.String("thread"),
			}

			verbose := cmd.Bool("verbose")
			startedAt := time.Now()
			if verbose {
				_, _ = fmt.Fprintln(cmd.Writer, verboseSendSummary(cmd, a))
			}

			if err = a.execute(); err != nil {
				return fmt.Errorf("failed to send message to chat ID %s: %w", a.chatID, err)
			}

			if verbose {
				_, _ = fmt.Fprintf(
					cmd.Writer,
					"Message sent successfully. Chat ID: %s. Duration: %s\n",
					a.chatID,
					time.Since(startedAt).Round(time.Millisecond),
				)
			}

			return nil
		},
	}
}

func printVersion(cmd *cli.Command) error {
	displayVersion := cmd.Version
	if !strings.HasPrefix(displayVersion, "v") {
		displayVersion = "v" + displayVersion
	}

	_, err := fmt.Fprintf(cmd.Writer, "%s %s\n%s\n", cmd.Name, displayVersion, author)
	return err
}

func verboseSendSummary(cmd *cli.Command, a *action) string {
	hasMessage := "no"
	if a.message != "" {
		hasMessage = "yes"
	}

	parts := []string{
		fmt.Sprintf("chat=%s", a.chatID),
		fmt.Sprintf("message=%s", hasMessage),
		fmt.Sprintf("attachments=%d", len(a.attachmentsPaths)),
	}

	if a.threadID != "" {
		parts = append(parts, fmt.Sprintf("thread=%s", a.threadID))
	}
	if a.MessageFormat != "" {
		parts = append(parts, fmt.Sprintf("format=%s", a.MessageFormat))
	}
	if cmd.String("proxy") != "" {
		parts = append(parts, "proxy=yes")
	}

	return "Sending Telegram message: " + strings.Join(parts, ", ")
}
