import type { ReactNode } from "react";

import { cn } from "../../lib/classnames";

type DataStateTone = "neutral" | "danger";

export function DataState({
  title,
  description,
  tone = "neutral",
  action
}: {
  title: string;
  description?: string;
  tone?: DataStateTone;
  action?: ReactNode;
}) {
  return (
    <div
      className={cn(
        "rounded-lg border p-4 text-sm",
        tone === "danger"
          ? "border-red-200 bg-red-50 text-red-800"
          : "border-slate-200 bg-white text-slate-700"
      )}
    >
      <div className="font-medium">{title}</div>
      {description ? <div className="mt-1 opacity-80">{description}</div> : null}
      {action ? <div className="mt-3">{action}</div> : null}
    </div>
  );
}
