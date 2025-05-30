package env

import (
	"github.com/caarlos0/env/v11"
	"strconv"
)

type appEnv struct {
	BotToken                     string `env:"BOT_TOKEN"`
	AppId                        string `env:"APP_ID"`
	GuildId                      string `env:"GUILD_ID"`
	GreetingChannelId            string `env:"GREETING_CHANNEL_ID"`
	DailyReportChannelId         string `env:"DAILY_REPORT_CHANNEL_ID"`
	DailyReportTargetUserId      string `env:"DAILY_REPORT_TARGET_USER_ID"`
	Env                          string `env:"GO_ENV"`
	YouTubeApiKey                string `env:"YT_API_KEY"`
	S3Bucket                     string `env:"S3_BUCKET"`
	GPTServerHost                string `env:"GPT_SERVER_HOST"`
	OllamaHost                   string `env:"OLLAMA_HOST"`
	OllamaModel                  string `env:"OLLAMA_MODEL"`
	OllamaVisionModel            string `env:"OLLAMA_VISION_MODEL"`
	OpenAIApiKey                 string `env:"OPENAI_API_KEY"`
	OpenAIAssistantID            string `env:"OPENAI_ASSISTANT_ID"`
	OpenAIAssistantVectorStoreID string `env:"OPENAI_ASSISTANT_VECTOR_STORE_ID"`
	AllMessagesReplyWorthy       string `env:"ALL_MESSAGES_REPLY_WORTHY"`
}

func (e *appEnv) AreAllMessagesReplyWorthy() bool {
	val, err := strconv.ParseBool(e.AllMessagesReplyWorthy)
	if err != nil {
		return false
	}
	return val
}

var Env appEnv

func Init() {
	if err := env.Parse(&Env); err != nil {
		panic(err)
	}
}
