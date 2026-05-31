import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";

import type { JourneyEvent, JourneyEventType } from "../../../api/session-api";
import { EventDetailPanel } from "../components/event-detail-panel";
import { EventTypeFilter } from "../components/event-type-filter";
import { JourneyEmptyState } from "../components/journey-empty-state";
import { JourneySessionSearch } from "../components/journey-session-search";
import { SessionContextCard } from "../components/session-context-card";
import { SessionTimeline } from "../components/session-timeline";
import { useSessionJourney } from "../hooks/use-session-journey";
import { filterJourneyEvents } from "../lib/journey-filter";

export function JourneyExplorerPage() {
  const navigate = useNavigate();
  const { sessionId } = useParams();
  const [selectedTypes, setSelectedTypes] = useState<JourneyEventType[]>([]);
  const [selectedEvent, setSelectedEvent] = useState<JourneyEvent | null>(null);
  const journeyQuery = useSessionJourney(sessionId);

  const filteredEvents = useMemo(
    () => filterJourneyEvents(journeyQuery.data?.events ?? [], selectedTypes),
    [journeyQuery.data?.events, selectedTypes]
  );

  useEffect(() => {
    setSelectedEvent(null);
  }, [sessionId]);

  return (
    <div className="p-4 lg:p-6">
      <div className="mb-5 flex flex-col gap-3 xl:flex-row xl:items-end xl:justify-between">
        <div>
          <p className="text-xs font-semibold uppercase text-emerald-700">
            Session diagnostics
          </p>
          <h1 className="mt-1 text-2xl font-semibold text-zinc-950">
            Journey explorer
          </h1>
        </div>

        <JourneySessionSearch
          initialValue={sessionId ?? ""}
          onSubmit={(nextSessionId) =>
            navigate(`/journey/${encodeURIComponent(nextSessionId)}`)
          }
        />
      </div>

      {!sessionId ? (
        <JourneyEmptyState />
      ) : (
        <div className="grid gap-5 xl:grid-cols-[minmax(0,1fr)_380px]">
          <section className="space-y-4">
            {journeyQuery.data ? (
              <SessionContextCard
                session={journeyQuery.data.session}
                summary={journeyQuery.data.summary}
              />
            ) : null}

            <EventTypeFilter
              selectedTypes={selectedTypes}
              onChange={setSelectedTypes}
            />

            <SessionTimeline
              error={journeyQuery.error}
              events={filteredEvents}
              isLoading={journeyQuery.isPending}
              onRetry={() => void journeyQuery.refetch()}
              onSelectEvent={setSelectedEvent}
              selectedEventId={selectedEvent?.eventId}
            />
          </section>

          <EventDetailPanel
            event={selectedEvent}
            onClose={() => setSelectedEvent(null)}
          />
        </div>
      )}
    </div>
  );
}
