import { APIInteraction, APIInteractionResponse } from 'discord-api-types/v10';
import { APIGatewayProxyEventV2 } from 'aws-lambda';

export enum Commands {
  Kolego = 'kolego',
}

export enum KolegoOptions {
  Question = 'pytanie',
}

export interface CommandHandlerResult {
  content: string;
}

export type CommandHandler = (
  interaction: APIInteraction,
  event: APIGatewayProxyEventV2
) => Promise<APIInteractionResponse>;