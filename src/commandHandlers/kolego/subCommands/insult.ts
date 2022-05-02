import { KolegoSubCommandHandler } from '../kolego.handler';
import {
  ApplicationCommandOptionType,
  ApplicationCommandType,
  InteractionResponseType,
  InteractionType,
} from 'discord-api-types/v10';
import { KolegoOptions } from '../../../command.types';
import { messages } from '../../../messages';
import { getRandomArrayElement } from '../../../shared/utils/array';
import { applyTokens } from '../../../shared/tokens';
import { mentionUser } from '../../../shared/mentions';

export const insultSubCommandHandler: KolegoSubCommandHandler =
  async interaction => {
    if (
      interaction.type === InteractionType.ApplicationCommand &&
      interaction.data.type === ApplicationCommandType.ChatInput
    ) {
      const option = interaction.data.options?.[0];

      if (
        option?.type === ApplicationCommandOptionType.User &&
        option.name === KolegoOptions.Insult
      ) {
        const tokens = {
          USER: mentionUser(option.value),
        };
        const message = applyTokens(
          getRandomArrayElement(messages.insults),
          tokens
        );

        return {
          type: InteractionResponseType.ChannelMessageWithSource,
          data: {
            content: message,
          },
        };
      }
    }

    return false;
  };
