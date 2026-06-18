type PlaceholderPageProps = {
  title: string;
};

export function PlaceholderPage({ title }: PlaceholderPageProps) {
  return (
    <section className="rounded-md border border-dashed border-slate-300 bg-white p-6 shadow-sm">
      <h1 className="text-2xl font-semibold text-slate-950">{title}</h1>
      <p className="mt-2 text-sm leading-6 text-slate-600">
        This page will be implemented in a later user app frontend task.
      </p>
    </section>
  );
}
