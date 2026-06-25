import { afterEach, describe, expect, it, vi } from "vitest";

import { http } from "./http";

describe("http client", () => {
afterEach(() => {
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
});
