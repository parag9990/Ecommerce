import { apiFetch } from "../../../lib/http";
import type {
  AdminUser,
  AdminUserFilters,
  AdminUserListResponse,
  JourneyResponse,
  SuccessResponse,
  UpdateUserStatusInput,
  UserSession,
  UserSessionListResponse,
  UserStatus
} from "../types";

const USERS_PATH = "/api/v1/admin/users";
const SESSIONS_PATH = "/api/v1/analytics/sessions";
const DEFAULT_USER_STATUS: UserStatus = "active";

function buildQueryString(values: Record<string, string | number | undefined>): string {
  const params = new URLSearchParams();

  Object.entries(values).forEach(([key, value]) => {
    if (value !== undefined && value !== "") {
      params.set(key, String(value));
    }
  });

  return params.toString();
}

function normalizeUser(user: AdminUser): AdminUser {
  return {
    ...user,
    status: user.status ?? DEFAULT_USER_STATUS,
    roles: user.roles?.length ? user.roles : ["buyer"]
  };
}

function normalizeSession(session: UserSession): UserSession {
  return {
    ...session,
    device: session.device ?? {}
  };
}

export function toAdminUsersQueryString(filters: AdminUserFilters): string {
  return buildQueryString({
    q: filters.q?.trim(),
    status: filters.status && filters.status !== "all" ? filters.status : undefined,
    page: filters.page,
    page_size: filters.limit
  });
}

export async function listAdminUsers(filters: AdminUserFilters): Promise<AdminUserListResponse> {
  const query = toAdminUsersQueryString(filters);
  const response = await apiFetch<AdminUserListResponse>(`${USERS_PATH}?${query}`);

  return {
    ...response,
    users: (response.users ?? []).map(normalizeUser),
    page: response.page ?? filters.page,
    limit: response.page_size ?? response.limit ?? filters.limit
  };
}

export async function getAdminUser(userId: string): Promise<AdminUser | null> {
  const response = await listAdminUsers({
    q: userId,
    status: "all",
    page: 1,
    limit: 1
  });

  return response.users.find((user) => user.user_id === userId) ?? null;
}

export async function updateUserStatus(input: UpdateUserStatusInput): Promise<SuccessResponse> {
  return apiFetch<SuccessResponse>(`${USERS_PATH}/${encodeURIComponent(input.userId)}/status`, {
    method: "PATCH",
    body: JSON.stringify({
      status: input.status,
      reason: input.reason
    })
  });
}

export async function listUserSessions(userId: string): Promise<UserSessionListResponse> {
  const query = buildQueryString({
    user_id: userId,
    page: 1,
    page_size: 20
  });
  const response = await apiFetch<UserSessionListResponse>(`${SESSIONS_PATH}?${query}`);

  return {
    ...response,
    sessions: (response.sessions ?? []).map(normalizeSession)
  };
}

export async function getSessionJourney(sessionId: string): Promise<JourneyResponse> {
  const response = await apiFetch<JourneyResponse>(
    `${SESSIONS_PATH}/${encodeURIComponent(sessionId)}/journey`
  );

  return {
    ...response,
    session: normalizeSession(response.session),
    events: response.events ?? []
  };
}
