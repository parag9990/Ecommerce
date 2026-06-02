import { cn } from "../../lib/classnames";

const statusClassName: Record<string, string> = {
  active: "border-emerald-200 bg-emerald-50 text-emerald-700",
  approved: "border-emerald-200 bg-emerald-50 text-emerald-700",
  delivered: "border-emerald-200 bg-emerald-50 text-emerald-700",
  paid: "border-emerald-200 bg-emerald-50 text-emerald-700",
  published: "border-emerald-200 bg-emerald-50 text-emerald-700",
  captured: "border-emerald-200 bg-emerald-50 text-emerald-700",
  matched: "border-emerald-200 bg-emerald-50 text-emerald-700",
  succeeded: "border-emerald-200 bg-emerald-50 text-emerald-700",
  authorized: "border-blue-200 bg-blue-50 text-blue-700",
  blocked: "border-red-200 bg-red-50 text-red-700",
  cancelled: "border-red-200 bg-red-50 text-red-700",
  failed: "border-red-200 bg-red-50 text-red-700",
  manual_review: "border-red-200 bg-red-50 text-red-700",
  mismatch: "border-red-200 bg-red-50 text-red-700",
  missing_local: "border-red-200 bg-red-50 text-red-700",
  missing_provider: "border-red-200 bg-red-50 text-red-700",
  payment_failed: "border-red-200 bg-red-50 text-red-700",
  suspended: "border-red-200 bg-red-50 text-red-700",
  rejected: "border-rose-200 bg-rose-50 text-rose-700",
  disputed: "border-orange-200 bg-orange-50 text-orange-800",
  retry_allowed: "border-orange-200 bg-orange-50 text-orange-800",
  pending: "border-amber-200 bg-amber-50 text-amber-800",
  pending_payment: "border-amber-200 bg-amber-50 text-amber-800",
  pending_review: "border-amber-200 bg-amber-50 text-amber-800",
  requires_action: "border-amber-200 bg-amber-50 text-amber-800",
  packed: "border-blue-200 bg-blue-50 text-blue-700",
  shipped: "border-indigo-200 bg-indigo-50 text-indigo-700",
  partially_refunded: "border-violet-200 bg-violet-50 text-violet-700",
  refunded: "border-violet-200 bg-violet-50 text-violet-700",
  resolved: "border-violet-200 bg-violet-50 text-violet-700",
  draft: "border-slate-300 bg-slate-100 text-slate-600",
  created: "border-slate-300 bg-slate-100 text-slate-600",
  deleted: "border-slate-300 bg-slate-100 text-slate-600",
  initiated: "border-slate-300 bg-slate-100 text-slate-600",
  none: "border-slate-300 bg-slate-100 text-slate-600",
  unpublished: "border-slate-300 bg-slate-100 text-slate-600",
  not_started: "border-slate-300 bg-slate-100 text-slate-600"
};

export function StatusBadge({ status }: { status: string }) {
  const label = status.replace(/_/g, " ");

  return (
    <span
      className={cn(
        "inline-flex h-6 items-center rounded-md border px-2 text-xs font-medium capitalize",
        statusClassName[status] ?? "border-slate-300 bg-white text-slate-700"
      )}
    >
      {label}
    </span>
  );
}
