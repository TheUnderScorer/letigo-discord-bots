import { RouteHandler } from '../aws.types';
import * as nacl from 'tweetnacl';
import { InteractionType, APIInteraction } from 'discord-api-types/v10';

export const handler: RouteHandler = async event => {
  const publicKey = process.env.PUBLIC_KEY as string;
  const signature = event.headers['x-signature-ed25519'] as string;
  const timestamp = event.headers['x-signature-timestamp'] as string;
  const strBody = event.body as string;

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

  const body = JSON.parse(strBody) as APIInteraction;

  console.log('Received request:', body);

  if (body.type === InteractionType.Ping) {
    return {
      statusCode: 200,
      body: JSON.stringify({ type: InteractionType.Ping }),
    };
  }

  return {
    statusCode: 404,
  };
};
