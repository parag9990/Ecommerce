import type { ReactNode } from "react";

import { cn } from "../../lib/classnames";

export type TimelineItem = {
  id: string;
  title: ReactNode;
  description?: ReactNode;
  timestamp?: ReactNode;
};

export function Timeline({ items, emptyLabel }: { items: TimelineItem[]; emptyLabel: string }) {
  if (items.length === 0) {
    return <p className="text-sm text-slate-600">{emptyLabel}</p>;
  }

  return (
    <ol className="space-y-4">
      {items.map((item, index) => (
        <li key={item.id} className="grid grid-cols-[1rem_minmax(0,1fr)] gap-3">
          <div className="relative flex justify-center">
            <span className="mt-1 h-2.5 w-2.5 rounded-full bg-slate-900" />
            {index < items.length - 1 ? (
              <span className="absolute bottom-[-1rem] top-4 w-px bg-slate-200" aria-hidden="true" />
            ) : null}
          </div>
          <div className={cn("min-w-0 pb-1", index < items.length - 1 && "border-b border-slate-100")}>
            <div className="text-sm font-medium text-slate-950">{item.title}</div>
            {item.description ? <div className="mt-1 text-sm text-slate-600">{item.description}</div> : null}
            {item.timestamp ? <div className="mt-1 text-xs text-slate-500">{item.timestamp}</div> : null}
          </div>
        </li>
      ))}
    </ol>
  );
}
