import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { AuditExportDialog } from "../../src/features/audit/components/audit-export-dialog";
import type { AuditLogFilters } from "../../src/features/audit/types";

const filters: AuditLogFilters = {
  resource_type: "refund",
  from: "2026-06-01T10:00",
  to: "2026-06-02T10:00",
  page: 1,
  page_size: 25
};

describe("AuditExportDialog", () => {
  it("requires a valid reason before exporting", async () => {
    const user = userEvent.setup();
    const onExport = vi.fn();

    render(
      <AuditExportDialog
        open
        filters={filters}
        canExport
        isExporting={false}
        error={null}
        onClose={vi.fn()}
        onExport={onExport}
      />
    );

    const exportButton = screen.getByRole("button", { name: /export csv/i });
    expect((exportButton as HTMLButtonElement).disabled).toBe(true);

    await user.type(screen.getByLabelText(/export reason/i), "Security review for refund logs");
    await user.click(exportButton);

    expect(onExport).toHaveBeenCalledWith("Security review for refund logs");
  });

  it("blocks export for readonly admins", () => {
    render(
      <AuditExportDialog
        open
        filters={filters}
        canExport={false}
        isExporting={false}
        error={null}
        onClose={vi.fn()}
        onExport={vi.fn()}
      />
    );

    expect(screen.getByText(/export requires superadmin/i)).toBeTruthy();
    expect((screen.getByRole("button", { name: /export csv/i }) as HTMLButtonElement).disabled).toBe(true);
  });
});
