import { AlertTriangle } from "lucide-react";

export type SettingChangeSummaryItem = {
  label: string;
  value: string;
};

export function SettingChangeSummary({
  title,
  items,
  tone = "neutral"
}: {
  title: string;
  items: readonly SettingChangeSummaryItem[];
  tone?: "neutral" | "warning";
}) {
  return (
    <div
      className={
        tone === "warning"
          ? "rounded-lg border border-amber-200 bg-amber-50 p-3 text-sm text-amber-900"
          : "rounded-lg border border-slate-200 bg-slate-50 p-3 text-sm text-slate-700"
      }
    >
      <div className="flex items-center gap-2 font-medium">
        {tone === "warning" ? <AlertTriangle className="h-4 w-4" aria-hidden="true" /> : null}
        {title}
      </div>
      <dl className="mt-3 grid gap-2">
        {items.map((item) => (
          <div key={item.label} className="grid gap-1 sm:grid-cols-[160px_1fr]">
            <dt className="text-xs font-medium uppercase text-slate-500">{item.label}</dt>
            <dd className="break-words text-slate-900">{item.value}</dd>
          </div>
        ))}
      </dl>
    </div>
  );
}
