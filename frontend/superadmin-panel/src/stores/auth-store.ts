import { create } from "zustand";

import { normalizeAdminRoles, type AdminRole } from "../lib/admin-rbac";

const SESSION_STORAGE_KEY = "superadmin.session";

export type AdminUser = {
  id: string;
  email: string;
  name: string;
  roles: string[];
};

export type AdminSession = {
  accessToken: string;
  user: AdminUser;
  expiresAt?: string;
};

type AuthState = {
  accessToken: string | null;
  user: AdminUser | null;
  expiresAt: string | null;
  isHydrated: boolean;
  adminRoles: () => AdminRole[];
  hydrateSession: () => void;
  setSession: (session: AdminSession) => void;
  clearSession: () => void;
};

function readStoredSession(): AdminSession | null {
  if (typeof window === "undefined") {
    return null;
  }

  const rawSession = window.sessionStorage.getItem(SESSION_STORAGE_KEY);

  if (!rawSession) {
    return null;
  }

  try {
    const parsedSession = JSON.parse(rawSession) as Partial<AdminSession>;

    if (!parsedSession.accessToken || !parsedSession.user) {
      return null;
    }

    return {
      accessToken: parsedSession.accessToken,
      user: parsedSession.user,
      expiresAt: parsedSession.expiresAt
    };
  } catch {
    window.sessionStorage.removeItem(SESSION_STORAGE_KEY);
    return null;
  }
}

function writeStoredSession(session: AdminSession): void {
  if (typeof window === "undefined") {
    return;
  }

  window.sessionStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(session));
}

function clearStoredSession(): void {
  if (typeof window === "undefined") {
    return;
  }

  window.sessionStorage.removeItem(SESSION_STORAGE_KEY);
}

export const useAuthStore = create<AuthState>((set, get) => ({
  accessToken: null,
  user: null,
  expiresAt: null,
  isHydrated: false,
  adminRoles: () => normalizeAdminRoles(get().user?.roles ?? []),
  hydrateSession: () => {
    const storedSession = readStoredSession();

    if (!storedSession) {
      set({ accessToken: null, user: null, expiresAt: null, isHydrated: true });
      return;
    }

    set({
      accessToken: storedSession.accessToken,
      user: storedSession.user,
      expiresAt: storedSession.expiresAt ?? null,
      isHydrated: true
    });
  },
  setSession: (session) => {
    writeStoredSession(session);

    set({
      accessToken: session.accessToken,
      user: session.user,
      expiresAt: session.expiresAt ?? null,
      isHydrated: true
    });
  },
  clearSession: () => {
    clearStoredSession();
    set({ accessToken: null, user: null, expiresAt: null, isHydrated: true });
  }
}));

export function resetAuthStoreForTest(): void {
  clearStoredSession();
  useAuthStore.setState({
    accessToken: null,
    user: null,
    expiresAt: null,
    isHydrated: true
  });
}
