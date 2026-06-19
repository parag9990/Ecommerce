import { MapPin } from "lucide-react";

import type { ActiveSessionBreakdownItem } from "../../../api/session-api";
import { normalizeBreakdownWidth } from "../../../lib/session-view";

type ActiveUserMapProps = {
  locations: ActiveSessionBreakdownItem[];
};

export function ActiveUserMap({ locations }: ActiveUserMapProps) {
  const topLocations = locations.slice(0, 6);

  return (
    <section className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel">
      <div className="mb-4 flex items-center justify-between gap-3">
        <h2 className="text-sm font-semibold text-zinc-950">Active locations</h2>
        <MapPin className="h-4 w-4 text-emerald-600" aria-hidden="true" />
      </div>

      {topLocations.length === 0 ? (
        <p className="text-sm text-zinc-500">No location data yet.</p>
      ) : (
        <div className="space-y-3">
          {topLocations.map((location) => (
            <div key={location.label}>
              <div className="mb-1 flex items-center justify-between gap-3 text-sm">
                <span className="truncate text-zinc-700">{location.label}</span>
                <span className="shrink-0 text-zinc-500">{location.count}</span>
              </div>
              <div className="h-2 rounded-full bg-zinc-100">
                <div
                  className="h-2 rounded-full bg-emerald-500"
                  style={{ width: normalizeBreakdownWidth(location.percentage) }}
                />
              </div>
            </div>
          ))}
        </div>
      )}
    </section>
  );
}
