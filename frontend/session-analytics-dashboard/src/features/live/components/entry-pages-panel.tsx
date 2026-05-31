import { CornerDownRight } from "lucide-react";

import type { ActiveSessionBreakdownItem } from "../../../api/session-api";
import { normalizeBreakdownWidth } from "../../../lib/session-view";

type EntryPagesPanelProps = {
  pages: ActiveSessionBreakdownItem[];
};

export function EntryPagesPanel({ pages }: EntryPagesPanelProps) {
  const topPages = pages.slice(0, 6);

  return (
    <section className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel">
      <div className="mb-4 flex items-center justify-between gap-3">
        <h2 className="text-sm font-semibold text-zinc-950">Top entry pages</h2>
        <CornerDownRight className="h-4 w-4 text-emerald-600" aria-hidden="true" />
      </div>

      {topPages.length === 0 ? (
        <p className="text-sm text-zinc-500">No entry page data yet.</p>
      ) : (
        <div className="space-y-3">
          {topPages.map((page) => (
            <div key={page.label}>
              <div className="mb-1 flex items-center justify-between gap-3 text-sm">
                <span className="truncate font-mono text-xs text-zinc-700">
                  {page.label}
                </span>
                <span className="shrink-0 text-zinc-500">{page.count}</span>
              </div>
              <div className="h-2 rounded-full bg-zinc-100">
                <div
                  className="h-2 rounded-full bg-sky-500"
                  style={{ width: normalizeBreakdownWidth(page.percentage) }}
                />
              </div>
            </div>
          ))}
        </div>
      )}
    </section>
  );
}
