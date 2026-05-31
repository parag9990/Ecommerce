import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis
} from "recharts";

import type { NewReturningBucket } from "../../../api/session-api";
import { formatRetentionCount } from "../lib/retention-format";

type NewReturningChartProps = {
  data: NewReturningBucket[];
};

export function NewReturningChart({ data }: NewReturningChartProps) {
  const chartData = data.map((bucket) => ({
    ...bucket,
    label: bucket.label ?? bucket.bucket
  }));

  return (
    <section className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel">
      <div className="mb-4">
        <p className="text-xs font-semibold uppercase text-sky-700">
          Active identity mix
        </p>
        <h2 className="mt-1 text-base font-semibold text-zinc-950">
          New vs returning
        </h2>
      </div>

      <div className="h-72 min-h-72">
        <ResponsiveContainer height="100%" width="100%">
          <BarChart
            data={chartData}
            margin={{ bottom: 8, left: 0, right: 8, top: 8 }}
          >
            <CartesianGrid stroke="#e4e4e7" strokeDasharray="3 3" vertical={false} />
            <XAxis
              dataKey="label"
              tick={{ fill: "#52525b", fontSize: 12 }}
              tickLine={false}
            />
            <YAxis
              tick={{ fill: "#52525b", fontSize: 12 }}
              tickFormatter={(value) => formatRetentionCount(Number(value))}
              tickLine={false}
            />
            <Tooltip
              cursor={{ fill: "#f4f4f5" }}
              formatter={(value, name) => [
                formatRetentionCount(Number(value)),
                name === "newUsers" ? "New users" : "Returning users"
              ]}
            />
            <Legend />
            <Bar
              dataKey="newUsers"
              fill="#38bdf8"
              name="New users"
              radius={[4, 4, 0, 0]}
              stackId="users"
            />
            <Bar
              dataKey="returningUsers"
              fill="#10b981"
              name="Returning users"
              radius={[4, 4, 0, 0]}
              stackId="users"
            />
          </BarChart>
        </ResponsiveContainer>
      </div>
    </section>
  );
}
