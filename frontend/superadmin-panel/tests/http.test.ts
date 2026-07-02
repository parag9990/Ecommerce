import { afterEach, describe, expect, it, vi } from "vitest";

import { apiFetch } from "../src/lib/http";

afterEach(() => {
  vi.unstubAllEnvs();
  vi.unstubAllGlobals();
});

describe("apiFetch URL composition", () => {
  it("joins the gateway origin and a versioned API path exactly once", async () => {
    vi.stubEnv("VITE_API_BASE_URL", "https://gateway.example.test/");

    const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) =>
      new Response(JSON.stringify({ success: true }), {
        status: 200,
        headers: { "Content-Type": "application/json" }
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    await apiFetch<{ success: boolean }>("/api/v1/admin/users");

    expect(fetchMock).toHaveBeenCalledTimes(1);
    const requestUrl = new URL(String(fetchMock.mock.calls[0]?.[0]));
    expect(requestUrl.origin).toBe("https://gateway.example.test");
    expect(requestUrl.pathname).toBe("/api/v1/admin/users");
  });

  it("surfaces nested auth and gateway error payload messages", async () => {
    vi.stubEnv("VITE_API_BASE_URL", "https://gateway.example.test/");

    const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) =>
      new Response(
        JSON.stringify({
          request_id: "req_nested",
          error: {
            code: "INVALID_REQUEST",
            message: "json: unknown field \"source\""
          }
        }),
        {
          status: 400,
          headers: { "Content-Type": "application/json" }
        }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(apiFetch("/api/v1/auth/login")).rejects.toMatchObject({
      status: 400,
      code: "INVALID_REQUEST",
      message: 'json: unknown field "source"',
      requestId: "req_nested"
    });
  });
});
