import type { CommandDefinition } from './command.types';
import { REST } from '@discordjs/rest';
import { Routes } from 'discord-api-types/v10';
import { kolegoCommand } from './kolego/kolego.command';
import { Collection } from 'discord.js';
import { playerCommand } from './player/player.command';

export const commands: CommandDefinition[] = [kolegoCommand, playerCommand];

export const commandsCollection = new Collection<string, CommandDefinition>();

commands.forEach(cmd => {
  commandsCollection.set(cmd.data.name, cmd);
});

export async function registerSlashCommands(
  token: string,
  applicationId: string,
  guildId: string
) {
  const rest = new REST({ version: '10' }).setToken(token);

  const commandsPayload = Object.values(commands).map(cmd => cmd.data.toJSON());

  await rest.put(Routes.applicationGuildCommands(applicationId, guildId), {
    body: commandsPayload,
  });

  console.log(`Registered ${commandsPayload.length} commands`);
}
