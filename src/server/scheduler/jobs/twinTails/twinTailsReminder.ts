import { ScheduledJobDefinition } from '../../scheduler.types';
import { TextChannel } from 'discord.js';
import { isMessageFromDate } from '../../../../shared/messages/isMessageFromDate';
import { applyTokens } from '../../../../shared/tokens';
import { getRandomArrayElement } from '../../../../shared/utils/array';
import { mentionUser } from '../../../../shared/mentions';
import { isTextChannel } from '../../../../shared/utils/channel';

export const createTwinTailsReminder = (
  channelId: string,
  userId: string
): ScheduledJobDefinition => ({
  cron: '00 23 * * *',
  name: 'Twin Tails reminder',
  execute: async ({ date, client, messages }) => {
    const channel = client.channels.cache.get(channelId);

    if (!isTextChannel(channel)) {
      return;
    }

    const recentMessage = await getRecentMessage(userId, channel);

    if (!recentMessage) {
      return;
    }

    if (!isMessageFromDate(recentMessage, date)) {
      await channel.send(
        applyTokensToReminder(
          getRandomArrayElement(messages.twinTailsReminder.night),
          userId
        )
      );
    }
  },
});

function applyTokensToReminder(msg: string, userId: string) {
  return applyTokens(msg, {
    mention: mentionUser(userId),
  });
}

async function getRecentMessage(userId: string, channel: TextChannel) {
  const channelMessages = isTextChannel(channel)
    ? await channel.messages
        .fetch({ limit: 1 })
        .then(res => Array.from(res.values()))
    : [];

  return channelMessages.filter(message => message.author.id === userId)?.[0];
}
