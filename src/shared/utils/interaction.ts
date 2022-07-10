import { CommandInteraction } from 'discord.js';

export async function deferredReply<T>(
  interaction: CommandInteraction,
  callback: () => Promise<T>
) {
  await interaction.deferReply();

  const result = await callback();

  await interaction.deleteReply();

  return result;
}
