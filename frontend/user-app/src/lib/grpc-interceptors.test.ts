import { create } from '@bufbuild/protobuf';
import {
  createContextValues,
  type UnaryRequest,
  type UnaryResponse,
} from '@connectrpc/connect';
import {
  IngestEventRequestSchema,
  IngestEventResponseSchema,
  SessionService,
} from '@ecommerce/proto-client/gen/ecommerce/session/v1/session_pb';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { persistAuthSession } from './auth-session';
import { grpcRequestContextInterceptor } from './grpc-interceptors';
import { useSessionStore } from '../stores/session-store';

type SessionUnaryRequest = UnaryRequest<
  typeof IngestEventRequestSchema,
  typeof IngestEventResponseSchema
>;

type SessionUnaryResponse = UnaryResponse<
  typeof IngestEventRequestSchema,
  typeof IngestEventResponseSchema
>;

afterEach(() => {
  window.sessionStorage.clear();
});

function createRequest(): SessionUnaryRequest {
  return {
    contextValues: createContextValues(),
    header: new Headers(),
    message: create(IngestEventRequestSchema),
    method: SessionService.method.ingestEvent,
    requestMethod: 'POST',
    service: SessionService,
    signal: new AbortController().signal,
    stream: false,
    url: 'http://localhost:8082/ecommerce.session.v1.SessionService/IngestEvent',
  };
}

function createResponse(): SessionUnaryResponse {
  return {
    header: new Headers(),
    message: create(IngestEventResponseSchema, {
      accepted: true,
      eventId: 'evt_test',
    }),
    method: SessionService.method.ingestEvent,
    service: SessionService,
    stream: false,
    trailer: new Headers(),
  };
}

describe('grpcRequestContextInterceptor', () => {
  it('adds auth, request, and session headers', async () => {
    persistAuthSession({
      tokens: {
        access_token: 'access_test',
        expires_in: 60,
      },
    });
    useSessionStore.setState({
      anonymousId: 'anon_test',
      sessionId: 'session_test',
      startedAt: '2026-06-01T00:00:00.000Z',
    });

    const request = createRequest();
    const next = vi.fn(() => Promise.resolve(createResponse()));

    await grpcRequestContextInterceptor(next)(request);

    expect(request.header.get('authorization')).toBe('Bearer access_test');
    expect(request.header.get('x-anonymous-id')).toBe('anon_test');
    expect(request.header.get('x-request-id')).toBeTruthy();
    expect(request.header.get('x-request-source')).toBe('user-app');
    expect(request.header.get('x-session-id')).toBe('session_test');
    expect(next).toHaveBeenCalledOnce();
  });
});
