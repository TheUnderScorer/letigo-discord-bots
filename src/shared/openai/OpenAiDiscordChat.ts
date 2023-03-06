import { ChatCompletionRequestMessage, OpenAIApi } from 'openai';
import { Client, Message } from 'discord.js';
import { Messages } from '../../messages/messages';
import { AxiosError } from 'axios/index';
import { BotError } from '../errors/BotError';
import { isTextChannel } from '../utils/channel';

export class OpenAiDiscordChat {
  private threadMessages: ChatCompletionRequestMessage[] = [];

  private currentRequestAbortController?: AbortController;

  private static nicknameMap = {
    ['Taliön']: 'Przemek',
    Husky: 'Marek',
    PureGold: 'Wojtek',
    Mawgan: 'Artur',
    Walter_441: 'Zachariasz',
    Ravutto: 'Rafał',
    Amaterasu: 'Paulina',
  };

  constructor(
    private readonly openAiClient: OpenAIApi,
    private readonly bot: Client<true>,
    private readonly messages: Messages,
    initialMessages: ChatCompletionRequestMessage[]
  ) {
    this.threadMessages = initialMessages;
  }

  clearMessages() {
    this.threadMessages = [];
  }

  async replyToMessage(message: Message) {
    this.currentRequestAbortController?.abort();

    const { channel } = message;

    if (!isTextChannel(channel)) {
      return;
    }

    const formattedMessage =
      message.content.endsWith('?') ||
      message.content.endsWith('!') ||
      message.content.endsWith('.')
        ? message.content
        : `${message.content}.`;

    this.threadMessages.push({
      content: formattedMessage,
      name:
        OpenAiDiscordChat.nicknameMap[message.author.username] ??
        message.author.username,
      role: 'user',
    });

    const abortController = new AbortController();
    this.currentRequestAbortController = abortController;

    await channel.sendTyping();

    try {
      const response = await this.openAiClient.createChatCompletion(
        {
          messages: this.threadMessages,
          temperature: 1.2,
          max_tokens: 1500,
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

          await channel.send(choice.message.content);
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
