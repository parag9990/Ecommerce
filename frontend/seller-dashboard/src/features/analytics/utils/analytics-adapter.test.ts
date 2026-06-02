import { describe, expect, it } from "vitest";

import { normalizeSellerAnalytics } from "./analytics-adapter";

describe("normalizeSellerAnalytics", () => {
  it("normalizes current CMS analytics response shape", () => {
    const analytics = normalizeSellerAnalytics({
      revenue: { amount: 125000, currency: "INR" },
      orders: 84,
      conversion_rate: 3.4,
      top_products: [
        {
          product_id: "prod-1",
          product_name: "Cotton Shirt",
          sku: "SHIRT-001",
          revenue: { amount: 52000, currency: "INR" },
          order_count: 18,
          quantity_sold: 23,
        },
        { name: "Broken row" },
      ],
    });

    expect(analytics.revenue).toEqual({ amount: 125000, currency: "INR" });
    expect(analytics.gmv).toBeNull();
    expect(analytics.orders).toBe(84);
    expect(analytics.conversionRate).toBe(3.4);
    expect(analytics.topProducts).toEqual([
      {
        product_id: "prod-1",
        name: "Cotton Shirt",
        sku: "SHIRT-001",
        revenue: { amount: 52000, currency: "INR" },
        orders: 18,
        units_sold: 23,
      },
    ]);
  });

  it("keeps optional trend fields only when the backend sends them", () => {
    const analytics = normalizeSellerAnalytics({
      series: [
        {
          date: "2026-06-01",
          revenue: { amount: 10000, currency: "INR" },
          orders: 2,
          conversion_rate: 1.5,
        },
        { orders: 3 },
      ],
    });

    expect(analytics.series).toEqual([
      {
        date: "2026-06-01",
        revenue: { amount: 10000, currency: "INR" },
        orders: 2,
        conversion_rate: 1.5,
      },
    ]);
  });
});
