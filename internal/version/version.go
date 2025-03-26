package version

// version value will be set during compilation.
var version = "v0.0.0@undefined"

func Version() string {
	return version
}
