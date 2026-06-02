import { Plus, Trash2 } from "lucide-react";

import { IconButton } from "../../../components/ui/icon-button";

type AttributeListEditorProps = {
  label: string;
  value?: Record<string, string>;
  onChange: (attributes: Record<string, string>) => void;
  compact?: boolean;
};

function createAttributeKey(attributes: Record<string, string>) {
  let index = Object.keys(attributes).length + 1;
  let key = `attribute_${index}`;

  while (key in attributes) {
    index += 1;
    key = `attribute_${index}`;
  }

  return key;
}

export function AttributeListEditor({
  label,
  value,
  onChange,
  compact = false,
}: AttributeListEditorProps) {
  const attributes = value ?? {};
  const entries = Object.entries(attributes);

  function addAttribute() {
    onChange({
      ...attributes,
      [createAttributeKey(attributes)]: "",
    });
  }

  function renameAttribute(oldKey: string, nextKey: string) {
    const trimmedKey = nextKey.trim();
    const next = { ...attributes };
    const attributeValue = next[oldKey] ?? "";

    delete next[oldKey];

    if (trimmedKey) {
      next[trimmedKey] = attributeValue;
    }

    onChange(next);
  }

  function updateAttributeValue(key: string, nextValue: string) {
    onChange({
      ...attributes,
      [key]: nextValue,
    });
  }

  function removeAttribute(key: string) {
    const next = { ...attributes };
    delete next[key];
    onChange(next);
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between gap-3">
        <h3 className={compact ? "text-xs font-semibold text-slate-700" : "text-sm font-semibold text-slate-950"}>
          {label}
        </h3>
        <button
          type="button"
          onClick={addAttribute}
          className="inline-flex h-8 items-center gap-1.5 rounded-md border border-slate-200 bg-white px-2.5 text-xs font-medium text-slate-700 transition hover:bg-slate-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
        >
          <Plus className="h-3.5 w-3.5" aria-hidden="true" />
          Add
        </button>
      </div>

      {entries.length ? (
        <div className="space-y-2">
          {entries.map(([key, attributeValue]) => (
            <div key={key} className="grid gap-2 sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto]">
              <input
                value={key}
                aria-label={`${label} key`}
                onChange={(event) => renameAttribute(key, event.target.value)}
                className="h-9 rounded-md border border-slate-300 px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
                placeholder="Key"
              />
              <input
                value={attributeValue}
                aria-label={`${label} value`}
                onChange={(event) => updateAttributeValue(key, event.target.value)}
                className="h-9 rounded-md border border-slate-300 px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
                placeholder="Value"
              />
              <IconButton
                label="Remove attribute"
                onClick={() => removeAttribute(key)}
                className="h-9 w-9 text-rose-600 hover:text-rose-700"
              >
                <Trash2 className="h-4 w-4" aria-hidden="true" />
              </IconButton>
            </div>
          ))}
        </div>
      ) : (
        <div className="rounded-md border border-dashed border-slate-200 px-3 py-2 text-xs text-slate-500">
          No attributes
        </div>
      )}
    </div>
  );
}
