import dotenv from 'dotenv';
import { Client } from 'discord.js';
import { parseServerCommand } from './parseMessage';
import { CommandContext, ServerCommand } from './server.types';
import { playerCommand } from './commands/player/player.command';
import { BotError } from '../shared/errors/BotError';
import * as messages from '../messages/messages.json';
import { ChannelPlayerManager } from './commands/player/ChannelPlayerManager';
import { applyTokens } from '../shared/tokens';
import { mapCommandsForHelp, quoteCommand } from '../shared/utils/commands';
import express from 'express';
import { makeInteractionsHandler } from './interactions';
import {
  registerSlashCommands,
  slashCommandsCollection,
} from './slashCommands';
import { initScheduler } from './scheduler/scheduler';

dotenv.config();

async function initBot(token: string) {
  const bot = new Client({
    intents: [
      'GUILDS',
      'GUILD_VOICE_STATES',
      'GUILD_MESSAGES',
      'GUILD_EMOJIS_AND_STICKERS',
      'GUILD_INTEGRATIONS',
      'GUILD_MESSAGE_TYPING',
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

  await registerSlashCommands(botToken, appId, guildId);
  initScheduler(bot, messages);

  const app = express();
  const port = process.env.PORT || 3000;

  console.log('Bot is ready!');

  app.get('/', (req, res) => {
    res.json({
      result: true,
    });
  });

  app.listen(port, () => {
    console.log('Server is running on port ', port);
  });

  bot.on(
    'interactionCreate',
    makeInteractionsHandler({
      ctx: {
        bot,
        messages,
      },
      commands: slashCommandsCollection,
    })
  );

  bot.on('messageCreate', async message => {
    if (message.author.id === bot.user.id) {
      return;
    }

    const parsedMessage = parseServerCommand(message);

    if (parsedMessage) {
      const ctx: CommandContext = {
        messages,
        bot,
        channelPlayerManager: channelPlayerManager,
      };

      try {
        switch (parsedMessage.command) {
          case ServerCommand.Player:
            await playerCommand(message, parsedMessage, ctx);

            break;

          case ServerCommand.Help:
            await message.reply(
              applyTokens(messages.server.help, {
                COMMANDS: mapCommandsForHelp(Object.values(ServerCommand)),
                PLAY_COMMAND: quoteCommand(ServerCommand.Player),
              })
            );
        }
      } catch (error) {
        console.error(error);

        if (error instanceof BotError) {
          await message.reply(error.message);
        } else {
          await message.reply(messages.unknownError);
        }
      }
    }
  });
}

main().catch(console.error);
