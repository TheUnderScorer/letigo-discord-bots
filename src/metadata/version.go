package metadata

var Version string

func GetVersion() string {
	if Version != "" {
		return Version
	}

	return "dev"
}
