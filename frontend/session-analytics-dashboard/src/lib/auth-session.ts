const sessionKey = "session-analytics.admin-session";

const allowedAdminRoles = new Set([
  "admin",
  "operations_admin",
  "finance_admin",
  "catalog_admin",
  "superadmin"
]);

export type AnalyticsAdminSession = {
  accessToken: string;
  expiresAt?: string;
  userId: string;
  roles: string[];
};

export function readAdminSession(): AnalyticsAdminSession | null {
  if (typeof window === "undefined") {
    return null;
  }

  const raw = window.sessionStorage.getItem(sessionKey);
  if (!raw) {
    return null;
  }

  try {
    const parsed = JSON.parse(raw) as Partial<AnalyticsAdminSession>;
    if (
      !parsed.accessToken ||
      !parsed.userId ||
      !Array.isArray(parsed.roles) ||
      !parsed.roles.some(isAdminRole)
    ) {
      clearAdminSession();
      return null;
    }
    if (parsed.expiresAt && Date.parse(parsed.expiresAt) <= Date.now()) {
      clearAdminSession();
      return null;
    }
    return {
      accessToken: parsed.accessToken,
      expiresAt: parsed.expiresAt,
      roles: parsed.roles,
      userId: parsed.userId
    };
  } catch {
    clearAdminSession();
    return null;
  }
}

export function writeAdminSession(session: AnalyticsAdminSession): void {
  window.sessionStorage.setItem(sessionKey, JSON.stringify(session));
}

export function clearAdminSession(): void {
  if (typeof window !== "undefined") {
    window.sessionStorage.removeItem(sessionKey);
  }
}

export function getAccessToken(): string | undefined {
  return readAdminSession()?.accessToken;
}

function isAdminRole(role: string): boolean {
  return allowedAdminRoles.has(role.trim().toLowerCase());
}
