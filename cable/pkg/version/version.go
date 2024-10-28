package version

var (
	version string
	sha     string
)

func init() {
	if version == "" {
		version = "0.2.0"
	}

	if sha != "" {
		version = version + "-" + sha
	}
}

// Version returns the current program version
func Version() string {
	return version
}

// SHA returns the build commit sha
func SHA() string {
	return sha
}
