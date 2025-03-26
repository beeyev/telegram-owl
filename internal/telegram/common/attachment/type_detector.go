package attachment

import (
	"path/filepath"
	"strings"
)

// DetectType determines the file type based on its extension.
// Falls back to Document if unknown.
func DetectType(fileName string) AType {
	if fileName == "" {
		panic("file name is empty")
	}

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(fileName)), ".")

	extensionToType := map[string]AType{
		// Images
		"jpg":  Photo,
		"jpeg": Photo,
		"png":  Photo,
		"bmp":  Photo,
		"webp": Photo,

		// Audio
		"mp3":  Audio,
		"m4a":  Audio,
		"wav":  Audio,
		"flac": Audio,
		"aac":  Audio,

		// Video
		"mp4": Video,
		"mov": Video,
		"mkv": Video,
		"gif": Video,
	}

	if attachmentType, ok := extensionToType[ext]; ok {
		return attachmentType
	}

	return Document
}
