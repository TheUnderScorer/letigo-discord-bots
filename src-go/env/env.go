package env

import (
	"github.com/caarlos0/env/v11"
	"os"
)

type appEnv struct {
	BotToken                string `env:"BOT_TOKEN"`
	AppId                   string `env:"APP_ID"`
	GuildId                 string `env:"GUILD_ID"`
	GreetingChannelId       string `env:"GREETING_CHANNEL_ID"`
	DailyReportChannelId    string `env:"DAILY_REPORT_CHANNEL_ID"`
	DailyReportTargetUserId string `env:"DAILY_REPORT_TARGET_USER_ID"`
	Env                     string `env:"GO_ENV"`
	YouTubeApiKey           string `env:"YT_API_KEY"`
}

var Cfg appEnv

func Init() {
	if err := env.Parse(&Cfg); err != nil {
		panic(err)
	}
}

func IsProd() bool {
	return os.Getenv("GO_ENV") == "production"
}
