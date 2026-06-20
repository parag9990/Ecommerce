import { ExternalLink, FileText, ShieldAlert } from "lucide-react";

import { formatDateTime } from "../../lib/format";
import { StatusBadge } from "./status-badge";

export function DocumentPreview({
  title,
  status,
  url,
  createdAt,
  reviewedAt,
  rejectionReason,
  masked = false
}: {
  title: string;
  status: string;
  url?: string | null;
  createdAt?: string | null;
  reviewedAt?: string | null;
  rejectionReason?: string | null;
  masked?: boolean;
}) {
  if (masked) {
    return (
      <div className="rounded-lg border border-slate-200 bg-slate-50 p-3 text-sm">
        <div className="flex items-center gap-2 font-medium text-slate-900">
          <ShieldAlert className="h-4 w-4 text-slate-500" aria-hidden="true" />
          KYC document details masked
        </div>
        <p className="mt-1 text-slate-600">Readonly admins can see seller status without document URLs.</p>
      </div>
    );
  }

  return (
    <article className="rounded-lg border border-slate-200 bg-white p-3 text-sm">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <div className="flex min-w-0 items-center gap-2">
            <FileText className="h-4 w-4 shrink-0 text-slate-500" aria-hidden="true" />
            <h3 className="truncate font-medium text-slate-950">{title}</h3>
          </div>
          <p className="mt-1 text-xs text-slate-500">Uploaded {formatDateTime(createdAt)}</p>
          {reviewedAt ? <p className="mt-1 text-xs text-slate-500">Reviewed {formatDateTime(reviewedAt)}</p> : null}
        </div>
        <StatusBadge status={status} />
      </div>

      {rejectionReason ? (
        <p className="mt-3 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-red-700">
          {rejectionReason}
        </p>
      ) : null}

      {url ? (
        <a
          href={url}
          target="_blank"
          rel="noreferrer"
          className="mt-3 inline-flex items-center gap-2 text-sm font-medium text-blue-700 hover:text-blue-900"
        >
          Open secure document
          <ExternalLink className="h-4 w-4" aria-hidden="true" />
        </a>
      ) : (
        <p className="mt-3 text-sm text-slate-500">Secure document URL unavailable.</p>
      )}
    </article>
  );
}
