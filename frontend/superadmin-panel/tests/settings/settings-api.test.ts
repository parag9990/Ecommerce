import { afterEach, describe, expect, it, vi } from "vitest";

import {
  listPlatformSettings,
  updatePlatformSetting
} from "../../src/features/settings/api/settings-api";
import {
  createSearchSynonym,
  deleteSearchSynonym,
  listSearchSynonyms,
  updateSearchSynonym
} from "../../src/features/settings/api/search-synonyms-api";
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

describe("settings API", () => {
  it("lists known platform settings and ignores unsupported keys", async () => {
    const fetchMock = prepareApiTest({
      settings: [
        {
          key: "commission.default_rate",
          value: { rate_percent: 12, applies_to: "all_sellers" },
          updated_at: "2026-06-02T10:00:00Z"
        },
        {
          key: "unsupported.secret",
          value: { token: "hidden" }
        }
      ]
    });

    const response = await listPlatformSettings();
    const { requestUrl, requestInit } = getFetchCall(fetchMock);

    expect(requestUrl.pathname).toBe("/api/v1/admin/settings");
    expect(requestInit.headers).toMatchObject({ Authorization: "Bearer admin-token" });
    expect(response.settings).toHaveLength(1);
    expect(response.settings[0]?.key).toBe("commission.default_rate");
  });

  it("updates one platform setting with encoded key and reason", async () => {
    const fetchMock = prepareApiTest({
      key: "platform.maintenance_mode",
      value: { enabled: false },
      updated_at: "2026-06-02T10:00:00Z"
    });

    await updatePlatformSetting("platform.maintenance_mode", {
      value: {
        enabled: false,
        message: "",
        starts_at: null,
        ends_at: null,
        allow_admin_bypass: true
      },
      reason: "Maintenance window cancelled"
    });

    const { requestUrl, requestInit } = getFetchCall(fetchMock);

    expect(requestUrl.pathname).toBe("/api/v1/admin/settings/platform.maintenance_mode");
    expect(requestInit.method).toBe("PATCH");
    expect(JSON.parse(requestInit.body as string)).toMatchObject({
      reason: "Maintenance window cancelled",
      value: { enabled: false }
    });
  });

  it("uses search synonym endpoints with normalized audit payloads", async () => {
    const fetchMock = prepareApiTest({
      synonym_id: "syn/1",
      root: "mobile",
      synonyms: ["phone"]
    });

    await listSearchSynonyms();
    await createSearchSynonym({
      root: " Mobile ",
      synonyms: [" Phone ", "phone"],
      reason: "Improve mobile product discovery"
    });
    await updateSearchSynonym("syn/1", {
      root: "phone",
      synonyms: ["mobile"],
      reason: "Align search wording"
    });
    await deleteSearchSynonym("syn/1", "Remove stale synonym mapping");

    expect(getFetchCall(fetchMock, 0).requestUrl.pathname).toBe("/api/v1/admin/search/synonyms");
    expect(getFetchCall(fetchMock, 1).requestInit.method).toBe("POST");
    expect(JSON.parse(getFetchCall(fetchMock, 1).requestInit.body as string)).toEqual({
      root: "mobile",
      synonyms: ["phone"],
      reason: "Improve mobile product discovery"
    });
    expect(getFetchCall(fetchMock, 2).requestUrl.pathname).toBe(
      "/api/v1/admin/search/synonyms/syn%2F1"
    );
    expect(getFetchCall(fetchMock, 2).requestInit.method).toBe("PATCH");
    expect(getFetchCall(fetchMock, 3).requestInit.method).toBe("DELETE");
    expect(JSON.parse(getFetchCall(fetchMock, 3).requestInit.body as string)).toEqual({
      reason: "Remove stale synonym mapping"
    });
  });
});
