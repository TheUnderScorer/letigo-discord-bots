import { initDiscord } from '../src/shared/initDiscord';
import {
  ApplicationCommandOptionType,
  ApplicationCommandType,
  RESTPostAPIApplicationGuildCommandsJSONBody,
} from 'discord-api-types/v10';
import {
  Commands,
  KolegoSubcommand,
} from '../src/server/commands/command.types';
import { config } from 'dotenv';
import * as path from 'path';

config({
  path: path.resolve(__dirname, '../.env'),
});

const { client } = initDiscord();

const appId = process.env.APP_ID as string;
const guildId = process.env.GUILD_ID as string;

const commands: RESTPostAPIApplicationGuildCommandsJSONBody[] = [
  {
    name: Commands.Kolego,
    type: ApplicationCommandType.ChatInput,
    description: 'Wywołaj Wojciecha',
    options: [
      {
        name: KolegoSubcommand.Question,
        description: 'Zadaj pytanie Wojciechowi',
        type: ApplicationCommandOptionType.String,
      },
      {
        name: KolegoSubcommand.Insult,
        description: 'Niech Wojciech kogoś obrazi!',
        type: ApplicationCommandOptionType.User,
      },
    ],
  },
  {
    name: Commands.CoTam,
    type: ApplicationCommandType.ChatInput,
    description: 'Zapytaj Wojciecha co słychać u niego',
  },
];

async function main() {
  try {
    const results = await client.registerGuildCommands(
      appId,
      guildId,
      commands
    );

    results.forEach(response => {
      if (response.data?.id) {
        console.log(`Command ${response.data.name} registered`);
      } else {
        console.error(`Command ${response.data.name} failed to register`);
      }
    });
  } catch (error) {
    console.error('Failed to register commands', error);

    throw error;
  }
}

main().catch(console.error);
