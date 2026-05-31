const legendItems = [
  { className: "border-zinc-200 bg-zinc-50", label: "0%" },
  { className: "border-amber-300 bg-amber-200", label: "1-19%" },
  { className: "border-sky-500 bg-sky-400", label: "20-39%" },
  { className: "border-teal-600 bg-teal-500", label: "40-59%" },
  { className: "border-emerald-700 bg-emerald-600", label: "60%+" }
];

export function RetentionLegend() {
  return (
    <div className="flex flex-wrap items-center gap-2 text-xs text-zinc-600">
      <span className="font-semibold uppercase text-zinc-500">
        Retention scale
      </span>
      {legendItems.map((item) => (
        <span className="inline-flex items-center gap-1.5" key={item.label}>
          <span
            className={`h-3 w-5 rounded border ${item.className}`}
            aria-hidden="true"
          />
          {item.label}
        </span>
      ))}
    </div>
  );
}
