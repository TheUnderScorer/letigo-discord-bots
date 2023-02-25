import { ChannelType } from 'discord.js';

export const createMockChannel = <T>(messages: T[]) => ({
  isText: () => true,
  type: ChannelType.GuildText,
  send: jest.fn(),
  messages: {
    fetch: jest.fn().mockResolvedValue({
      values: () => messages,
    }),
  },
});
