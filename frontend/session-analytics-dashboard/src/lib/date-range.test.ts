import { describe, expect, it } from "vitest";

import {
  getDefaultDateRange,
  getDateRangeForPreset,
  validateDateRange
} from "./date-range";

describe("date range helpers", () => {
  const now = new Date("2026-05-28T12:00:00.000Z");

  it("defaults to the last seven calendar days", () => {
    expect(getDefaultDateRange(now)).toEqual({
      preset: "7d",
      from: "2026-05-22",
      to: "2026-05-28"
    });
  });

  it("builds the 30 day preset", () => {
    expect(getDateRangeForPreset("30d", now)).toEqual({
      preset: "30d",
      from: "2026-04-29",
      to: "2026-05-28"
    });
  });

  it("rejects inverted ranges", () => {
    expect(
      validateDateRange({
        preset: "custom",
        from: "2026-05-29",
        to: "2026-05-28"
      })
    ).toEqual({
      ok: false,
      message: "Start date must be before or equal to end date."
    });
  });
});
