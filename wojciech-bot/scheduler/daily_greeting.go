package scheduler

import (
	"go.uber.org/zap"
	"lib/discord"
	"lib/logging"
	"lib/util/arrayutil"
	"time"
	"wojciech-bot/env"
	"wojciech-bot/messages"
)

var log = logging.Get().Named("DailyGreeting")

func DailyGreeting(bot *discord.Bot) {
	cid := env.Env.GreetingChannelId
	if cid == "" {
		log.Error("no daily report channel id set")
		return
	}

	day := int(time.Now().Weekday())
	dayMessages := messages.Messages.Greetings[day]
	if !arrayutil.IsValidArray(dayMessages) {
		log.Error("no messages found for given day", zap.Int("day", day))
		return
	}

	message := arrayutil.RandomElement(dayMessages)
	bot.SendMessageAndForget(cid, message)
}
