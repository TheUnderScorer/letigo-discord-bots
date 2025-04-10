package env

import (
	"github.com/caarlos0/env/v11"
	"os"
)

type appEnv struct {
	WojciechBotToken             string `env:"WOJCIECH_BOT_TOKEN"`
	TadeuszBotToken              string `env:"TADEUSZ_BOT_TOKEN"`
	WojciechBotAppId             string `env:"WOJCIECH_BOT_APP_ID"`
	TadeuszBotAppId              string `env:"TADEUSZ_BOT_APP_ID"`
	GuildId                      string `env:"GUILD_ID"`
	GreetingChannelId            string `env:"GREETING_CHANNEL_ID"`
	DailyReportChannelId         string `env:"DAILY_REPORT_CHANNEL_ID"`
	DailyReportTargetUserId      string `env:"DAILY_REPORT_TARGET_USER_ID"`
	Env                          string `env:"GO_ENV"`
	YouTubeApiKey                string `env:"YT_API_KEY"`
	TTSHost                      string `env:"TTS_HOST"`
	S3Bucket                     string `env:"S3_BUCKET"`
	GPTServerHost                string `env:"GPT_SERVER_HOST"`
	OllamaHost                   string `env:"OLLAMA_HOST"`
	OllamaModel                  string `env:"OLLAMA_MODEL"`
	OpenAIApiKey                 string `env:"OPENAI_API_KEY"`
	OpenAIAssistantID            string `env:"OPENAI_ASSISTANT_ID"`
	OpenAIAssistantVectorStoreID string `env:"OPENAI_ASSISTANT_VECTOR_STORE_ID"`
}

var Env appEnv

func Init() {
	if err := env.Parse(&Env); err != nil {
		panic(err)
	}
}

func IsProd() bool {
	return os.Getenv("GO_ENV") == "production"
}
