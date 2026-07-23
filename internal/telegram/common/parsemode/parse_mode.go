// Package parsemode maps CLI format names to Telegram Bot API parse modes.
package parsemode

// Normalize maps the CLI's "markdown" value to Telegram's MarkdownV2 wire
// value. Other values, including "html" and an empty string, pass through.
func Normalize(parseMode string) string {
	if parseMode == "markdown" {
		return "MarkdownV2"
	}

	return parseMode
}
