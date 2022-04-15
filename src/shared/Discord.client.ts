import {
  RESTGetAPIChannelMessagesResult,
  RESTPostAPIChannelMessageJSONBody,
  RESTPostAPIChannelMessageResult,
  Routes,
} from 'discord-api-types/v10';
import axios, { AxiosInstance } from 'axios';

export class DiscordClient {
  private static baseUrl = 'https://discord.com/api/v10';

  constructor(
    token: string,
    protected httpClient: AxiosInstance = axios.create({
      baseURL: DiscordClient.baseUrl,
      headers: {
        Authorization: `Bot ${token}`,
      },
    })
  ) {}

  async getChannelMessages(channelId: string) {
    return await this.httpClient.get<RESTGetAPIChannelMessagesResult>(
      Routes.channelMessages(channelId)
    );
  }

  async sendMessageToChannel(
    channelId: string,
    body: RESTPostAPIChannelMessageJSONBody
  ) {
    return this.httpClient.post<RESTPostAPIChannelMessageResult>(
      Routes.channelMessages(channelId),
      body
    );
  }
}
