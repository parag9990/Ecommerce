import { describe, expect, it } from "vitest";

import {
  formatByMaskingMode,
  formatLocationByGranularity,
  maskIdentifier
} from "./privacy-format";

describe("privacy format helpers", () => {
  it("masks identifiers without exposing the full value", () => {
    expect(maskIdentifier("user_123456789")).toBe("user_1...6789");
    expect(maskIdentifier("short")).toBe("****");
    expect(maskIdentifier("")).toBe("-");
  });

  it("formats values by masking mode", () => {
    expect(formatByMaskingMode("user_123456789", "hidden")).toBe("Hidden");
    expect(formatByMaskingMode("user_123456789", "masked")).toBe(
      "user_1...6789"
    );
    expect(formatByMaskingMode("user_123456789", "full")).toBe(
      "user_123456789"
    );
  });

  it("formats location by configured granularity", () => {
    const geo = { city: "Delhi", country: "India", region: "Delhi" };

    expect(formatLocationByGranularity(geo, "none")).toBe("Hidden");
    expect(formatLocationByGranularity(geo, "country")).toBe("India");
    expect(formatLocationByGranularity(geo, "city")).toBe(
      "Delhi, Delhi, India"
    );
  });
});
