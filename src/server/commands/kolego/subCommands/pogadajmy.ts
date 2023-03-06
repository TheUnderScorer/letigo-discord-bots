import { CommandHandler } from '../../command.types';
import { isTextChannel } from '../../../../shared/utils/channel';
import { createOpenAiThread } from '../../../../shared/openai/openAiThread';
import { OpenAiDiscordChat } from '../../../../shared/openai/OpenAiDiscordChat';
import { openAiContext } from '../../../../shared/openai/context';

export const pogadajmySubCommandHandler: CommandHandler = async (
  interaction,
  { bot, messages, openAiClient }
) => {
  if (interaction.isRepliable() && isTextChannel(interaction.channel)) {
    await interaction.reply('No dobra kolego');

    const thread = await interaction.channel.threads.create({
      name: 'Chat z Wojtkiem',
      autoArchiveDuration: 60,
    });

    const chat = new OpenAiDiscordChat(openAiClient, bot, messages, [
      {
        content: openAiContext,
        role: 'system',
      },
    ]);

    createOpenAiThread(thread, bot, chat);
  }
};
