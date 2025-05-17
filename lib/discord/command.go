package discord

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"lib/util/arrayutil"
)

type Command struct {
	Name        string
	Description string
	SubCommands []SubCommand
}

func (b *Command) ToApplicationCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        b.Name,
		Description: b.Description,
		Options: arrayutil.Map(b.SubCommands, func(subCommand SubCommand) *discordgo.ApplicationCommandOption {
			return subCommand.ToApplicationCommandOption()
		}),
	}
}

func (b *Command) Handle(bot *Bot, interaction *discordgo.InteractionCreate) {
	if interaction.Type != discordgo.InteractionApplicationCommand {
		return
	}

	ctx, cancel := NewInteractionContext(context.Background())
	defer cancel()
	data := interaction.ApplicationCommandData()
	name := data.Name

	if name == b.Name {
		options := interaction.ApplicationCommandData().Options

		for _, option := range options {
			subCommandName := option.Name
			subCommand, ok := arrayutil.Find(b.SubCommands, func(subCommand SubCommand) bool {
				return subCommand.Name == subCommandName
			})

			if ok {
				bot.StartLoadingInteractionAndForget(interaction.Interaction)
				cmdInteraction := CommandInteractionOptions{
					optionsMap:         make(map[string]*ResolvedCommandOption),
					interactionOptions: option.Options,
				}

				log.Debug("handling subcommand", zap.Any("subcommand", subCommand), zap.Any("interaction", interaction))

				err := subCommand.Handle(ctx, cmdInteraction, interaction)
				if err != nil {
					log.Error("interaction failed", zap.Error(err), zap.Any("interaction", interaction))

					bot.FollowUpInteractionErrorReply(err, interaction.Interaction)
				}
			}
		}
	}
}
