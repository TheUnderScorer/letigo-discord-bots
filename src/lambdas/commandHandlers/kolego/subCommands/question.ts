import { KolegoSubCommandHandler } from '../kolego.handler';
import {
  ApplicationCommandOptionType,
  ApplicationCommandType,
  InteractionResponseType,
  InteractionType,
} from 'discord-api-types/v10';
import { KolegoOptions } from '../../../command.types';
import { messages } from '../../../../messages/messages';
import { getRandomArrayElement } from '../../../../shared/utils/array';

export const questionSubCommandHandler: KolegoSubCommandHandler =
  async interaction => {
    if (
      interaction.type === InteractionType.ApplicationCommand &&
      interaction.data.type === ApplicationCommandType.ChatInput
    ) {
      const option = interaction.data.options?.[0];

      if (
        option?.type === ApplicationCommandOptionType.String &&
        option.name === KolegoOptions.Question
      ) {
        const question = option.value;

        if (!question.endsWith('?')) {
          return {
            type: InteractionResponseType.ChannelMessageWithSource,
            data: {
              content: messages.notAQuestion,
            },
          };
        }

        return {
          type: InteractionResponseType.ChannelMessageWithSource,
          data: {
            content: getRandomArrayElement(messages.answers),
          },
        };
      }
    }

    return false;
  };
