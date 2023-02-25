import { GuildEmoji, Message } from 'discord.js';
import { MessageCreateContext } from '../messageCreate.types';
import { getRandomArrayElement } from '../../../shared/utils/array';

function resolveReactions(message: Message) {
  if (!message.guild) {
    return [];
  }

  return [
    message.guild.emojis.cache.find(emoji => emoji.name === 'tamakohappy'),
    '👍',
  ].filter(Boolean) as Array<string | GuildEmoji>;
}

export async function twinTailsReact(
  message: Message,
  ctx: MessageCreateContext
) {
  if (
    message.channelId !== ctx.twinTailsChannelId ||
    message.author.id !== ctx.twinTailsUserId
  ) {
    return;
  }

  const reactions = resolveReactions(message);

  await message.react(getRandomArrayElement(reactions));
}
