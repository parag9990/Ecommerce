import { describe, expect, it } from "vitest";

import { sanitizeDisplayText, sanitizeEventProperties } from "./event-privacy";

describe("sanitizeEventProperties", () => {
  it("masks sensitive journey event fields recursively", () => {
    expect(
      sanitizeEventProperties({
        email: "buyer@example.com",
        product_id: "prod_123",
        query: "buyer@example.com",
        shipping: { address: "1 Secret Street" },
        user_id: "user_123"
      })
    ).toEqual({
      email: "[masked]",
      product_id: "prod_123",
      query: "[masked-email]",
      shipping: { address: "[masked]" },
      user_id: "[masked]"
    });
  });

  it("masks sensitive values inside display text", () => {
    expect(sanitizeDisplayText("/lookup?email=buyer@example.com")).toBe(
      "/lookup?email=[masked-email]"
    );
  });
});
