import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type AuthUser = {
  email?: string | undefined;
  name?: string | undefined;
  roles: string[];
  userId: string;
};

type AuthState = {
  clearUser: () => void;
  isHydrated: boolean;
  markHydrated: () => void;
  sessionId?: string | undefined;
  setUser: (user: AuthUser, sessionId?: string) => void;
  user?: AuthUser | undefined;
};

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      clearUser: () => {
        set({ sessionId: undefined, user: undefined });
      },
      isHydrated: false,
      markHydrated: () => {
        set({ isHydrated: true });
      },
      setUser: (user, sessionId) => {
        set({ sessionId, user });
      },
    }),
    {
      name: 'user-app-auth',
      onRehydrateStorage: () => (state) => {
        state?.markHydrated();
      },
      partialize: (state) => ({
        sessionId: state.sessionId,
        user: state.user,
      }),
    },
  ),
);
