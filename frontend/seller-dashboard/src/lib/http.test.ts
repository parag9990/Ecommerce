import { afterEach, describe, expect, it, vi } from "vitest";

import { clearSellerAuthSession, writeSellerAuthSession } from "./auth-session";
import { http } from "./http";

describe("http client", () => {
  afterEach(() => {
    clearSellerAuthSession();
    vi.restoreAllMocks();
    vi.unstubAllEnvs();
  });

  it("does not duplicate the API base path when env includes it", async () => {
    vi.stubEnv("VITE_API_BASE_URL", "http://localhost:8080/api/v1");
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ data: { ok: true }, request_id: "req_1" }), {
        status: 200,
        headers: { "content-type": "application/json" },
      }),
    );

    await expect(http<{ ok: boolean }>("/api/v1/seller/session")).resolves.toEqual({
      ok: true,
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/seller/session",
      expect.any(Object),
    );
    fetchMock.mockRestore();
  });

  it("attaches the stored seller access token", async () => {
    writeSellerAuthSession({
      accessToken: "seller-token",
      roles: ["seller"],
      sellerId: "seller_1",
      userId: "user_1",
    });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ data: { ok: true } }), {
        status: 200,
        headers: { "content-type": "application/json" },
      }),
    );

    await expect(http<{ ok: boolean }>("/api/v1/seller/session")).resolves.toEqual({
      ok: true,
    });

    const requestOptions = fetchMock.mock.calls[0]?.[1] as RequestInit | undefined;
    const headers = requestOptions?.headers as Headers;
    expect(headers.get("authorization")).toBe("Bearer seller-token");
  });
});
