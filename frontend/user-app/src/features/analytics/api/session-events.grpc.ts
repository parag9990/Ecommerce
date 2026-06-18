import type { CallOptions } from '@connectrpc/connect';
import type { SessionEventMetadata } from '@ecommerce/proto-client';

import { grpcClients, type GrpcClients } from '../../../lib/grpc-client';
import { normalizeGrpcError } from '../../../lib/grpc-errors';
import { ApiError } from '../../../lib/http';
import { useSessionStore } from '../../../stores/session-store';

export type TrackSessionEventInput = Readonly<{
  cartId?: string | undefined;
  eventName: string;
  metadata?: Record<string, unknown> | undefined;
  pageUrl: string;
  productId?: string | undefined;
}>;

type TrackSessionEventOptions = Readonly<{
  client?: Pick<GrpcClients['session'], 'ingestEvent'> | undefined;
  signal?: AbortSignal | undefined;
  timeoutMs?: number | undefined;
}>;

function requireNonEmpty(value: string, fieldName: string): string {
  const trimmedValue = value.trim();

  if (!trimmedValue) {
    throw new ApiError(`${fieldName} is required.`, 'VALIDATION_ERROR', 400);
  }

  return trimmedValue;
}

function stringifyMetadata(
  metadata: Record<string, unknown> | undefined,
): SessionEventMetadata {
  if (!metadata) {
    return {};
  }

  return Object.fromEntries(
    Object.entries(metadata)
      .filter(([key, value]) => key.trim().length > 0 && value !== undefined)
      .map(([key, value]) => [key, String(value)]),
  );
}

function buildCallOptions(options: TrackSessionEventOptions): CallOptions {
  const callOptions: CallOptions = {};

  if (options.signal) {
    callOptions.signal = options.signal;
  }

  if (options.timeoutMs !== undefined) {
    callOptions.timeoutMs = options.timeoutMs;
  }

  return callOptions;
}

export async function trackSessionEvent(
  input: TrackSessionEventInput,
  options: TrackSessionEventOptions = {},
): Promise<void> {
  const { anonymousId, sessionId } = useSessionStore.getState();
  const client = options.client ?? grpcClients.session;

  try {
    await client.ingestEvent(
      {
        anonymousId,
        cartId: input.cartId ?? '',
        eventName: requireNonEmpty(input.eventName, 'Event name'),
        metadata: stringifyMetadata(input.metadata),
        occurredAt: new Date().toISOString(),
        pageUrl: requireNonEmpty(input.pageUrl, 'Page URL'),
        productId: input.productId ?? '',
        sessionId,
      },
      buildCallOptions(options),
    );
  } catch (error) {
    throw normalizeGrpcError(error);
  }
}
