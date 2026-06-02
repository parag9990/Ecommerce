import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactElement } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { ManualReviewDrawer } from "../../src/features/orders/components/manual-review-drawer";
import { useAuthStore } from "../../src/stores/auth-store";

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
    new Response(JSON.stringify({ success: true }), {
      status: 200,
      headers: { "Content-Type": "application/json" }
    })
  );
  vi.stubGlobal("fetch", fetchMock);

  return fetchMock;
}

afterEach(() => {
  vi.unstubAllEnvs();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe("ManualReviewDrawer", () => {
  it("hides manual review actions from readonly admins", () => {
    setAdminSession(["readonly_admin"]);

    renderWithQueryClient(
      <ManualReviewDrawer orderId="order_123" reviewStatus="manual_review" roles={["readonly_admin"]} />
    );

    expect(screen.queryByRole("button", { name: /mark reviewing/i })).toBeNull();
    expect(screen.queryByRole("button", { name: /resolve review/i })).toBeNull();
    expect(screen.getByText(/cannot submit manual review decisions/i)).toBeTruthy();
  });

  it("requires a review reason and submits internal admin notes", async () => {
    const fetchMock = prepareMutationTest();
    setAdminSession(["superadmin"]);
    const user = userEvent.setup();

    renderWithQueryClient(
      <ManualReviewDrawer orderId="order/123" reviewStatus="manual_review" roles={["superadmin"]} />
    );

    await user.click(screen.getByRole("button", { name: /^Resolve review$/i }));

    const dialog = screen.getByRole("dialog");
    const confirmButton = within(dialog).getByRole("button", { name: /^Resolve review$/i });

    expect((confirmButton as HTMLButtonElement).disabled).toBe(true);

    await user.type(within(dialog).getByLabelText(/review reason/i), "short");
    expect((confirmButton as HTMLButtonElement).disabled).toBe(true);

    await user.clear(within(dialog).getByLabelText(/review reason/i));
    await user.type(within(dialog).getByLabelText(/review reason/i), "Shipment evidence matches seller timeline");
    await user.type(
      within(dialog).getByLabelText(/internal admin note/i),
      "Support ticket SUPPORT-123 has customer confirmation."
    );

    expect((confirmButton as HTMLButtonElement).disabled).toBe(false);

    await user.click(confirmButton);

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));

    const [input, requestInit = {}] = fetchMock.mock.calls[0] ?? [];
    const requestUrl = new URL(String(input));

    expect(requestUrl.pathname).toBe("/api/v1/admin/orders/order%2F123/review");
    expect(JSON.parse(requestInit.body as string)).toEqual({
      decision: "resolve",
      reason: "Shipment evidence matches seller timeline",
      internal_note: "Support ticket SUPPORT-123 has customer confirmation."
    });
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
    expect(screen.getByText(/resolve review submitted/i)).toBeTruthy();
  });
});
