import { describe, expect, it } from "vitest";

import {
  canReviewSellerCatalog,
  canReviewSellerKyc,
  canSuspendSeller,
  canViewMaskedSellerKyc,
  canViewSellerCatalog,
  canViewSellers
} from "../../src/features/sellers/permissions";

describe("seller management permissions", () => {
  it("allows documented roles to view seller management", () => {
    expect(canViewSellers(["superadmin"])).toBe(true);
    expect(canViewSellers(["operations_admin"])).toBe(true);
    expect(canViewSellers(["catalog_admin"])).toBe(true);
    expect(canViewSellers(["readonly_admin"])).toBe(true);
  });

  it("keeps finance admins out of seller management", () => {
    expect(canViewSellers(["finance_admin"])).toBe(false);
  });

  it("limits KYC lifecycle actions to superadmin and operations admins", () => {
    expect(canReviewSellerKyc(["superadmin"])).toBe(true);
    expect(canReviewSellerKyc(["operations_admin"])).toBe(true);
    expect(canReviewSellerKyc(["catalog_admin"])).toBe(false);
    expect(canReviewSellerKyc(["readonly_admin"])).toBe(false);
  });

  it("limits suspension actions to superadmin and operations admins", () => {
    expect(canSuspendSeller(["superadmin"])).toBe(true);
    expect(canSuspendSeller(["operations_admin"])).toBe(true);
    expect(canSuspendSeller(["catalog_admin"])).toBe(false);
    expect(canSuspendSeller(["readonly_admin"])).toBe(false);
  });

  it("allows catalog review without granting KYC document access", () => {
    expect(canReviewSellerCatalog(["catalog_admin"])).toBe(true);
    expect(canViewSellerCatalog(["readonly_admin"])).toBe(true);
    expect(canViewMaskedSellerKyc(["readonly_admin"])).toBe(true);
  });
});
