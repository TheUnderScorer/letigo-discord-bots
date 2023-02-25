import { Message } from 'discord.js';
import { MessageCreateContext } from './messageCreate.types';
import { dailyReportReply } from './handlers/dailyReportReply';
import { twinTailsReact } from './handlers/twinTailsReact';

interface Dependencies {
  ctx: MessageCreateContext;
}

export const makeMessageCreateHandler =
  ({ ctx }: Dependencies) =>
  async (message: Message) => {
    await Promise.all([
      dailyReportReply(message, ctx).catch(error =>
        handleError(message, error)
      ),
      twinTailsReact(message, ctx).catch(error => handleError(message, error)),
    ]);
  };

function handleError(message: Message, error: Error) {
  console.error(`Failed to reply to message ${message.id}`, error);
}
