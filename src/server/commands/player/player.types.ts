import type ytdl from 'ytdl-core';

export enum PlayerCommandArg {
  Queue = 'queue',
  List = 'list',
  ClearQueue = 'clear-queue',
  Help = 'help',
  Pause = 'pause',
  Play = 'play',
  Next = 'next',
}

export interface PlayerSong {
  url: string;
  name: string;
  format: ytdl.videoFormat;
}
