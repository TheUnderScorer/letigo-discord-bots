package main

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"lib/discord"
	"wojciech-bot/player"
)

const DjQueueOptionSong = "piosenka"

func NewDJCommand(interactions *player.Interactions) discord.Command {
	return discord.Command{
		Name:        "dj",
		Description: "Pobaw się w DJa!",
		SubCommands: []discord.SubCommand{
			{
				Name:        "list",
				Description: "Pokaż listę utworów",
				Handler: func(ctx context.Context, options discord.CommandInteractionOptions, interaction *discordgo.InteractionCreate) error {
					return interactions.List(ctx, interaction.Interaction)
				},
			},
			{
				Name:        "dodaj",
				Description: "Dodaj utwór do kolejki",
				Options: []discord.CommandOption{
					{
						Name:        DjQueueOptionSong,
						Description: "URL do utworu z youtube",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
				},
				Handler: func(ctx context.Context, options discord.CommandInteractionOptions, interaction *discordgo.InteractionCreate) error {
					songUrl := options.Option(DjQueueOptionSong).String()
					return interactions.Queue(ctx, interaction.Interaction, songUrl)
				},
			},
			{
				Name:        "next",
				Description: "Odtwórz następny utwór w kolejce",
				Handler: func(ctx context.Context, options discord.CommandInteractionOptions, interaction *discordgo.InteractionCreate) error {
					return interactions.Next(ctx, interaction.Interaction)
				},
			},
			{
				Name:        "odtwarzaj",
				Description: "Rozpocznij odtwarzanie utworu/kolejki",
				Handler: func(ctx context.Context, options discord.CommandInteractionOptions, interaction *discordgo.InteractionCreate) error {
					return interactions.Play(ctx, interaction.Interaction)
				},
			},
			{
				Name:        "pauza",
				Description: "Wstrzymaj odtwarzanie",
				Handler: func(ctx context.Context, options discord.CommandInteractionOptions, interaction *discordgo.InteractionCreate) error {
					return interactions.Pause(ctx, interaction.Interaction)
				},
			},
			{
				Name:        "wyczysc-kolejke",
				Description: "Wyczyść kolejkę",
				Handler: func(ctx context.Context, options discord.CommandInteractionOptions, interaction *discordgo.InteractionCreate) error {
					return interactions.ClearQueue(ctx, interaction.Interaction)
				},
			},
			{
				Name:        "odtwarzacz",
				Description: "Pokaż obecny status odtwarzacza",
				Handler: func(ctx context.Context, options discord.CommandInteractionOptions, interaction *discordgo.InteractionCreate) error {
					return interactions.Player(ctx, interaction.Interaction)
				},
			},
		},
	}
}
