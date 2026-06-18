import { useCallback } from 'react';

import { trackSessionEvent } from '../api/session-events.grpc';

function currentPageUrl(): string {
  return `${window.location.pathname}${window.location.search}`;
}

export function useTrackEvent() {
  return useCallback((eventName: string, metadata?: Record<string, unknown>) => {
    void trackSessionEvent({
      eventName,
      metadata,
      pageUrl: currentPageUrl(),
    }).catch(() => {
      // Analytics transport failure must not block the buyer journey.
    });
  }, []);
}
