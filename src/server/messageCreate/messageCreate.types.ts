import { Client } from 'discord.js';
import { Messages } from '../../messages/messages';

export interface MessageCreateContext {
  bot: Client<true>;
  messages: Messages;
  dailyReportChannelId: string;
  dailyReportTargetUserId: string;
  twinTailsChannelId: string;
  twinTailsUserId: string;
}
