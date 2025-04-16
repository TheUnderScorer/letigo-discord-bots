package interaction

import (
	"app/bots"
	"app/env"
	"app/logging"
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

func Init(bots []*bots.Bot) {
	var registeredCommands []*discordgo.ApplicationCommand

	for _, bot := range bots {
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

func Handle(session *discordgo.Session, botName bots.BotName, interaction *discordgo.InteractionCreate, container *CommandsContainer) {
	if interaction.Type == discordgo.InteractionMessageComponent {
		logger.Info("got component")

		HandleComponentInteraction(container.ComponentInteractionHandlers, session, interaction)

		return
	}

	data := interaction.ApplicationCommandData()
	name := data.Name
	log := logger.With(zap.String("command", name)).With(zap.Any("data", data))

	log.Info("got command")
	if h, ok := commandHandlers[botName][name]; ok {
		h(session, interaction, container)
	} else {
		log.Error("unknown command")
	}
}
