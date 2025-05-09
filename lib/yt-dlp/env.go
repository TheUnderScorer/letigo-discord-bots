package ytdlp

import envutil "github.com/caarlos0/env/v11"

type env struct {
	CookiesTxtPath string `env:"COOKIES_TXT_PATH"`
}

var cachedEnv env
var didParse = false

func getEnv() env {
	if !didParse {
		err := envutil.Parse(&cachedEnv)
		if err != nil {
			panic(err)
		}
		didParse = true
	}

	return cachedEnv
}
