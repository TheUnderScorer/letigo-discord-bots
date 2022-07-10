import { StageChannel, VoiceChannel } from 'discord.js';
import {
  entersState,
  getVoiceConnection,
  joinVoiceChannel,
  VoiceConnectionStatus,
} from '@discordjs/voice';

export async function retrieveVoiceConnection(
  channel: StageChannel | VoiceChannel
) {
  const existingConnection = getVoiceConnection(channel.guildId);

  return existingConnection ?? (await createConnection(channel));
}

async function createConnection(channel: StageChannel | VoiceChannel) {
  const connection = joinVoiceChannel({
    channelId: channel.id,
    guildId: channel.guildId,
    adapterCreator: channel.guild.voiceAdapterCreator,
    selfMute: false,
    selfDeaf: false,
  });

  try {
    return entersState(connection, VoiceConnectionStatus.Ready, 30_000);
  } catch (error) {
    connection.destroy();

    throw error;
  }
}
