import { RefreshCw } from "lucide-react";

import { cn } from "../../lib/cn";

type RefreshingNoticeProps = {
  show: boolean;
  failed?: boolean;
  message?: string;
  onRetry?: () => void;
};

export function RefreshingNotice({
  show,
  failed = false,
  message,
  onRetry,
}: RefreshingNoticeProps) {
  if (!show) {
    return null;
  }

  return (
    <div
      className={cn(
        "inline-flex min-h-8 flex-wrap items-center gap-2 rounded-md border px-3 py-1.5 text-xs font-medium",
        failed
          ? "border-yellow-200 bg-yellow-50 text-yellow-800"
          : "border-slate-200 bg-white text-slate-500",
      )}
      role={failed ? "alert" : "status"}
      aria-live={failed ? "assertive" : "polite"}
    >
      <RefreshCw className={cn("h-3.5 w-3.5", !failed && "animate-spin")} aria-hidden="true" />
      <span>{message ?? (failed ? "Latest refresh failed. Showing saved data." : "Refreshing latest data...")}</span>
      {failed && onRetry ? (
        <button
          type="button"
          onClick={onRetry}
          className="rounded border border-yellow-300 bg-white px-2 py-0.5 text-[11px] font-semibold text-yellow-800 transition hover:bg-yellow-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-yellow-600"
        >
          Retry
        </button>
      ) : null}
    </div>
  );
}
