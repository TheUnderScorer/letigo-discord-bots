package scheduler

import (
	"app/bots"
	"app/discord"
	"app/env"
	"app/logging"
	"app/messages"
	"app/util"
	"context"
	"go.uber.org/zap"
	"time"
)

var log = logging.Get().Named("DailyGreeting")

func DailyGreeting(ctx context.Context) {
	cid := env.Env.GreetingChannelId
	if cid == "" {
		log.Error("no daily report channel id set")
		return
	}

	bot := ctx.Value(bots.BotNameWojciech).(*bots.Bot)

	day := int(time.Now().Weekday())
	dayMessages := messages.Messages.Greetings[day]
	if !util.IsValidArray(dayMessages) {
		log.Error("no messages found for given day", zap.Int("day", day))
		return
	}

	message := util.RandomElement(dayMessages)
	discord.SendMessageAndForget(bot.Session, cid, message)
}
