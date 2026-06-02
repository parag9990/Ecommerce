import { HIGH_RISK_AUDIT_ACTIONS } from "../constants";
import { cn } from "../../../lib/classnames";

export function isHighRiskAuditAction(action: string): boolean {
  return HIGH_RISK_AUDIT_ACTIONS.includes(action as (typeof HIGH_RISK_AUDIT_ACTIONS)[number]);
}

export function AuditStatusBadge({ action }: { action: string }) {
  const isHighRisk = isHighRiskAuditAction(action);

  return (
    <span
      className={cn(
        "inline-flex max-w-[220px] items-center rounded-md px-2 py-1 text-xs font-semibold",
        isHighRisk
          ? "bg-red-50 text-red-700 ring-1 ring-red-200"
          : "bg-slate-100 text-slate-700"
      )}
      title={action}
    >
      <span className="truncate">{action}</span>
    </span>
  );
}
