import { APIInteraction, APIInteractionResponse } from 'discord-api-types/v10';
import { APIGatewayProxyEventV2 } from 'aws-lambda';

export enum Commands {
  Kolego = 'kolego',
  CoTam = 'cotam',
}

export enum KolegoOptions {
  Question = 'pytanie',
  Insult = 'obraź',
}

export interface CommandHandlerResult {
  content: string;
}

export type CommandHandler = (
  interaction: APIInteraction,
  event: APIGatewayProxyEventV2
) => Promise<APIInteractionResponse>;
