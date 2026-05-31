import { X } from "lucide-react";

import {
  journeyEventTypes,
  type JourneyEventType
} from "../../../api/session-api";
import { getEventIcon, getEventLabel } from "../lib/event-format";

type EventTypeFilterProps = {
  selectedTypes: JourneyEventType[];
  onChange: (types: JourneyEventType[]) => void;
};

export function EventTypeFilter({
  selectedTypes,
  onChange
}: EventTypeFilterProps) {
  function toggleType(type: JourneyEventType) {
    if (selectedTypes.includes(type)) {
      onChange(selectedTypes.filter((item) => item !== type));
      return;
    }

    onChange([...selectedTypes, type]);
  }

  return (
    <section className="rounded-lg border border-zinc-200 bg-white p-3 shadow-panel">
      <div className="flex flex-wrap items-center gap-2">
        {journeyEventTypes.map((type) => {
          const selected = selectedTypes.includes(type);
          const Icon = getEventIcon(type);

          return (
            <button
              aria-pressed={selected}
              className={[
                "inline-flex h-9 items-center gap-2 rounded-md border px-3 text-xs font-medium transition-colors",
                selected
                  ? "border-zinc-950 bg-zinc-950 text-white"
                  : "border-zinc-200 bg-zinc-50 text-zinc-700 hover:border-zinc-300 hover:bg-zinc-100"
              ].join(" ")}
              key={type}
              onClick={() => toggleType(type)}
              title={`Filter ${getEventLabel(type)} events`}
              type="button"
            >
              <Icon className="h-3.5 w-3.5" aria-hidden="true" />
              {getEventLabel(type)}
            </button>
          );
        })}

        {selectedTypes.length > 0 ? (
          <button
            className="inline-flex h-9 items-center gap-2 rounded-md border border-zinc-200 bg-white px-3 text-xs font-medium text-zinc-600 transition-colors hover:bg-zinc-50 hover:text-zinc-950"
            onClick={() => onChange([])}
            type="button"
          >
            <X className="h-3.5 w-3.5" aria-hidden="true" />
            Clear
          </button>
        ) : null}
      </div>
    </section>
  );
}
