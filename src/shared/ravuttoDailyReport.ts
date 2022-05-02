import type { APIMessage } from 'discord-api-types/v10';
import type { DiscordClient } from './Discord.client';
import { mentionUser } from './mentions';

type TokenHandler = (msg: string, userId: string) => string;

const tokens: Record<string, TokenHandler> = {
  '{{MENTION}}': (msg, userId) => mentionUser(userId),
};

export const isDailyReport = (content: string) =>
  content.toLowerCase().startsWith('[dzień');

export const isValidAuthor = (message: APIMessage, targetUserId: string) =>
  message.author?.id === targetUserId;

export const isMessageFromDate = (message: APIMessage, date = new Date()) => {
  const messageDate = new Date(message.timestamp);

  return (
    messageDate.getDate() === date.getDate() &&
    messageDate.getMonth() === date.getMonth() &&
    messageDate.getFullYear() === date.getFullYear()
  );
};

export function applyTokensToMessage(msg: string, targetUserId: string) {
  const tokensArray = Object.entries(tokens);

  return tokensArray.reduce(
    (acc, [token, replacer]) =>
      acc.replace(new RegExp(token, 'g'), () => replacer(msg, targetUserId)),
    msg
  );
}

export async function getDailyReportForDay(
  channelId: string,
  targetUserId: string,
  client: DiscordClient,
  date = new Date()
) {
  const response = await client.getChannelMessages(channelId);

  return response.data?.find(
    message =>
      isValidAuthor(message, targetUserId) &&
      isDailyReport(message.content) &&
      isMessageFromDate(message, date)
  );
}
