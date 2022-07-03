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

dotenv.config();

async function initBot() {
  const bot = new Client({
    intents: [
      'GUILDS',
      'GUILD_VOICE_STATES',
      'GUILD_MESSAGES',
      'GUILD_EMOJIS_AND_STICKERS',
    ],
  });

  await bot.login(process.env.BOT_TOKEN);

  return new Promise<Client<true>>(resolve => {
    bot.once('ready', async readyBot => {
      resolve(readyBot);
    });
  });
}

async function main() {
  const bot = await initBot();
  const channelPlayerManager = new ChannelPlayerManager(messages);

  console.log('Bot is ready!');

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
