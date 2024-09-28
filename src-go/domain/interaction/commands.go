package interaction

import (
	"context"
	errors2 "errors"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"src-go/bots"
	"src-go/discord"
	"src-go/domain/player"
	"src-go/domain/trivia"
	"src-go/errors"
	"src-go/messages"
	"src-go/util"
	"strconv"
)

var commands = map[bots.BotName][]*discordgo.ApplicationCommand{
	bots.BotNameTadeuszSznuk: {
		{
			Name:        string(CommandTrivia),
			Description: "Rozpocznij Jeden z dziesięciu (beta).",
			Type:        discordgo.ChatApplicationCommand,
		},
	},
	bots.BotNameWojciech: {
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
					Name:        string(DjSubCommandPlayer),
					Description: "Pokaż odtwarzacz",
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
	},
}

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, ctx context.Context){
	string(CommandTrivia): func(s *discordgo.Session, i *discordgo.InteractionCreate, ctx context.Context) {
		triviaManager := ctx.Value(trivia.ManagerContextKey).(*trivia.Manager)
		if triviaManager == nil {
			logger.Error("trivia manager is nil")
			return
		}

		trivia, err := triviaManager.GetOrCreate(s, i.ChannelID)
		if err != nil {
			logger.Error("failed to get trivia", zap.Error(err))
			return
		}

		go discord.StartLoadingInteractionAndForget(s, i.Interaction)

		err = trivia.Start()
		if err != nil {
			var userFriendly *errors.UserFriendlyError
			if errors2.As(err, &userFriendly) {
				discord.FollowupInteractionErrorAndForget(s, i.Interaction, err)
			} else {
				discord.DeleteFollowupAndForget(s, i.Interaction)
			}

			return
		}
	},
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

		channelPlayer, err := playerManager.GetOrCreate(s, i.ChannelID)
		if err != nil {
			go discord.FollowupInteractionMessageAndForget(s, i.Interaction, &discord.InteractionReply{
				Content:   messages.Messages.Player.FailedToQueue,
				Ephemeral: true,
			})

			log.Error("failed to get channelPlayer", zap.Error(err))
			return
		}

		switch options[0].Name {
		case string(DjSubCommandClearQueue):
			go discord.StartLoadingInteractionAndForget(s, i.Interaction)
			channelPlayer.ClearQueue()
			go discord.DeleteFollowupAndForget(s, i.Interaction)

		case string(DjSubCommandPause):
			go discord.StartLoadingInteractionAndForget(s, i.Interaction)
			err = channelPlayer.Pause()
			if err != nil {
				log.Error("failed to pause", zap.Error(err))
			}
			go discord.DeleteFollowupAndForget(s, i.Interaction)

		case string(DjSubCommandList):
			discord.StartLoadingInteractionAndForget(s, i.Interaction)
			message := channelPlayer.ListQueueForDisplay()
			if message == "" {
				message = messages.Messages.Player.NoMoreSongs
			}
			discord.FollowupInteractionMessageAndForget(s, i.Interaction, &discord.InteractionReply{
				Content: message,
			})

		case string(DjSubCommandNext):
			go discord.StartLoadingInteractionAndForget(s, i.Interaction)
			err = channelPlayer.Next()

			if err != nil {
				log.Error("failed to play next", zap.Error(err))

				var userFriendly *errors.UserFriendlyError

				if errors2.As(err, &userFriendly) {
					discord.FollowupInteractionErrorAndForget(s, i.Interaction, err)

					return
				}
			}
			go discord.DeleteFollowupAndForget(s, i.Interaction)

		case string(DjSubCommandPlayer):
			go discord.StartLoadingInteractionAndForget(s, i.Interaction)
			component, err := player.GetPlayerComponent(channelPlayer)
			if err != nil {
				log.Error("failed to get player component", zap.Error(err))

				go discord.FollowupInteractionMessageAndForget(s, i.Interaction, &discord.InteractionReply{
					Content: messages.Messages.UnknownError,
				})

				return
			}

			go discord.SendMessageComplexAndForget(s, i.Interaction.ChannelID, &discordgo.MessageSend{
				Components: *component,
			})
			go discord.DeleteFollowupAndForget(s, i.Interaction)

		case string(DjSubCommandPlay):
			go discord.StartLoadingInteractionAndForget(s, i.Interaction)
			err = channelPlayer.Play()
			if err != nil {
				log.Error("failed to play", zap.Error(err))

				var userFriendly *errors.UserFriendlyError

				if errors2.As(err, &userFriendly) {
					discord.FollowupInteractionErrorAndForget(s, i.Interaction, err)

					return
				}
			}

			go discord.DeleteFollowupAndForget(s, i.Interaction)

		case string(DjSubCommandQueue):
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options[0].Options))
			for _, opt := range options[0].Options {
				optionMap[opt.Name] = opt
			}

			go discord.StartLoadingInteractionAndForget(s, i.Interaction)

			if opt, ok := optionMap[string(DjQueueOptionSong)]; ok {
				order, err := channelPlayer.AddToQueue(opt.StringValue(), i.Member.User.ID)

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
