import type { JourneyEvent, JourneyEventType } from "../../../api/session-api";

export function filterJourneyEvents(
  events: JourneyEvent[],
  selectedTypes: JourneyEventType[]
): JourneyEvent[] {
  if (selectedTypes.length === 0) {
    return events;
  }

  return events.filter((event) => selectedTypes.includes(event.eventType));
}
