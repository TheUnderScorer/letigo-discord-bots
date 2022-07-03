import { Message, VoiceChannel } from 'discord.js';
import { ParsedMessage } from '../../parseMessage';
import { BotError } from '../../../shared/errors/BotError';
import { CommandContext } from '../../server.types';
import { PlayerCommandArg } from './player.types';
import { applyTokens } from '../../../shared/tokens';
import { mapCommandsForHelp } from '../../../shared/utils/commands';
import { getRandomArrayElement } from '../../../shared/utils/array';

export async function playerCommand(
  message: Message,
  { args }: ParsedMessage,
  ctx: CommandContext
) {
  if (!message.member?.voice?.channel) {
    throw new BotError(ctx.messages.mustBeInVoiceChannel);
  }

  const [arg] = args;

  if (arg === PlayerCommandArg.Help) {
    await message.reply({
      content: applyTokens(ctx.messages.server.player.availableCommands, {
        COMMANDS: mapCommandsForHelp(Object.values(PlayerCommandArg)),
      }),
    });

    return;
  }

  const channel = message.member.voice.channel;

  if (!channel.isVoice()) {
    throw new Error('Channel is not a voice channel');
  }

  const player = await ctx.channelPlayerManager.getOrCreateChannelPlayer(
    channel as VoiceChannel
  );

  switch (arg) {
    case PlayerCommandArg.Queue: {
      const { entryIndex, isPlaying } = await player.queue(args[1]);

      if (!isPlaying) {
        const reply =
          entryIndex > 0
            ? applyTokens(
                getRandomArrayElement(ctx.messages.server.player.addedToQueue),
                {
                  INDEX: (entryIndex + 1).toString(),
                }
              )
            : ctx.messages.server.player.addedToQueueAsNext;

        await message.reply(reply);
      }

      break;
    }

    case PlayerCommandArg.Pause:
      await player.pause();
      break;

    case PlayerCommandArg.Play:
      await player.play();
      break;

    case PlayerCommandArg.Next:
      await player.next();
      break;

    case PlayerCommandArg.ClearQueue:
      await player.clearQueue();

      await message.reply(ctx.messages.server.player.clearedQueue);

      break;

    case PlayerCommandArg.List: {
      const songs = player.songQueue;

      if (!songs.length) {
        await message.reply(ctx.messages.server.player.noMoreSongs);

        return;
      }

      await message.reply({
        content: songs
          .map((song, index) => `${index + 1}. ${song.name}`)
          .join('\n'),
      });

      break;
    }

    default:
      throw new BotError(ctx.messages.unknownCommand);
  }
}
