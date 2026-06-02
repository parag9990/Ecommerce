import { formatPercentage } from "../utils/analytics-formatters";
import { AnalyticsEmptyState } from "./analytics-empty-state";

type ConversionChartProps = {
  conversionRate: number | null;
};

export function ConversionChart({ conversionRate }: ConversionChartProps) {
  if (conversionRate === null) {
    return (
      <AnalyticsEmptyState
        title="Conversion rate"
        description="Conversion metric is not available from the current CMS analytics response."
      />
    );
  }

  const safeRate = Math.max(0, Math.min(conversionRate ?? 0, 100));

  return (
    <section className="rounded-md border border-slate-200 bg-white p-4 shadow-sm">
      <div>
        <h2 className="text-base font-semibold text-slate-950">Conversion rate</h2>
        <p className="mt-1 text-xs text-slate-500">Selected range performance.</p>
      </div>

      <div className="mt-6">
        <div className="flex items-end justify-between gap-4">
          <p className="text-3xl font-semibold text-slate-950">
            {formatPercentage(conversionRate)}
          </p>
          <span className="rounded border border-slate-200 px-2 py-1 text-xs font-medium text-slate-500">
            0-100%
          </span>
        </div>

        <div
          className="mt-5 h-3 rounded-full bg-slate-100"
          role="meter"
          aria-label="Conversion rate"
          aria-valuemin={0}
          aria-valuemax={100}
          aria-valuenow={safeRate}
        >
          <div
            className="h-3 rounded-full bg-blue-600"
            style={{ width: `${safeRate}%` }}
          />
        </div>

        <div className="mt-2 flex justify-between text-[11px] text-slate-400">
          <span>0%</span>
          <span>50%</span>
          <span>100%</span>
        </div>
      </div>
    </section>
  );
}
