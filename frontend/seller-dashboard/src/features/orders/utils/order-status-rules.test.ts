import { describe, expect, it } from "vitest";

import { canUpdateFulfillment, nextFulfillmentStatuses } from "./order-status-rules";

describe("order-status-rules", () => {
  it("allows fulfillment updates only for seller-editable statuses", () => {
    expect(canUpdateFulfillment("paid")).toBe(true);
    expect(canUpdateFulfillment("packed")).toBe(true);
    expect(canUpdateFulfillment("shipped")).toBe(true);
    expect(canUpdateFulfillment("created")).toBe(false);
    expect(canUpdateFulfillment("refunded")).toBe(false);
  });

  it("returns the next seller fulfillment statuses", () => {
    expect(nextFulfillmentStatuses("paid")).toEqual(["packed", "shipped"]);
    expect(nextFulfillmentStatuses("packed")).toEqual(["shipped"]);
    expect(nextFulfillmentStatuses("shipped")).toEqual(["delivered"]);
    expect(nextFulfillmentStatuses("delivered")).toEqual([]);
  });
});
