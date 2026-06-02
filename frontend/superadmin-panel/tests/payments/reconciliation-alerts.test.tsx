import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ReconciliationTable } from "../../src/features/payments/components/reconciliation-table";
import type { ReconciliationAlert } from "../../src/features/payments/types";

const alerts: ReconciliationAlert[] = [
  {
    reconciliation_id: "recon_123",
    payment_id: "pay_123",
    provider: "stripe",
    status: "mismatch",
    local_amount: { amount: 100000, currency: "INR" },
    provider_amount: { amount: 95000, currency: "INR" },
    settlement_id: "set_123",
    detected_at: "2026-06-01T10:00:00Z"
  }
];

describe("ReconciliationTable", () => {
  it("renders mismatch alerts and emits inspect selection", () => {
    const onSelect = vi.fn();

    render(
      <ReconciliationTable
        alerts={alerts}
        isLoading={false}
        isFetching={false}
        error={null}
        page={1}
        limit={5}
        compact
        onPageChange={vi.fn()}
        onRetry={vi.fn()}
        onSelect={onSelect}
      />
    );

    expect(screen.getByText("recon_123")).toBeTruthy();
    expect(screen.getByText(/mismatch/i)).toBeTruthy();
    expect(screen.getByText(/medium risk/i)).toBeTruthy();

    fireEvent.click(screen.getByRole("button", { name: /inspect/i }));

    expect(onSelect).toHaveBeenCalledWith(alerts[0]);
  });
});
