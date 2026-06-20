import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { OrderFilterBar } from "../../src/features/orders/components/order-filter-bar";
import type { AdminOrderFilters } from "../../src/features/orders/types";

const filters: AdminOrderFilters = {
  q: "",
  status: "all",
  review_status: "all",
  user_id: "",
  seller_id: "",
  from: "",
  to: "",
  page: 1,
  limit: 25
};

describe("OrderFilterBar", () => {
  it("emits order search, status, id, and date filter patches", () => {
    const onChange = vi.fn();
    const onReset = vi.fn();

    render(<OrderFilterBar filters={filters} onChange={onChange} onReset={onReset} />);

    fireEvent.change(screen.getByPlaceholderText(/order id or support token/i), {
      target: { value: "order_123" }
    });
    fireEvent.change(screen.getByLabelText(/order status/i), {
      target: { value: "paid" }
    });
    fireEvent.change(screen.getByLabelText(/review status/i), {
      target: { value: "manual_review" }
    });
    fireEvent.change(screen.getByPlaceholderText(/buyer user id/i), {
      target: { value: "user_123" }
    });
    fireEvent.change(screen.getByPlaceholderText(/seller id/i), {
      target: { value: "seller_123" }
    });
    fireEvent.change(screen.getByLabelText(/created from date/i), {
      target: { value: "2026-06-01" }
    });
    fireEvent.click(screen.getByRole("button", { name: /reset/i }));

    expect(onChange).toHaveBeenCalledWith({ q: "order_123" });
    expect(onChange).toHaveBeenCalledWith({ status: "paid" });
    expect(onChange).toHaveBeenCalledWith({ review_status: "manual_review" });
    expect(onChange).toHaveBeenCalledWith({ user_id: "user_123" });
    expect(onChange).toHaveBeenCalledWith({ seller_id: "seller_123" });
    expect(onChange).toHaveBeenCalledWith({ from: "2026-06-01" });
    expect(onReset).toHaveBeenCalledTimes(1);
  });
});
