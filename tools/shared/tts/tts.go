package tts

type Manifest map[string]ManifestEntry

type ManifestEntry struct {
	// FileName contains file path relative to outDir
	FileName string `json:"file_name"`
}
