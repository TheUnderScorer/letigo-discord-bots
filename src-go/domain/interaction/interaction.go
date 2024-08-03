package interaction

import (
	"context"
	"errors"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"src-go/discord"
	"src-go/domain/player"
	"src-go/env"
	errors2 "src-go/errors"
	"src-go/logging"
	"src-go/messages"
	"src-go/util"
	"strconv"
)

type Command string
type DjSubCommand string
type DjQueueOption string

const (
	CommandDj Command = "dj"

	DjSubCommandList       DjSubCommand = "list"
	DjSubCommandNext       DjSubCommand = "next"
	DjSubCommandPlay       DjSubCommand = "odtwarzaj"
	DjSubCommandPause      DjSubCommand = "pauza"
	DjSubCommandClearQueue DjSubCommand = "wyczysc-kolejke"
	DjSubCommandQueue      DjSubCommand = "dodaj"

	DjQueueOptionSong DjQueueOption = "piosenka"
)

var logger = logging.Get().Named("interaction")

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        string(CommandDj),
		Description: "Pobaw się w DJa!",
		Type:        discordgo.ChatApplicationCommand,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        string(DjSubCommandList),
				Description: "Pokaż listę utworów",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
			{
				Name:        string(DjSubCommandQueue),
				Description: "Dodaj utwór do kolejki",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        string(DjQueueOptionSong),
						Description: "URL do utworu z youtube",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
				},
			},
			{
				Name:        string(DjSubCommandNext),
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Description: "Zagraj następny utwór",
			},
			{
				Name:        string(DjSubCommandPlay),
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Description: "Zagraj utwór",
			},
			{
				Name:        string(DjSubCommandPause),
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Description: "Zatrzymaj odtwarzanie",
			},
			{
				Name:        string(DjSubCommandClearQueue),
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Description: "Wyczyść kolejkę",
			},
		},
	},
}

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, ctx context.Context){
	string(CommandDj): func(s *discordgo.Session, i *discordgo.InteractionCreate, ctx context.Context) {
		playerManager := ctx.Value(player.ChannelPlayerContextKey).(*player.ChannelPlayerManager)
		var log = logger.With(zap.String("command", string(CommandDj)))

		options := i.ApplicationCommandData().Options
		log.Info("handling command", zap.Any("options", options))

		channel, err := s.Channel(i.ChannelID)
		if err != nil {
			log.Error("failed to get channel", zap.Error(err))
			return
		}

		if channel.Type != discordgo.ChannelTypeGuildVoice {
			log.Error("channel is not a voice channel")

			discord.ReplyToInteractionAndForget(s, i.Interaction, &discord.InteractionReply{
				Content:   messages.Messages.MustBeInVoiceChannel,
				Ephemeral: true,
			})

			return
		}

		player, err := playerManager.GetOrCreate(s, i.ChannelID)
		if err != nil {
			go discord.FollowupInteractionMessageAndForget(s, i.Interaction, &discord.InteractionReply{
				Content:   messages.Messages.Player.FailedToQueue,
				Ephemeral: true,
			})

			log.Error("failed to get player", zap.Error(err))
			return
		}

		switch options[0].Name {
		case string(DjSubCommandClearQueue):
			go discord.StartLoadingInteractionAndForget(s, i.Interaction)
			player.ClearQueue()
			go discord.DeleteFollowupAndForget(s, i.Interaction)

		case string(DjSubCommandPause):
			go discord.StartLoadingInteractionAndForget(s, i.Interaction)
			err = player.Pause()
			if err != nil {
				log.Error("failed to pause", zap.Error(err))
			}
			go discord.DeleteFollowupAndForget(s, i.Interaction)

		case string(DjSubCommandList):
			discord.StartLoadingInteractionAndForget(s, i.Interaction)
			message := player.ListQueueForDisplay()
			if message == "" {
				message = messages.Messages.Player.NoMoreSongs
			}
			discord.FollowupInteractionMessageAndForget(s, i.Interaction, &discord.InteractionReply{
				Content: message,
			})

		case string(DjSubCommandNext):
			go discord.StartLoadingInteractionAndForget(s, i.Interaction)
			err = player.Next()

			if err != nil {
				log.Error("failed to play next", zap.Error(err))

				var userFriendly *errors2.UserFriendlyError

				if errors.As(err, &userFriendly) {
					discord.FollowupInteractionErrorAndForget(s, i.Interaction, err)

					return
				}
			}
			go discord.DeleteFollowupAndForget(s, i.Interaction)

		case string(DjSubCommandPlay):
			go discord.StartLoadingInteractionAndForget(s, i.Interaction)
			err = player.Play()
			if err != nil {
				log.Error("failed to play", zap.Error(err))

				var userFriendly *errors2.UserFriendlyError

				if errors.As(err, &userFriendly) {
					discord.FollowupInteractionErrorAndForget(s, i.Interaction, err)

					return
				}
			}

			go discord.DeleteFollowupAndForget(s, i.Interaction)

		case string(DjSubCommandQueue):
			// Or convert the slice into a map
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options[0].Options))
			for _, opt := range options[0].Options {
				optionMap[opt.Name] = opt
			}

			go discord.StartLoadingInteractionAndForget(s, i.Interaction)

			if opt, ok := optionMap[string(DjQueueOptionSong)]; ok {
				order, err := player.AddToQueue(opt.StringValue(), i.Member.User.ID)

				if err != nil {
					log.Error("failed to queue song", zap.Error(err))

					go discord.FollowupInteractionMessageAndForget(s, i.Interaction, &discord.InteractionReply{
						Content:   messages.Messages.Player.FailedToQueue,
						Ephemeral: true,
					})
					return
				}

				var message string
				if order == 0 {
					message = messages.Messages.Player.AddedToQueueAsNext
				} else {
					message = util.ApplyTokens(util.RandomElement(messages.Messages.Player.AddedToQueue), map[string]string{
						"INDEX": strconv.Itoa(order),
					})
				}

				go discord.FollowupInteractionMessageAndForget(s, i.Interaction, &discord.InteractionReply{
					Content: message,
				})
			}

		}
	},
}

func Init(s *discordgo.Session) {
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, env.Cfg.GuildId, v)
		if err != nil {
			logger.Fatal("failed to create command", zap.String("command", v.Name), zap.Error(err))
		}
		registeredCommands[i] = cmd
	}
}

func Handle(s *discordgo.Session, i *discordgo.InteractionCreate, ctx context.Context) {
	data := i.ApplicationCommandData()
	name := data.Name
	log := logger.With(zap.String("command", name)).With(zap.Any("data", data))

	log.Info("got command")
	if h, ok := commandHandlers[name]; ok {
		h(s, i, ctx)
	} else {
		log.Error("unknown command")
	}
}
