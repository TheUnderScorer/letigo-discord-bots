import { DailyReport } from '../types';
import { fetchChronographData } from './fetchChronographData';
import { arrayToOrRegex } from '../../utils/regex';
import ytSearch from 'youtube-search';

interface ValueExtractorParams {
  msg: string;
  msgSplit: string[];
}

type ValueExtractors = {
  [Key in keyof DailyReport]: (
    params: ValueExtractorParams
  ) => DailyReport[Key] | Promise<DailyReport[Key]>;
};

const definitions = {
  timeSpent: {
    tokens: ['czas spędzony na it:'],
    hoursRegex: /\s(\d)h|\s(\d.\d)h/,
    urlRegex:
      /(http|ftp|https):\/\/([\w_-]+(?:\.[\w_-]+)+)([\w.,@?^=%&:\\/~+#-]*[\w@?^=%&\\/~+#-])/,
  },
  mentalScore: {
    tokens: ['mental:'],
    split: '/',
  },
  day: {
    tokens: ['dzien', 'dzień'],
    bracketsRegex: /\[(.*?)]/,
    extractNumberRegex: /[a-zA-ZŃń]/g,
  },
  song: {
    tokens: ['song dnia:'],
  },
};

function findByTokens(line: string, tokens: string[], lowerCase = true) {
  const parsedLine = lowerCase ? line.toLowerCase() : line;

  return tokens.some(token => parsedLine.includes(token));
}

export const valueExtractors: ValueExtractors = {
  song: async ({ msgSplit }) => {
    const line = msgSplit.find(line =>
      findByTokens(line, definitions.song.tokens)
    );

    if (line) {
      const songName = line
        .toLowerCase()
        .replace(arrayToOrRegex(definitions.song.tokens), '')
        .trim();

      if (songName) {
        const result = await ytSearch(songName, {
          maxResults: 1,
          // TODO Get key from context
          key: process.env.YT_API_KEY,
        }).catch(err => {
          console.error(err);

          return undefined;
        });

        return {
          url: result?.results[0].link,
          name: songName,
        };
      }
    }

    return undefined;
  },
  day: ({ msgSplit }) => {
    const line = msgSplit.find(line =>
      findByTokens(line, definitions.day.tokens)
    );

    if (line) {
      const dayStr = line
        .match(definitions.day.bracketsRegex)?.[1]
        .replace(definitions.day.extractNumberRegex, '')
        .trim();

      if (dayStr) {
        return parseInt(dayStr, 10);
      }
    }

    return undefined;
  },
  mentalScore: ({ msgSplit }) => {
    const line = msgSplit.find(line =>
      findByTokens(line, definitions.mentalScore.tokens)
    );

    if (line) {
      const [mentalScore] = line
        .toLowerCase()
        .replace(arrayToOrRegex(definitions.mentalScore.tokens), '')
        .trim()
        .split(definitions.mentalScore.split)
        .map(val => parseFloat(val.replace(/,/g, '.')))
        .filter(val => !Number.isNaN(val));

      return mentalScore;
    }

    return undefined;
  },
  timeSpentSeconds: async ({ msgSplit }) => {
    const line = msgSplit.find(line =>
      findByTokens(line, definitions.timeSpent.tokens)
    );

    if (line) {
      const [start] = line.split(':');
      const details = line.replace(start, '');

      const chronographMatch = details.match(definitions.timeSpent.urlRegex);
      const chronographUrl = chronographMatch?.[0];

      if (chronographUrl) {
        const chronographData = await fetchChronographData(
          chronographUrl
        ).catch(error => {
          console.error(error);

          return null;
        });

        if (chronographData) {
          return chronographData;
        }
      }

      const hoursMatch = details.match(definitions.timeSpent.hoursRegex);
      const hours = parseFloat(
        (hoursMatch?.[1] || hoursMatch?.[2])?.trim() ?? ''
      );

      return Number.isNaN(hours)
        ? undefined
        : {
            seconds: hours * 3600,
          };
    }

    return undefined;
  },
};
