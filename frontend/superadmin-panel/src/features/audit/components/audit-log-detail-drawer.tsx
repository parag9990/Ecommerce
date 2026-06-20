import { X } from "lucide-react";
import type { ReactNode } from "react";

import { formatDateTime } from "../../../lib/format";
import { AuditSummaryDiff } from "./audit-summary-diff";
import { AuditStatusBadge } from "./audit-status-badge";
import type { AdminAuditLog } from "../types";

function utcTitle(value: string | null): string | undefined {
  if (!value) {
    return undefined;
  }

  const date = new Date(value);

  return Number.isNaN(date.getTime()) ? undefined : date.toISOString();
}

function DetailField({
  label,
  children,
  mono = false
}: {
  label: string;
  children: ReactNode;
  mono?: boolean;
}) {
  return (
    <div className="min-w-0">
      <dt className="text-xs font-medium uppercase text-slate-500">{label}</dt>
      <dd className={mono ? "mt-1 break-all font-mono text-xs text-slate-900" : "mt-1 text-sm text-slate-900"}>
        {children}
      </dd>
    </div>
  );
}

export function AuditLogDetailDrawer({
  log,
  onClose
}: {
  log: AdminAuditLog | null;
  onClose: () => void;
}) {
  if (!log) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-40">
      <button
        type="button"
        className="absolute inset-0 bg-slate-950/30"
        aria-label="Close audit log detail"
        onClick={onClose}
      />
      <aside className="absolute inset-y-0 right-0 flex w-full max-w-2xl flex-col overflow-hidden border-l border-slate-200 bg-slate-50 shadow-xl">
        <header className="border-b border-slate-200 bg-white px-4 py-4">
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <h2 className="text-lg font-semibold text-slate-950">Audit Log Detail</h2>
              <p className="mt-1 break-all font-mono text-xs text-slate-500">{log.id}</p>
            </div>
            <button
              type="button"
              onClick={onClose}
              title="Close audit log detail"
              className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-100"
            >
              <X className="h-4 w-4" aria-hidden="true" />
              <span className="sr-only">Close audit log detail</span>
            </button>
          </div>
        </header>

        <div className="min-h-0 flex-1 overflow-auto p-4">
          <section className="rounded-lg border border-slate-200 bg-white p-4">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <h3 className="text-base font-semibold text-slate-950">Action</h3>
                <div className="mt-2">
                  <AuditStatusBadge action={log.action} />
                </div>
              </div>
              <time
                dateTime={log.created_at ?? undefined}
                title={utcTitle(log.created_at)}
                className="text-sm text-slate-600"
              >
                {formatDateTime(log.created_at)}
              </time>
            </div>

            <dl className="mt-4 grid gap-4 sm:grid-cols-2">
              <DetailField label="Actor">{log.actor_admin_id}</DetailField>
              <DetailField label="Role">{log.actor_role}</DetailField>
              <DetailField label="Resource Type">{log.resource_type}</DetailField>
              <DetailField label="Resource ID" mono>
                {log.resource_id}
              </DetailField>
              <DetailField label="Request ID" mono>
                {log.request_id}
              </DetailField>
              <DetailField label="IP Hash" mono>
                {log.ip_hash}
              </DetailField>
            </dl>
          </section>

          <section className="mt-4 rounded-lg border border-amber-200 bg-amber-50 p-4">
            <h3 className="text-sm font-semibold text-amber-950">Admin Reason</h3>
            <p className="mt-1 text-sm text-amber-900">{log.reason ?? "No reason recorded."}</p>
          </section>

          <section className="mt-4 rounded-lg border border-slate-200 bg-white p-4">
            <AuditSummaryDiff before={log.before_summary} after={log.after_summary} />
          </section>
        </div>
      </aside>
    </div>
  );
}
