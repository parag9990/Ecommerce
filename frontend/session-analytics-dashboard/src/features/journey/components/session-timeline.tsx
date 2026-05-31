import { AlertTriangle, RefreshCcw } from "lucide-react";

import type { JourneyEvent } from "../../../api/session-api";
import { ApiError } from "../../../lib/http";
import { TimelineEventItem } from "./timeline-event-item";

type SessionTimelineProps = {
  error: unknown;
  events: JourneyEvent[];
  isLoading: boolean;
  onRetry: () => void;
  onSelectEvent: (event: JourneyEvent) => void;
  selectedEventId?: string;
};

export function SessionTimeline({
  error,
  events,
  isLoading,
  onRetry,
  onSelectEvent,
  selectedEventId
}: SessionTimelineProps) {
  if (isLoading) {
    return <TimelineLoadingState />;
  }

  if (error) {
    return <TimelineErrorState error={error} onRetry={onRetry} />;
  }

  if (events.length === 0) {
    return (
      <section className="rounded-lg border border-zinc-200 bg-white p-6 shadow-panel">
        <h2 className="text-sm font-semibold text-zinc-950">No journey events</h2>
        <p className="mt-1 text-sm text-zinc-500">
          No ordered session events match the current view.
        </p>
      </section>
    );
  }

  return (
    <section
      aria-label="Session event timeline"
      className="rounded-lg border border-zinc-200 bg-zinc-50 p-4 shadow-panel"
    >
      <ol className="relative space-y-3 border-l border-zinc-300 pl-5">
        {events.map((event) => (
          <TimelineEventItem
            event={event}
            isSelected={event.eventId === selectedEventId}
            key={event.eventId}
            onSelect={() => onSelectEvent(event)}
          />
        ))}
      </ol>
    </section>
  );
}

function TimelineLoadingState() {
  return (
    <section
      aria-label="Loading journey"
      className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel"
    >
      <div className="space-y-3">
        {Array.from({ length: 5 }).map((_, index) => (
          <div
            aria-hidden="true"
            className="h-24 animate-pulse rounded-lg bg-zinc-100"
            key={index}
          />
        ))}
      </div>
    </section>
  );
}

function TimelineErrorState({
  error,
  onRetry
}: {
  error: unknown;
  onRetry: () => void;
}) {
  const notFound = error instanceof ApiError && error.status === 404;

  return (
    <section className="rounded-lg border border-red-200 bg-red-50 p-4 text-red-900">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex gap-3">
          <AlertTriangle className="mt-0.5 h-5 w-5 shrink-0" aria-hidden="true" />
          <div>
            <h2 className="text-sm font-semibold">
              {notFound ? "Session not found" : "Journey unavailable"}
            </h2>
            <p className="mt-1 text-sm text-red-800">
              {notFound
                ? "The analytics service could not find that session id."
                : "The analytics service could not return this session journey."}
            </p>
          </div>
        </div>
        <button
          className="inline-flex h-9 items-center justify-center gap-2 rounded-md border border-red-300 bg-white px-3 text-sm font-medium text-red-900 transition-colors hover:bg-red-100"
          onClick={onRetry}
          type="button"
        >
          <RefreshCcw className="h-4 w-4" aria-hidden="true" />
          Retry
        </button>
      </div>
    </section>
  );
}
