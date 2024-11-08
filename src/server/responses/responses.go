package responses

type VersionInfo struct {
	Result  bool   `json:"result"`
	Version string `json:"version"`
}
