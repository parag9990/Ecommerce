import {
  analyticsReportTypeOptions,
  type AnalyticsReportType
} from "../../../api/session-api";
import { reportDescriptions } from "../lib/report-labels";

type ReportTypeSelectProps = {
  id?: string;
  value: AnalyticsReportType;
  onChange: (value: AnalyticsReportType) => void;
};

export function ReportTypeSelect({
  id = "report-type",
  value,
  onChange
}: ReportTypeSelectProps) {
  return (
    <label className="block" htmlFor={id}>
      <span className="mb-1 block text-xs font-semibold uppercase text-zinc-500">
        Report type
      </span>
      <select
        className="h-10 w-full rounded-md border border-zinc-300 bg-white px-3 text-sm outline-none transition-colors focus:border-zinc-950 focus:ring-2 focus:ring-zinc-950/10"
        id={id}
        onChange={(event) => onChange(event.target.value as AnalyticsReportType)}
        value={value}
      >
        {analyticsReportTypeOptions.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
      <span className="mt-1 block text-xs text-zinc-500">
        {reportDescriptions[value]}
      </span>
    </label>
  );
}
