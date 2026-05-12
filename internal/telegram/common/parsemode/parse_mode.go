package parsemode

func Normalize(parseMode string) string {
	if parseMode == "markdown" {
		return "MarkdownV2"
	}

	return parseMode
}
