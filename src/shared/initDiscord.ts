import { DiscordClient } from './Discord.client';
import { AxiosInstance } from 'axios';

export interface InitDiscordParams {
  axios?: AxiosInstance;
}

export interface InitDiscordResult {
  client: DiscordClient;
  targetUserId: string;
  channelId: string;
}

export function initDiscord({
  axios,
}: InitDiscordParams = {}): InitDiscordResult {
  const channelId = process.env.DAILY_REPORT_CHANNEL_ID as string;
  const targetUserId = process.env.DAILY_REPORT_TARGET_USER_ID as string;

  const client = new DiscordClient(process.env.BOT_TOKEN as string, axios);

  return { channelId, client, targetUserId };
}
