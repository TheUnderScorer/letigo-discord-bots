import type { Client } from 'discord.js';
import type { Messages } from '../../messages/messages';

export interface ScheduledJobContext {
  client: Client<true>;
  date: Date;
  messages: Messages;
}

export interface ScheduledJobDefinition {
  cron: string;
  name: string;
  execute: (ctx: ScheduledJobContext) => Promise<void>;
}
