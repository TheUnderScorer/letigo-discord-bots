import * as nacl from 'tweetnacl';
import { Middleware } from 'middy';
import { APIGatewayProxyEventV2 } from 'aws-lambda';
import { response } from '../aws.types';

export const verifyDiscordSignature: Middleware<
  void,
  APIGatewayProxyEventV2
> = () => ({
  before: async ({ event }) => {
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
        return response({
          statusCode: 401,
          body: JSON.stringify('invalid request signature'),
        });
      }
    } catch (error) {
      console.error(error);

      return response({
        statusCode: 500,
      });
    }
  },
});
