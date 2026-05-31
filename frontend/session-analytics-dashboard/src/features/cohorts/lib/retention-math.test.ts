import { describe, expect, it } from "vitest";

import type { RetentionCohort } from "../../../api/session-api";
import {
  getAverageRetentionForOffset,
  hasRetentionData,
  safePercent
} from "./retention-math";

describe("retention math", () => {
  it("returns zero when percentage denominator is empty", () => {
    expect(safePercent(5, 0)).toBe(0);
  });

  it("calculates rounded percentages", () => {
    expect(safePercent(33, 120)).toBe(27.5);
  });

  it("averages visible retention rates for an offset", () => {
    expect(getAverageRetentionForOffset(cohorts, 1)).toBe(34);
  });

  it("detects data from cohort sizes and retained buckets", () => {
    expect(hasRetentionData(cohorts)).toBe(true);
    expect(hasRetentionData([])).toBe(false);
  });
});

const cohorts: RetentionCohort[] = [
  {
    buckets: [
      { label: "W0", offset: 0, rate: 100, suppressed: false, users: 100 },
      { label: "W1", offset: 1, rate: 32, suppressed: false, users: 32 }
    ],
    cohortKey: "2026-W20",
    cohortLabel: "May 11 - May 17",
    cohortSize: 100,
    suppressed: false
  },
  {
    buckets: [
      { label: "W0", offset: 0, rate: 100, suppressed: false, users: 80 },
      { label: "W1", offset: 1, rate: 36, suppressed: false, users: 29 }
    ],
    cohortKey: "2026-W21",
    cohortLabel: "May 18 - May 24",
    cohortSize: 80,
    suppressed: false
  },
  {
    buckets: [
      { label: "W0", offset: 0, rate: 100, suppressed: false, users: 3 },
      { label: "W1", offset: 1, rate: 66, suppressed: true, users: 2 }
    ],
    cohortKey: "2026-W22",
    cohortLabel: "May 25 - May 31",
    cohortSize: 3,
    suppressed: true
  }
];
