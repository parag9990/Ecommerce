export const SESSION_RISK_LEVELS = ["low", "medium", "high"] as const;
export const SESSION_STATUSES = ["active", "revoked", "expired"] as const;

export type SessionRiskLevel = (typeof SESSION_RISK_LEVELS)[number];
export type AdminSessionStatus = (typeof SESSION_STATUSES)[number];

export type SessionDevice = {
  browser?: string | null;
  os?: string | null;
  operating_system?: string | null;
  device_type?: "desktop" | "mobile" | "tablet" | "unknown" | string | null;
  country?: string | null;
  city?: string | null;
  language?: string | null;
  ip_hash?: string | null;
  device_fingerprint_hash?: string | null;
};

export type AdminSession = {
  session_id: string;
  anonymous_id?: string | null;
  user_id?: string | null;
  started_at: string;
  last_seen_at: string;
  revoked_at?: string | null;
  status?: AdminSessionStatus | string | null;
  device?: SessionDevice;
  risk_score?: number | null;
  risk_level?: SessionRiskLevel | null;
  risk_flags?: string[] | null;
  event_count?: number | null;
  events_per_minute?: number | null;
  checkout_step_count?: number | null;
  cart_action_count?: number | null;
  payment_failure_count?: number | null;
  active_session_count?: number | null;
};

export type SessionEvent = {
  event_type: string;
  anonymous_id: string;
  session_id: string;
  user_id?: string | null;
  occurred_at: string;
  path?: string | null;
  properties?: Record<string, unknown> | null;
};

export type LiveMetrics = {
  active_users: number;
  active_sessions: number;
  events_per_minute: number;
  suspicious_sessions?: number;
  updated_at?: string | null;
};

export type AdminSessionFilters = {
  user_id?: string;
  from?: string;
  to?: string;
  status: AdminSessionStatus | "all";
  risk_level: SessionRiskLevel | "all";
  page: number;
  limit: number;
};

export type SessionListResponse = {
  sessions: AdminSession[];
  total?: number;
  page?: number;
  limit?: number;
};

export type SessionJourney = {
  session: AdminSession;
  events: SessionEvent[];
};

export type SessionRisk = {
  score: number;
  level: SessionRiskLevel;
  reasons: string[];
};
