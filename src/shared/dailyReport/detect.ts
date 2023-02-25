import { Message } from 'discord.js';

export const isDailyReport = (content: string) =>
  content.split('\n')[0].toLowerCase().includes('[dzień');

export const isValidAuthor = (message: Message, targetUserId: string) =>
  message.author?.id === targetUserId;
