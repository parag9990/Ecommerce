const ACCESS_TOKEN_KEY = 'user_app_access_token';
const ACCESS_TOKEN_EXPIRES_AT_KEY = 'user_app_access_token_expires_at';
const AUTH_HINT_KEY = 'user_app_auth_hint';

type TokenLike = {
  access_token?: string | undefined;
  expires_in?: number | undefined;
};

type AuthSessionLike = {
  session_id?: string | undefined;
  tokens?: TokenLike | undefined;
  user?: unknown;
};

function readSessionValue(key: string): string | null {
  try {
    return window.sessionStorage.getItem(key);
  } catch {
    return null;
  }
}

function writeSessionValue(key: string, value: string) {
  try {
    window.sessionStorage.setItem(key, value);
  } catch {
    // Route guards still fall back to cookie-backed requests when storage is unavailable.
  }
}

function removeSessionValue(key: string) {
  try {
    window.sessionStorage.removeItem(key);
  } catch {
    // Ignore storage failures; the backend remains the source of truth.
  }
}

export function persistAuthSession(session: AuthSessionLike | undefined) {
  if (!session) {
    return;
  }

  if (session.tokens?.access_token) {
    writeSessionValue(ACCESS_TOKEN_KEY, session.tokens.access_token);

    if (session.tokens.expires_in && session.tokens.expires_in > 0) {
      const expiresAt = Date.now() + session.tokens.expires_in * 1000;
      writeSessionValue(ACCESS_TOKEN_EXPIRES_AT_KEY, String(expiresAt));
    }
  }

  if (session.session_id || session.user || session.tokens?.access_token) {
    writeSessionValue(AUTH_HINT_KEY, '1');
  }
}

export function getAccessToken(): string | undefined {
  const token = readSessionValue(ACCESS_TOKEN_KEY);

  if (!token) {
    return undefined;
  }

  const expiresAt = Number(readSessionValue(ACCESS_TOKEN_EXPIRES_AT_KEY));

  if (Number.isFinite(expiresAt) && expiresAt > 0 && expiresAt <= Date.now()) {
    removeSessionValue(ACCESS_TOKEN_KEY);
    removeSessionValue(ACCESS_TOKEN_EXPIRES_AT_KEY);
    return undefined;
  }

  return token;
}

export function hasAuthSession() {
  return Boolean(getAccessToken() ?? readSessionValue(AUTH_HINT_KEY));
}

export function clearAuthSession() {
  removeSessionValue(ACCESS_TOKEN_KEY);
  removeSessionValue(ACCESS_TOKEN_EXPIRES_AT_KEY);
  removeSessionValue(AUTH_HINT_KEY);
}
