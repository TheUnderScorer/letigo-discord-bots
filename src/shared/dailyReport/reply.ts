import {
  DailyReport,
  isSkippedDailyReport,
  isTimeSpentChronograph,
  isTimeSpentPlain,
  PassedDailyReport,
} from './types';
import { Messages } from '../../messages/messages';
import { applyTokens } from '../tokens';
import { getRandomArrayElement } from '../utils/array';

const definitions = {
  mental: {
    low: [0, 4],
    medium: [5, 7],
    high: [8, 10],
  },
  timeSpentSeconds: {
    low: [0, 60 * 30],
    medium: [60 * 30, 60 * 60 * 2],
    high: [60 * 60 * 2],
  },
};

function matchValueToThreshold<T extends Record<string, number[]>>(
  thresholds: T,
  value: number
): keyof T | undefined {
  return Object.entries(thresholds).find(([, [min, max]]) => {
    if (typeof max === 'undefined') {
      return value >= min;
    }

    return value >= min && value <= max;
  })?.[0];
}

function getSpentSeconds(report: PassedDailyReport) {
  if (isTimeSpentChronograph(report.timeSpentSeconds)) {
    return report.timeSpentSeconds.grossSeconds;
  }

  if (isTimeSpentPlain(report.timeSpentSeconds)) {
    return report.timeSpentSeconds.seconds;
  }

  return undefined;
}

// TODO Add tests
export function generateDailyReportReply(
  report: DailyReport,
  messages: Messages
) {
  if (isSkippedDailyReport(report)) {
    return getRandomArrayElement(messages.dailyReportReplies.skipped);
  }

  const mentalScore =
    typeof report.mentalScore === 'number'
      ? matchValueToThreshold(definitions.mental, report.mentalScore)
      : undefined;

  const timeSpent = getSpentSeconds(report);
  const timeSpentScore =
    typeof timeSpent === 'number'
      ? matchValueToThreshold(definitions.timeSpentSeconds, timeSpent)
      : undefined;

  if (!report.day) {
    return undefined;
  }

  const messageParts: string[] = [
    applyTokens(getRandomArrayElement(messages.dailyReportReplies.greeting), {
      DAY: report.day.toString(),
    }),
  ];

  if (mentalScore) {
    messageParts.push(
      getRandomArrayElement(
        messages.dailyReportReplies.mentalComments[mentalScore]
      )
    );
  }

  if (timeSpentScore) {
    messageParts.push(
      getRandomArrayElement(
        messages.dailyReportReplies.timeSpentComments[timeSpentScore]
      )
    );
  }

  if (report.song?.url) {
    messageParts.push(
      '',
      applyTokens(getRandomArrayElement(messages.dailyReportReplies.song), {
        SONG_URL: report.song.url,
      })
    );
  }

  return messageParts.join('\n');
}
