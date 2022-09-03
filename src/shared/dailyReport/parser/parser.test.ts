import { DailyReport } from '../types';
import { parseDailyReport } from './parser';

interface TestCase {
  input: string;
  expected: DailyReport | null;
}

const testCases: TestCase[] = [
  {
    input: `🏋️‍♂️  [DZIEŃ 143] - 31.08.2022r. 🏋️‍♂️ 
Czas spędzony na IT: 2h ip
Co zostało zrobione:
Plan na dziś: Tworzenie własnych mini komponentów (spiracone Udemy)

Przemyślenia: Te mini-komponenty dają jakąś satysfakcję, a pod tym względem była mega posucha ostatnio. Dlatego na razie będę je cisnąć dalej, aż poczuję się swobodnie i pewnie w tym.

Mental: 6,1/10

SONG dnia: Bon Jovi - It's My Life

It's my life
It's now or never
But I ain't gonna live forever
I just want to live while I'm alive

Tomorrow's getting harder, make no mistake
Luck ain't even lucky, got to make your own breaks

Better stand tall when they're calling you out
Don't bend, don't break, baby, don't back down`,
    expected: {
      day: 143,
      mentalScore: 6.1,
      timeSpentSeconds: {
        seconds: 7200,
      },
      song: {
        name: "bon jovi - it's my life",
        url: 'https://www.youtube.com/watch?v=vx2u5uUu3DE',
      },
    },
  },

  {
    input: `🏋️‍♂️  [DZIEŃ 143] - 31.08.2022r. 🏋️‍♂️ 
Czas spędzony na IT: 2h https://chronograph.io/jMcvjZc
Co zostało zrobione:
Plan na dziś: Tworzenie własnych mini komponentów (spiracone Udemy)

Przemyślenia: Te mini-komponenty dają jakąś satysfakcję, a pod tym względem była mega posucha ostatnio. Dlatego na razie będę je cisnąć dalej, aż poczuję się swobodnie i pewnie w tym.

Mental: 6,1/10

SONG dnia: Bon Jovi - It's My Life

It's my life
It's now or never
But I ain't gonna live forever
I just want to live while I'm alive

Tomorrow's getting harder, make no mistake
Luck ain't even lucky, got to make your own breaks

Better stand tall when they're calling you out
Don't bend, don't break, baby, don't back down`,
    expected: {
      day: 143,
      mentalScore: 6.1,
      timeSpentSeconds: {
        seconds: 7200,
      },
      song: {
        name: "bon jovi - it's my life",
        url: 'https://www.youtube.com/watch?v=vx2u5uUu3DE',
      },
    },
  },
];

describe('Daily report parser', () => {
  testCases.forEach((testCase, index) => {
    it(`should parse daily report #${index}`, async () => {
      expect(await parseDailyReport(testCase.input)).toEqual(testCase.expected);
    }, 10000);
  });
});
