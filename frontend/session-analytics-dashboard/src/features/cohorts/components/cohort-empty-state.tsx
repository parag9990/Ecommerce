import { Repeat } from "lucide-react";

export function CohortEmptyState() {
  return (
    <section className="rounded-lg border border-zinc-200 bg-white p-6 text-center shadow-panel">
      <div className="mx-auto flex h-10 w-10 items-center justify-center rounded-md bg-zinc-100 text-zinc-500">
        <Repeat className="h-5 w-5" aria-hidden="true" />
      </div>
      <h3 className="mt-4 text-sm font-semibold text-zinc-950">
        No retention data for this selection
      </h3>
      <p className="mt-1 text-sm text-zinc-500">
        Try a wider range or fewer segment filters.
      </p>
    </section>
  );
}
