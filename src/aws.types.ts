import { APIGatewayProxyStructuredResultV2 } from 'aws-lambda/trigger/api-gateway-proxy';

export const response = (data: APIGatewayProxyStructuredResultV2) => data;
