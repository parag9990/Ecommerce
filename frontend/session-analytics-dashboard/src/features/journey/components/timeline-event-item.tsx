import type { JourneyEvent } from "../../../api/session-api";
import { formatTime } from "../../../lib/format";
import {
  describeEvent,
  getEventIcon,
  getEventLabel,
  getEventTone,
  summarizeEventProperties
} from "../lib/event-format";
import { sanitizeDisplayText } from "../lib/event-privacy";

type TimelineEventItemProps = {
  event: JourneyEvent;
  isSelected: boolean;
  onSelect: () => void;
};

export function TimelineEventItem({
  event,
  isSelected,
  onSelect
}: TimelineEventItemProps) {
  const Icon = getEventIcon(event.eventType);
  const properties = summarizeEventProperties(event);

  return (
    <li className="relative">
      <span className="absolute -left-[29px] top-5 flex h-4 w-4 items-center justify-center rounded-full border border-zinc-300 bg-white">
        <span className="h-1.5 w-1.5 rounded-full bg-zinc-950" />
      </span>

      <button
        className={[
          "w-full rounded-lg border bg-white p-4 text-left shadow-panel transition-colors",
          isSelected
            ? "border-zinc-950 ring-2 ring-zinc-950/10"
            : "border-zinc-200 hover:border-zinc-300 hover:bg-zinc-50"
        ].join(" ")}
        onClick={onSelect}
        type="button"
      >
        <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-2">
              <span
                className={[
                  "inline-flex h-7 items-center gap-1.5 rounded-md border px-2 text-xs font-medium",
                  getEventTone(event.eventType)
                ].join(" ")}
              >
                <Icon className="h-3.5 w-3.5" aria-hidden="true" />
                {getEventLabel(event.eventType)}
              </span>
              <span className="font-mono text-xs text-zinc-500">
                {formatTime(event.occurredAt)}
              </span>
            </div>

            <p className="mt-2 text-sm font-medium text-zinc-950">
              {describeEvent(event)}
            </p>
            <p className="mt-1 truncate font-mono text-xs text-zinc-500">
              {sanitizeDisplayText(event.path)}
            </p>
          </div>

          {properties.length > 0 ? (
            <div className="flex max-w-full flex-wrap gap-1.5 sm:max-w-md sm:justify-end">
              {properties.map((property) => (
                <span
                  className="max-w-full truncate rounded-md bg-zinc-100 px-2 py-1 font-mono text-[11px] text-zinc-600"
                  key={property}
                >
                  {property}
                </span>
              ))}
            </div>
          ) : null}
        </div>
      </button>
    </li>
  );
}
