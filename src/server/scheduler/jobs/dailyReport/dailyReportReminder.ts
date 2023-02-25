import { ScheduledJobDefinition } from '../../scheduler.types';
import { applyTokens } from '../../../../shared/tokens';
import { mentionUser } from '../../../../shared/mentions';
import { getDailyReportForDay } from '../../../../shared/dailyReport/getDailyReportForDay';
import { getRandomArrayElement } from '../../../../shared/utils/array';
import { isTextChannel } from '../../../../shared/utils/channel';

export const createDailyReportReminder = (
  channelId: string,
  targetUserId: string,
  messages: string[],
  cron: string
): ScheduledJobDefinition => ({
  cron,
  name: 'Daily report reminder',
  execute: async ({ client, date }) => {
    const todayMessage = await getDailyReportForDay(
      channelId,
      targetUserId,
      client,
      date
    );

    if (!todayMessage) {
      const channel = client.channels.cache.get(channelId);

      if (isTextChannel(channel)) {
        await channel.send(
          applyTokensToDailyReportReminder(
            getRandomArrayElement(messages),
            targetUserId
          )
        );
      }
    }
  },
});

export function applyTokensToDailyReportReminder(
  msg: string,
  targetUserId: string
) {
  const tokens = {
    MENTION: mentionUser(targetUserId),
  };

  return applyTokens(msg, tokens);
}
