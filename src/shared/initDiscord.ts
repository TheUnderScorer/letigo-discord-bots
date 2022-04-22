import { DiscordClient } from './Discord.client';
import { AxiosInstance } from 'axios';

export interface InitDiscordParams {
  axios?: AxiosInstance;
}

export interface InitDiscordResult {
  client: DiscordClient;
  dailyReportTargetUserID: string;
  dailyReportChannelId: string;
  greetingChannelId: string;
}

export function initDiscord({
  axios,
}: InitDiscordParams = {}): InitDiscordResult {
  const dailyReportChannelId = process.env.DAILY_REPORT_CHANNEL_ID as string;
  const dailyReportTargetUserID = process.env
    .DAILY_REPORT_TARGET_USER_ID as string;

  const greetingChannelId = process.env.GREETING_CHANNEL_ID as string;

  const client = new DiscordClient(process.env.BOT_TOKEN as string, axios);

  return {
    dailyReportChannelId,
    client,
    dailyReportTargetUserID,
    greetingChannelId,
  };
}
