/* eslint-disable @typescript-eslint/no-explicit-any */
import { Commands, KolegoSubcommand } from './command.types';
import { mentionUser } from '../../shared/mentions';
import { makeInteractionsHandler } from './interactions';
import { messages } from '../../messages/messages';
import { commandsCollection } from './commands';
import { applyTokens } from '../../shared/tokens';

function createKolegoInteraction(
  subcommand: KolegoSubcommand,
  options: Record<string, any>
) {
  return {
    commandName: Commands.Kolego,
    isCommand: () => true,
    isRepliable: () => true,
    reply: jest.fn(),
    options: {
      getSubcommand: () => subcommand,
      ...options,
    },
  };
}

describe('interactions', () => {
  const handler = makeInteractionsHandler({
    ctx: {
      bot: {} as any,
      messages,
      channelPlayerManager: {} as any,
    },
    commands: commandsCollection,
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
