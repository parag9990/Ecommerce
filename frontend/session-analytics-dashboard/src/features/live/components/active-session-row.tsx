import { Clock3, Route } from "lucide-react";
import { Link } from "react-router-dom";

import type { ActiveSession } from "../../../api/session-api";
import {
  formatDuration,
  formatNumber,
  formatRelativeTime
} from "../../../lib/format";
import {
  formatDeviceLabel,
  formatLocation,
  maskAnonymousId,
  maskSessionId,
  maskUserId
} from "../../../lib/session-view";

type ActiveSessionRowProps = {
  session: ActiveSession;
};

export function ActiveSessionRow({ session }: ActiveSessionRowProps) {
  return (
    <tr className="hover:bg-zinc-50">
      <td className="min-w-56 px-4 py-3 align-top">
        <div className="font-medium text-zinc-950">
          {maskAnonymousId(session.anonymousId)}
        </div>
        <div className="mt-1 text-xs text-zinc-500">
          {maskUserId(session.maskedUserId)}
        </div>
        <div className="mt-1 max-w-56 truncate font-mono text-xs text-zinc-400">
          {maskSessionId(session.sessionId)}
        </div>
      </td>
      <td className="min-w-40 px-4 py-3 align-top text-sm">
        <div className="capitalize text-zinc-800">{session.device.type}</div>
        <div className="mt-1 text-xs text-zinc-500">
          {formatDeviceLabel(session.device)}
        </div>
      </td>
      <td className="min-w-44 px-4 py-3 align-top text-sm text-zinc-700">
        {formatLocation(session.location)}
      </td>
      <td className="min-w-48 px-4 py-3 align-top">
        <PathValue value={session.entryPage} />
      </td>
      <td className="min-w-48 px-4 py-3 align-top">
        <PathValue value={session.currentPage ?? "-"} />
      </td>
      <td className="min-w-32 px-4 py-3 align-top text-sm text-zinc-700">
        <div>{formatDuration(session.durationSeconds)}</div>
        <div className="mt-1 text-xs text-zinc-500">
          {formatNumber(session.eventCount)} events
        </div>
      </td>
      <td className="min-w-32 px-4 py-3 align-top text-sm text-zinc-700">
        <div className="flex items-center gap-1.5">
          <Clock3
            className="h-3.5 w-3.5 text-emerald-600"
            aria-hidden="true"
          />
          {formatRelativeTime(session.lastSeenAt)}
        </div>
      </td>
      <td className="min-w-28 px-4 py-3 align-top text-sm">
        <Link
          aria-label="Open journey for selected session"
          className="inline-flex h-8 items-center gap-2 rounded-md border border-zinc-300 bg-white px-3 text-xs font-medium text-zinc-700 transition-colors hover:bg-zinc-50 hover:text-zinc-950"
          to={`/journey/${encodeURIComponent(session.sessionId)}`}
        >
          <Route className="h-3.5 w-3.5" aria-hidden="true" />
          Journey
        </Link>
      </td>
    </tr>
  );
}

function PathValue({ value }: { value: string }) {
  return (
    <span className="block max-w-64 truncate font-mono text-xs text-zinc-700">
      {value}
    </span>
  );
}
