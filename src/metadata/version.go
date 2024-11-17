package metadata

// Version usually is set during build via: -ldflags "-X app/metadata.Version=${VERSION}"
var Version string

func GetVersion() string {
	if Version != "" {
		return Version
	}

	return "dev"
}
