import { initDiscord } from '../src/shared/initDiscord';
import {
  ApplicationCommandType,
  RESTPostAPIApplicationGuildCommandsJSONBody,
} from 'discord-api-types/v10';
import { Commands } from '../src/command.types';
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
    description: 'Odpowiada mądrościami życiowymi Wojciecha',
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
