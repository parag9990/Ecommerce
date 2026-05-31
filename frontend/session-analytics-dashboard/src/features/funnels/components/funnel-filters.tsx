import { ArrowRight } from "lucide-react";

import type { FunnelStepKey } from "../../../api/session-api";
import { FiltersBar } from "../../../layout/filters-bar";
import type { DateRange } from "../../../lib/date-range";
import type { SegmentFilters } from "../../shell/types";
import { funnelStepLabels } from "../lib/funnel-format";

type FunnelFiltersProps = {
  dateRange: DateRange;
  dateRangeError?: string;
  filters: SegmentFilters;
  steps: FunnelStepKey[];
  onDateRangeChange: (range: DateRange) => void;
  onFiltersChange: (filters: SegmentFilters) => void;
};

export function FunnelFilters({
  dateRange,
  dateRangeError,
  filters,
  steps,
  onDateRangeChange,
  onFiltersChange
}: FunnelFiltersProps) {
  return (
    <div className="sticky top-0 z-10 border-b border-zinc-200 bg-white/95 backdrop-blur">
      <FiltersBar
        dateRange={dateRange}
        dateRangeError={dateRangeError}
        description="Ordered conversion from product views to paid sessions."
        filters={filters}
        sticky={false}
        title="Funnel Analysis"
        onDateRangeChange={onDateRangeChange}
        onFiltersChange={onFiltersChange}
      />

      <div className="flex flex-wrap items-center gap-2 px-4 pb-4 lg:px-6">
        <span className="text-xs font-semibold uppercase text-zinc-500">
          Steps
        </span>
        {steps.map((step, index) => (
          <div className="flex items-center gap-2" key={step}>
            {index > 0 ? (
              <ArrowRight className="h-3.5 w-3.5 text-zinc-400" aria-hidden="true" />
            ) : null}
            <span className="rounded-md border border-zinc-200 bg-zinc-50 px-2.5 py-1 text-xs font-medium text-zinc-700">
              {funnelStepLabels[step]}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
