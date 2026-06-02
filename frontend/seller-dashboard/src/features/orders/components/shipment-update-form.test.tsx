import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { useFulfillmentUpdate } from "../hooks/use-fulfillment-update";
import type { Order } from "../types";
import { ShipmentUpdateForm } from "./shipment-update-form";

vi.mock("../hooks/use-fulfillment-update", () => ({
  useFulfillmentUpdate: vi.fn(),
}));

const order: Order = {
  order_id: "order-1",
  status: "paid",
  items: [],
  total: { amount: 5000, currency: "INR" },
  created_at: "2026-05-01T10:00:00Z",
  shipments: [],
};

describe("ShipmentUpdateForm", () => {
  it("submits the selected fulfillment payload", async () => {
    const mutateAsync = vi.fn().mockResolvedValue({ ...order, status: "shipped" });
    const user = userEvent.setup();

    vi.mocked(useFulfillmentUpdate).mockReturnValue({
      mutateAsync,
      isPending: false,
      isError: false,
      isSuccess: false,
      error: null,
    } as unknown as ReturnType<typeof useFulfillmentUpdate>);

    render(<ShipmentUpdateForm order={order} />);

    await user.selectOptions(screen.getByLabelText(/next status/i), "shipped");
    await user.type(screen.getByLabelText(/carrier/i), "Blue Dart");
    await user.type(screen.getByLabelText(/tracking number/i), "TRK-123");
    await user.click(screen.getByRole("button", { name: /update shipment/i }));

    await waitFor(() => {
      expect(mutateAsync).toHaveBeenCalledWith({
        status: "shipped",
        carrier: "Blue Dart",
        tracking_number: "TRK-123",
      });
    });
  });
});
