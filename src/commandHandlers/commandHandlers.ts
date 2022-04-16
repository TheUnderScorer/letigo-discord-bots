import { CommandHandler, Commands } from '../command.types';
import { kolegoHandler } from './kolego/kolego';

export const commandHandlers: Record<Commands, CommandHandler> = {
  [Commands.Kolego]: kolegoHandler,
};
