import { FileSpreadsheet } from "lucide-react";

export function ReportFormatBadge() {
  return (
    <span className="inline-flex h-7 items-center gap-1.5 rounded-md border border-emerald-200 bg-emerald-50 px-2 text-xs font-semibold uppercase text-emerald-800">
      <FileSpreadsheet className="h-3.5 w-3.5" aria-hidden="true" />
      CSV
    </span>
  );
}
