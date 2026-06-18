import { describe, expect, it } from "vitest";

import {
  formatAuditTime,
  formatAuditValue,
  formatRelativeAuditTime,
  getChangedFields,
  shouldRedactAuditField,
} from "./audit-formatters";

describe("audit-formatters", () => {
  it("detects changed fields between before and after snapshots", () => {
    expect(getChangedFields({ price: 100 }, { price: 90 })).toEqual([
      {
        field: "price",
        before: 100,
        after: 90,
      },
    ]);
  });

  it("returns no changed fields when one snapshot is absent", () => {
    expect(getChangedFields(null, { status: "active" })).toEqual([]);
  });

  it("formats relative time from a stable clock", () => {
    expect(
      formatRelativeAuditTime(
        "2026-06-02T09:55:00.000Z",
        new Date("2026-06-02T10:00:00.000Z").getTime(),
      ),
    ).toBe("5 minutes ago");
  });

  it("formats invalid timestamps as unknown", () => {
    expect(formatAuditTime("not-a-date")).toBe("Unknown time");
    expect(formatRelativeAuditTime("not-a-date")).toBe("Unknown time");
  });

  it("formats empty and sensitive values safely", () => {
    expect(formatAuditValue(null)).toBe("Empty");
    expect(formatAuditValue(true)).toBe("Yes");
    expect(shouldRedactAuditField("refresh_token")).toBe(true);
    expect(formatAuditValue("secret-value", "refresh_token")).toBe("Redacted");
  });
});
