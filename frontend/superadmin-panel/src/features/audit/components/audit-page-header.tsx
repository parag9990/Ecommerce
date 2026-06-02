import { Download, ShieldCheck } from "lucide-react";

export function AuditPageHeader({
  canExport,
  isExporting,
  onExport
}: {
  canExport: boolean;
  isExporting: boolean;
  onExport: () => void;
}) {
  return (
    <header className="border-b border-slate-200 bg-white px-4 py-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="min-w-0">
          <div className="flex items-center gap-2 text-sm font-semibold text-slate-600">
            <ShieldCheck className="h-4 w-4" aria-hidden="true" />
            <span>Security review</span>
          </div>
          <h1 className="mt-1 text-xl font-semibold text-slate-950">Audit Logs</h1>
          <p className="mt-1 text-sm text-slate-600">
            Admin actions, request correlation, change summaries, and controlled CSV export.
          </p>
        </div>

        <button
          type="button"
          onClick={onExport}
          title={canExport ? "Export audit logs" : "Export requires superadmin"}
          className="inline-flex h-10 items-center justify-center gap-2 rounded-lg bg-slate-950 px-3 text-sm font-medium text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-70"
          disabled={isExporting}
        >
          <Download className="h-4 w-4" aria-hidden="true" />
          {isExporting ? "Exporting" : "Export CSV"}
        </button>
      </div>
    </header>
  );
}
