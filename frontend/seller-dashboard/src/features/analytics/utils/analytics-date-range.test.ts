import { describe, expect, it } from "vitest";

import {
  createCustomRange,
  createPresetRange,
  formatDateRangeLabel,
  toDateInputValue,
} from "./analytics-date-range";

describe("analytics date ranges", () => {
  it("builds inclusive UTC preset ranges", () => {
    const range = createPresetRange("30d", new Date("2026-06-01T12:00:00.000Z"));

    expect(range).toEqual({
      preset: "30d",
      from: "2026-05-03T00:00:00.000Z",
      to: "2026-06-01T23:59:59.999Z",
    });
  });

  it("normalizes custom ranges into chronological order", () => {
    const range = createCustomRange("2026-06-10", "2026-06-01");

    expect(range.from).toBe("2026-06-01T00:00:00.000Z");
    expect(range.to).toBe("2026-06-10T23:59:59.999Z");
    expect(range.preset).toBe("custom");
  });

  it("formats range labels from ISO values", () => {
    const range = createCustomRange("2026-05-01", "2026-06-01");

    expect(toDateInputValue(range.from)).toBe("2026-05-01");
    expect(formatDateRangeLabel(range)).toBe("2026-05-01 to 2026-06-01");
  });
});
