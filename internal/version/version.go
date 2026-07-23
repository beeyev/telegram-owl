// Package version exposes build-time version metadata.
package version

// version is replaced by GoReleaser through -ldflags -X. The fallback makes
// local, unversioned builds explicit instead of looking like a release.
var version = "v0.0.0@undefined"

// Version returns the build version exactly as injected by the linker.
func Version() string {
	return version
}
