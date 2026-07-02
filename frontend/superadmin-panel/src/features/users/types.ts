export const USER_STATUSES = ["active", "blocked", "deleted"] as const;

export type UserStatus = (typeof USER_STATUSES)[number];

export type AdminUser = {
  user_id: string;
  email?: string | null;
  phone?: string | null;
  full_name?: string | null;
  status: UserStatus;
  roles: string[];
  created_at?: string | null;
  last_login_at?: string | null;
};

export type AdminUserFilters = {
  q?: string;
  status?: UserStatus | "all";
  page: number;
  limit: number;
};

export type AdminUserListResponse = {
  users: AdminUser[];
  total?: number;
  page?: number;
  limit?: number;
  page_size?: number;
};

export type UpdateUserStatusInput = {
  userId: string;
  status: Extract<UserStatus, "active" | "blocked">;
  reason: string;
};

export type SuccessResponse = {
  success: boolean;
};

export type UserSessionDevice = {
  browser?: string;
  os?: string;
  device_type?: string;
};

export type UserSession = {
  session_id: string;
  anonymous_id: string;
  user_id?: string;
  started_at: string;
  last_seen_at: string;
  device?: UserSessionDevice;
};

export type UserSessionListResponse = {
  sessions: UserSession[];
  total?: number;
};

export type SessionEvent = {
  event_type: string;
  anonymous_id: string;
  session_id: string;
  user_id?: string;
  occurred_at: string;
  path?: string;
  properties?: Record<string, unknown>;
};

export type JourneyResponse = {
  session: UserSession;
  events: SessionEvent[];
};
