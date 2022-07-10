import type ytdl from 'ytdl-core';

export enum PlayerSubcommands {
  Queue = 'queue',
  List = 'list',
  ClearQueue = 'clear-queue',
  Pause = 'pause',
  Play = 'play',
  Next = 'next',
}

export interface PlayerSong {
  url: string;
  name: string;
  format: ytdl.videoFormat;
}

export enum PlayerQueueOptions {
  song = 'piosenka',
}
