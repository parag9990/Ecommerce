import type { LucideIcon } from "lucide-react";

type MetricTone = "neutral" | "positive" | "warning";

type MetricCardProps = {
  helper: string;
  icon: LucideIcon;
  label: string;
  tone?: MetricTone;
  value: string;
};

const toneStyles: Record<MetricTone, string> = {
  neutral: "border-zinc-200 bg-white text-zinc-950",
  positive: "border-emerald-200 bg-emerald-50 text-emerald-950",
  warning: "border-amber-200 bg-amber-50 text-amber-950"
};

export function MetricCard({
  helper,
  icon: Icon,
  label,
  tone = "neutral",
  value
}: MetricCardProps) {
  return (
    <article
      className={[
        "h-32 rounded-lg border p-4 shadow-panel",
        toneStyles[tone]
      ].join(" ")}
    >
      <div className="flex items-center justify-between gap-3">
        <p className="truncate text-sm font-medium text-zinc-600">{label}</p>
        <Icon className="h-4 w-4 shrink-0 text-zinc-500" aria-hidden="true" />
      </div>
      <p className="mt-3 truncate text-2xl font-semibold tracking-normal">
        {value}
      </p>
      <p className="mt-1 line-clamp-2 text-sm text-zinc-500">{helper}</p>
    </article>
  );
}
