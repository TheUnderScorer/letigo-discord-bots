import {
  createAudioResource,
  PlayerSubscription,
  VoiceConnectionStatus,
} from '@discordjs/voice';
import { VoiceChannel } from 'discord.js';
import { PlayerSong } from './player.types';
import ytdl from '@distube/ytdl-core';

import { TypedEmitter } from 'tiny-typed-emitter';
import { Messages } from '../../../messages/messages';
import { BotError } from '../../../shared/errors/BotError';
import { findDesiredFormat } from './findDesiredFormat';

import { getRandomIPv6 } from '@distube/ytdl-core/lib/utils';

export interface PlayerQueueEvents {
  finished: (channel: VoiceChannel) => unknown;
  nextSong: (song: PlayerSong) => unknown;
  playStarted: (song: PlayerSong, channel: VoiceChannel) => unknown;
}

export class ChannelPlayer extends TypedEmitter<PlayerQueueEvents> {
  private playerQueue: PlayerSong[] = [];

  constructor(
    private readonly voiceChannel: VoiceChannel,
    private readonly playerSubscription: PlayerSubscription,
    private readonly messages: Messages
  ) {
    super();

    this.setupEvents();
  }

  get songQueue() {
    return [...this.playerQueue] as ReadonlyArray<PlayerSong>;
  }

  private get playerState() {
    return this.playerSubscription.player.state;
  }

  private setupEvents() {
    const { player } = this.playerSubscription;

    player.on('stateChange', async (oldState, newState) => {
      if (oldState.status === 'playing' && newState.status === 'idle') {
        await this.next(false);
      }
    });
  }

  async queue(url: string) {
    if (!ytdl.validateURL(url)) {
      throw new BotError(this.messages.invalidUrl);
    }

    let isPlaying = false;

    const state = this.playerState;
    const info = await ytdl.getInfo(url, {
      agent: this.getAgent(),
    });

    const bestFormat = findDesiredFormat(info);

    const existingSong = this.playerQueue.find(song => song.url === url);

    if (existingSong) {
      throw new BotError(this.messages.player.alreadyQueued);
    }

    const song: PlayerSong = {
      url,
      name: info.videoDetails.media.song ?? info.videoDetails.title,
      format: bestFormat,
    };

    const index = this.playerQueue.push(song);

    if (state.status !== 'playing') {
      isPlaying = true;
      await this.next(false);
    }

    return {
      entryIndex: index - 1,
      isPlaying,
    };
  }

  async pause() {
    this.playerSubscription.player.pause();
  }

  async stop() {
    this.playerSubscription.player.stop(true);
  }

  async play() {
    this.playerSubscription.player.unpause();
  }

  async next(throwErrorOnEmpty = true) {
    const song = this.playerQueue.shift();

    if (song) {
      await this.playSong(song);

      this.emit('nextSong', song);
    } else {
      if (throwErrorOnEmpty) {
        throw new BotError(this.messages.player.noMoreSongs);
      }

      await this.stop();
      this.emit('finished', this.voiceChannel);
    }
  }

  async playSong(song: PlayerSong) {
    const stream = ytdl(song.url, {
      filter: 'audioonly',
      // 32mb
      highWaterMark: 1 << 25,
      quality: song.format?.itag,
      agent: this.getAgent(),
    });

    const { player } = this.playerSubscription;

    const resource = createAudioResource(stream);

    player.play(resource);

    this.emit('playStarted', song, this.voiceChannel);
  }

  async clearQueue() {
    await this.stop();

    this.playerQueue = [];
  }

  async dispose() {
    if (
      this.playerSubscription.connection.state.status !==
      VoiceConnectionStatus.Destroyed
    ) {
      console.info(
        'Destroying player connection for channel:',
        `${this.voiceChannel.name}`
      );

      this.removeAllListeners();
      this.playerSubscription.connection.destroy();

      await this.clearQueue();
    }
  }

  private getAgent() {
    return ytdl.createAgent(undefined, {
      localAddress: getRandomIPv6('2001:2::/48'),
    } as any);
  }
}
