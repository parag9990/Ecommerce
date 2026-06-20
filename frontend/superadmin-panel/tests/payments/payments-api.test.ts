import { afterEach, describe, expect, it, vi } from "vitest";

import {
  getAdminPayment,
  listAdminPayments,
  listReconciliationAlerts,
  listRefunds,
  reviewRefund,
  toAdminPaymentsQueryString
} from "../../src/features/payments/api/payments-api";
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
      roles: ["finance_admin"]
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

describe("payments API", () => {
  it("builds payment query strings without all-filter noise", () => {
    expect(
      toAdminPaymentsQueryString({
        order_id: " order_123 ",
        status: "all",
        provider: "razorpay",
        page: 2,
        page_size: 25
      })
    ).toBe("page=2&page_size=25&provider=razorpay&order_id=order_123");
  });

  it("lists admin payments with filters, auth headers, and normalized records", async () => {
    const fetchMock = prepareApiTest({
      payments: [
        {
          payment_id: "pay_123",
          order_id: "order_123",
          provider: "razorpay",
          status: "captured",
          amount: { amount: 159900, currency: "INR" },
          created_at: "2026-06-01T10:00:00Z"
        }
      ],
      total: 1
    });

    const response = await listAdminPayments({
      order_id: "order_123",
      status: "captured",
      provider: "razorpay",
      page: 1,
      page_size: 25
    });

    const { requestUrl, requestInit } = getFetchCall(fetchMock);

    expect(requestUrl.pathname).toBe("/api/v1/admin/payments");
    expect(requestUrl.searchParams.get("order_id")).toBe("order_123");
    expect(requestUrl.searchParams.get("status")).toBe("captured");
    expect(requestUrl.searchParams.get("provider")).toBe("razorpay");
    expect(requestInit.headers).toMatchObject({
      Authorization: "Bearer admin-token"
    });
    expect(response.payments[0].status).toBe("captured");
    expect(response.payments[0].amount.currency).toBe("INR");
  });

  it("uses admin-safe payment detail, refund queue, and reconciliation endpoints", async () => {
    const fetchMock = prepareApiTest({
      payment: {
        payment_id: "pay/123",
        order_id: "order_123",
        provider: "stripe",
        status: "captured",
        amount: { amount: 99900, currency: "INR" }
      },
      attempts: [{ attempt_id: "attempt_1", status: "captured" }],
      refunds: [{ refund_id: "refund_1", payment_id: "pay/123", status: "requested", amount: { amount: 9900 } }],
      alerts: [{ reconciliation_id: "recon_1", status: "mismatch", provider: "stripe" }]
    });

    await getAdminPayment("pay/123");
    await listRefunds({ status: "pending_review", page: 1, page_size: 25 });
    await listReconciliationAlerts({ status: "mismatch", page: 1, page_size: 5 });

    expect(getFetchCall(fetchMock, 0).requestUrl.pathname).toBe("/api/v1/admin/payments/pay%2F123");
    expect(getFetchCall(fetchMock, 1).requestUrl.pathname).toBe("/api/v1/admin/refunds");
    expect(getFetchCall(fetchMock, 1).requestUrl.searchParams.get("status")).toBe("pending_review");
    expect(getFetchCall(fetchMock, 2).requestUrl.pathname).toBe("/api/v1/admin/payment-reconciliations");
    expect(getFetchCall(fetchMock, 2).requestUrl.searchParams.get("status")).toBe("mismatch");
  });

  it("submits refund review decisions without provider-direct calls", async () => {
    const fetchMock = prepareApiTest({
      refund_id: "refund/123",
      payment_id: "pay_123",
      status: "approved",
      amount: { amount: 25000, currency: "INR" },
      reason: "Duplicate charge"
    });

    await reviewRefund("refund/123", {
      decision: "approved",
      reason: "Payment duplicate confirmed by support ticket"
    });

    const { requestUrl, requestInit } = getFetchCall(fetchMock);

    expect(requestUrl.pathname).toBe("/api/v1/admin/refunds/refund%2F123/review");
    expect(requestUrl.pathname).not.toContain("/webhooks/");
    expect(requestUrl.pathname).not.toContain("/payments/pay_123/refund");
    expect(requestInit.method).toBe("POST");
    expect(JSON.parse(requestInit.body as string)).toEqual({
      decision: "approved",
      reason: "Payment duplicate confirmed by support ticket"
    });
  });
});
