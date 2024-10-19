import type ytdl from '@distube/ytdl-core';

export enum PlayerSubcommands {
  Queue = 'dodaj',
  List = 'list',
  ClearQueue = 'wyczysc-kolejke',
  Pause = 'pauza',
  Play = 'odtwarzaj',
  Next = 'next',
}

export interface PlayerSong {
  url: string;
  name: string;
  format?: ytdl.videoFormat;
}

export enum PlayerQueueOptions {
  song = 'piosenka',
}
