package interaction

import (
	"app/aws"
	"app/bots"
	"app/discord"
	"app/domain/player"
	"app/domain/trivia"
	"app/errors"
	"app/messages"
	"app/util"
	"app/util/arrayutil"
	errors2 "errors"
	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"go.uber.org/zap"
	"strconv"
)

type CommandsContainer struct {
	TriviaManager                *trivia.Manager
	S3                           *aws.S3
	ChannelPlayerManager         *player.ChannelPlayerManager
	ComponentInteractionHandlers []discord.ComponentInteractionHandler
}

var commands = map[bots.BotName][]*discordgo.ApplicationCommand{
	bots.BotNameTadeuszSznuk: {
		{
			Name:        string(CommandTrivia),
			Description: "Jeden z dziesięciu (beta).",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        string(TriviaSubCommandStart),
					Description: "Rozpocznij 1 dzban z 10 (beta).",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        string(TriviaSubCommandPoints),
					Description: "Pokaż ilość punktów.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
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

type CommandHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, container *CommandsContainer)

var commandHandlers = map[bots.BotName]CommandHandlers{
	bots.BotNameTadeuszSznuk: {
		string(CommandTrivia): func(s *discordgo.Session, i *discordgo.InteractionCreate, container *CommandsContainer) {
			var err error
			defer ReplyToError(err, s, i.Interaction)

			discord.StartLoadingInteractionAndForget(s, i.Interaction)

			options := i.ApplicationCommandData().Options
			log.Info("handling trivia command", zap.Any("options", options))

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

			triviaManager := container.TriviaManager
			trivia, err := triviaManager.GetOrCreate(container.S3, s, i.ChannelID)
			if err != nil {
				logger.Error("failed to get trivia", zap.Error(err))

				return
			}

			switch options[0].Name {
			case string(TriviaSubCommandStart):
				err = trivia.Start()

				logger.Info("started trivia")

				if err == nil {
					discord.DeleteFollowupAndForget(s, i.Interaction)
				}

			case string(TriviaSubCommandPoints):
				trivia.SendPointsMessage()

				discord.DeleteFollowupAndForget(s, i.Interaction)
			}
		},
	},
	bots.BotNameWojciech: {
		string(CommandDj): func(s *discordgo.Session, i *discordgo.InteractionCreate, container *CommandsContainer) {
			var err error
			defer ReplyToError(err, s, i.Interaction)

			discord.StartLoadingInteractionAndForget(s, i.Interaction)

			playerManager := container.ChannelPlayerManager
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
				discord.FollowupInteractionMessageAndForget(s, i.Interaction, &discord.InteractionReply{
					Content:   messages.Messages.Player.FailedToQueue,
					Ephemeral: true,
				})

				log.Error("failed to get channelPlayer", zap.Error(err))
				return
			}

			switch options[0].Name {
			case string(DjSubCommandClearQueue):
				channelPlayer.ClearQueue()
				discord.DeleteFollowupAndForget(s, i.Interaction)

			case string(DjSubCommandPause):
				err = channelPlayer.Pause()
				if err != nil {
					log.Error("failed to pause", zap.Error(err))
				}
				discord.DeleteFollowupAndForget(s, i.Interaction)

			case string(DjSubCommandList):
				message := channelPlayer.ListQueueForDisplay()
				if message == "" {
					message = messages.Messages.Player.NoMoreSongs
				}
				discord.FollowupInteractionMessageAndForget(s, i.Interaction, &discord.InteractionReply{
					Content: message,
				})

			case string(DjSubCommandNext):
				err = channelPlayer.Next()

				if err != nil {
					log.Error("failed to play next", zap.Error(err))

					var userFriendly *errors.PublicError

					if errors2.As(err, &userFriendly) {
						discord.FollowupInteractionErrorAndForget(s, i.Interaction, err)

						return
					}
				}
				discord.DeleteFollowupAndForget(s, i.Interaction)

			case string(DjSubCommandPlayer):
				component, err := player.GetPlayerComponent(channelPlayer)
				if err != nil {
					log.Error("failed to get player component", zap.Error(err))

					go discord.FollowupInteractionMessageAndForget(s, i.Interaction, &discord.InteractionReply{
						Content: messages.Messages.UnknownError,
					})

					return
				}

				discord.SendMessageComplexAndForget(s, i.ChannelID, &discordgo.MessageSend{
					Components: *component,
				})
				discord.DeleteFollowupAndForget(s, i.Interaction)

			case string(DjSubCommandPlay):
				err = channelPlayer.Play()
				if err != nil {
					log.Error("failed to play", zap.Error(err))

					var userFriendly *errors.PublicError

					if errors2.As(err, &userFriendly) {
						discord.FollowupInteractionErrorAndForget(s, i.Interaction, err)

						return
					}
				}

				discord.DeleteFollowupAndForget(s, i.Interaction)

			case string(DjSubCommandQueue):
				optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options[0].Options))
				for _, opt := range options[0].Options {
					optionMap[opt.Name] = opt
				}

				if opt, ok := optionMap[string(DjQueueOptionSong)]; ok {
					order, err := channelPlayer.AddToQueue(opt.StringValue(), i.Member.User.ID)

					if err != nil {
						log.Error("failed to queue song", zap.Error(err))

						discord.FollowupInteractionMessageAndForget(s, i.Interaction, &discord.InteractionReply{
							Content:   messages.Messages.Player.FailedToQueue,
							Ephemeral: true,
						})
						return
					}

					var message string
					if order == 0 {
						message = messages.Messages.Player.AddedToQueueAsNext
					} else {
						message = util.ApplyTokens(arrayutil.RandomElement(messages.Messages.Player.AddedToQueue), map[string]string{
							"INDEX": strconv.Itoa(order),
						})
					}

					discord.FollowupInteractionMessageAndForget(s, i.Interaction, &discord.InteractionReply{
						Content: message,
					})
				}

			}
		},
	},
}
