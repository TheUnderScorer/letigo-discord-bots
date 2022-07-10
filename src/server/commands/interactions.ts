import type { Collection, Interaction } from 'discord.js';
import type { CommandDefinition, CommandHandlerContext } from './command.types';
import { BotError } from '../../shared/errors/BotError';

interface Dependencies {
  ctx: CommandHandlerContext;
  commands: Collection<string, CommandDefinition>;
}

export const makeInteractionsHandler =
  ({ ctx, commands }: Dependencies) =>
  async (interaction: Interaction) => {
    if (!interaction.isCommand()) {
      return;
    }

    const handler = commands.get(interaction.commandName);

    if (handler) {
      try {
        await handler.execute(interaction, ctx);
      } catch (error) {
        console.error(error);

        let reply: string;

        if (error instanceof BotError) {
          reply = error.message;
        } else {
          reply = ctx.messages.unknownError;
        }

        if (interaction.deferred) {
          await interaction.editReply(reply);
        } else {
          await interaction.reply(reply);
        }
      }
    }
  };
