import {
  applyTokensToMessage,
  getDailyReportForToday,
} from '../shared/ravuttoDailyReport';
import { initDiscord } from '../shared/initDiscord';

export async function handler() {
  const messageToSend = process.env.MESSAGE_TO_SEND as string;

  const { channelId, client, targetUserId } = initDiscord();

  const todayMessage = await getDailyReportForToday(
    channelId,
    targetUserId,
    client
  );

  if (!todayMessage) {
    await client.sendMessageToChannel(channelId, {
      content: applyTokensToMessage(messageToSend, targetUserId),
    });
  }
}
