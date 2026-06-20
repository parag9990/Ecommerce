import { describe, expect, it } from "vitest";

import {
  getMaintenanceModeStatus,
  normalizeSearchSynonymInput,
  validateCommissionCategoryOverrides,
  validateCommissionDefaultRate,
  validateFeatureFlags,
  validateMaintenanceMode,
  validateReason
} from "../../src/features/settings/validators";

describe("settings validators", () => {
  it("rejects commission rates outside the platform range", () => {
    expect(
      validateCommissionDefaultRate({ rate_percent: -1, applies_to: "all_sellers" }).valid
    ).toBe(false);
    expect(
      validateCommissionDefaultRate({ rate_percent: 51, applies_to: "all_sellers" }).valid
    ).toBe(false);
    expect(
      validateCommissionDefaultRate({ rate_percent: 12.5, applies_to: "all_sellers" }).valid
    ).toBe(true);
  });

  it("rejects duplicate category overrides", () => {
    const validation = validateCommissionCategoryOverrides({
      overrides: [
        { category_id: "cat_mobile", rate_percent: 8 },
        { category_id: "cat_mobile", rate_percent: 9 }
      ]
    });

    expect(validation.valid).toBe(false);
    expect(validation.errors.join(" ")).toContain("Duplicate category override");
  });

  it("requires lowercase snake_case feature flag keys", () => {
    expect(validateFeatureFlags({ flags: { new_checkout: true } }).valid).toBe(true);
    expect(validateFeatureFlags({ flags: { "New-Checkout": true } }).valid).toBe(false);
  });

  it("requires reason and safe maintenance windows", () => {
    expect(validateReason("too short").valid).toBe(false);
    expect(
      validateMaintenanceMode({
        enabled: true,
        message: "",
        starts_at: "2026-06-03T02:00:00Z",
        ends_at: "2026-06-03T01:00:00Z",
        allow_admin_bypass: true
      }).valid
    ).toBe(false);
  });

  it("calculates maintenance status from enabled state and window", () => {
    const now = new Date("2026-06-03T01:30:00Z");

    expect(
      getMaintenanceModeStatus(
        {
          enabled: true,
          message: "Scheduled maintenance",
          starts_at: "2026-06-03T02:00:00Z",
          ends_at: "2026-06-03T03:00:00Z",
          allow_admin_bypass: true
        },
        now
      )
    ).toBe("scheduled");
    expect(
      getMaintenanceModeStatus(
        {
          enabled: true,
          message: "Scheduled maintenance",
          starts_at: "2026-06-03T01:00:00Z",
          ends_at: "2026-06-03T02:00:00Z",
          allow_admin_bypass: true
        },
        now
      )
    ).toBe("active");
  });

  it("normalizes search synonym input", () => {
    expect(
      normalizeSearchSynonymInput({
        root: " Mobile ",
        synonyms: [" Phone ", "Smartphone", "phone", "mobile"]
      })
    ).toEqual({
      root: "mobile",
      synonyms: ["phone", "smartphone"]
    });
  });
});
