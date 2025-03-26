package cli

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

type inputValues struct {
	cmd *cli.Command
}

func (iv *inputValues) validate() error {
	if iv.cmd.String("token") == "" {
		//nolint:stylecheck // Probably i need to render this message in a different way
		return errors.New(`missing required flag: --token
Set it via the --token flag or the TELEGRAM_SEND_TOKEN environment variable.

Example:
  telegram-owl --token=123:ABC

Run with --help to see all options.`)
	}

	if iv.cmd.String("chat") == "" {
		//nolint:stylecheck // Probably i need to render this message in a different way
		return errors.New(`missing required flag: --chat
Set it via the --chat flag or the TELEGRAM_SEND_CHAT environment variable.

Example:
  telegram-owl --chat=31337

Run with --help to see all options.`)
	}

	return nil
}

func (iv *inputValues) getMessage() string {
	// Return the message if provided via `--message`
	if msg := iv.cmd.String("message"); msg != "" {
		return msg
	}

	// Read message from stdin if `--stdin` flag is used
	if !iv.cmd.Bool("stdin") {
		return ""
	}

	// Check if stdin is actually piped data
	stat, err := os.Stdin.Stat()
	if err != nil || (stat.Mode()&os.ModeCharDevice) != 0 {
		return ""
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}
