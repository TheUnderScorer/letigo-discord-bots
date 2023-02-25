import { Channel, ChannelType, TextChannel, VoiceChannel } from 'discord.js';

export function isTextChannel(channel?: Channel): channel is TextChannel {
  return channel?.type === ChannelType.GuildText;
}

export function isVoiceChannel(channel?: Channel): channel is VoiceChannel {
  return channel?.type === ChannelType.GuildVoice;
}
