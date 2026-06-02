import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { SellerOrderFilters } from "../types";
import { OrderFiltersBar } from "./order-filters-bar";

const filters: SellerOrderFilters = {
  status: "shipped",
  q: "ORD-1",
  page: 3,
  page_size: 20,
  date_from: "2026-05-01",
};

describe("OrderFiltersBar", () => {
  it("resets pagination when status changes", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();

    render(<OrderFiltersBar filters={filters} onChange={onChange} />);

    await user.click(screen.getByRole("button", { name: /paid/i }));

    expect(onChange).toHaveBeenCalledWith({
      ...filters,
      status: "paid",
      page: 1,
    });
  });

  it("clears optional filters without changing the page size", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();

    render(<OrderFiltersBar filters={filters} onChange={onChange} />);

    await user.click(screen.getByRole("button", { name: /reset/i }));

    expect(onChange).toHaveBeenCalledWith({
      status: "all",
      page: 1,
      page_size: 20,
    });
  });
});
