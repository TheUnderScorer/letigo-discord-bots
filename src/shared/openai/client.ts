import { Configuration, OpenAIApi } from 'openai';

export function createOpenAiClient(apiKey: string) {
  const configuration = new Configuration({
    apiKey,
  });

  return new OpenAIApi(configuration);
}
