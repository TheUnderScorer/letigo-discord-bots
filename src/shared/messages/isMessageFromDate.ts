import { Message } from 'discord.js';

export const isMessageFromDate = (message: Message, date = new Date()) => {
  const messageDate = message.createdAt;

  return (
    messageDate.getDate() === date.getDate() &&
    messageDate.getMonth() === date.getMonth() &&
    messageDate.getFullYear() === date.getFullYear()
  );
};
