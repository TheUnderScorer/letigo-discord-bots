/* eslint-disable @typescript-eslint/no-explicit-any */
import {
  APIApplicationCommandInteractionDataOption,
  ApplicationCommandOptionType,
  ApplicationCommandType,
  InteractionType,
} from 'discord-api-types/v10';
import { Commands, KolegoOptions } from '../command.types';
import { handler } from './handleInteractions';
import { messages } from '../messages';

function triggerInteraction(
  command: Commands,
  options?: APIApplicationCommandInteractionDataOption[]
) {
  const body = {
    type: InteractionType.ApplicationCommand,
    data: {
      name: command,
      type: ApplicationCommandType.ChatInput,
      options,
    },
  };

  return handler({ body: JSON.stringify(body) } as any);
}

describe('Handle interactions', () => {
  describe('/kolego pytanie', () => {
    it('should insult if question doesnt contain question mark', async () => {
      const response = await triggerInteraction(Commands.Kolego, [
        {
          name: KolegoOptions.Question,
          type: ApplicationCommandOptionType.String,
          value: 'pytanie',
        },
      ]);
      const body = JSON.parse(response.body as string);

      expect(body.data.content).toEqual(messages.notAQuestion);
    });

    it('should reply with random answer', async () => {
      const response = await triggerInteraction(Commands.Kolego, [
        {
          name: KolegoOptions.Question,
          type: ApplicationCommandOptionType.String,
          value: 'Jaki jest sens życia?',
        },
      ]);
      const body = JSON.parse(response.body as string);

      expect(messages.answers).toContain(body.data.content);
    });
  });

  describe('/cotam', () => {
    it('should return random message', async () => {
      const response = await triggerInteraction(Commands.CoTam);
      const body = JSON.parse(response.body as string);

      expect(messages.whatsUpReplies).toContain(body.data.content);
    });
  });
});
