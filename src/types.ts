import type {
  APIGatewayProxyEventV2,
  APIGatewayProxyResultV2,
  Context,
} from 'aws-lambda';

export type RouteHandler<Req = APIGatewayProxyEventV2> = (
  request: Req,
  context: Context
) => Promise<APIGatewayProxyResultV2>;
