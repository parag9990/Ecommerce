import { afterEach, describe, expect, it, vi } from "vitest";

import {
  listAdminSellers,
  listSellerCatalog,
  listSellerKycDocuments,
  toAdminSellersQueryString,
  updateSellerStatus
} from "../../src/features/sellers/api/sellers-api";
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

describe("sellers API", () => {
  it("builds seller list query strings with page_size and without all-status noise", () => {
    expect(
      toAdminSellersQueryString({
        q: " acme ",
        status: "all",
        page: 2,
        page_size: 20
      })
    ).toBe("q=acme&page=2&page_size=20");
  });

  it("lists admin sellers with search, status, pagination, and auth headers", async () => {
    const fetchMock = prepareApiTest({
      sellers: [
        {
          seller_id: "seller_123",
          user_id: "user_123",
          store_name: "Acme Store",
          status: "pending_review",
          gst_number: "27ABCDE1234F1Z5"
        }
      ],
      total: 1
    });

    const response = await listAdminSellers({
      q: "acme",
      status: "pending_review",
      page: 1,
      page_size: 20
    });

    const { requestUrl, requestInit } = getFetchCall(fetchMock);

    expect(requestUrl.pathname).toBe("/api/v1/admin/sellers");
    expect(requestUrl.searchParams.get("q")).toBe("acme");
    expect(requestUrl.searchParams.get("status")).toBe("pending_review");
    expect(requestUrl.searchParams.get("page_size")).toBe("20");
    expect(requestInit.headers).toMatchObject({
      Authorization: "Bearer admin-token"
    });
    expect(response.sellers[0].store_name).toBe("Acme Store");
  });

  it("patches seller status with the documented status and reason payload", async () => {
    const fetchMock = prepareApiTest({ success: true });

    await updateSellerStatus({
      sellerId: "seller/123",
      status: "suspended",
      reason: "policy_violation: Counterfeit product investigation"
    });

    const { requestUrl, requestInit } = getFetchCall(fetchMock);

    expect(requestUrl.pathname).toBe("/api/v1/admin/sellers/seller%2F123/status");
    expect(requestInit.method).toBe("PATCH");
    expect(JSON.parse(requestInit.body as string)).toEqual({
      status: "suspended",
      reason: "policy_violation: Counterfeit product investigation"
    });
  });

  it("uses admin-safe KYC and catalog review endpoints", async () => {
    const fetchMock = prepareApiTest({ documents: [], products: [] });

    await listSellerKycDocuments("seller_123");
    await listSellerCatalog("seller_123");

    const kycUrl = new URL(String(fetchMock.mock.calls[0]?.[0]));
    const catalogUrl = new URL(String(fetchMock.mock.calls[1]?.[0]));

    expect(kycUrl.pathname).toBe("/api/v1/admin/sellers/seller_123/kyc-documents");
    expect(catalogUrl.pathname).toBe("/api/v1/admin/sellers/seller_123/catalog");
  });
});
