import type { Message } from 'discord.js';

export const messagePrefix = '!kolego ';

export interface ParsedMessage {
  command: string;
  args: string[];
}

export function parseServerCommand(message: Message): ParsedMessage | null {
  if (!message.content.startsWith(messagePrefix)) {
    return null;
  }

  const command = message.content.slice(messagePrefix.length).split(' ')[0];
  const args = message.content
    .slice(messagePrefix.length + command.length + 1)
    .split(' ');

  return {
    command,
    args,
  };
}
