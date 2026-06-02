import { afterEach, describe, expect, it, vi } from "vitest";

import {
  getAdminSessionJourney,
  getLiveMetrics,
  listAdminSessions,
  toAdminSessionsQueryString
} from "../../src/features/sessions/api/sessions-api";
import { useAuthStore } from "../../src/stores/auth-store";

function mockJsonResponse(body: unknown) {
  return new Response(JSON.stringify(body), {
    status: 200,
    headers: {
      "Content-Type": "application/json"
    }
  });
}

function prepareApiTest(responseBody: unknown) {
  vi.stubEnv("VITE_API_BASE_URL", "https://api.example.test");
  useAuthStore.getState().setSession({
    accessToken: "admin-token",
    user: {
      id: "admin-1",
      email: "admin@example.com",
      name: "Admin User",
      roles: ["operations_admin"]
    }
  });

  const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) =>
    mockJsonResponse(responseBody)
  );
  vi.stubGlobal("fetch", fetchMock);

  return fetchMock;
}

function getFetchCall(fetchMock: ReturnType<typeof prepareApiTest>, index = 0) {
  const call = fetchMock.mock.calls[index];

  if (!call) {
    throw new Error("Expected fetch to be called.");
  }

  const [input, init = {}] = call;

  return {
    requestUrl: new URL(String(input)),
    requestInit: init
  };
}

afterEach(() => {
  vi.unstubAllEnvs();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe("sessions API", () => {
  it("builds list query strings from supported server filters", () => {
    expect(
      toAdminSessionsQueryString({
        user_id: " user_123 ",
        from: "2026-06-01T10:00",
        to: "2026-06-01T12:00",
        page: 2,
        limit: 25
      })
    ).toBe("page=2&limit=25&user_id=user_123&from=2026-06-01T10%3A00&to=2026-06-01T12%3A00");
  });

  it("loads live metrics with auth headers and normalized defaults", async () => {
    const fetchMock = prepareApiTest({
      active_users: 12,
      active_sessions: 18,
      events_per_minute: 90
    });

    const response = await getLiveMetrics();
    const { requestUrl, requestInit } = getFetchCall(fetchMock);

    expect(requestUrl.pathname).toBe("/api/v1/analytics/live");
    expect(requestInit.headers).toMatchObject({
      Authorization: "Bearer admin-token"
    });
    expect(response.suspicious_sessions).toBe(0);
  });

  it("lists admin sessions through analytics endpoint and normalizes devices", async () => {
    const fetchMock = prepareApiTest({
      sessions: [
        {
          session_id: "sess_123",
          anonymous_id: "anon_123",
          started_at: "2026-06-01T10:00:00Z",
          last_seen_at: "2026-06-01T10:05:00Z",
          device: {
            browser: "Chrome",
            os: "Windows",
            device_type: "desktop"
          }
        }
      ],
      total: 1
    });

    const response = await listAdminSessions({
      user_id: "user_123",
      from: "2026-06-01T10:00",
      to: "2026-06-01T12:00",
      page: 1,
      limit: 25
    });
    const { requestUrl } = getFetchCall(fetchMock);

    expect(requestUrl.pathname).toBe("/api/v1/analytics/sessions");
    expect(requestUrl.searchParams.get("user_id")).toBe("user_123");
    expect(response.sessions[0].device?.operating_system).toBe("Windows");
    expect(response.sessions[0].status).toBe("active");
  });

  it("loads journey detail with encoded session id and ordered events", async () => {
    const fetchMock = prepareApiTest({
      session: {
        session_id: "sess/123",
        anonymous_id: "anon_123",
        started_at: "2026-06-01T10:00:00Z",
        last_seen_at: "2026-06-01T10:05:00Z"
      },
      events: [
        {
          event_type: "checkout_step",
          anonymous_id: "anon_123",
          session_id: "sess/123",
          occurred_at: "2026-06-01T10:05:00Z"
        },
        {
          event_type: "page_view",
          anonymous_id: "anon_123",
          session_id: "sess/123",
          occurred_at: "2026-06-01T10:01:00Z"
        }
      ]
    });

    const response = await getAdminSessionJourney("sess/123");
    const { requestUrl } = getFetchCall(fetchMock);

    expect(requestUrl.pathname).toBe("/api/v1/analytics/sessions/sess%2F123/journey");
    expect(response.events.map((event) => event.event_type)).toEqual(["page_view", "checkout_step"]);
  });
});
