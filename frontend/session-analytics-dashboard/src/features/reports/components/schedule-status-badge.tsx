import type { ReportScheduleStatus } from "../../../api/session-api";

const statusClassName: Record<ReportScheduleStatus, string> = {
  active: "border-emerald-200 bg-emerald-50 text-emerald-800",
  failed: "border-red-200 bg-red-50 text-red-800",
  paused: "border-zinc-200 bg-zinc-100 text-zinc-700"
};

export function ScheduleStatusBadge({
  status
}: {
  status: ReportScheduleStatus;
}) {
  return (
    <span
      className={[
        "inline-flex h-7 items-center rounded-md border px-2 text-xs font-semibold capitalize",
        statusClassName[status]
      ].join(" ")}
    >
      {status}
    </span>
  );
}
