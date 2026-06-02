import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactElement } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { RefundReviewDrawer } from "../../src/features/payments/components/refund-review-drawer";
import type { Refund } from "../../src/features/payments/types";
import { useAuthStore } from "../../src/stores/auth-store";

const mockRefund: Refund = {
  refund_id: "refund_123",
  payment_id: "pay_123",
  order_id: "order_123",
  status: "pending_review",
  amount: { amount: 25000, currency: "INR" },
  reason: "Buyer was charged twice",
  requested_by: "support",
  created_at: "2026-06-01T10:00:00Z"
};

function setAdminSession(roles: string[]) {
  useAuthStore.getState().setSession({
    accessToken: "admin-token",
    user: {
      id: "admin-1",
      email: "admin@example.com",
      name: "Admin User",
      roles
    }
  });
}

function renderWithQueryClient(ui: ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false }
    }
  });

  render(<QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>);
}

function prepareMutationTest() {
  vi.stubEnv("VITE_API_BASE_URL", "https://api.example.test");
  const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) =>
    new Response(
      JSON.stringify({
        ...mockRefund,
        status: "approved"
      }),
      {
        status: 200,
        headers: { "Content-Type": "application/json" }
      }
    )
  );
  vi.stubGlobal("fetch", fetchMock);

  return fetchMock;
}

afterEach(() => {
  vi.unstubAllEnvs();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe("RefundReviewDrawer", () => {
  it("hides approve and reject actions from readonly admins", () => {
    setAdminSession(["readonly_admin"]);

    renderWithQueryClient(
      <RefundReviewDrawer refund={mockRefund} canReview={false} onClose={vi.fn()} />
    );

    expect(screen.queryByRole("button", { name: /approve refund/i })).toBeNull();
    expect(screen.queryByRole("button", { name: /reject refund/i })).toBeNull();
    expect(screen.getByText(/cannot approve or reject/i)).toBeTruthy();
  });

  it("requires a finance review reason before approving a refund", async () => {
    const fetchMock = prepareMutationTest();
    setAdminSession(["finance_admin"]);
    const user = userEvent.setup();

    renderWithQueryClient(
      <RefundReviewDrawer refund={mockRefund} canReview onClose={vi.fn()} />
    );

    await user.click(screen.getByRole("button", { name: /^Approve refund$/i }));

    const dialog = screen.getByRole("dialog");
    const confirmButton = within(dialog).getByRole("button", { name: /^Approve refund$/i });

    expect((confirmButton as HTMLButtonElement).disabled).toBe(true);

    await user.type(within(dialog).getByLabelText(/finance review reason/i), "short");
    expect((confirmButton as HTMLButtonElement).disabled).toBe(true);

    await user.clear(within(dialog).getByLabelText(/finance review reason/i));
    await user.type(
      within(dialog).getByLabelText(/finance review reason/i),
      "Duplicate payment confirmed by provider report"
    );

    expect((confirmButton as HTMLButtonElement).disabled).toBe(false);

    await user.click(confirmButton);

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));

    const [input, requestInit = {}] = fetchMock.mock.calls[0] ?? [];
    const requestUrl = new URL(String(input));

    expect(requestUrl.pathname).toBe("/api/v1/admin/refunds/refund_123/review");
    expect(JSON.parse(requestInit.body as string)).toEqual({
      decision: "approved",
      reason: "Duplicate payment confirmed by provider report"
    });
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
    expect(screen.getByText(/approve refund submitted/i)).toBeTruthy();
  });
});
