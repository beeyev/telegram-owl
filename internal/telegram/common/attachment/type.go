package attachment

type AType string

const (
	Document AType = "document"
	Video    AType = "video"
	Audio    AType = "audio"
	Photo    AType = "photo"
)

func (a AType) String() string {
	return string(a)
}
