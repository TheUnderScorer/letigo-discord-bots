import { Client, Message, ThreadChannel } from 'discord.js';
import { Messages } from '../../messages/messages';
import { openAiContext } from './context';
import { OpenAIApi } from 'openai';
import { AxiosError } from 'axios';
import { BotError } from '../errors/BotError';
import { ChatCompletionRequestMessage } from 'openai/api';

export class OpenAiThread {
  private threadMessages: ChatCompletionRequestMessage[] = [
    {
      content: openAiContext,
      role: 'system',
    },
  ];

  private static nicknameMap = {
    ['Taliön']: 'Przemek',
    Husky: 'Marek',
    PureGold: 'Wojtek',
    Mawgan: 'Artur',
    Walter_441: 'Zachariasz',
    Ravutto: 'Rafał',
  };

  private currentRequestAbortController?: AbortController;

  constructor(
    private readonly thread: ThreadChannel,
    private readonly bot: Client<true>,
    private readonly messages: Messages,
    private readonly openAiClient: OpenAIApi
  ) {}

  init() {
    const messageCreateHandler = async (message: Message) => {
      if (
        message.channel.id === this.thread.id &&
        message.author.id !== this.bot.user?.id
      ) {
        await this.handleIncomingMessage(message);
      }
    };

    const threadDeletedHandler = (thread: ThreadChannel) => {
      if (thread.id === this.thread.id) {
        this.bot.off('messageCreate', messageCreateHandler);
        this.bot.off('threadDelete', threadDeletedHandler);

        console.info('OpenAiThread: Thread deleted, removing handlers');
      }
    };

    this.bot.on('messageCreate', messageCreateHandler);
    this.bot.on('threadDelete', threadDeletedHandler);
  }

  private async handleIncomingMessage(message: Message) {
    this.currentRequestAbortController?.abort();

    const formattedMessage =
      message.content.endsWith('?') ||
      message.content.endsWith('!') ||
      message.content.endsWith('.')
        ? message.content
        : `${message.content}.`;

    this.threadMessages.push({
      content: formattedMessage,
      name:
        OpenAiThread.nicknameMap[message.author.username] ??
        message.author.username,
      role: 'user',
    });

    const abortController = new AbortController();
    this.currentRequestAbortController = abortController;

    await this.thread.sendTyping();

    try {
      const response = await this.openAiClient.createChatCompletion(
        {
          messages: this.threadMessages,
          temperature: 0.5,
          model: 'gpt-3.5-turbo',
        },
        {
          signal: abortController.signal,
        }
      );

      const choices = response.data.choices;

      if (choices.length) {
        const choice = choices.find(choice => Boolean(choice.message?.content));

        if (choice?.message) {
          this.threadMessages.push(choice.message);
          await this.thread.send(choice.message.content);
        } else {
          await message.react('❌');
        }
      }
    } catch (error) {
      if (error.name === 'AbortError' || error.message === 'canceled') {
        console.info('OpenAiThread: Request aborted');

        return;
      }

      console.error('OpenAiThread: Error while generating response', error);

      if (error instanceof AxiosError) {
        const { response } = error;

        console.error('OpenAiThread: Error response', response?.data);

        if (response) {
          const error = new BotError(
            this.messages.unknownError,
            JSON.stringify(response.data, null, 2)
          );

          await message.reply(error.messageContent);

          return;
        }
      }

      await message.reply(this.messages.unknownError);
    }
  }
}
