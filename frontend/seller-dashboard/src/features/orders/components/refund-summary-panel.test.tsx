import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { RefundSummaryPanel } from "./refund-summary-panel";

describe("RefundSummaryPanel", () => {
  it("renders refund details without seller action buttons", () => {
    render(
      <RefundSummaryPanel
        refunds={[
          {
            refund_id: "refund-1",
            payment_id: "payment-1",
            status: "requested",
            amount: { amount: 2500, currency: "INR" },
            reason: "Damaged item",
          },
        ]}
      />,
    );

    expect(screen.getByText("Refunds")).toBeTruthy();
    expect(screen.getByText("Damaged item")).toBeTruthy();
    expect(screen.queryByRole("button")).toBeNull();
  });
});
