import {
  CommandHandler,
  KolegoInsultOptions,
} from '../../../../lambdas/command.types';
import { getRandomArrayElement } from '../../../../shared/utils/array';
import { applyTokens } from '../../../../shared/tokens';

export const insultSubCommandHandler: CommandHandler = async (
  interaction,
  context
) => {
  const user = interaction.options.getUser(KolegoInsultOptions.User);

  if (!user) {
    throw new Error('User not found');
  }

  const tokens = {
    USER: user.toString(),
  };
  const message = applyTokens(
    getRandomArrayElement(context.messages.insults),
    tokens
  );

  if (interaction.isRepliable()) {
    await interaction.reply(message);
  }
};
