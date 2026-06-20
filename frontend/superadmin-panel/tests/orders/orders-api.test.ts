import { afterEach, describe, expect, it, vi } from "vitest";

import {
  getAdminOrder,
  listAdminOrders,
  listOrderDisputes,
  submitOrderReview,
  toAdminOrdersQueryString
} from "../../src/features/orders/api/orders-api";
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

describe("orders API", () => {
  it("builds order query strings without all-status noise", () => {
    expect(
      toAdminOrdersQueryString({
        q: " order_123 ",
        status: "all",
        review_status: "manual_review",
        user_id: " user_123 ",
        seller_id: "",
        from: "2026-06-01",
        to: "2026-06-02",
        page: 2,
        limit: 25
      })
    ).toBe("q=order_123&review_status=manual_review&user_id=user_123&from=2026-06-01&to=2026-06-02&page=2&limit=25");
  });

  it("lists admin orders with filters, auth headers, and normalized records", async () => {
    const fetchMock = prepareApiTest({
      orders: [
        {
          order_id: "order_123",
          user_id: "user_123",
          status: "paid",
          review_status: "disputed",
          total: { amount: 159900, currency: "INR" },
          items: [
            {
              item_id: "item_1",
              product_id: "prod_1",
              seller_id: "seller_1",
              title: "Running shoes",
              quantity: 2,
              unit_price: { amount: 79950, currency: "INR" }
            }
          ],
          created_at: "2026-06-01T10:00:00Z"
        }
      ],
      total: 1
    });

    const response = await listAdminOrders({
      q: "order_123",
      status: "paid",
      review_status: "disputed",
      user_id: "user_123",
      seller_id: "seller_1",
      from: "2026-06-01",
      to: "2026-06-02",
      page: 1,
      limit: 25
    });

    const { requestUrl, requestInit } = getFetchCall(fetchMock);

    expect(requestUrl.pathname).toBe("/api/v1/admin/orders");
    expect(requestUrl.searchParams.get("q")).toBe("order_123");
    expect(requestUrl.searchParams.get("status")).toBe("paid");
    expect(requestUrl.searchParams.get("review_status")).toBe("disputed");
    expect(requestUrl.searchParams.get("user_id")).toBe("user_123");
    expect(requestUrl.searchParams.get("seller_id")).toBe("seller_1");
    expect(requestInit.headers).toMatchObject({
      Authorization: "Bearer admin-token"
    });
    expect(response.orders[0].items[0].seller_id).toBe("seller_1");
  });

  it("uses admin-safe order detail and dispute endpoints", async () => {
    const fetchMock = prepareApiTest({
      order: {
        order_id: "order/123",
        user_id: "user_123",
        status: "shipped",
        total: { amount: 99900, currency: "INR" },
        items: []
      },
      status_history: [{ status: "paid", actor_type: "payment", created_at: "2026-06-01T10:00:00Z" }],
      shipments: [{ shipment_id: "ship_123", status: "shipped", carrier: "Delhivery" }],
      disputes: []
    });

    const detail = await getAdminOrder("order/123");
    await listOrderDisputes("order/123");

    const detailUrl = getFetchCall(fetchMock, 0).requestUrl;
    const disputesUrl = getFetchCall(fetchMock, 1).requestUrl;

    expect(detailUrl.pathname).toBe("/api/v1/admin/orders/order%2F123");
    expect(disputesUrl.pathname).toBe("/api/v1/admin/orders/order%2F123/disputes");
    expect(detail.status_history[0].actor_type).toBe("payment");
    expect(detail.shipments[0].carrier).toBe("Delhivery");
  });

  it("submits manual review decisions without calling refund review APIs", async () => {
    const fetchMock = prepareApiTest({ success: true });

    await submitOrderReview("order/123", {
      decision: "escalate",
      reason: "Status mismatch needs senior support review",
      internal_note: "Buyer reported shipment delivered to wrong address."
    });

    const { requestUrl, requestInit } = getFetchCall(fetchMock);

    expect(requestUrl.pathname).toBe("/api/v1/admin/orders/order%2F123/review");
    expect(requestUrl.pathname).not.toContain("/refunds/");
    expect(requestInit.method).toBe("POST");
    expect(JSON.parse(requestInit.body as string)).toEqual({
      decision: "escalate",
      reason: "Status mismatch needs senior support review",
      internal_note: "Buyer reported shipment delivered to wrong address."
    });
  });
});
