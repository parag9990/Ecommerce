import type { AnalyticsSeriesPoint } from "../types";
import { toRevenueGmvRows } from "../utils/analytics-chart-data";
import { formatCompactNumber, formatShortDate } from "../utils/analytics-formatters";
import { AnalyticsEmptyState } from "./analytics-empty-state";

type RevenueGmvChartProps = {
  series: AnalyticsSeriesPoint[];
};

const chart = {
  width: 680,
  height: 280,
  left: 56,
  right: 20,
  top: 20,
  bottom: 38,
};

const innerWidth = chart.width - chart.left - chart.right;
const innerHeight = chart.height - chart.top - chart.bottom;

function getX(index: number, total: number) {
  if (total <= 1) {
    return chart.left + innerWidth / 2;
  }

  return chart.left + (index / (total - 1)) * innerWidth;
}

function getY(value: number, maxValue: number) {
  return chart.top + innerHeight - (value / maxValue) * innerHeight;
}

function buildPath(
  rows: ReturnType<typeof toRevenueGmvRows>,
  key: "revenue" | "gmv",
  maxValue: number,
) {
  const points = rows
    .map((row, index) => ({ index, value: row[key] }))
    .filter((point): point is { index: number; value: number } => point.value !== undefined);

  return points
    .map((point, pointIndex) => {
      const command = pointIndex === 0 ? "M" : "L";
      return `${command} ${getX(point.index, rows.length).toFixed(1)} ${getY(
        point.value,
        maxValue,
      ).toFixed(1)}`;
    })
    .join(" ");
}

export function RevenueGmvChart({ series }: RevenueGmvChartProps) {
  const rows = toRevenueGmvRows(series);
  const values = rows.flatMap((row) => [row.revenue, row.gmv]).filter((value): value is number => {
    return value !== undefined;
  });

  if (rows.length === 0 || values.length === 0) {
    return (
      <AnalyticsEmptyState
        title="Revenue and GMV trend"
        description="Trend series is not available from the current CMS analytics response."
        className="min-h-80"
      />
    );
  }

  const maxValue = Math.max(...values, 1);
  const revenuePath = buildPath(rows, "revenue", maxValue);
  const gmvPath = buildPath(rows, "gmv", maxValue);
  const yTicks = [maxValue, maxValue / 2, 0];
  const labelStep = Math.max(1, Math.ceil(rows.length / 5));

  return (
    <section className="rounded-md border border-slate-200 bg-white p-4 shadow-sm">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-base font-semibold text-slate-950">Revenue and GMV trend</h2>
          <p className="mt-1 text-xs text-slate-500">Real trend points returned by CMS.</p>
        </div>
        <div className="flex items-center gap-3 text-xs text-slate-500">
          <span className="inline-flex items-center gap-1">
            <span className="h-2 w-2 rounded-full bg-blue-600" />
            Revenue
          </span>
          <span className="inline-flex items-center gap-1">
            <span className="h-2 w-2 rounded-full bg-emerald-600" />
            GMV
          </span>
        </div>
      </div>

      <div className="mt-4 overflow-x-auto">
        <svg
          viewBox={`0 0 ${chart.width} ${chart.height}`}
          role="img"
          aria-label="Revenue and GMV trend chart"
          className="h-72 min-w-[42rem] overflow-visible"
        >
          {yTicks.map((tick) => {
            const y = getY(tick, maxValue);

            return (
              <g key={tick}>
                <line
                  x1={chart.left}
                  x2={chart.width - chart.right}
                  y1={y}
                  y2={y}
                  stroke="#e2e8f0"
                  strokeWidth="1"
                />
                <text x="0" y={y + 4} fill="#64748b" fontSize="12">
                  {formatCompactNumber(tick)}
                </text>
              </g>
            );
          })}

          {rows.map((row, index) =>
            index % labelStep === 0 || index === rows.length - 1 ? (
              <text
                key={`${row.label}-${index}`}
                x={getX(index, rows.length)}
                y={chart.height - 8}
                textAnchor="middle"
                fill="#64748b"
                fontSize="12"
              >
                {formatShortDate(row.label)}
              </text>
            ) : null,
          )}

          {revenuePath ? (
            <path
              d={revenuePath}
              fill="none"
              stroke="#2563eb"
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth="3"
            />
          ) : null}
          {gmvPath ? (
            <path
              d={gmvPath}
              fill="none"
              stroke="#059669"
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth="3"
            />
          ) : null}
        </svg>
      </div>
    </section>
  );
}
