import { describe, expect, it } from "vitest";

import { formatMoney, formatNumber, formatPercentage } from "./analytics-formatters";

describe("analytics formatters", () => {
  it("formats money from minor units", () => {
    expect(formatMoney({ amount: 125000, currency: "INR" })).toBe("₹1,250");
  });

  it("keeps zero distinct from unavailable metrics", () => {
    expect(formatNumber(0)).toBe("0");
    expect(formatPercentage(0)).toBe("0.00%");
    expect(formatNumber(null)).toBe("Not available");
  });
});
