import { initDiscord, InitDiscordParams } from '../../shared/initDiscord';
import { getRandomArrayElement } from '../../shared/utils/array';
import { messages } from '../../messages/messages';

interface Dependencies extends InitDiscordParams {
  now?: () => Date;
}

export const createHandler =
  ({ now = () => new Date(), ...rest }: Dependencies = {}) =>
  async () => {
    const today = now();
    const day = today.getDay();

    const message = getRandomArrayElement(messages.greetings[day]);

    const { client, greetingChannelId } = initDiscord(rest);

    if (message) {
      await client.sendMessageToChannel(greetingChannelId, {
        content: message,
      });
    }
  };

export default createHandler();
