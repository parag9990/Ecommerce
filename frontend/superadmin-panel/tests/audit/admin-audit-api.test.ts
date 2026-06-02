import { afterEach, describe, expect, it, vi } from "vitest";

import {
  exportAdminAuditLogs,
  listAdminAuditLogs,
  toAdminAuditLogQueryString
} from "../../src/features/audit/api/admin-audit-api";
import { useAuthStore } from "../../src/stores/auth-store";

function mockJsonResponse(body: unknown) {
  return new Response(JSON.stringify(body), {
    status: 200,
    headers: {
      "Content-Type": "application/json"
    }
  });
}

function prepareApiTest(responseBody: unknown | Response) {
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
    responseBody instanceof Response ? responseBody : mockJsonResponse(responseBody)
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

describe("admin audit API", () => {
  it("builds audit query strings without empty or all-filter noise", () => {
    const query = toAdminAuditLogQueryString({
      actor_id: " admin_1 ",
      action: "",
      resource_type: "all",
      resource_id: "seller_1",
      request_id: " req_1 ",
      from: "2026-06-01T10:00:00Z",
      to: "",
      page: 2,
      page_size: 50
    });
    const params = new URLSearchParams(query);

    expect(params.get("page")).toBe("2");
    expect(params.get("page_size")).toBe("50");
    expect(params.get("actor_id")).toBe("admin_1");
    expect(params.get("resource_type")).toBeNull();
    expect(params.get("resource_id")).toBe("seller_1");
    expect(params.get("request_id")).toBe("req_1");
    expect(params.get("from")).toBe("2026-06-01T10:00:00.000Z");
  });

  it("lists audit logs with auth headers and normalized records", async () => {
    const fetchMock = prepareApiTest({
      logs: [
        {
          id: 123,
          actor_id: "admin_1",
          actor_role: "superadmin",
          action: "seller.suspended",
          resource_type: "seller",
          resource_id: "seller_1",
          request_id: "req_1",
          ip_hash: "sha256:hash",
          before_summary: { status: "active" },
          after_summary: { status: "suspended" },
          reason: "Policy violation confirmed",
          created_at: "2026-06-02T10:00:00Z"
        }
      ],
      total: 1
    });

    const response = await listAdminAuditLogs({
      actor_id: "admin_1",
      page: 1,
      page_size: 25
    });
    const { requestUrl, requestInit } = getFetchCall(fetchMock);

    expect(requestUrl.pathname).toBe("/api/v1/admin/audit-logs");
    expect(requestUrl.searchParams.get("actor_id")).toBe("admin_1");
    expect(requestInit.headers).toMatchObject({ Authorization: "Bearer admin-token" });
    expect(response.logs).toHaveLength(1);
    expect(response.logs[0]?.id).toBe("123");
    expect(response.logs[0]?.actor_admin_id).toBe("admin_1");
    expect(response.logs[0]?.ip_hash).toBe("sha256:hash");
    expect(response.total).toBe(1);
  });

  it("posts audit export requests as CSV blob requests with reason", async () => {
    const fetchMock = prepareApiTest(
      new Response("id,action\nlog_1,audit_logs.exported", {
        status: 200,
        headers: { "Content-Type": "text/csv" }
      })
    );

    const blob = await exportAdminAuditLogs({
      filters: {
        resource_type: "refund",
        page: 1,
        page_size: 25
      },
      reason: "Quarterly finance audit export"
    });
    const { requestUrl, requestInit } = getFetchCall(fetchMock);

    expect(requestUrl.pathname).toBe("/api/v1/admin/audit-logs/export");
    expect(requestInit.method).toBe("POST");
    expect(requestInit.headers).toMatchObject({
      Authorization: "Bearer admin-token",
      Accept: "text/csv,application/octet-stream,*/*"
    });
    expect(JSON.parse(requestInit.body as string)).toEqual({
      filters: {
        page: 1,
        page_size: 25,
        resource_type: "refund"
      },
      reason: "Quarterly finance audit export"
    });
    expect(await blob.text()).toContain("audit_logs.exported");
  });
});
