import { ScheduledJobDefinition } from '../../scheduler.types';
import { applyTokens } from '../../../../shared/tokens';
import { mentionUser } from '../../../../shared/mentions';
import { getDailyReportForDay } from '../../../../shared/dailyReport/getDailyReportForDay';

export const createDailyReportReminder = (
  channelId: string,
  targetUserId: string,
  message: string,
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

      if (channel?.isText()) {
        await channel.send(
          applyTokensToDailyReportReminder(message, targetUserId)
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
