package interaction

import (
	"app/bots"
	"app/discord"
	"app/env"
	"app/logging"
	"context"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type Command string
type DjSubCommand string
type DjQueueOption string
type TriviaSubCommand string

const (
	CommandDj     Command = "dj"
	CommandTrivia Command = "teleturniej"

	DjSubCommandList       DjSubCommand = "list"
	DjSubCommandNext       DjSubCommand = "next"
	DjSubCommandPlay       DjSubCommand = "odtwarzaj"
	DjSubCommandPause      DjSubCommand = "pauza"
	DjSubCommandClearQueue DjSubCommand = "wyczysc-kolejke"
	DjSubCommandQueue      DjSubCommand = "dodaj"
	DjSubCommandPlayer     DjSubCommand = "odtwarzacz"

	TriviaSubCommandStart  TriviaSubCommand = "start"
	TriviaSubCommandPoints TriviaSubCommand = "punkty"

	DjQueueOptionSong DjQueueOption = "piosenka"
)

var logger = logging.Get().Named("interaction")

func Init(ctx context.Context) {
	var registeredCommands []*discordgo.ApplicationCommand

	for _, botName := range bots.Bots {
		bot := ctx.Value(botName).(*bots.Bot)
		commands := commands[bot.Name]

		unregisterCommands(bot.Session, commands)

		for _, v := range commands {
			cmd, err := bot.Session.ApplicationCommandCreate(bot.Session.State.User.ID, env.Env.GuildId, v)
			if err != nil {
				logger.Error("failed to create command", zap.String("command", v.Name), zap.Error(err))
			}
			registeredCommands = append(registeredCommands, cmd)
		}

	}

	logger.Info("registered commands", zap.Any("commands", registeredCommands))
}

func unregisterCommands(s *discordgo.Session, commands []*discordgo.ApplicationCommand) {
	for _, v := range commands {
		err := s.ApplicationCommandDelete(s.State.User.ID, env.Env.GuildId, v.ID)
		if err != nil {
			logger.Error("failed to delete command", zap.String("command", v.Name), zap.Error(err))
		}
	}
}

func Handle(s *discordgo.Session, botName bots.BotName, i *discordgo.InteractionCreate, ctx context.Context) {
	if i.Type == discordgo.InteractionMessageComponent {
		logger.Info("got component")

		go discord.RespondToInteractionAndForget(s, i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})

		HandleComponentInteraction(ctx, s, i.ChannelID, i)

		return
	}

	data := i.ApplicationCommandData()
	name := data.Name
	log := logger.With(zap.String("command", name)).With(zap.Any("data", data))

	log.Info("got command")
	if h, ok := commandHandlers[botName][name]; ok {
		h(s, i, ctx)
	} else {
		log.Error("unknown command")
	}
}
