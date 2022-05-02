import axios, { AxiosRequestConfig, AxiosResponse } from 'axios';
import { createHandler } from './ravuttoDailyReportReminder';
import * as getMessagesResponse from '../__tests__/mockResponses/getMessages.json';

describe('Ravutto daily report reminder', () => {
  const adapter = jest.fn<Promise<AxiosResponse>, [AxiosRequestConfig]>();

  const message = '{{MENTION}} kolego, ale pamiętaj o daily raporcie dzisiaj';
  Object.assign(process.env, {
    MESSAGE_TO_SEND: message,
  });

  const axiosInstance = axios.create({
    adapter,
  });

  beforeEach(() => {
    adapter.mockClear();
  });

  it('should not send message if daily report is present', async () => {
    const now = new Date('2022-04-16T00:00:00.000Z');

    const handler = createHandler({ axios: axiosInstance, now: () => now });

    Object.assign(process.env, {
      MESSAGE_TO_SEND: message,
    });

    adapter.mockImplementation(async request => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      let data: any;

      if (request.url?.includes('channels')) {
        data = Object.values(getMessagesResponse);
      }

      return {
        config: request,
        data,
        status: 200,
        statusText: 'OK',
        request,
        headers: {},
      };
    });

    await handler();

    expect(adapter).toHaveBeenCalledTimes(1);
  });

  it('should send message if daily report is missing', async () => {
    const now = new Date('2022-04-17T00:00:00.000Z');

    const handler = createHandler({ axios: axiosInstance, now: () => now });

    adapter.mockImplementation(async request => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      let data: any;

      if (request.url?.includes('channels')) {
        data = Object.values(getMessagesResponse);
      }

      return {
        config: request,
        data,
        status: 200,
        statusText: 'OK',
        request,
        headers: {},
      };
    });

    await handler();

    expect(adapter).toHaveBeenCalledTimes(2);

    const call = adapter.mock.calls[1];
    expect(call[0].data).toMatchInlineSnapshot(
      // eslint-disable-next-line
      `"{\\"content\\":\\"<@300692223769575425> kolego, ale pamiętaj o daily raporcie dzisiaj\\"}"`
    );
    expect(call[0].url).toMatchInlineSnapshot(
      // eslint-disable-next-line
      `"/channels/963085194276007966/messages"`
    );
  });
});
