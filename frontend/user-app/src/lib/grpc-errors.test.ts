import { Code, ConnectError } from '@connectrpc/connect';
import { describe, expect, it } from 'vitest';

import { ApiError } from './http';
import { grpcCodeToHttpStatus, normalizeGrpcError } from './grpc-errors';

describe('normalizeGrpcError', () => {
  it('maps Connect errors to ApiError with request id metadata', () => {
    const error = new ConnectError('Login required', Code.Unauthenticated, {
      'x-request-id': 'req_test',
    });

    const normalizedError = normalizeGrpcError(error);

    expect(normalizedError).toBeInstanceOf(ApiError);
    expect(normalizedError.code).toBe('GRPC_UNAUTHENTICATED');
    expect(normalizedError.message).toBe('Login required');
    expect(normalizedError.requestId).toBe('req_test');
    expect(normalizedError.status).toBe(401);
  });

  it('keeps existing ApiError instances untouched', () => {
    const apiError = new ApiError('Invalid input', 'VALIDATION_ERROR', 400);

    expect(normalizeGrpcError(apiError)).toBe(apiError);
  });
});

describe('grpcCodeToHttpStatus', () => {
  it('marks unavailable errors as retryable server failures', () => {
    expect(grpcCodeToHttpStatus(Code.Unavailable)).toBe(503);
  });
});
