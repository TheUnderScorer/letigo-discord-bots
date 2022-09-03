import { Message } from 'discord.js';

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
