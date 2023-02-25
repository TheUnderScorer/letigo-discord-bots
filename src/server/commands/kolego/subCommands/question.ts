import { CommandHandler, KolegoQuestionOptions } from '../../command.types';
import { getRandomArrayElement } from '../../../../shared/utils/array';

export const questionSubCommandHandler: CommandHandler = async (
  interaction,
  context
) => {
  if (!interaction.isChatInputCommand()) {
    return;
  }

  const question = interaction.options.getString(
    KolegoQuestionOptions.Question
  );

  if (!question) {
    throw new Error('No question provided');
  }

  if (interaction.isRepliable()) {
    if (!question.endsWith('?')) {
      await interaction.reply(context.messages.notAQuestion);

      return;
    }

    await interaction.reply(getRandomArrayElement(context.messages.answers));
  }
};
