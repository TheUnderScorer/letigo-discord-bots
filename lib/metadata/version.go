package metadata

// Version usually is set during build via: -ldflags "-X lib/metadata.Version=${VERSION}"
var Version string

func GetVersion() string {
	if Version != "" {
		return Version
	}

	return "dev"
}
