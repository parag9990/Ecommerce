import { formatFunnelCount, formatFunnelPercent } from "../lib/funnel-format";
import type { FunnelReport } from "../lib/funnel-math";

type FunnelDropoffTableProps = {
  report: FunnelReport;
};

export function FunnelDropoffTable({ report }: FunnelDropoffTableProps) {
  const rows = report.steps.slice(1).map((step, index) => ({
    from: report.steps[index],
    to: step
  }));

  return (
    <section className="rounded-lg border border-zinc-200 bg-white shadow-panel">
      <div className="border-b border-zinc-200 px-4 py-3">
        <h2 className="text-sm font-semibold text-zinc-950">Step drop-off</h2>
        <p className="mt-1 text-sm text-zinc-500">
          Loss and conversion from one funnel step to the next.
        </p>
      </div>

      <div className="overflow-x-auto">
        <table className="min-w-[760px] text-left text-sm">
          <thead className="bg-zinc-50 text-xs uppercase text-zinc-500">
            <tr>
              <th className="px-4 py-3 font-semibold">From</th>
              <th className="px-4 py-3 font-semibold">To</th>
              <th className="px-4 py-3 font-semibold">Lost</th>
              <th className="px-4 py-3 font-semibold">Drop-off rate</th>
              <th className="px-4 py-3 font-semibold">Conversion</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-zinc-200">
            {rows.map(({ from, to }) => (
              <tr key={`${from.key}-${to.key}`}>
                <td className="px-4 py-3 font-medium text-zinc-900">
                  {from.label}
                </td>
                <td className="px-4 py-3 text-zinc-700">{to.label}</td>
                <td className="px-4 py-3 font-medium text-red-700">
                  {formatFunnelCount(to.dropoffFromPrevious ?? 0, {
                    threshold: report.minSegmentSize
                  })}
                </td>
                <td className="px-4 py-3 text-red-700">
                  {formatFunnelPercent(to.dropoffRateFromPrevious)}
                </td>
                <td className="px-4 py-3 text-emerald-700">
                  {formatFunnelPercent(to.conversionFromPrevious)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
