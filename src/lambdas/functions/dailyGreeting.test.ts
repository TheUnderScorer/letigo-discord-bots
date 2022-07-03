import axios, { AxiosRequestConfig, AxiosResponse } from 'axios';
import { createHandler } from './dailyGreeting';
import { messages } from '../../messages/messages';

const days = Array.from({ length: 7 }, (v, k) => k);

describe('Daily greeting', () => {
  const adapter = jest.fn<Promise<AxiosResponse>, [AxiosRequestConfig]>();

  const axiosInstance = axios.create({
    adapter,
  });

  beforeEach(() => {
    adapter.mockClear();

    adapter.mockImplementation(async request => {
      return {
        config: request,
        data: {},
        status: 200,
        statusText: 'OK',
        request,
        headers: {},
      };
    });
  });

  it.each(days)('should return a greeting for day %d', async day => {
    const date = new Date();
    const dateSpy = jest.spyOn(date, 'getDay');

    dateSpy.mockReturnValue(day);

    const handler = createHandler({
      axios: axiosInstance,
      now: () => date,
    });

    await handler();

    expect(adapter).toHaveBeenCalledTimes(1);

    const call = adapter.mock.calls[0];
    const body = JSON.parse(call[0].data);

    const dayMessages = messages.greetings[day];
    const isMessageValid = dayMessages.some(msg => body.content === msg);

    expect(isMessageValid).toBe(true);
  });
});
