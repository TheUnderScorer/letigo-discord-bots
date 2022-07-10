import {
  CommandDefinition,
  Commands,
  KolegoInsultOptions,
  KolegoQuestionOptions,
  KolegoSubcommand,
} from '../command.types';
import { SlashCommandBuilder } from '@discordjs/builders';
import { insultSubCommandHandler } from './subCommands/insult';
import { questionSubCommandHandler } from './subCommands/question';
import { coTamSubCommandHandler } from './subCommands/cotam';

export const kolegoCommand: CommandDefinition = {
  data: new SlashCommandBuilder()
    .setName(Commands.Kolego)
    .setDescription('Wywołaj Wojciecha')
    .addSubcommand(subCommand =>
      subCommand
        .setName(KolegoSubcommand.Question)
        .setDescription('Zadaj pytanie Wojciechowi')
        .addStringOption(option =>
          option
            .setName(KolegoQuestionOptions.Question)
            .setDescription('Twoje pytanie')
            .setRequired(true)
        )
    )
    .addSubcommand(subcommand =>
      subcommand
        .setName(KolegoSubcommand.Insult)
        .setDescription('Niech Wojciech kogoś obrazi!')
        .addUserOption(option =>
          option
            .setName(KolegoInsultOptions.User)
            .setDescription('Użytkownik do obrażenia')
            .setRequired(true)
        )
    )
    .addSubcommand(subcommand =>
      subcommand
        .setName(KolegoSubcommand.CoTam)
        .setDescription('Zapytaj Wojciecha co słychać u niego')
    ),
  execute: async (interaction, context) => {
    const subcommand = interaction.options.getSubcommand();

    switch (subcommand) {
      case KolegoSubcommand.Insult:
        await insultSubCommandHandler(interaction, context);

        break;

      case KolegoSubcommand.Question:
        await questionSubCommandHandler(interaction, context);

        break;

      case KolegoSubcommand.CoTam:
        await coTamSubCommandHandler(interaction, context);

        break;
    }
  },
};
