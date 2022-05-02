import { CommandHandler } from '../../command.types';
import { InteractionResponseType } from 'discord-api-types/v10';
import { messages } from '../../messages';
import { getRandomArrayElement } from '../../shared/utils/array';

export const coTamHandler: CommandHandler = async () => {
  return {
    type: InteractionResponseType.ChannelMessageWithSource,
    data: {
      content: getRandomArrayElement(messages.whatsUpReplies),
    },
  };
};
