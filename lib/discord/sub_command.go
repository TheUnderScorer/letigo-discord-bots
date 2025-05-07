package discord

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"lib/util/arrayutil"
)

type SubCommandHandler func(ctx context.Context, options CommandInteractionOptions, interaction *discordgo.InteractionCreate) error

type SubCommand struct {
	Name        string
	Description string
	Options     []CommandOption
	Handler     SubCommandHandler
}

func (c *SubCommand) Handle(ctx context.Context, options CommandInteractionOptions, interaction *discordgo.InteractionCreate) error {
	return c.Handler(ctx, options, interaction)
}

func (c *SubCommand) ToApplicationCommandOption() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Name:        c.Name,
		Description: c.Description,
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Options: arrayutil.Map(c.Options, func(option CommandOption) *discordgo.ApplicationCommandOption {
			return option.ToApplicationCommandOption()
		}),
	}
}
