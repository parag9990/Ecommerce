import { Route } from "lucide-react";

export function JourneyEmptyState() {
  return (
    <section className="flex min-h-[420px] items-center justify-center rounded-lg border border-dashed border-zinc-300 bg-white p-6 shadow-panel">
      <div className="max-w-sm text-center">
        <div className="mx-auto flex h-11 w-11 items-center justify-center rounded-md bg-zinc-950 text-white">
          <Route className="h-5 w-5" aria-hidden="true" />
        </div>
        <h2 className="mt-4 text-base font-semibold text-zinc-950">
          Select a session
        </h2>
        <p className="mt-2 text-sm text-zinc-500">
          Open a row from Live Sessions or enter a session id above.
        </p>
      </div>
    </section>
  );
}
