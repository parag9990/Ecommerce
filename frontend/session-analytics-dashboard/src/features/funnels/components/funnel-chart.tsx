import {
  Cell,
  Funnel,
  FunnelChart as RechartsFunnelChart,
  LabelList,
  ResponsiveContainer,
  Tooltip
} from "recharts";

import { funnelStepColors } from "../../../lib/chart-theme";
import { formatFunnelCount, formatFunnelPercent } from "../lib/funnel-format";
import type { FunnelReport } from "../lib/funnel-math";

type FunnelChartProps = {
  report: FunnelReport;
};

export function FunnelChart({ report }: FunnelChartProps) {
  const data = report.steps.map((step) => ({
    conversion: step.conversionFromStart,
    name: step.label,
    value: step.count
  }));

  return (
    <section className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel">
      <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h2 className="text-sm font-semibold text-zinc-950">
            Conversion funnel
          </h2>
          <p className="mt-1 text-sm text-zinc-500">
            Unique sessions that reached each ordered step.
          </p>
        </div>
        <div className="rounded-md bg-amber-50 px-3 py-2 text-sm text-amber-900">
          <span className="block text-xs font-medium uppercase text-amber-700">
            Overall
          </span>
          <span className="text-lg font-semibold">
            {formatFunnelPercent(report.overallConversionRate)}
          </span>
        </div>
      </div>

      <div className="h-[320px] w-full">
        <ResponsiveContainer height="100%" width="100%">
          <RechartsFunnelChart margin={{ bottom: 20, left: 16, right: 96, top: 8 }}>
            <Tooltip
              formatter={(value) => [
                formatFunnelCount(Number(value)),
                "Sessions"
              ]}
              labelFormatter={(label) => String(label)}
            />
            <Funnel data={data} dataKey="value" isAnimationActive={false}>
              <LabelList
                dataKey="name"
                fill="#3f3f46"
                position="right"
                stroke="none"
              />
              {data.map((entry, index) => (
                <Cell
                  fill={funnelStepColors[index % funnelStepColors.length]}
                  key={entry.name}
                />
              ))}
            </Funnel>
          </RechartsFunnelChart>
        </ResponsiveContainer>
      </div>

      {report.hasSmallCounts ? (
        <p className="mt-3 text-xs text-zinc-500">
          Small buckets are masked below {report.minSegmentSize} sessions.
        </p>
      ) : null}
    </section>
  );
}
