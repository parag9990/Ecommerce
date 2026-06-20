import { describe, expect, it } from "vitest";

import {
  canMutateUserStatus,
  canViewUserManagement,
  canViewUserProfile,
  canViewUserSessions
} from "../../src/lib/admin-permissions";
import { maskPhone } from "../../src/lib/format";

describe("user management permissions", () => {
  it("allows user management views for documented roles", () => {
    expect(canViewUserManagement(["superadmin"])).toBe(true);
    expect(canViewUserProfile(["operations_admin"])).toBe(true);
    expect(canViewUserSessions(["readonly_admin"])).toBe(true);
  });

  it("keeps finance and catalog admins out of user management", () => {
    expect(canViewUserManagement(["finance_admin"])).toBe(false);
    expect(canViewUserManagement(["catalog_admin"])).toBe(false);
  });

  it("allows only superadmin and operations admins to update user status", () => {
    expect(canMutateUserStatus(["superadmin"])).toBe(true);
    expect(canMutateUserStatus(["operations_admin"])).toBe(true);
    expect(canMutateUserStatus(["readonly_admin"])).toBe(false);
  });
});

describe("user privacy formatting", () => {
  it("masks phone numbers to the last four digits", () => {
    expect(maskPhone("+919999999999")).toBe("********9999");
    expect(maskPhone("")).toBe("Not added");
  });
});
