import { Client, Message, ThreadChannel } from 'discord.js';
import { Messages } from '../../messages/messages';
import { openAiContext } from './context';
import { OpenAIApi } from 'openai';
import { AxiosError } from 'axios';
import { BotError } from '../errors/BotError';

export class OpenAiThread {
  private threadContext = [openAiContext];

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

    this.threadContext.push(
      message.content.endsWith('?') ||
        message.content.endsWith('!') ||
        message.content.endsWith('.')
        ? message.content
        : `${message.content}.`
    );

    const abortController = new AbortController();
    this.currentRequestAbortController = abortController;

    await this.thread.sendTyping();

    try {
      const response = await this.openAiClient.createCompletion(
        {
          prompt: this.threadContext.join('\n'),
          max_tokens: 750,
          temperature: 0.5,
          model: 'text-davinci-003',
        },
        {
          signal: abortController.signal,
        }
      );

      const choices = response.data.choices;

      if (choices.length) {
        const [choice] = choices;

        if (choice.text) {
          this.threadContext.push(choice.text);
          await this.thread.send(choice.text);
        } else {
          await message.react('❌');
        }
      }
    } catch (error) {
      if (error.name === 'AbortError') {
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
