import { describe, expect, it } from "vitest";

import { formatDateTime, formatMoney, formatOrderStatus } from "./order-formatters";

describe("order-formatters", () => {
  it("formats minor-unit money values", () => {
    expect(formatMoney({ amount: 123456, currency: "INR" })).toContain("1,234.56");
  });

  it("returns a placeholder for invalid dates", () => {
    expect(formatDateTime("not-a-date")).toBe("-");
  });

  it("formats snake-case statuses", () => {
    expect(formatOrderStatus("pending_payment")).toBe("pending payment");
  });
});
