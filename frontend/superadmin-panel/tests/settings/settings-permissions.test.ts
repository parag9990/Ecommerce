import { describe, expect, it } from "vitest";

import {
  canManageSearchSynonyms,
  canViewPlatformSettings,
  canWritePlatformSettings
} from "../../src/features/settings/permissions";

describe("settings permissions", () => {
  it("allows superadmin and readonly admins to view platform settings", () => {
    expect(canViewPlatformSettings(["superadmin"])).toBe(true);
    expect(canViewPlatformSettings(["readonly_admin"])).toBe(true);
  });

  it("allows only superadmin to write platform settings", () => {
    expect(canWritePlatformSettings(["superadmin"])).toBe(true);
    expect(canWritePlatformSettings(["catalog_admin"])).toBe(false);
    expect(canWritePlatformSettings(["finance_admin"])).toBe(false);
    expect(canWritePlatformSettings(["operations_admin"])).toBe(false);
    expect(canWritePlatformSettings(["readonly_admin"])).toBe(false);
  });

  it("allows superadmin and catalog admins to manage search synonyms", () => {
    expect(canManageSearchSynonyms(["superadmin"])).toBe(true);
    expect(canManageSearchSynonyms(["catalog_admin"])).toBe(true);
    expect(canManageSearchSynonyms(["operations_admin"])).toBe(false);
    expect(canManageSearchSynonyms(["finance_admin"])).toBe(false);
  });
});
