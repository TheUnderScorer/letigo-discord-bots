import dotenv from 'dotenv';
import { Client } from 'discord.js';
import * as messages from '../messages/messages.json';
import { ChannelPlayerManager } from './commands/player/ChannelPlayerManager';
import express from 'express';
import { makeInteractionsHandler } from './commands/interactions';
import { commandsCollection, registerSlashCommands } from './commands/commands';
import { initScheduler } from './scheduler/scheduler';
import pkg from '../../package.json';
import { makeMessageCreateHandler } from './messageCreate/messageCreate';

dotenv.config();

async function initBot(token: string) {
  const bot = new Client({
    intents: [
      'Guilds',
      'GuildVoiceStates',
      'GuildMessages',
      'GuildEmojisAndStickers',
      'GuildIntegrations',
      'GuildMessageTyping',
    ],
  });

  await bot.login(token);

  return new Promise<Client<true>>(resolve => {
    bot.once('ready', async readyBot => {
      resolve(readyBot);
    });
  });
}

async function main() {
  const botToken = process.env.BOT_TOKEN as string;
  const appId = process.env.APP_ID as string;
  const guildId = process.env.GUILD_ID as string;
  const bot = await initBot(botToken);
  const channelPlayerManager = new ChannelPlayerManager(messages);

  const greetingChannelId = process.env.GREETING_CHANNEL_ID as string;
  const dailyReportChannelId = process.env.DAILY_REPORT_CHANNEL_ID as string;
  const dailyReportTargetUserId = process.env
    .DAILY_REPORT_TARGET_USER_ID as string;
  const twinTailsChannelId = process.env.TWIN_TAILS_CHANNEL_ID as string;
  const twinTailsUserId = process.env.TWIN_TAILS_USER_ID as string;

  await registerSlashCommands(botToken, appId, guildId);
  initScheduler({
    client: bot,
    messages,
    dailyReportChannelId,
    dailyReportTargetUserId,
    greetingChannelId,
    twinTailsChannelId,
    twinTailsUserId,
  });

  const app = express();
  const port = process.env.PORT || 3000;

  console.log('Bot is ready!');

  app.get('/', (req, res) => {
    res.json({
      result: true,
      version: pkg.version,
    });
  });

  app.listen(port, () => {
    console.log('Server is running on port ', port);
  });

  bot.on(
    'messageCreate',
    makeMessageCreateHandler({
      ctx: {
        bot,
        messages,
        dailyReportChannelId,
        dailyReportTargetUserId,
        twinTailsUserId,
        twinTailsChannelId,
      },
    })
  );

  bot.on(
    'interactionCreate',
    makeInteractionsHandler({
      ctx: {
        bot,
        messages,
        channelPlayerManager,
      },
      commands: commandsCollection,
    })
  );
}

main().catch(console.error);
