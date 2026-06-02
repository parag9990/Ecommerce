import { describe, expect, it } from "vitest";

import {
  hasSellerPermission,
  isAssignableSellerStaffRole,
  isSellerStaffRole,
} from "./seller-permissions";

describe("seller permissions", () => {
  it("allows owner and manager team controls", () => {
    expect(hasSellerPermission(["seller"], "team:invite")).toBe(true);
    expect(hasSellerPermission(["seller_manager"], "team:disable")).toBe(true);
  });

  it("keeps catalog editors out of team and order controls", () => {
    expect(hasSellerPermission(["seller_catalog_editor"], "team:invite")).toBe(false);
    expect(hasSellerPermission(["seller_catalog_editor"], "orders:update_fulfillment")).toBe(false);
    expect(hasSellerPermission(["seller_catalog_editor"], "products:write")).toBe(true);
  });

  it("allows order managers to update fulfillment and view analytics", () => {
    expect(hasSellerPermission(["seller_order_manager"], "orders:update_fulfillment")).toBe(true);
    expect(hasSellerPermission(["seller_order_manager"], "analytics:view")).toBe(true);
    expect(hasSellerPermission(["seller_order_manager"], "offers:write")).toBe(false);
  });

  it("validates known and assignable seller staff roles", () => {
    expect(isSellerStaffRole("seller")).toBe(true);
    expect(isAssignableSellerStaffRole("seller")).toBe(false);
    expect(isAssignableSellerStaffRole("seller_manager")).toBe(true);
    expect(isSellerStaffRole("admin")).toBe(false);
  });
});
