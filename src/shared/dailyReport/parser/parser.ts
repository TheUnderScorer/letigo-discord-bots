import { DailyReport } from '../types';
import { isDailyReport } from '../detect';
import { valueExtractors } from './valueExtractors';

export async function parseDailyReport(
  msg: string
): Promise<DailyReport | null> {
  if (!isDailyReport(msg)) {
    return null;
  }

  const msgSplit = msg.split('\n');

  const parsedReport: DailyReport = {};

  for (const [prop, handler] of Object.entries(valueExtractors)) {
    parsedReport[prop] = await handler({
      msgSplit,
      msg,
    });
  }

  return parsedReport;
}
