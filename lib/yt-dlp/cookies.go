package ytdlp

import (
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

func getCookiesArgs() []string {
	cookiesTxtPath := getEnv().CookiesTxtPath
	if cookiesTxtPath == "" {
		return []string{}
	}

	_, err := os.ReadFile(cookiesTxtPath)
	if err != nil {
		log.Warn("provided cookies file does not exist")
		return []string{}
	}

	absolutePath, err := filepath.Abs(cookiesTxtPath)
	if err != nil {
		log.Warn("failed to get absolute path of cookies file", zap.Error(err))
		return []string{}
	}

	return []string{
		"--cookies", absolutePath,
	}
}
