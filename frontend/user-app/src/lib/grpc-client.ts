import { createClient, type Client, type Transport } from '@connectrpc/connect';
import { createGrpcWebTransport } from '@connectrpc/connect-web';
import {
  RecommendationService,
  SearchService,
  SessionService,
} from '@ecommerce/proto-client';

import { env } from './env';
import { grpcInterceptors } from './grpc-interceptors';

export type GrpcClients = Readonly<{
  recommendation: Client<typeof RecommendationService>;
  search: Client<typeof SearchService>;
  session: Client<typeof SessionService>;
}>;

function normalizeBaseUrl(baseUrl: string): string {
  return baseUrl.replace(/\/+$/, '');
}

const fetchWithCredentials: typeof fetch = (input, init) =>
  fetch(input, {
    ...init,
    credentials: 'include',
  });

export function createUserGrpcTransport(): Transport {
  return createGrpcWebTransport({
    baseUrl: normalizeBaseUrl(env.grpcWebBaseUrl),
    defaultTimeoutMs: env.grpcWebTimeoutMs,
    fetch: fetchWithCredentials,
    interceptors: grpcInterceptors,
  });
}

export function createGrpcClients(
  transport: Transport = createUserGrpcTransport(),
): GrpcClients {
  return {
    recommendation: createClient(RecommendationService, transport),
    search: createClient(SearchService, transport),
    session: createClient(SessionService, transport),
  };
}

export const grpcClients = createGrpcClients();
