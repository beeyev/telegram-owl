package attachment

// AType is the Telegram Bot API media type used in an InputMedia payload.
type AType string

const (
	// Document is also the safe fallback for unknown or incompatible media.
	Document AType = "document"
	Video    AType = "video"
	Audio    AType = "audio"
	Photo    AType = "photo"
)

// String returns the exact wire value expected by the Telegram Bot API.
func (a AType) String() string {
	return string(a)
}
