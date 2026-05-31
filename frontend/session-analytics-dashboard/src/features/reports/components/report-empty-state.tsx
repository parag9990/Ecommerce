import { CalendarClock } from "lucide-react";

type ReportEmptyStateProps = {
  title: string;
  description: string;
};

export function ReportEmptyState({ description, title }: ReportEmptyStateProps) {
  return (
    <div className="rounded-lg border border-dashed border-zinc-300 bg-white p-6">
      <div className="flex gap-3">
        <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-zinc-100 text-zinc-600">
          <CalendarClock className="h-5 w-5" aria-hidden="true" />
        </div>
        <div>
          <h3 className="text-sm font-semibold text-zinc-950">{title}</h3>
          <p className="mt-1 text-sm text-zinc-500">{description}</p>
        </div>
      </div>
    </div>
  );
}
