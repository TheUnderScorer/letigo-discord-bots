/* eslint-disable @typescript-eslint/no-explicit-any */
import { messages } from '../../../../messages/messages';
import { createTwinTailsReminder } from './twinTailsReminder';
import { createMockChannel } from '../../../../__tests__/mocks';

const targetUserId = '#targetUserId';
const channelId = '#channelId';

const createMockMessage = (date: Date) => ({
  createdAt: date,
  content: 'hehe',
  author: {
    id: targetUserId,
  },
});

describe('twinTailsReminder', () => {
  it('should send message it last found message is not from today', async () => {
    const today = new Date();
    const yesterday = new Date(today.getTime() - 24 * 60 * 60 * 1000);

    const schedule = createTwinTailsReminder(channelId, targetUserId);

    const channel = createMockChannel([createMockMessage(yesterday)]);
    const client = {
      channels: {
        cache: {
          get: () => channel,
        },
      },
    };

    await schedule.execute({
      client: client as any,
      date: new Date(),
      messages,
    });

    expect(channel.send).toHaveBeenCalledTimes(1);
  });
});
