import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";

import { useSessionJourney } from "./use-session-journey";

describe("useSessionJourney", () => {
  it("does not fetch without a session id", () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const { result } = renderHook(() => useSessionJourney(undefined), {
      wrapper: createWrapper()
    });

    expect(result.current.fetchStatus).toBe("idle");
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("fetches the journey when a session id is provided", async () => {
    const fetchMock = vi.fn(async () => jsonResponse(journeyResponse()));
    vi.stubGlobal("fetch", fetchMock);

    const { result } = renderHook(() => useSessionJourney("sess_123"), {
      wrapper: createWrapper()
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.session.sessionId).toBe("sess_123");
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
});

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false
      }
    }
  });

  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
  };
}

function jsonResponse(body: unknown, init: ResponseInit = {}) {
  return new Response(JSON.stringify(body), {
    status: 200,
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...init.headers
    }
  });
}

function journeyResponse() {
  return {
    events: [
      {
        _id: "evt_1",
        event_type: "page_view",
        occurred_at: "2026-05-18T00:00:02Z",
        path: "/",
        properties: { title: "Home" },
        session_id: "sess_123"
      }
    ],
    session: {
      anonymous_id: "anon_1234567890",
      device: { type: "mobile" },
      last_seen_at: "2026-05-18T00:01:00Z",
      session_id: "sess_123",
      started_at: "2026-05-18T00:00:00Z"
    }
  };
}
