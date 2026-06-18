import { create } from 'zustand';
import { persist } from 'zustand/middleware';

type SessionSnapshot = {
  sessionId: string;
  startedAt: string;
};

type SessionState = SessionSnapshot & {
  anonymousId: string;
  resetSession: () => void;
};

function createSession(): SessionSnapshot {
  return {
    sessionId: crypto.randomUUID(),
    startedAt: new Date().toISOString(),
  };
}

export const useSessionStore = create<SessionState>()(
  persist(
    (set) => ({
      anonymousId: crypto.randomUUID(),
      ...createSession(),
      resetSession: () => {
        set(createSession());
      },
    }),
    {
      name: 'user-app-session',
    },
  ),
);
