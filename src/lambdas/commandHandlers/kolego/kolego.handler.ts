import { CommandHandler } from '../../command.types';
import { kolegoSubCommands } from './subCommands';
import { CommandInteraction } from 'discord.js';

export type KolegoSubCommandHandler = (
  interaction: CommandInteraction
) => Promise<boolean>;

export const kolegoHandler: CommandHandler = async interaction => {
  for (const subCommandHandler of kolegoSubCommands) {
    const result = await subCommandHandler(interaction);

    if (result) {
      break;
    }
  }
};
