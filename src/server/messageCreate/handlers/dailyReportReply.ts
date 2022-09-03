import { Message } from 'discord.js';
import { MessageCreateContext } from '../messageCreate.types';
import { parseDailyReport } from '../../../shared/dailyReport/parser/parser';
import { generateDailyReportReply } from '../../../shared/dailyReport/reply';

export async function dailyReportReply(
  message: Message,
  ctx: MessageCreateContext
) {
  if (
    message.channelId !== ctx.dailyReportChannelId ||
    message.author.id !== ctx.dailyReportTargetUserId
  ) {
    return;
  }

  const report = await parseDailyReport(message.content);

  if (!report) {
    return;
  }

  await message.react('👍');

  const reply = generateDailyReportReply(report, ctx.messages);

  if (reply) {
    await message.reply(reply);
  }
}
