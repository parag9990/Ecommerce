import { Info, X } from "lucide-react";
import { useMemo } from "react";

import type { JourneyEvent } from "../../../api/session-api";
import { formatDateTime } from "../../../lib/format";
import { getEventLabel } from "../lib/event-format";
import {
  sanitizeDisplayText,
  sanitizeEventProperties
} from "../lib/event-privacy";

type EventDetailPanelProps = {
  event: JourneyEvent | null;
  onClose: () => void;
};

export function EventDetailPanel({ event, onClose }: EventDetailPanelProps) {
  const propertiesJson = useMemo(() => {
    if (!event) {
      return "";
    }

    return JSON.stringify(sanitizeEventProperties(event.properties), null, 2);
  }, [event]);

  if (!event) {
    return (
      <aside className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel">
        <div className="flex items-start gap-3 text-sm text-zinc-500">
          <Info className="mt-0.5 h-4 w-4 shrink-0 text-zinc-400" aria-hidden="true" />
          <p>Select a timeline event to inspect its sanitized properties.</p>
        </div>
      </aside>
    );
  }

  return (
    <aside className="overflow-hidden rounded-lg border border-zinc-200 bg-white shadow-panel xl:sticky xl:top-6">
      <div className="flex items-center justify-between gap-3 border-b border-zinc-200 px-4 py-3">
        <div className="min-w-0">
          <p className="text-sm font-semibold text-zinc-950">
            {getEventLabel(event.eventType)}
          </p>
          <p className="truncate font-mono text-xs text-zinc-500">
            {event.eventId}
          </p>
        </div>
        <button
          aria-label="Close event detail"
          className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-zinc-500 transition-colors hover:bg-zinc-100 hover:text-zinc-950"
          onClick={onClose}
          title="Close event detail"
          type="button"
        >
          <X className="h-4 w-4" aria-hidden="true" />
        </button>
      </div>

      <dl className="space-y-4 px-4 py-4 text-sm">
        <div>
          <dt className="text-xs font-semibold uppercase text-zinc-500">Path</dt>
          <dd className="mt-1 break-all font-mono text-xs text-zinc-800">
            {sanitizeDisplayText(event.path)}
          </dd>
        </div>
        <div>
          <dt className="text-xs font-semibold uppercase text-zinc-500">
            Occurred
          </dt>
          <dd className="mt-1 text-zinc-800">
            {formatDateTime(event.occurredAt)}
          </dd>
        </div>
        <div>
          <dt className="text-xs font-semibold uppercase text-zinc-500">
            Properties
          </dt>
          <dd className="mt-2 rounded-md bg-zinc-950 p-3">
            <pre className="max-h-[520px] overflow-auto whitespace-pre-wrap break-words text-xs leading-5 text-zinc-100">
              {propertiesJson}
            </pre>
          </dd>
        </div>
      </dl>
    </aside>
  );
}
