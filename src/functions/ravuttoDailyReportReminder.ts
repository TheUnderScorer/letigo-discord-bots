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

  const { dailyReportChannelId, client, dailyReportTargetUserID } =
    initDiscord(dependencies);

  const todayMessage = await getDailyReportForDay(
    dailyReportChannelId,
    dailyReportTargetUserID,
    client,
    dependencies?.now?.()
  );

  if (!todayMessage) {
    await client.sendMessageToChannel(dailyReportChannelId, {
      content: applyTokensToMessage(messageToSend, dailyReportTargetUserID),
    });
  }
};

export default createHandler();
