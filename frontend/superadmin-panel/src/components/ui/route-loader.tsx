export function RouteLoader({ label }: { label: string }) {
  return (
    <main className="flex min-h-screen items-center justify-center bg-slate-50 px-4">
      <div className="flex items-center gap-3 rounded-lg border border-slate-200 bg-white px-4 py-3 shadow-sm">
        <span className="h-2.5 w-2.5 animate-pulse rounded-full bg-slate-900" aria-hidden="true" />
        <p className="text-sm font-medium text-slate-700">{label}...</p>
      </div>
    </main>
  );
}
