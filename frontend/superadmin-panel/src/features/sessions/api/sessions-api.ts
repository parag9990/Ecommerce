import { apiFetch } from "../../../lib/http";
import type {
  AdminSession,
  AdminSessionFilters,
  LiveMetrics,
  SessionDevice,
  SessionEvent,
  SessionJourney,
  SessionListResponse
} from "../types";

const LIVE_METRICS_PATH = "/api/v1/analytics/live";
const SESSIONS_PATH = "/api/v1/analytics/sessions";

type SessionQueryParams = Pick<AdminSessionFilters, "page" | "limit" | "user_id" | "from" | "to">;

function buildQueryString(values: Record<string, string | number | undefined>): string {
  const params = new URLSearchParams();

  Object.entries(values).forEach(([key, value]) => {
    if (value !== undefined && value !== "") {
      params.set(key, String(value));
    }
  });

  return params.toString();
}

function normalizeDevice(device?: SessionDevice | null): SessionDevice {
  if (!device) {
    return {
      device_type: "unknown"
    };
  }

  return {
    ...device,
    operating_system: device.operating_system ?? device.os ?? null,
    device_type: device.device_type ?? "unknown"
  };
}

function normalizeSession(session: AdminSession): AdminSession {
  return {
    ...session,
    anonymous_id: session.anonymous_id ?? null,
    user_id: session.user_id ?? null,
    status: session.status ?? (session.revoked_at ? "revoked" : "active"),
    device: normalizeDevice(session.device)
  };
}

function toOrderedEvents(events: SessionEvent[] = []): SessionEvent[] {
  return [...events].sort((left, right) => {
    const leftTime = new Date(left.occurred_at).getTime();
    const rightTime = new Date(right.occurred_at).getTime();

    if (Number.isNaN(leftTime) || Number.isNaN(rightTime)) {
      return 0;
    }

    return leftTime - rightTime;
  });
}

export function toAdminSessionsQueryString(filters: SessionQueryParams): string {
  return buildQueryString({
    page: filters.page,
    page_size: filters.limit,
    user_id: filters.user_id?.trim(),
    from: filters.from,
    to: filters.to
  });
}

export async function getLiveMetrics(): Promise<LiveMetrics> {
  const response = await apiFetch<Partial<LiveMetrics>>(LIVE_METRICS_PATH);

  return {
    active_users: response.active_users ?? 0,
    active_sessions: response.active_sessions ?? 0,
    events_per_minute: response.events_per_minute ?? 0,
    suspicious_sessions: response.suspicious_sessions ?? 0,
    updated_at: response.updated_at ?? null
  };
}

export async function listAdminSessions(filters: SessionQueryParams): Promise<SessionListResponse> {
  const query = toAdminSessionsQueryString(filters);
  const response = await apiFetch<SessionListResponse>(`${SESSIONS_PATH}?${query}`);

  return {
    ...response,
    sessions: (response.sessions ?? []).map(normalizeSession),
    page: response.page ?? filters.page,
    limit: response.page_size ?? response.limit ?? filters.limit
  };
}

export async function getAdminSessionJourney(sessionId: string): Promise<SessionJourney> {
  const response = await apiFetch<SessionJourney>(
    `${SESSIONS_PATH}/${encodeURIComponent(sessionId)}/journey`
  );

  return {
    session: normalizeSession(response.session),
    events: toOrderedEvents(response.events ?? [])
  };
}
