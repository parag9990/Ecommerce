import { describe, expect, it } from "vitest";

import { canExportAuditLogs, canViewAuditLogs } from "../../src/features/audit/permissions";

describe("audit permissions", () => {
  it("allows superadmin and readonly admins to view audit logs", () => {
    expect(canViewAuditLogs(["superadmin"])).toBe(true);
    expect(canViewAuditLogs(["readonly_admin"])).toBe(true);
    expect(canViewAuditLogs(["finance_admin"])).toBe(false);
    expect(canViewAuditLogs(["operations_admin"])).toBe(false);
    expect(canViewAuditLogs(["catalog_admin"])).toBe(false);
  });

  it("allows only superadmin to export audit logs", () => {
    expect(canExportAuditLogs(["superadmin"])).toBe(true);
    expect(canExportAuditLogs(["readonly_admin"])).toBe(false);
    expect(canExportAuditLogs(["finance_admin"])).toBe(false);
  });
});
