import { afterEach, describe, expect, it, vi } from "vitest";

import {
  listAdminUsers,
  listUserSessions,
  toAdminUsersQueryString,
  updateUserStatus
} from "../../src/features/users/api/users-api";
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
      roles: ["superadmin"]
    }
  });

  const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) =>
    mockJsonResponse(responseBody)
  );
  vi.stubGlobal("fetch", fetchMock);

  return fetchMock;
}

function getFetchCall(fetchMock: ReturnType<typeof prepareApiTest>) {
  const call = fetchMock.mock.calls[0];

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

describe("users API", () => {
  it("builds list query strings without all-status noise", () => {
    expect(
      toAdminUsersQueryString({
        q: " rahul ",
        status: "all",
        page: 2,
        limit: 25
      })
    ).toBe("q=rahul&page=2&page_size=25");
  });

  it("lists admin users with search, status, pagination, and auth headers", async () => {
    const fetchMock = prepareApiTest({
      users: [
        {
          user_id: "user_123",
          email: "rahul@example.com",
          phone: "+919999999999",
          full_name: "Rahul Sharma",
          status: "active",
          roles: ["buyer"]
        }
      ]
    });

    const response = await listAdminUsers({
      q: "rahul",
      status: "active",
      page: 1,
      limit: 25
    });

    const { requestUrl, requestInit } = getFetchCall(fetchMock);

    expect(requestUrl.pathname).toBe("/api/v1/admin/users");
    expect(requestUrl.searchParams.get("q")).toBe("rahul");
    expect(requestUrl.searchParams.get("status")).toBe("active");
    expect(requestUrl.searchParams.get("page_size")).toBe("25");
    expect(requestInit.headers).toMatchObject({
      Authorization: "Bearer admin-token"
    });
    expect(response.users[0].roles).toEqual(["buyer"]);
  });

  it("patches user status with a required reason payload", async () => {
    const fetchMock = prepareApiTest({ success: true });

    await updateUserStatus({
      userId: "user/123",
      status: "blocked",
      reason: "Repeated suspicious checkout attempts"
    });

    const { requestUrl, requestInit } = getFetchCall(fetchMock);

    expect(requestUrl.pathname).toBe("/api/v1/admin/users/user%2F123/status");
    expect(requestInit.method).toBe("PATCH");
    expect(JSON.parse(requestInit.body as string)).toEqual({
      status: "blocked",
      reason: "Repeated suspicious checkout attempts"
    });
  });

  it("uses the session management admin analytics endpoint for user sessions", async () => {
    const fetchMock = prepareApiTest({
      sessions: [
        {
          session_id: "sess_123",
          anonymous_id: "anon_456",
          user_id: "user_123",
          started_at: "2026-06-01T10:00:00Z",
          last_seen_at: "2026-06-01T10:42:00Z",
          device: {
            browser: "Chrome",
            os: "Windows",
            device_type: "desktop"
          }
        }
      ]
    });

    await listUserSessions("user_123");

    const { requestUrl } = getFetchCall(fetchMock);

    expect(requestUrl.pathname).toBe("/api/v1/analytics/sessions");
    expect(requestUrl.searchParams.get("user_id")).toBe("user_123");
    expect(requestUrl.searchParams.get("page_size")).toBe("20");
  });
});
