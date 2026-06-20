export function AdminModulePlaceholderPage({
  title,
  description
}: {
  title: string;
  description: string;
}) {
  return (
    <section>
      <div className="rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
        <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Module Shell</p>
        <h1 className="mt-1 text-xl font-semibold text-slate-950">{title}</h1>
        <p className="mt-1 text-sm text-slate-500">{description}</p>
        <p className="mt-4 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800">
          Data tables, filters, forms, mutations, exports, and operational workflows are reserved for
          later Superadmin Panel tasks.
        </p>
      </div>
    </section>
  );
}
