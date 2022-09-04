import { DailyReport } from '../types';
import { isDailyReport } from '../detect';
import { valueExtractors } from './valueExtractors';

const freeDayToken = 'wolne';

export async function parseDailyReport(
  msg: string
): Promise<DailyReport | null> {
  if (!isDailyReport(msg)) {
    return null;
  }

  const msgSplit = msg.split('\n');

  const parsedReport: DailyReport = {
    skipped: Boolean(msgSplit[1]?.toLowerCase().includes(freeDayToken)),
  };

  for (const [prop, handler] of Object.entries(valueExtractors)) {
    parsedReport[prop] = await handler({
      msgSplit,
      msg,
    });
  }

  return parsedReport;
}
