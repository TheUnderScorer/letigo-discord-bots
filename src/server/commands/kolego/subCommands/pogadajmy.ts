import { CommandHandler } from '../../command.types';
import { isTextChannel } from '../../../../shared/utils/channel';
import { OpenAiThread } from '../../../../shared/openai/OpenAiThread';

export const pogadajmySubCommandHandler: CommandHandler = async (
  interaction,
  { messages, bot, openAiClient }
) => {
  if (interaction.isRepliable()) {
    await interaction.reply('No dobra kolego');

    if (interaction.channel && isTextChannel(interaction.channel)) {
      const thread = await interaction.channel.threads.create({
        name: 'Chat z Wojtkiem',
        autoArchiveDuration: 60,
      });

      new OpenAiThread(thread, bot, messages, openAiClient).init();
    }
  }
};
