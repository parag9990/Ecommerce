import { describe, expect, it } from "vitest";

import {
  formatDuration,
  formatNumber,
  formatPercent,
  formatRelativeTime
} from "./format";

describe("format helpers", () => {
  it("formats large numbers consistently", () => {
    expect(formatNumber(1280)).toBe("1,280");
  });

  it("formats percentages with one decimal place", () => {
    expect(formatPercent(3.84)).toBe("3.8%");
  });

  it("formats durations in minutes and seconds", () => {
    expect(formatDuration(125)).toBe("2m 5s");
  });

  it("formats relative timestamps", () => {
    expect(
      formatRelativeTime(
        "2026-05-28T09:55:00.000Z",
        new Date("2026-05-28T10:00:00.000Z")
      )
    ).toBe("5m ago");
  });
});
