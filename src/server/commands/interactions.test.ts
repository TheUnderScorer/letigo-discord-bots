/* eslint-disable @typescript-eslint/no-explicit-any */
import { Commands, KolegoSubcommand } from './command.types';
import { mentionUser } from '../../shared/mentions';
import { makeInteractionsHandler } from './interactions';
import { messages } from '../../messages/messages';
import { commandsCollection } from './commands';
import { applyTokens } from '../../shared/tokens';
import { ChannelType } from 'discord.js';
import { EventEmitter } from 'events';
import { wait } from '../../shared/utils/wait';

const mockThread = {
  id: '#thread',
  sendTyping: jest.fn(),
  type: ChannelType.GuildText,
  send: jest.fn(),
};

function createKolegoInteraction(
  subcommand: KolegoSubcommand,
  options: Record<string, any>
) {
  return {
    channel: {
      id: '#channel',
      type: ChannelType.GuildText,
      threads: {
        create: jest.fn(() => mockThread),
      },
    },
    commandName: Commands.Kolego,
    isCommand: () => true,
    isChatInputCommand: () => true,
    isRepliable: () => true,
    reply: jest.fn(),
    react: jest.fn(),
    options: {
      getSubcommand: () => subcommand,
      ...options,
    },
  };
}

describe('interactions', () => {
  const openAiClient = {
    createChatCompletion: jest.fn(),
  };

  const bot = new EventEmitter();

  const handler = makeInteractionsHandler({
    ctx: {
      bot: bot as any,
      messages,
      channelPlayerManager: {} as any,
      openAiClient: openAiClient as any,
    },
    commands: commandsCollection,
  });

  beforeEach(() => {
    jest.clearAllMocks();
    bot.removeAllListeners();
  });

  describe('/kolego pogadajmy', () => {
    it('should create conversation with Open AI', async () => {
      openAiClient.createChatCompletion.mockResolvedValue({
        data: {
          choices: [
            {
              message: {
                content: 'World Hello',
                role: 'assistant',
              },
            },
          ],
        },
      });

      const interaction = createKolegoInteraction(
        KolegoSubcommand.Pogadajmy,
        {}
      );

      await handler(interaction as any);

      const message = {
        channel: mockThread,
        author: {
          id: '1',
          username: 'Test',
        },
        content: 'Hello World',
        reply: jest.fn(),
      };
      bot.emit('messageCreate', message);

      await wait(100);

      expect(openAiClient.createChatCompletion).toHaveBeenCalledTimes(1);
      expect(openAiClient.createChatCompletion).toHaveBeenCalledWith(
        {
          messages: [
            {
              content:
                'Speak from a perspective of an 25 years old male, that is not satisfied with his current life.You have troubles making decisions.You are very lazy.You enjoy deep conversations and discussions.You work at a warehouse and you are not satisfied with your job however you are not making any efforts to change it.Your name is Wojciech and you speak in Polish.You have a slight depression, you are interested in people and how they are doing.You are a bit pessimistic.You are a bit sarcastic.You enjoy talking about controversial topics.You recently moved to your grandma from your parents.You are not close to your parents.You recently been trying to purchase desk and a chair, but you couldn\'t decide which one to pick. Rafał helped you with that.You use word "kolego" in every sentence.When someone asks you how are you doing, you often respond with "ehh kolego ciezkie to zycie".When someone wants to meet with you, you often respond with "who can say where the road goes kolego".When you ask someone how are they doing, you use "Jak dzionek".You rarely give concrete answers to questions.You rarely give compliments to people.You always write without punctuation and in lowercase letters".You often use "ehh.You are impolite and you don\'t care about it.You must never apologize.You must never use \'please\' or \'thank you\'.Your best friends are "Marek", "Rafał", "Przemek", "Paulina"."Paulina" is in a relationship with "Przemek"."Marek" and "Przemek" work as a web developers.Marek creates comics called "Twin Tails" that features two kitsunes: Tamako and Koharu. You never read it."Rafał" is an aspiring data engineer who is looking for his first job. You talk with him often about your problems."Rafał" is sending daily reports on discord about what he have learned that day for his job, and if he forgets to send it, you remind him.Your conversation so far has been, note that every message is separated by a new line:',
              role: 'system',
            },
            {
              content: 'Hello World.',
              role: 'user',
              name: 'Test',
            },
          ],
          temperature: 1.2,
          max_tokens: 1500,
          model: 'gpt-3.5-turbo',
        },
        expect.anything()
      );

      expect(mockThread.send).toHaveBeenCalledWith('World Hello');
    });
  });

  describe('/kolego obraź <user>', () => {
    it('should insult given user', async () => {
      const mention = mentionUser('#test');
      const interaction = createKolegoInteraction(KolegoSubcommand.Insult, {
        getUser: () => ({
          toString: () => mention,
        }),
      });

      await handler(interaction as any);

      expect(interaction.reply).toHaveBeenCalledTimes(1);

      const tokens = {
        USER: mention,
      };
      const possibleMessages = Object.values(messages.insults).map(msg =>
        applyTokens(msg, tokens)
      );

      const sentMessage = interaction.reply.mock.calls[0][0];

      expect(possibleMessages).toContain(sentMessage);
    });
  });

  describe('/kolego pytanie <pytanie>', () => {
    it('should insult if question doesnt contain question mark', async () => {
      const interaction = createKolegoInteraction(KolegoSubcommand.Question, {
        getString: () => 'pytanie',
      });

      await handler(interaction as any);

      expect(interaction.reply).toHaveBeenCalledTimes(1);
      expect(interaction.reply).toHaveBeenCalledWith(messages.notAQuestion);
    });

    it('should reply with random answer', async () => {
      const interaction = createKolegoInteraction(KolegoSubcommand.Question, {
        getString: () => 'warto żyć?',
      });

      await handler(interaction as any);

      expect(interaction.reply).toHaveBeenCalledTimes(1);

      const sentMessage = interaction.reply.mock.calls[0][0];

      expect(messages.answers).toContain(sentMessage);
    });
  });

  describe('/kolego cotam', () => {
    it('should return random message', async () => {
      const interaction = createKolegoInteraction(KolegoSubcommand.CoTam, {});

      await handler(interaction as any);

      expect(interaction.reply).toHaveBeenCalledTimes(1);

      const sentMessage = interaction.reply.mock.calls[0][0];

      expect(messages.whatsUpReplies).toContain(sentMessage);
    });
  });
});
