import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { PaymentFilterBar } from "../../src/features/payments/components/payment-filter-bar";
import type { PaymentFilters } from "../../src/features/payments/types";

const filters: PaymentFilters = {
  order_id: "",
  status: "all",
  provider: "all",
  page: 1,
  page_size: 25
};

describe("PaymentFilterBar", () => {
  it("emits order id, status, provider, and reset patches", () => {
    const onChange = vi.fn();
    const onReset = vi.fn();

    render(<PaymentFilterBar filters={filters} onChange={onChange} onReset={onReset} />);

    fireEvent.change(screen.getByPlaceholderText(/order id/i), {
      target: { value: "order_123" }
    });
    fireEvent.change(screen.getByLabelText(/payment status/i), {
      target: { value: "captured" }
    });
    fireEvent.change(screen.getByLabelText(/payment provider/i), {
      target: { value: "razorpay" }
    });
    fireEvent.click(screen.getByRole("button", { name: /reset/i }));

    expect(onChange).toHaveBeenCalledWith({ order_id: "order_123" });
    expect(onChange).toHaveBeenCalledWith({ status: "captured" });
    expect(onChange).toHaveBeenCalledWith({ provider: "razorpay" });
    expect(onReset).toHaveBeenCalledTimes(1);
  });
});
