import { ChannelPlayer } from './ChannelPlayer';
import { VoiceChannel } from 'discord.js';
import {
  createAudioPlayer,
  NoSubscriberBehavior,
  VoiceConnection,
  VoiceConnectionStatus,
} from '@discordjs/voice';
import { retrieveVoiceConnection } from './voiceConnection';
import { Messages } from '../../../messages/messages';
import { applyTokens } from '../../../shared/tokens';
import { getRandomArrayElement } from '../../../shared/utils/array';

export class ChannelPlayerManager {
  private readonly players: Map<string, ChannelPlayer> = new Map();

  constructor(private readonly messages: Messages) {}

  async getOrCreateChannelPlayer(channel: VoiceChannel) {
    if (this.players.has(channel.guildId)) {
      return this.players.get(channel.guildId) as ChannelPlayer;
    }

    const connection = await retrieveVoiceConnection(channel);
    const audioPlayer = createAudioPlayer({
      behaviors: {
        noSubscriber: NoSubscriberBehavior.Stop,
        maxMissedFrames: 1,
      },
    });
    const subscription = connection.subscribe(audioPlayer);

    if (!subscription) {
      throw new Error('Failed to subscribe to voice connection');
    }

    const player = new ChannelPlayer(channel, subscription, this.messages);

    this.setupEvents(player, connection, channel);

    this.players.set(channel.guildId, player);

    return player;
  }

  setupEvents(
    player: ChannelPlayer,
    connection: VoiceConnection,
    channel: VoiceChannel
  ) {
    connection.once(VoiceConnectionStatus.Disconnected, async () => {
      console.log('Voice connection destroyed');

      await player.dispose();

      this.players.delete(channel.guildId);
    });

    player.on('playStarted', async (song, channel) => {
      const message = applyTokens(
        getRandomArrayElement(this.messages.server.player.nowPlaying),
        {
          SONG_NAME: song.name,
        }
      );

      await channel.send(message);
    });

    player.on('finished', async channel => {
      const message = getRandomArrayElement(this.messages.server.player.ended);

      await channel.send(message);
    });
  }
}
