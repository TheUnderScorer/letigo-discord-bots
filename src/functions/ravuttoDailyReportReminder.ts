import {
  applyTokensToMessage,
  getDailyReportForDay,
} from '../shared/ravuttoDailyReport';
import { initDiscord, InitDiscordParams } from '../shared/initDiscord';

interface Dependencies extends InitDiscordParams {
  now?: () => Date;
}

export const createHandler = (dependencies?: Dependencies) => async () => {
  const messageToSend = process.env.MESSAGE_TO_SEND as string;

  const { channelId, client, targetUserId } = initDiscord(dependencies);

  const todayMessage = await getDailyReportForDay(
    channelId,
    targetUserId,
    client,
    dependencies?.now?.()
  );

  if (!todayMessage) {
    await client.sendMessageToChannel(channelId, {
      content: applyTokensToMessage(messageToSend, targetUserId),
    });
  }
};

export default createHandler();
