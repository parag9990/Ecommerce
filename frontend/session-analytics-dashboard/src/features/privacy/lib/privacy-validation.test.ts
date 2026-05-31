import { describe, expect, it } from "vitest";

import {
  validateDeletionReason,
  validateDeletionTargetValue,
  validateRetentionSettings
} from "./privacy-validation";

describe("privacy validation helpers", () => {
  it("validates retention policy ranges", () => {
    expect(
      validateRetentionSettings({
        activeSessionTtlMinutes: 45,
        analyticsAggregatesMonths: 36,
        deletionRequestLogDays: 730,
        heatmapAggregatesDays: 365,
        journeySummariesDays: 365,
        rawEventsDays: 90
      })
    ).toEqual({});

    expect(
      validateRetentionSettings({
        activeSessionTtlMinutes: 45,
        analyticsAggregatesMonths: 36,
        deletionRequestLogDays: 730,
        heatmapAggregatesDays: 365,
        journeySummariesDays: 365,
        rawEventsDays: 2
      })
    ).toHaveProperty("rawEventsDays");
  });

  it("requires audit reasons for high-risk actions", () => {
    expect(validateDeletionReason("short")).toMatch(/at least 10/);
    expect(validateDeletionReason("Support ticket SUP-123")).toBeUndefined();
  });

  it("validates deletion target values", () => {
    expect(validateDeletionTargetValue("user_id", "user_123")).toBeUndefined();
    expect(validateDeletionTargetValue("session_id", "sess/123")).toMatch(
      /unsupported/
    );
  });
});
