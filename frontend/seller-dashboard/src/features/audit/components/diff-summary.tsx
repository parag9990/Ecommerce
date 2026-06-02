import type { AuditJson } from "../types";
import { formatAuditValue, getChangedFields } from "../utils/audit-formatters";

type DiffSummaryProps = {
  before: AuditJson;
  after: AuditJson;
};

export function DiffSummary({ before, after }: DiffSummaryProps) {
  const changedFields = getChangedFields(before, after);
  const visibleFields = changedFields.slice(0, 4);

  if (visibleFields.length === 0) {
    return null;
  }

  return (
    <div className="mt-3 rounded-md bg-slate-50 p-3">
      <div className="mb-2 text-xs font-medium text-slate-500">Changed fields</div>

      <dl className="grid gap-2 text-sm md:grid-cols-2">
        {visibleFields.map((change) => (
          <div key={change.field} className="min-w-0">
            <dt className="font-medium text-slate-700">{change.field}</dt>
            <dd className="mt-1 min-w-0 truncate text-slate-600">
              <span className="line-through decoration-red-400">
                {formatAuditValue(change.before, change.field)}
              </span>
              <span className="mx-2 text-slate-400">to</span>
              <span className="font-medium text-emerald-700">
                {formatAuditValue(change.after, change.field)}
              </span>
            </dd>
          </div>
        ))}
      </dl>

      {changedFields.length > visibleFields.length ? (
        <p className="mt-2 text-xs text-slate-500">
          {changedFields.length - visibleFields.length} more changed fields
        </p>
      ) : null}
    </div>
  );
}
