import { CommandHandler } from '../../command.types';
import {
  APIInteraction,
  APIInteractionResponse,
  InteractionResponseType,
} from 'discord-api-types/v10';
import { kolegoSubCommands } from './subCommands';
import { messages } from '../../messages';

export type KolegoSubCommandHandler = (
  interaction: APIInteraction
) => Promise<APIInteractionResponse | false>;

export const kolegoHandler: CommandHandler = async interaction => {
  for (const subCommandHandler of kolegoSubCommands) {
    const result = await subCommandHandler(interaction);

    if (result) {
      return result;
    }
  }

  return {
    type: InteractionResponseType.ChannelMessageWithSource,
    data: {
      content: messages.whatAreYouSaying,
    },
  };
};
