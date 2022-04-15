import { DiscordClient } from './Discord.client';

export function initDiscord() {
  const channelId = process.env.CHANNEL_ID as string;
  const targetUserId = process.env.TARGET_USER_ID as string;

  const client = new DiscordClient(process.env.BOT_TOKEN as string);

  return { channelId, client, targetUserId };
}
