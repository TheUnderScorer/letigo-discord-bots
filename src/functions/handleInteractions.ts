import { RouteHandler } from '../aws.types';
import * as nacl from 'tweetnacl';
import {
  APIInteraction,
  InteractionResponseType,
  InteractionType,
} from 'discord-api-types/v10';
import * as middy from 'middy';
import httpHeaderNormalizer from '@middy/http-header-normalizer';
import { commandHandlers } from '../commandHandlers/commandHandlers';
import { Commands } from '../command.types';

export const handler: RouteHandler = async event => {
  const strBody = event.body as string;

  try {
    const publicKey = process.env.PUBLIC_KEY as string;
    const signature = event.headers['x-signature-ed25519'] as string;
    const timestamp = event.headers['x-signature-timestamp'] as string;

    const isVerified = nacl.sign.detached.verify(
      Buffer.from(timestamp + strBody),
      Buffer.from(signature, 'hex'),
      Buffer.from(publicKey, 'hex')
    );

    if (!isVerified) {
      return {
        statusCode: 401,
        body: JSON.stringify('invalid request signature'),
      };
    }
  } catch (error) {
    console.error(error);

    return {
      statusCode: 500,
    };
  }

  const body = JSON.parse(strBody) as APIInteraction;

  console.log('Received request:', body);

  if (body.type === InteractionType.Ping) {
    return {
      statusCode: 200,
      body: JSON.stringify({ type: InteractionResponseType.Pong }),
    };
  }

  if (body.type === InteractionType.ApplicationCommand) {
    const handler = commandHandlers[body.data.name as Commands];

    if (handler) {
      const response = await handler(body, event);

      return {
        statusCode: 200,
        body: JSON.stringify(response),
      };
    }
  }

  return {
    statusCode: 404,
  };
};

export default middy(handler).use(httpHeaderNormalizer());
