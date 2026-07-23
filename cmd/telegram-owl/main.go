// Command telegram-owl sends text messages and file attachments through the
// Telegram Bot API.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/beeyev/telegram-owl/internal/cli"
)

const apiBotURL = "https://api.telegram.org"

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())

		os.Exit(1)
	}
}

// run keeps signal cleanup on a normal return path. main calls [os.Exit] only
// after run's deferred functions have completed.
func run() error {
	// Cancel in-flight HTTP requests on both interactive interrupts and
	// container/process termination.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	return cli.NewApp(apiBotURL).Run(ctx, os.Args)
}
