import { describe, expect, it } from "vitest";

import { inviteStaffSchema } from "./team-validation";

describe("team validation", () => {
  it("normalizes valid invite input", () => {
    expect(
      inviteStaffSchema.parse({
        email: " Catalog.Editor@Example.COM ",
        role: "seller_catalog_editor",
      }),
    ).toEqual({
      email: "catalog.editor@example.com",
      role: "seller_catalog_editor",
    });
  });

  it("rejects invalid email addresses", () => {
    const result = inviteStaffSchema.safeParse({
      email: "not-email",
      role: "seller_manager",
    });

    expect(result.success).toBe(false);
  });

  it("does not allow inviting another owner role", () => {
    const result = inviteStaffSchema.safeParse({
      email: "owner@example.com",
      role: "seller",
    });

    expect(result.success).toBe(false);
  });
});
