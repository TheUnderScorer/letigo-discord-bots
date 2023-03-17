import { Client, Message, ThreadChannel } from 'discord.js';
import { OpenAiDiscordChat } from './OpenAiDiscordChat';

export function createOpenAiThread(
  thread: ThreadChannel,
  bot: Client<true>,
  openAiChat: OpenAiDiscordChat
) {
  const messageCreateHandler = async (message: Message) => {
    if (
      message.channel.id === thread.id &&
      message.author.id !== bot.user?.id
    ) {
      await openAiChat.replyToMessage(message);
    }
  };

  const threadDeletedHandler = (thread: ThreadChannel) => {
    if (thread.id === this.thread.id) {
      bot.off('messageCreate', messageCreateHandler);
      bot.off('threadDelete', threadDeletedHandler);

      openAiChat.clearMessages();

      console.info('OpenAiThread: Thread deleted, removing handlers');
    }
  };

  bot.on('messageCreate', messageCreateHandler);
  bot.on('threadDelete', threadDeletedHandler);
}
