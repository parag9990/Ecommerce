import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { getSellerSession } from "../api/seller-session-api";
import type { SellerSessionResponse } from "../api/seller-session-api";
import { useSellerStore } from "../stores/seller-store";
import { SellerSwitcher } from "./seller-switcher";

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

function renderSwitcher() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  render(
    <QueryClientProvider client={queryClient}>
      <SellerSwitcher />
    </QueryClientProvider>,
  );
}

describe("SellerSwitcher", () => {
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
      {
        seller_id: "seller-2",
        display_name: "South Store",
        status: "active",
      },
      {
        seller_id: "seller-3",
        display_name: "Paused Store",
        status: "suspended",
      },
    ],
  };

  beforeEach(() => {
    useSellerStore.getState().reset();
    useSellerStore.getState().setActiveSeller(session.active_seller);
    mockedGetSellerSession.mockResolvedValue(session);
  });

  it("updates active seller state when an active seller is selected", async () => {
    const user = userEvent.setup();

    renderSwitcher();

    const select = await screen.findByLabelText("Seller");
    await screen.findByText("South Store");
    await user.selectOptions(select, "seller-2");

    expect(useSellerStore.getState().activeSeller?.seller_id).toBe("seller-2");
  });
});
