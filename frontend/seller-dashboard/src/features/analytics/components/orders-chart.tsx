import type { AnalyticsSeriesPoint } from "../types";
import { toOrdersRows } from "../utils/analytics-chart-data";
import { formatNumber, formatShortDate } from "../utils/analytics-formatters";
import { AnalyticsEmptyState } from "./analytics-empty-state";

type OrdersChartProps = {
  series: AnalyticsSeriesPoint[];
};

export function OrdersChart({ series }: OrdersChartProps) {
  const rows = toOrdersRows(series);

  if (rows.length === 0) {
    return (
      <AnalyticsEmptyState
        title="Orders trend"
        description="Orders trend points are not available from the current CMS analytics response."
      />
    );
  }

  const maxOrders = Math.max(...rows.map((row) => row.orders), 1);
  const labelStep = Math.max(1, Math.ceil(rows.length / 6));

  return (
    <section className="rounded-md border border-slate-200 bg-white p-4 shadow-sm">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-base font-semibold text-slate-950">Orders trend</h2>
          <p className="mt-1 text-xs text-slate-500">Period-wise order count.</p>
        </div>
        <span className="text-xs font-medium text-slate-500">
          Peak {formatNumber(maxOrders)}
        </span>
      </div>

      <div className="mt-4 overflow-x-auto">
        <div className="min-w-[34rem]">
          <div className="flex h-56 items-end gap-2 rounded-md border border-slate-200 bg-slate-50 px-3 pb-3 pt-4">
            {rows.map((row) => {
              const height = Math.max(4, (row.orders / maxOrders) * 100);

              return (
                <div
                  key={row.label}
                  className="flex h-full min-w-7 flex-1 items-end"
                  title={`${formatShortDate(row.label)}: ${formatNumber(row.orders)} orders`}
                >
                  <div
                    className="w-full rounded-t bg-emerald-500 transition hover:bg-emerald-600"
                    style={{ height: `${height}%` }}
                    aria-label={`${formatNumber(row.orders)} orders on ${formatShortDate(
                      row.label,
                    )}`}
                  />
                </div>
              );
            })}
          </div>

          <div className="mt-2 flex gap-2 px-3 text-[11px] text-slate-500">
            {rows.map((row, index) => (
              <span key={row.label} className="min-w-7 flex-1 text-center">
                {index % labelStep === 0 || index === rows.length - 1
                  ? formatShortDate(row.label)
                  : ""}
              </span>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
