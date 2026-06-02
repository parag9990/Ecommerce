import { Loader2 } from "lucide-react";

import { TimelineSkeleton } from "../../../components/state/loading-skeleton";
import type { SellerAuditLog } from "../types";
import { ActivityEventCard } from "./activity-event-card";

type ActivityTimelineProps = {
  logs: SellerAuditLog[];
  isLoading: boolean;
  hasNextPage: boolean;
  isFetchingNextPage: boolean;
  onLoadMore: () => void;
};

export function ActivityTimeline({
  logs,
  isLoading,
  hasNextPage,
  isFetchingNextPage,
  onLoadMore,
}: ActivityTimelineProps) {
  if (isLoading) {
    return <TimelineSkeleton rows={5} />;
  }

  return (
    <div className="space-y-3">
      <div className="relative space-y-3 before:absolute before:left-3 before:top-2 before:h-[calc(100%-1rem)] before:w-px before:bg-slate-200">
        {logs.map((log) => (
          <ActivityEventCard key={log.audit_id} log={log} />
        ))}
      </div>

      {hasNextPage ? (
        <div className="flex justify-center pt-1">
          <button
            type="button"
            onClick={onLoadMore}
            disabled={isFetchingNextPage}
            className="inline-flex h-9 items-center gap-2 rounded-md border border-slate-300 bg-white px-3 text-sm font-medium text-slate-700 transition hover:bg-slate-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {isFetchingNextPage ? (
              <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />
            ) : null}
            Load more
          </button>
        </div>
      ) : null}
    </div>
  );
}
