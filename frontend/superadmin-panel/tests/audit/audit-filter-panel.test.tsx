import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { AuditFilterPanel } from "../../src/features/audit/components/audit-filter-panel";
import type { AuditLogFilters } from "../../src/features/audit/types";

const filters: AuditLogFilters = {
  actor_id: "",
  action: "",
  resource_type: "all",
  resource_id: "",
  request_id: "",
  from: "2026-06-01T10:00",
  to: "2026-06-02T10:00",
  page: 1,
  page_size: 25
};

describe("AuditFilterPanel", () => {
  it("emits audit filter and reset changes", () => {
    const onChange = vi.fn();
    const onReset = vi.fn();

    render(<AuditFilterPanel filters={filters} onChange={onChange} onReset={onReset} />);

    fireEvent.change(screen.getByPlaceholderText(/actor admin id/i), {
      target: { value: "admin_123" }
    });
    fireEvent.change(screen.getByLabelText(/audit action/i), {
      target: { value: "refund.approved" }
    });
    fireEvent.change(screen.getByLabelText(/audit resource type/i), {
      target: { value: "refund" }
    });
    fireEvent.change(screen.getByPlaceholderText(/resource id/i), {
      target: { value: "refund_123" }
    });
    fireEvent.change(screen.getByPlaceholderText(/request id/i), {
      target: { value: "req_123" }
    });
    fireEvent.change(screen.getByLabelText(/audit from date/i), {
      target: { value: "2026-06-02T09:00" }
    });
    fireEvent.change(screen.getByLabelText(/audit to date/i), {
      target: { value: "2026-06-02T12:00" }
    });
    fireEvent.change(screen.getByLabelText(/audit page size/i), {
      target: { value: "50" }
    });
    fireEvent.click(screen.getByRole("button", { name: /reset/i }));

    expect(onChange).toHaveBeenCalledWith({ actor_id: "admin_123" });
    expect(onChange).toHaveBeenCalledWith({ action: "refund.approved" });
    expect(onChange).toHaveBeenCalledWith({ resource_type: "refund" });
    expect(onChange).toHaveBeenCalledWith({ resource_id: "refund_123" });
    expect(onChange).toHaveBeenCalledWith({ request_id: "req_123" });
    expect(onChange).toHaveBeenCalledWith({ from: "2026-06-02T09:00" });
    expect(onChange).toHaveBeenCalledWith({ to: "2026-06-02T12:00" });
    expect(onChange).toHaveBeenCalledWith({ page_size: 50 });
    expect(onReset).toHaveBeenCalledTimes(1);
  });
});
