import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import type { ReactElement } from "react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { SellerListPage } from "../../src/features/sellers/pages/seller-list-page";
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

function renderWithProviders(ui: ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false }
    }
  });

  render(
    <MemoryRouter>
      <QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>
    </MemoryRouter>
  );
}

function prepareSellerListTest() {
  vi.stubEnv("VITE_API_BASE_URL", "https://api.example.test");

  const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) =>
    new Response(
      JSON.stringify({
        sellers: [
          {
            seller_id: "seller_123",
            user_id: "user_123",
            store_name: "Acme Store",
            status: "pending_review",
            kyc_status: "pending",
            document_count: 2,
            pending_document_count: 1,
            product_count: 5,
            pending_catalog_count: 2
          }
        ],
        total: 1,
        page: 1,
        page_size: 20
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

describe("SellerListPage", () => {
  it("renders the sellers route for superadmins without an auth selector render loop", async () => {
    const fetchMock = prepareSellerListTest();
    setAdminSession(["superadmin"]);

    renderWithProviders(<SellerListPage />);

    expect(await screen.findByRole("heading", { name: "Sellers" })).toBeTruthy();
    expect(await screen.findByRole("link", { name: "Acme Store" })).toBeTruthy();

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));

    const requestUrl = new URL(String(fetchMock.mock.calls[0]?.[0]));

    expect(requestUrl.pathname).toBe("/api/v1/admin/sellers");
    expect(requestUrl.searchParams.get("status")).toBe("pending_review");
  });
});
