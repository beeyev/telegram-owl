package attachment

import "math"

// BytesPerMegabyte uses the binary unit expected by the configured attachment
// limits and their user-facing MB messages.
const BytesPerMegabyte = 1024 * 1024

// bytesToMegabytes converts bytes to human-readable megabytes.
func bytesToMegabytes(size int64) int {
	return int(math.Round(float64(size) / float64(BytesPerMegabyte)))
}

// isOnlyPhotoOrVideo reports whether Telegram can send the detected types in a
// single mixed media group. Telegram permits photos and videos to be mixed,
// while documents and audio require a uniform group.
func isOnlyPhotoOrVideo(types map[AType]bool) bool {
	for t := range types {
		if t != Photo && t != Video {
			return false
		}
	}

	return true
}
