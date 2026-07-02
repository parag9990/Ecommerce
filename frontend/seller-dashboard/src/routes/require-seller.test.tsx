import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import { createMemoryRouter, RouterProvider } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { getSellerSession } from "../api/seller-session-api";
import type { SellerSessionResponse } from "../api/seller-session-api";
import { useSellerStore } from "../stores/seller-store";
import { sellerRoutes } from "./seller-routes";

vi.mock("../api/seller-session-api", async () => {
  const actual = await vi.importActual<typeof import("../api/seller-session-api")>(
    "../api/seller-session-api",
  );

  return {
    ...actual,
    getSellerSession: vi.fn(),
  };
});

const mockedGetSellerSession = vi.mocked(getSellerSession);

function renderSellerRoute(initialPath = "/seller") {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });
  const router = createMemoryRouter(sellerRoutes, {
    initialEntries: [initialPath],
  });

  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );

  return router;
}

describe("RequireSeller", () => {
  beforeEach(() => {
    useSellerStore.getState().reset();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("redirects unauthenticated users to the login redirect page", async () => {
    mockedGetSellerSession.mockResolvedValue({
      authenticated: false,
      active_seller: null,
      sellers: [],
    });

    renderSellerRoute();

    expect(await screen.findByRole("heading", { name: "Seller login" })).toBeTruthy();
  });

  it("blocks authenticated users without an active seller", async () => {
    mockedGetSellerSession.mockResolvedValue({
      authenticated: true,
      active_seller: {
        seller_id: "seller-pending",
        display_name: "Pending Store",
        status: "pending",
      },
      sellers: [],
    });

    renderSellerRoute();

    expect(await screen.findByText("Seller approval pending")).toBeTruthy();
  });

  it("renders the dashboard shell for active sellers", async () => {
    const session: SellerSessionResponse = {
      authenticated: true,
      active_seller: {
        seller_id: "seller-1",
        display_name: "North Store",
        status: "active",
      },
      sellers: [
        {
          seller_id: "seller-1",
          display_name: "North Store",
          status: "active",
        },
      ],
    };
    mockedGetSellerSession.mockResolvedValue(session);

    renderSellerRoute();

    expect(await screen.findByRole("heading", { name: "Seller Dashboard" })).toBeTruthy();
    expect(screen.getByText("Overview")).toBeTruthy();
    expect(screen.getByText("Products")).toBeTruthy();
    expect(screen.getAllByText("North Store").length).toBeGreaterThan(0);

    await waitFor(() => {
      expect(useSellerStore.getState().activeSeller?.seller_id).toBe("seller-1");
    });
  });

  it("renders the revenue analytics route for active sellers", async () => {
    mockedGetSellerSession.mockResolvedValue({
      authenticated: true,
      active_seller: {
        seller_id: "seller-1",
        display_name: "North Store",
        status: "active",
      },
      sellers: [
        {
          seller_id: "seller-1",
          display_name: "North Store",
          status: "active",
        },
      ],
    });
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            revenue: { amount: 125000, currency: "INR" },
            orders: 84,
            conversion_rate: 3.4,
            top_products: [
              {
                product_id: "prod-1",
                name: "Cotton Shirt",
                sku: "SHIRT-001",
                revenue: { amount: 52000, currency: "INR" },
                orders: 18,
                units_sold: 23,
                conversion_rate: 4.2,
              },
            ],
          },
        }),
        {
          status: 200,
          headers: { "content-type": "application/json" },
        },
      ),
    );

    renderSellerRoute("/seller/analytics");

    expect(await screen.findByRole("heading", { name: "Revenue analytics" })).toBeTruthy();
    expect(screen.getByText("₹1,250")).toBeTruthy();
    expect(screen.getByText("84")).toBeTruthy();
    expect(screen.getAllByText("3.40%").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Cotton Shirt").length).toBeGreaterThan(0);
  });
});
