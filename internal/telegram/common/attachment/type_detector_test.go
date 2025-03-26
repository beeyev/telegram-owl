package attachment_test

import (
	"testing"

	"github.com/beeyev/telegram-owl/internal/telegram/common/attachment"
	"github.com/stretchr/testify/assert"
)

func TestDetectType(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     attachment.AType
	}{
		// Images
		{"jpg extension", "image.jpg", attachment.Photo},
		{"jpeg extension", "image.jpeg", attachment.Photo},
		{"png extension", "image.png", attachment.Photo},
		{"bmp extension", "image.bmp", attachment.Photo},
		{"webp extension", "image.webp", attachment.Photo},

		// Audio
		{"mp3 extension", "audio.mp3", attachment.Audio},
		{"m4a extension", "audio.m4a", attachment.Audio},
		{"wav extension", "audio.wav", attachment.Audio},
		{"flac extension", "audio.flac", attachment.Audio},
		{"aac extension", "audio.aac", attachment.Audio},

		// Video
		{"mp4 extension", "video.mp4", attachment.Video},
		{"mov extension", "video.mov", attachment.Video},
		{"mkv extension", "video.mkv", attachment.Video},
		{"gif extension", "video.gif", attachment.Video},

		// Mixed-case / uppercase extensions
		{"upper JPG extension", "picture.JPG", attachment.Photo},
		{"mixed-case JpEg", "picture.JpEg", attachment.Photo},
		{"upper MP3 extension", "sound.MP3", attachment.Audio},
		{"mixed-case MoV", "clip.MoV", attachment.Video},

		// No extension
		{"no extension", "filename", attachment.Document},
		{"dot but no extension", "filename.", attachment.Document},
		{"hidden file in UNIX style", ".hiddenfile", attachment.Document},

		// Unknown extension
		{"unknown extension", "document.xyz", attachment.Document},
		{"another unknown extension", "archive.zip", attachment.Document},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := attachment.DetectType(tt.fileName)
			assert.Equal(t, tt.want, got, "DetectType(%q) should return %v", tt.fileName, tt.want)
		})
	}
}
