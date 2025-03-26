package attachment

import "math"

const BytesPerMegabyte = 1024 * 1024

// bytesToMegabytes converts bytes to human-readable megabytes.
func bytesToMegabytes(size int64) int {
	return int(math.Round(float64(size) / float64(BytesPerMegabyte)))
}

// isOnlyPhotoOrVideo returns true if *every* type in the set is either Photo or Video.
func isOnlyPhotoOrVideo(types map[AType]bool) bool {
	for t := range types {
		if t != Photo && t != Video {
			return false
		}
	}

	return true
}
