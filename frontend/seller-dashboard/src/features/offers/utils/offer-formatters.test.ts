import { describe, expect, it } from "vitest";

import { formatDiscount, formatWindow } from "./offer-formatters";

describe("offer formatters", () => {
  it("formats percentage and fixed discounts", () => {
    expect(
      formatDiscount({
        discount_type: "percentage",
        discount_value: 15,
      }),
    ).toBe("15% off");
    expect(
      formatDiscount({
        discount_type: "fixed",
        discount_value: 250,
        min_cart_amount: { amount: 999, currency: "INR" },
      }),
    ).toContain("250");
  });

  it("formats open and bounded windows", () => {
    expect(formatWindow()).toBe("Always available");
    expect(formatWindow("2026-06-01T00:00:00.000Z", "2026-06-30T00:00:00.000Z")).toContain("-");
  });
});
