import { response } from '../aws.types';
import {
  APIInteraction,
  InteractionResponseType,
  InteractionType,
} from 'discord-api-types/v10';
import middy from 'middy';
import httpHeaderNormalizer from '@middy/http-header-normalizer';
import { commandHandlers } from '../commandHandlers/commandHandlers';
import { Commands } from '../../server/commands/command.types';
import { verifyDiscordSignature } from '../middleware/verifyDiscordSignature';
import { APIGatewayProxyEventV2 } from 'aws-lambda';

export async function handler(event: APIGatewayProxyEventV2) {
  const body = JSON.parse(event.body as string) as APIInteraction;

  console.log('Received request:', body);

  if (body.type === InteractionType.Ping) {
    return response({
      statusCode: 200,
      body: JSON.stringify({ type: InteractionResponseType.Pong }),
    });
  }

  if (body.type === InteractionType.ApplicationCommand) {
    const handler = commandHandlers[body.data.name as Commands];

    if (handler) {
      const result = await handler(body, event);

      return response({
        statusCode: 200,
        body: JSON.stringify(result),
      });
    }
  }

  return response({
    statusCode: 404,
  });
}

export default middy(handler)
  .use(httpHeaderNormalizer())
  .use(verifyDiscordSignature());
