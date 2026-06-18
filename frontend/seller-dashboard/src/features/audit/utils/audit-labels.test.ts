import { describe, expect, it } from "vitest";

import { getActionLabel, getResourceLabel } from "./audit-labels";

describe("audit-labels", () => {
  it("returns known action and resource labels", () => {
    expect(getActionLabel("product.updated")).toBe("Updated product");
    expect(getResourceLabel("coupon")).toBe("Coupon");
  });

  it("falls back to readable labels for unknown values", () => {
    expect(getActionLabel("custom.workflow_ran")).toBe("Custom Workflow Ran");
    expect(getResourceLabel("risk_review")).toBe("Risk Review");
  });
});
