/* eslint-disable @typescript-eslint/no-explicit-any */
import { createDailyGreeting } from './dailyGreeting';
import { messages } from '../../../messages/messages';

const days = Array.from({ length: 7 }, (v, k) => k);

const channelId = 'test';

const mockChannel = {
  isText: () => true,
  send: jest.fn(),
};

const mockClient = {
  channels: {
    cache: {
      get: jest.fn(),
    },
  },
};

describe('Daily greeting', () => {
  beforeEach(() => {
    mockChannel.send.mockClear();
    mockClient.channels.cache.get
      .mockClear()
      .mockImplementation(() => mockChannel);
  });

  it.each(days)('should return a greeting for day %d', async day => {
    const date = new Date();
    const dateSpy = jest.spyOn(date, 'getDay');

    dateSpy.mockReturnValue(day);

    const handler = createDailyGreeting(channelId);

    await handler.execute({
      date,
      client: mockClient as any,
      messages,
    });

    expect(mockChannel.send).toHaveBeenCalledTimes(1);

    const call = mockChannel.send.mock.calls[0];
    const sentMessage = call[0];

    const dayMessages = messages.greetings[day];

    const isMessageValid = dayMessages.some(msg => sentMessage === msg);

    expect(isMessageValid).toBe(true);
  });
});
