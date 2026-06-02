import { describe, expect, it } from "vitest";

import { campaignFormSchema, couponFormSchema } from "./offer-validation";

describe("offer validation", () => {
  it("accepts a valid percentage coupon", () => {
    const result = couponFormSchema.safeParse({
      code: "SAVE10",
      discount_type: "percentage",
      discount_value: 10,
      min_cart_amount: {
        amount: 499,
        currency: "INR",
      },
      starts_at: "2026-06-01T09:00",
      ends_at: "2026-06-15T21:00",
      usage_limit: 100,
    });

    expect(result.success).toBe(true);
  });

  it("rejects percentage coupons above 100", () => {
    const result = couponFormSchema.safeParse({
      code: "SAVE101",
      discount_type: "percentage",
      discount_value: 101,
      min_cart_amount: {
        amount: undefined,
        currency: "INR",
      },
    });

    expect(result.success).toBe(false);
  });

  it("rejects coupon end dates before start dates", () => {
    const result = couponFormSchema.safeParse({
      code: "FLASH",
      discount_type: "fixed",
      discount_value: 200,
      min_cart_amount: {
        amount: undefined,
        currency: "INR",
      },
      starts_at: "2026-06-10T09:00",
      ends_at: "2026-06-09T09:00",
    });

    expect(result.success).toBe(false);
  });

  it("rejects campaigns with invalid date windows", () => {
    const result = campaignFormSchema.safeParse({
      name: "Festive Sale",
      starts_at: "2026-06-20T09:00",
      ends_at: "2026-06-01T09:00",
      budget: {
        amount: undefined,
        currency: "INR",
      },
    });

    expect(result.success).toBe(false);
  });
});
