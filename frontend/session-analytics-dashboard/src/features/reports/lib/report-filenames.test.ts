import { describe, expect, it, vi } from "vitest";

import { buildCsvFilename, sanitizeCsvFilename } from "./report-filenames";

describe("report filenames", () => {
  it("builds safe csv filenames with report type and date range", () => {
    vi.useFakeTimers();
    try {
      vi.setSystemTime(new Date("2026-05-31T10:15:30.000Z"));

      expect(
        buildCsvFilename({
          from: "2026-05-01",
          reportType: "funnel",
          to: "2026-05-28"
        })
      ).toBe("funnel_2026-05-01_2026-05-28_2026-05-31-10-15-30.csv");
    } finally {
      vi.useRealTimers();
    }
  });

  it("sanitizes server-provided filenames", () => {
    expect(sanitizeCsvFilename("bad:name?.txt")).toBe("bad_name_.txt.csv");
    expect(sanitizeCsvFilename(" funnel.csv ")).toBe("funnel.csv");
  });
});
