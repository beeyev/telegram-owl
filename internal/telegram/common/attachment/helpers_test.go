package attachment //nolint:testpackage // Yes, I want to test private functions, I know what I'm doing.

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBytesToMegabytes(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected int
	}{
		{"exact MB", 11 * 1024 * 1024, 11},
		{"zero bytes", 0, 0},
		{"round down", 1_500_000, 1},
		{"round up", 1_600_000, 2},
		{"large value", 3_435_973_836, 3277},
		{"partial MB", 512_345, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bytesToMegabytes(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsOnlyPhotoOrVideo(t *testing.T) {
	//nolint:exhaustive // We don't need to test all types.
	tests := []struct {
		name     string
		input    map[AType]bool
		expected bool
	}{
		{"only photo", map[AType]bool{Photo: true}, true},
		{"only video", map[AType]bool{Video: true}, true},
		{"photo and video", map[AType]bool{Photo: true, Video: true}, true},
		{"with document", map[AType]bool{Photo: true, Document: true}, false},
		{"with audio", map[AType]bool{Video: true, Audio: true}, false},
		{"empty map", map[AType]bool{}, true},
		{"nil map", nil, true},
		{"multiple invalid types", map[AType]bool{Audio: true, Document: true}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isOnlyPhotoOrVideo(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
