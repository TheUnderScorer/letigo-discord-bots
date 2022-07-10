/* eslint-disable @typescript-eslint/no-explicit-any */
import { createDailyReportReminder } from './dailyReportReminder';
import { messages } from '../../../../messages/messages';

const targetUserId = '#targetUserId';
const channelId = '#channelId';
const message = '{{MENTION}} kolego, ale pamiętaj o daily raporcie dzisiaj';

const createMockMessage = (date: Date) => ({
  createdAt: date,
  content:
    '[DZIEŃ 6] - 16.04.2022r.\nCzas spędzony na IT: 1h\nCo zostało zrobione:\n- odnalazłem mój tutorialowy projekt z CSS i przejrzałem\n- kilka zadań na Sololearn\n- kilka fiszek na AnkiDroid\n\nPrzemyślenia:\n- ogółem jeszcze badam teren, próbuję po trochu różnych metod, żeby znaleźć optymalne dla siebie podejście\n- na razie podchodzę do tego trochę jak pies do jeża, czuję potrzebę wielu więcej godzin i większej koncentracji, ale na razie święta trochę mnie rozpraszają\n- przydałby się ten tutorial gitowy od Przemka\n- ogółem mam pewną wizję utworzenia projektu składającego się z wielu malutkich podprojektów, tylko chciałbym go mądrze zaprojektować\n\nKrótko mówiąc, na razie jestem na etapie konceptualnym. I zwykle w przeszłości był on dla mnie najtrudniejszy. Ale najważniejszy.\n\nSONG dnia:\nMike Shinoda - Open Door',
  author: {
    id: targetUserId,
  },
});

const createMockChannel = (messages: any[]) => ({
  isText: () => true,
  send: jest.fn(),
  messages: {
    fetch: jest.fn().mockResolvedValue({
      values: () => messages,
    }),
  },
});

describe('dailyReportReminder', () => {
  it('should send message if daily report is missing', async () => {
    const today = new Date();
    const yesterday = new Date(today.getTime() - 24 * 60 * 60 * 1000);

    const schedule = createDailyReportReminder(
      channelId,
      targetUserId,
      message,
      ''
    );

    const channel = createMockChannel([createMockMessage(yesterday)]);

    const client = {
      channels: {
        cache: {
          get: () => channel,
        },
      },
    };

    await schedule.execute({
      client: client as any,
      date: new Date(),
      messages,
    });

    expect(channel.send).toHaveBeenCalledTimes(1);
    expect(channel.send).toHaveBeenCalledWith(
      '<@#targetUserId> kolego, ale pamiętaj o daily raporcie dzisiaj'
    );
  });

  it('should not send message if daily report is present for today', async () => {
    const today = new Date();

    const schedule = createDailyReportReminder(
      channelId,
      targetUserId,
      message,
      ''
    );

    const channel = createMockChannel([createMockMessage(today)]);

    const client = {
      channels: {
        cache: {
          get: () => channel,
        },
      },
    };

    await schedule.execute({
      client: client as any,
      date: new Date(),
      messages,
    });

    expect(channel.send).toHaveBeenCalledTimes(0);
  });
});
