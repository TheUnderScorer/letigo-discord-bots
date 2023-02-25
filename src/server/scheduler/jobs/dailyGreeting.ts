import { ScheduledJobDefinition } from '../scheduler.types';
import { getRandomArrayElement } from '../../../shared/utils/array';
import { isTextChannel } from '../../../shared/utils/channel';

export const createDailyGreeting = (
  channelId: string
): ScheduledJobDefinition => ({
  cron: '0 9 * * *',
  name: 'Daily greeting',
  execute: async ({ client, messages, date }) => {
    const day = date.getDay();

    const message = getRandomArrayElement(messages.greetings[day]);

    if (message) {
      const channel = client.channels.cache.get(channelId);

      if (isTextChannel(channel)) {
        await channel.send(message);
      }
    }
  },
});
