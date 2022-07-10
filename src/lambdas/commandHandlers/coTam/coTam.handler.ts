import { CommandHandler } from '../../command.types';
import { getRandomArrayElement } from '../../../shared/utils/array';

export const coTamHandler: CommandHandler = async (
  interaction,
  { messages }
) => {
  if (interaction.isRepliable()) {
    await interaction.reply(getRandomArrayElement(messages.whatsUpReplies));
  }
};
