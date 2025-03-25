package domain

import (
	"app/bots"
	"app/domain/chat"
	"app/domain/interaction"
	"app/logging"
	"context"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

var logger = logging.Get().Named("domain")

func Init(ctx context.Context) {
	for _, bot := range bots.GetAllFromContext(ctx) {
		session := bot.Session

		session.AddHandler(func(s *discordgo.Session, m *discordgo.InteractionCreate) {
			go interaction.Handle(s, bot.Name, m, ctx)
		})

		session.AddHandler(func(s *discordgo.Session, r *discordgo.GuildMembersChunk) {
			logger.Info("got members chunk", zap.Any("r", r))
		})

		session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
			logger.Info("connected")
		})
	}

	InitWojciechBot(ctx)
}

func InitWojciechBot(ctx context.Context) {
	bot := ctx.Value(bots.BotNameWojciech).(*bots.Bot)
	session := bot.Session

	session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		chat.HandleMessageCreate(ctx, m)
	})
}
