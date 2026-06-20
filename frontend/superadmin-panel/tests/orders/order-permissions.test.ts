import { describe, expect, it } from "vitest";

import {
  canReviewOrders,
  canViewOrderDetail,
  canViewOrderDisputes,
  canViewOrders
} from "../../src/features/orders/permissions";

describe("order operations permissions", () => {
  it("allows documented order roles to view orders and disputes", () => {
    expect(canViewOrders(["superadmin"])).toBe(true);
    expect(canViewOrders(["operations_admin"])).toBe(true);
    expect(canViewOrders(["readonly_admin"])).toBe(true);
    expect(canViewOrderDetail(["operations_admin"])).toBe(true);
    expect(canViewOrderDisputes(["readonly_admin"])).toBe(true);
  });

  it("limits manual review to superadmin and operations admins", () => {
    expect(canReviewOrders(["superadmin"])).toBe(true);
    expect(canReviewOrders(["operations_admin"])).toBe(true);
    expect(canReviewOrders(["readonly_admin"])).toBe(false);
  });

  it("keeps finance and catalog admins out of Task 4 order operations", () => {
    expect(canViewOrders(["finance_admin"])).toBe(false);
    expect(canViewOrders(["catalog_admin"])).toBe(false);
    expect(canReviewOrders(["finance_admin"])).toBe(false);
    expect(canReviewOrders(["catalog_admin"])).toBe(false);
  });
});
