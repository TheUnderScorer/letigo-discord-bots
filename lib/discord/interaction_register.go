package discord

import (
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"lib/util/arrayutil"
)

func RegisterCommands(bot *Bot, guildId string, commands ...Command) {
	var registeredCommands []*discordgo.ApplicationCommand

	discordCommands := arrayutil.Map(commands, func(command Command) *discordgo.ApplicationCommand {
		return command.ToApplicationCommand()
	})

	unregisterCommands(bot, guildId, discordCommands)

	for _, command := range discordCommands {
		cmd, err := bot.ApplicationCommandCreate(bot.Session.State.User.ID, guildId, command)
		if err != nil {
			logger.Error("failed to create command", zap.String("command", command.Name), zap.Error(err))
		}
		registeredCommands = append(registeredCommands, cmd)
	}

	logger.Info("registered commands", zap.Any("commands", registeredCommands))
}

func unregisterCommands(bot *Bot, guildId string, commands []*discordgo.ApplicationCommand) {
	for _, v := range commands {
		err := bot.ApplicationCommandDelete(bot.State.User.ID, guildId, v.ID)
		if err != nil {
			log.Error("failed to delete command", zap.String("command", v.Name), zap.Error(err))
		}
	}
}
