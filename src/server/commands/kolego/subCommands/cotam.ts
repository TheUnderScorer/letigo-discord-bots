import { getRandomArrayElement } from '../../../../shared/utils/array';
import { CommandHandler } from '../../command.types';

export const coTamSubCommandHandler: CommandHandler = async (
  interaction,
  { messages }
) => {
  if (interaction.isRepliable()) {
    await interaction.reply(getRandomArrayElement(messages.whatsUpReplies));
  }
};
