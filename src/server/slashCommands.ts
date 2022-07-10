import { CommandDefinition } from '../lambdas/command.types';
import { REST } from '@discordjs/rest';
import { Routes } from 'discord-api-types/v10';
import { kolegoCommand } from './commands/kolego/kolego.command';
import { Collection } from 'discord.js';

export const slashCommands: CommandDefinition[] = [kolegoCommand];

export const slashCommandsCollection = new Collection<
  string,
  CommandDefinition
>();

slashCommands.forEach(cmd => {
  slashCommandsCollection.set(cmd.data.name, cmd);
});

export async function registerSlashCommands(
  token: string,
  applicationId: string,
  guildId: string
) {
  const rest = new REST({ version: '10' }).setToken(token);

  const commandsPayload = Object.values(slashCommands).map(cmd =>
    cmd.data.toJSON()
  );

  await rest.put(Routes.applicationGuildCommands(applicationId, guildId), {
    body: commandsPayload,
  });

  console.log(`Registered ${commandsPayload.length} commands`);
}
