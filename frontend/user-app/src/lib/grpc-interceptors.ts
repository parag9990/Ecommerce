import type { Interceptor } from '@connectrpc/connect';

import { useSessionStore } from '../stores/session-store';
import { getAccessToken } from './auth-session';

export const USER_APP_GRPC_REQUEST_SOURCE = 'user-app';

export function createRequestId(): string {
  if (globalThis.crypto && 'randomUUID' in globalThis.crypto) {
    return globalThis.crypto.randomUUID();
  }

  return `req_${Date.now()}_${Math.random().toString(16).slice(2)}`;
}

export function readGrpcSessionHeaders(): Record<string, string> {
  const { anonymousId, sessionId } = useSessionStore.getState();

  return {
    'x-anonymous-id': anonymousId,
    'x-session-id': sessionId,
  };
}

export const grpcRequestContextInterceptor: Interceptor =
  (next) => async (request) => {
    const token = getAccessToken();

    request.header.set('x-request-id', createRequestId());
    request.header.set('x-request-source', USER_APP_GRPC_REQUEST_SOURCE);

    for (const [key, value] of Object.entries(readGrpcSessionHeaders())) {
      request.header.set(key, value);
    }

    if (token) {
      request.header.set('authorization', `Bearer ${token}`);
    }

    return next(request);
  };

export const grpcInterceptors: Interceptor[] = [grpcRequestContextInterceptor];
