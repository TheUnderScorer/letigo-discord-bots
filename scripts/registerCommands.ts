import { initDiscord } from '../src/shared/initDiscord';
import {
  ApplicationCommandType,
  RESTPostAPIApplicationGuildCommandsJSONBody,
  ApplicationCommandOptionType,
} from 'discord-api-types/v10';
import { Commands, KolegoOptions } from '../src/command.types';
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
        name: KolegoOptions.Question,
        description: 'Zadaj pytanie Wojciechowi',
        type: ApplicationCommandOptionType.String,
      },
    ],
  },
];

async function main() {
  const results = await client.registerGuildCommands(appId, guildId, commands);

  results.forEach(response => {
    if (response.data?.id) {
      console.log(`Command ${response.data.name} registered`);
    } else {
      console.error(`Command ${response.data.name} failed to register`);
    }
  });
}

main().catch(console.error);
