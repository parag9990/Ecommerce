import { describe, expect, it } from "vitest";

import { formatDateTime, formatMemberName } from "./team-formatters";

describe("team formatters", () => {
  it("formats ISO dates into readable display text", () => {
    const formatted = formatDateTime("2026-06-01T10:00:00Z");

    expect(formatted).toContain("2026");
    expect(formatted).not.toContain("T10:00:00Z");
  });

  it("falls back to Unknown for invalid dates", () => {
    expect(formatDateTime("not-a-date")).toBe("Unknown");
  });

  it("prefers full name and falls back to email", () => {
    expect(formatMemberName("Catalog Editor", "catalog@example.com")).toBe(
      "Catalog Editor",
    );
    expect(formatMemberName(" ", "catalog@example.com")).toBe("catalog@example.com");
    expect(formatMemberName(null, "catalog@example.com")).toBe("catalog@example.com");
  });
});
