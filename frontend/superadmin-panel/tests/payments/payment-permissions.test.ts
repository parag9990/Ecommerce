import { describe, expect, it } from "vitest";

import {
  canExportPayments,
  canReviewRefunds,
  canViewPayments,
  canViewReconciliation,
  canViewRefunds
} from "../../src/features/payments/permissions";

describe("payment permissions", () => {
  it("allows superadmin, finance admin, and readonly admin to view payment operations", () => {
    expect(canViewPayments(["superadmin"])).toBe(true);
    expect(canViewPayments(["finance_admin"])).toBe(true);
    expect(canViewPayments(["readonly_admin"])).toBe(true);
    expect(canViewRefunds(["readonly_admin"])).toBe(true);
    expect(canViewReconciliation(["readonly_admin"])).toBe(true);
  });

  it("blocks non-finance operational roles from payment operations", () => {
    expect(canViewPayments(["operations_admin"])).toBe(false);
    expect(canViewPayments(["catalog_admin"])).toBe(false);
    expect(canReviewRefunds(["operations_admin"])).toBe(false);
  });

  it("keeps refund review and export actions finance-only", () => {
    expect(canReviewRefunds(["superadmin"])).toBe(true);
    expect(canReviewRefunds(["finance_admin"])).toBe(true);
    expect(canReviewRefunds(["readonly_admin"])).toBe(false);
    expect(canExportPayments(["finance_admin"])).toBe(true);
    expect(canExportPayments(["readonly_admin"])).toBe(false);
  });
});
