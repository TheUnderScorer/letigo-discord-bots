import { ScheduledJobDefinition } from '../../scheduler.types';
import { applyTokens } from '../../../../shared/tokens';
import { mentionUser } from '../../../../shared/mentions';
import { Client, Message } from 'discord.js';

export const createDailyReportReminder = (
  channelId: string,
  targetUserId: string,
  message: string,
  cron: string
): ScheduledJobDefinition => ({
  cron,
  name: 'Daily report reminder',
  execute: async ({ client, date }) => {
    const todayMessage = await getDailyReportForDay(
      channelId,
      targetUserId,
      client,
      date
    );

    if (!todayMessage) {
      const channel = client.channels.cache.get(channelId);

      if (channel?.isText()) {
        await channel.send(applyTokensToMessage(message, targetUserId));
      }
    }
  },
});

export const isDailyReport = (content: string) =>
  content.split('\n')[0].toLowerCase().includes('[dzień');

export const isValidAuthor = (message: Message, targetUserId: string) =>
  message.author?.id === targetUserId;

export const isMessageFromDate = (message: Message, date = new Date()) => {
  const messageDate = message.createdAt;

  return (
    messageDate.getDate() === date.getDate() &&
    messageDate.getMonth() === date.getMonth() &&
    messageDate.getFullYear() === date.getFullYear()
  );
};

export function applyTokensToMessage(msg: string, targetUserId: string) {
  const tokens = {
    MENTION: mentionUser(targetUserId),
  };

  return applyTokens(msg, tokens);
}

export async function getDailyReportForDay(
  channelId: string,
  targetUserId: string,
  client: Client<true>,
  date = new Date()
) {
  const channel = client.channels.cache.get(channelId);
  const channelMessages = channel?.isText()
    ? await channel.messages
        .fetch({ limit: 100 })
        .then(res => Array.from(res.values()))
    : [];

  return channelMessages.find(
    message =>
      isValidAuthor(message, targetUserId) &&
      isDailyReport(message.content) &&
      isMessageFromDate(message, date)
  );
}
