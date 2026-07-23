package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/beeyev/telegram-owl/internal/telegram/method/sendrichmessage"
)

type inputValues struct {
	cmd *cli.Command
}

func (iv *inputValues) validate() error {
	if iv.cmd.String("token") == "" {
		//nolint:revive,staticcheck // Multiline CLI guidance intentionally uses sentence casing and punctuation.
		return errors.New(`missing required flag: --token
Set it via the --token flag or the TELEGRAM_OWL_TOKEN environment variable.

Example:
  telegram-owl --token=123:ABC

Run with --help to see all options.`)
	}

	if iv.cmd.String("chat") == "" {
		//nolint:revive,staticcheck // Multiline CLI guidance intentionally uses sentence casing and punctuation.
		return errors.New(`missing required flag: --chat
Set it via the --chat flag or the TELEGRAM_OWL_CHAT environment variable.

Example:
  telegram-owl --chat=31337

Run with --help to see all options.`)
	}

	format := iv.cmd.String("format")

	if format != "" && format != "markdown" && format != "html" && !sendrichmessage.IsFormat(format) {
		return errors.New(
			`incorrect value for --format flag, possible values: markdown, html, rich-markdown, rich-html`,
		)
	}

	if sendrichmessage.IsFormat(format) && iv.cmd.Bool("no-link-preview") {
		return errors.New("--no-link-preview is not supported with rich message formats")
	}

	return nil
}

func (iv *inputValues) getMessage() (string, error) {
	// An explicit flag wins over stdin. This avoids blocking or consuming a
	// pipeline when both inputs are provided.
	if msg := iv.cmd.String("message"); msg != "" {
		return msg, nil
	}

	if !iv.cmd.Bool("stdin") {
		return "", nil
	}

	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", fmt.Errorf("inspect stdin: %w", err)
	}
	if stat.Mode()&os.ModeCharDevice != 0 {
		// --stdin may accompany attachments to request an optional caption. An
		// interactive stdin therefore means "no caption" when attachments exist,
		// but is an error for a text-only send where it would send nothing.
		if len(iv.cmd.StringSlice("attach")) > 0 {
			return "", nil
		}

		return "", errors.New("stdin does not contain piped data")
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}

	// Shell pipelines commonly append one line ending. Remove only that ending
	// so indentation and other intentional whitespace remain available to rich
	// Markdown and HTML.
	message := strings.TrimSuffix(string(data), "\n")
	message = strings.TrimSuffix(message, "\r")

	return message, nil
}
