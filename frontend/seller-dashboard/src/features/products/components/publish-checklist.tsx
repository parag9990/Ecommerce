import { CheckCircle2, Circle } from "lucide-react";

type PublishChecklistProps = {
  hasTitle: boolean;
  hasCategory: boolean;
  hasVariant: boolean;
  hasImage: boolean;
};

export function PublishChecklist({
  hasTitle,
  hasCategory,
  hasVariant,
  hasImage,
}: PublishChecklistProps) {
  const items = [
    { label: "Title", done: hasTitle },
    { label: "Category", done: hasCategory },
    { label: "Variant", done: hasVariant },
    { label: "Image", done: hasImage },
  ];

  return (
    <section className="space-y-2 rounded-md border border-slate-200 bg-slate-50 p-3">
      <h2 className="text-sm font-semibold text-slate-950">Publish checklist</h2>
      <ul className="grid gap-2 sm:grid-cols-2">
        {items.map((item) => {
          const Icon = item.done ? CheckCircle2 : Circle;

          return (
            <li
              key={item.label}
              className={item.done ? "flex items-center gap-2 text-sm text-emerald-700" : "flex items-center gap-2 text-sm text-slate-500"}
            >
              <Icon className="h-4 w-4" aria-hidden="true" />
              {item.label}
            </li>
          );
        })}
      </ul>
    </section>
  );
}
