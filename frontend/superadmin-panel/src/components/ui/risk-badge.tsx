import { cn } from "../../lib/classnames";

const riskClassName: Record<string, string> = {
  high: "border-red-200 bg-red-50 text-red-700",
  medium: "border-amber-200 bg-amber-50 text-amber-800",
  low: "border-slate-300 bg-slate-100 text-slate-700"
};

export function RiskBadge({ level }: { level: "high" | "medium" | "low" }) {
  return (
    <span
      className={cn(
        "inline-flex h-6 items-center rounded-md border px-2 text-xs font-medium capitalize",
        riskClassName[level]
      )}
    >
      {level} risk
    </span>
  );
}
