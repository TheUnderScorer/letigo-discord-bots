import { Messages } from '../messages/messages';
import { Client } from 'discord.js';
import { ChannelPlayerManager } from './commands/player/ChannelPlayerManager';

export enum ServerCommand {
  Player = 'player',
  Help = 'help',
}

export interface CommandContext {
  messages: Messages;
  bot: Client;
  channelPlayerManager: ChannelPlayerManager;
}
