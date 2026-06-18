import { Code, ConnectError } from '@connectrpc/connect';

import { ApiError } from './http';

function grpcCodeName(code: Code): string {
  return Code[code]?.replace(/([a-z])([A-Z])/g, '$1_$2').toUpperCase() ?? 'UNKNOWN';
}

export function grpcCodeToHttpStatus(code: Code): number {
  switch (code) {
    case Code.Canceled:
      return 499;
    case Code.InvalidArgument:
    case Code.FailedPrecondition:
    case Code.OutOfRange:
      return 400;
    case Code.Unauthenticated:
      return 401;
    case Code.PermissionDenied:
      return 403;
    case Code.NotFound:
      return 404;
    case Code.AlreadyExists:
      return 409;
    case Code.ResourceExhausted:
      return 429;
    case Code.DeadlineExceeded:
      return 504;
    case Code.Unimplemented:
      return 501;
    case Code.Unavailable:
      return 503;
    case Code.Aborted:
    case Code.DataLoss:
    case Code.Internal:
    case Code.Unknown:
      return 500;
  }
}

export function normalizeGrpcError(error: unknown): ApiError {
  if (error instanceof ApiError) {
    return error;
  }

  if (error instanceof ConnectError) {
    const requestId = error.metadata.get('x-request-id') ?? undefined;
    const message = error.rawMessage || 'Request failed. Please try again.';

    return new ApiError(
      message,
      `GRPC_${grpcCodeName(error.code)}`,
      grpcCodeToHttpStatus(error.code),
      requestId,
    );
  }

  return new ApiError(
    'Network error. Please check your connection and try again.',
    'NETWORK_ERROR',
    0,
  );
}
