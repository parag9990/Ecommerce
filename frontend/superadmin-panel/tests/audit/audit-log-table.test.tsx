import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { AuditLogTable } from "../../src/features/audit/components/audit-log-table";
import type { AdminAuditLog } from "../../src/features/audit/types";

const logs: AdminAuditLog[] = [
  {
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
  }
];

describe("AuditLogTable", () => {
  it("renders audit investigation fields and opens selected log", () => {
    const onSelect = vi.fn();

    render(
      <AuditLogTable
        logs={logs}
        isLoading={false}
        isFetching={false}
        error={null}
        page={1}
        pageSize={25}
        totalCount={1}
        onPageChange={vi.fn()}
        onRetry={vi.fn()}
        onSelect={onSelect}
      />
    );

    expect(screen.getByText("admin_1")).toBeTruthy();
    expect(screen.getByText("refund.approved")).toBeTruthy();
    expect(screen.getByText("refund_1")).toBeTruthy();
    expect(screen.getByText("req_1")).toBeTruthy();
    expect(screen.getByText("sha256:hash_1")).toBeTruthy();

    fireEvent.click(screen.getByRole("button", { name: /view/i }));

    expect(onSelect).toHaveBeenCalledWith(logs[0]);
  });

  it("emits pagination changes", () => {
    const onPageChange = vi.fn();

    render(
      <AuditLogTable
        logs={logs}
        isLoading={false}
        isFetching={false}
        error={null}
        page={1}
        pageSize={1}
        totalCount={2}
        onPageChange={onPageChange}
        onRetry={vi.fn()}
        onSelect={vi.fn()}
      />
    );

    fireEvent.click(screen.getByRole("button", { name: /next/i }));

    expect(onPageChange).toHaveBeenCalledWith(2);
  });
});
