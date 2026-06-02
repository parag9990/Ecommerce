import { describe, expect, it } from "vitest";

import {
  auditExportFilename,
  auditLogsToCsv,
  sanitizeCsvCell
} from "../../src/features/audit/audit-csv";
import type { AdminAuditLog } from "../../src/features/audit/types";

const auditLog: AdminAuditLog = {
  id: "log_1",
  actor_admin_id: "admin_1",
  actor_role: "superadmin",
  action: "refund.approved",
  resource_type: "refund",
  resource_id: "refund_1",
  request_id: "req_1",
  ip_hash: "sha256:hash_1",
  before_summary: { status: "pending" },
  after_summary: { status: "approved" },
  reason: "Finance investigation approved",
  created_at: "2026-06-02T10:00:00Z"
};

describe("audit CSV helpers", () => {
  it("exports safe audit columns without before and after summaries", () => {
    const csv = auditLogsToCsv([auditLog]);

    expect(csv).toContain("refund.approved");
    expect(csv).toContain("sha256:hash_1");
    expect(csv).not.toContain("before_summary");
    expect(csv).not.toContain("after_summary");
    expect(csv).not.toContain("pending");
  });

  it("quotes CSV cells and blocks spreadsheet formula injection", () => {
    const csv = auditLogsToCsv([
      {
        ...auditLog,
        action: "=IMPORTDATA(\"https://example.test\")",
        reason: "Contains comma, quote \" and newline\nsafely"
      }
    ]);

    expect(sanitizeCsvCell("=1+1")).toBe("'=1+1");
    expect(csv).toContain("'=IMPORTDATA");
    expect(csv).toContain('"Contains comma, quote "" and newline\nsafely"');
  });

  it("creates date-stamped export filenames", () => {
    expect(auditExportFilename(new Date("2026-06-02T10:00:00Z"))).toBe("audit-logs-2026-06-02.csv");
  });
});
