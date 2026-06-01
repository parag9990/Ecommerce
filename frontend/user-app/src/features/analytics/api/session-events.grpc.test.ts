import type {
  IngestEventResponse,
} from '@ecommerce/proto-client/gen/ecommerce/session/v1/session_pb';
import { describe, expect, it, vi } from 'vitest';

import type { GrpcClients } from '../../../lib/grpc-client';
import { useSessionStore } from '../../../stores/session-store';
import { trackSessionEvent } from './session-events.grpc';

describe('trackSessionEvent', () => {
  it('sends normalized analytics payloads through the session client', async () => {
    useSessionStore.setState({
      anonymousId: 'anon_test',
      sessionId: 'session_test',
      startedAt: '2026-06-01T00:00:00.000Z',
    });

    const ingestEvent = vi
      .fn<GrpcClients['session']['ingestEvent']>()
      .mockResolvedValue({
        $typeName: 'ecommerce.session.v1.IngestEventResponse',
        accepted: true,
        eventId: 'evt_test',
      } satisfies IngestEventResponse);
    const client: Pick<GrpcClients['session'], 'ingestEvent'> = { ingestEvent };

    await trackSessionEvent(
      {
        eventName: 'product_view',
        metadata: {
          position: 2,
          source: 'home',
        },
        pageUrl: '/products/prod_123',
        productId: 'prod_123',
      },
      { client },
    );

    expect(ingestEvent).toHaveBeenCalledWith(
      expect.objectContaining({
        anonymousId: 'anon_test',
        eventName: 'product_view',
        metadata: {
          position: '2',
          source: 'home',
        },
        productId: 'prod_123',
        sessionId: 'session_test',
      }),
      {},
    );
  });
});
