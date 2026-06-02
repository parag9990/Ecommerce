import { useEffect, useMemo, useState } from "react";
import { Download, ShieldAlert, X } from "lucide-react";

import { validateAuditExportReason } from "../validators";
import type { AuditLogFilters } from "../types";

function activeFilterCount(filters: AuditLogFilters): number {
  return Object.entries(filters).filter(([key, value]) => {
    if (key === "page" || key === "page_size") {
      return false;
    }

    return value !== undefined && value !== "" && value !== "all";
  }).length;
}

export function AuditExportDialog({
  open,
  filters,
  canExport,
  isExporting,
  error,
  onClose,
  onExport
}: {
  open: boolean;
  filters: AuditLogFilters;
  canExport: boolean;
  isExporting: boolean;
  error: unknown;
  onClose: () => void;
  onExport: (reason: string) => void;
}) {
  const [reason, setReason] = useState("");
  const validation = useMemo(() => validateAuditExportReason(reason), [reason]);
  const filterCount = activeFilterCount(filters);

  useEffect(() => {
    if (!open) {
      setReason("");
    }
  }, [open]);

  if (!open) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-50 grid place-items-center bg-slate-950/40 p-4">
      <section className="w-full max-w-lg rounded-lg border border-slate-200 bg-white shadow-xl">
        <header className="flex items-start justify-between gap-3 border-b border-slate-200 px-4 py-4">
          <div>
            <h2 className="text-lg font-semibold text-slate-950">Export Audit Logs</h2>
            <p className="mt-1 text-sm text-slate-600">
              {filterCount > 0 ? `${filterCount} active filters` : "Current date range"} · page size{" "}
              {filters.page_size}
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            title="Close export dialog"
            className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-100"
          >
            <X className="h-4 w-4" aria-hidden="true" />
            <span className="sr-only">Close export dialog</span>
          </button>
        </header>

        <form
          className="p-4"
          onSubmit={(event) => {
            event.preventDefault();

            if (canExport && validation.valid && !isExporting) {
              onExport(reason);
            }
          }}
        >
          {!canExport ? (
            <div className="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-800">
              <div className="flex items-center gap-2 font-semibold">
                <ShieldAlert className="h-4 w-4" aria-hidden="true" />
                Export requires superadmin
              </div>
              <p className="mt-1 opacity-80">Readonly admins can inspect audit logs but cannot export them.</p>
            </div>
          ) : (
            <label className="block">
              <span className="text-sm font-medium text-slate-700">Export reason</span>
              <textarea
                value={reason}
                onChange={(event) => setReason(event.target.value)}
                className="mt-1 min-h-28 w-full resize-y rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-950 outline-none focus:border-slate-900 focus:ring-2 focus:ring-slate-900/10"
                placeholder="Security review for refund approval investigation"
              />
              {!validation.valid && reason.length > 0 ? (
                <span className="mt-1 block text-xs font-medium text-red-700">
                  {validation.errors[0]}
                </span>
              ) : null}
            </label>
          )}

          {error ? (
            <div className="mt-4 rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-800">
              {error instanceof Error ? error.message : "Audit export failed."}
            </div>
          ) : null}

          <div className="mt-5 flex justify-end gap-2">
            <button
              type="button"
              onClick={onClose}
              className="h-10 rounded-lg border border-slate-300 px-3 text-sm font-medium text-slate-700 hover:bg-slate-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={!canExport || !validation.valid || isExporting}
              className="inline-flex h-10 items-center justify-center gap-2 rounded-lg bg-slate-950 px-3 text-sm font-medium text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-50"
            >
              <Download className="h-4 w-4" aria-hidden="true" />
              {isExporting ? "Exporting" : "Export CSV"}
            </button>
          </div>
        </form>
      </section>
    </div>
  );
}
