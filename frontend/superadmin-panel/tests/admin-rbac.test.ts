import { describe, expect, it } from "vitest";

import { adminMenu } from "../src/config/admin-menu";
import {
  canViewMenuItem,
  filterAdminMenu,
  isAdminRole,
  isAdminUser,
  normalizeAdminRoles
} from "../src/lib/admin-rbac";

describe("admin RBAC", () => {
  it("recognizes supported admin roles", () => {
    expect(isAdminRole("superadmin")).toBe(true);
    expect(isAdminRole("operations_admin")).toBe(true);
    expect(isAdminRole("buyer")).toBe(false);
  });

  it("normalizes mixed role lists to known admin roles", () => {
    expect(normalizeAdminRoles(["buyer", "finance_admin", "unknown"])).toEqual(["finance_admin"]);
  });

  it("allows supported admin roles and blocks non-admin roles", () => {
    expect(isAdminUser(["superadmin"])).toBe(true);
    expect(isAdminUser(["readonly_admin"])).toBe(true);
    expect(isAdminUser(["buyer", "seller"])).toBe(false);
  });

  it("shows finance menu items to finance admins", () => {
    const payments = adminMenu.find((item) => item.id === "payments");

    expect(payments).toBeTruthy();
    expect(canViewMenuItem(["finance_admin"], payments!)).toBe(true);
    expect(canViewMenuItem(["operations_admin"], payments!)).toBe(false);
  });

  it("filters the menu from the documented visibility matrix", () => {
    expect(filterAdminMenu(["operations_admin"], adminMenu).map((item) => item.id)).toEqual([
      "overview",
      "users",
      "sellers",
      "orders",
      "sessions"
    ]);

    expect(filterAdminMenu(["catalog_admin"], adminMenu).map((item) => item.id)).toEqual([
      "overview",
      "sellers",
      "search"
    ]);

    expect(filterAdminMenu(["readonly_admin"], adminMenu).map((item) => item.id)).toEqual([
      "overview",
      "users",
      "sellers",
      "orders",
      "payments",
      "sessions",
      "search",
      "audit-logs"
    ]);

    expect(filterAdminMenu(["superadmin"], adminMenu)).toHaveLength(adminMenu.length);
  });
});
