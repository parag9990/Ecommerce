import { describe, expect, it } from "vitest";

import {
  canViewLiveMetrics,
  canViewSessionJourney,
  canViewSessions,
  canViewUnmaskedSessionPii
} from "../../src/features/sessions/permissions";

describe("session permissions", () => {
  it("allows superadmin, operations admin, and readonly admin", () => {
    expect(canViewSessions(["superadmin"])).toBe(true);
    expect(canViewLiveMetrics(["operations_admin"])).toBe(true);
    expect(canViewSessionJourney(["readonly_admin"])).toBe(true);
  });

  it("denies finance and catalog admins", () => {
    expect(canViewSessions(["finance_admin"])).toBe(false);
    expect(canViewSessions(["catalog_admin"])).toBe(false);
  });

  it("keeps unmasked PII disabled for this read-only module", () => {
    expect(canViewUnmaskedSessionPii(["superadmin"])).toBe(false);
  });
});
