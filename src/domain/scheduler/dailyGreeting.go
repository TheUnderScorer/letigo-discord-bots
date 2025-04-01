package scheduler

import (
	"app/bots"
	"app/discord"
	"app/env"
	"app/logging"
	"app/messages"
	"app/util/arrayutil"
	"go.uber.org/zap"
	"time"
)

var log = logging.Get().Named("DailyGreeting")

func DailyGreeting(bot *bots.Bot) {
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
	discord.SendMessageAndForget(bot.Session, cid, message)
}
