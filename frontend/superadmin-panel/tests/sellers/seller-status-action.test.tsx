import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactElement } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { SellerActionBar } from "../../src/features/sellers/components/seller-action-bar";
import type { AdminSeller } from "../../src/features/sellers/types";
import { useAuthStore } from "../../src/stores/auth-store";

const pendingSeller: AdminSeller = {
  seller_id: "seller_123",
  user_id: "user_123",
  store_name: "Acme Store",
  status: "pending_review",
  gst_number: "27ABCDE1234F1Z5"
};

const activeSeller: AdminSeller = {
  ...pendingSeller,
  status: "active"
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

describe("SellerActionBar status actions", () => {
  it("hides seller lifecycle actions from catalog admins", () => {
    setAdminSession(["catalog_admin"]);

    renderWithQueryClient(<SellerActionBar seller={pendingSeller} roles={["catalog_admin"]} />);

    expect(screen.queryByRole("button", { name: /approve seller/i })).toBeNull();
    expect(screen.queryByRole("button", { name: /reject seller/i })).toBeNull();
  });

  it("requires a reason before approving a pending seller", async () => {
    const fetchMock = prepareMutationTest();
    setAdminSession(["superadmin"]);
    const user = userEvent.setup();

    renderWithQueryClient(<SellerActionBar seller={pendingSeller} roles={["superadmin"]} />);

    await user.click(screen.getByRole("button", { name: /approve seller/i }));

    const confirmButton = screen.getByRole("button", { name: "Approve" });
    expect((confirmButton as HTMLButtonElement).disabled).toBe(true);

    await user.type(screen.getByLabelText(/approval reason/i), "short");
    expect((confirmButton as HTMLButtonElement).disabled).toBe(true);

    await user.clear(screen.getByLabelText(/approval reason/i));
    await user.type(screen.getByLabelText(/approval reason/i), "GST and business documents verified");
    expect((confirmButton as HTMLButtonElement).disabled).toBe(false);

    await user.click(confirmButton);

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));

    const [input, requestInit = {}] = fetchMock.mock.calls[0] ?? [];
    const requestUrl = new URL(String(input));

    expect(requestUrl.pathname).toBe("/api/v1/admin/sellers/seller_123/status");
    expect(JSON.parse(requestInit.body as string)).toEqual({
      status: "active",
      reason: "kyc_verified: GST and business documents verified"
    });
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
  });

  it("submits a suspension reason code for active sellers", async () => {
    const fetchMock = prepareMutationTest();
    setAdminSession(["operations_admin"]);
    const user = userEvent.setup();

    renderWithQueryClient(<SellerActionBar seller={activeSeller} roles={["operations_admin"]} />);

    await user.click(screen.getByRole("button", { name: /suspend seller/i }));
    await user.selectOptions(screen.getByLabelText(/reason code/i), "counterfeit_risk");
    await user.type(screen.getByLabelText(/suspension reason/i), "Repeated counterfeit product reports");
    await user.click(screen.getByRole("button", { name: "Suspend" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));

    const [, requestInit = {}] = fetchMock.mock.calls[0] ?? [];

    expect(JSON.parse(requestInit.body as string)).toEqual({
      status: "suspended",
      reason: "counterfeit_risk: Repeated counterfeit product reports"
    });
  });
});
