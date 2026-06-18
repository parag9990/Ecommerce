const CHECKOUT_KEY = 'user_app_checkout_idempotency_key';

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
    // The backend still enforces idempotency; storage only helps client retries.
  }
}

function removeSessionValue(key: string) {
  try {
    window.sessionStorage.removeItem(key);
  } catch {
    // Ignore storage failures.
  }
}

function createIdempotencyKey() {
  if (globalThis.crypto.randomUUID) {
    return globalThis.crypto.randomUUID();
  }

  return `checkout_${Date.now()}_${Math.random().toString(36).slice(2)}`;
}

export function getCheckoutIdempotencyKey() {
  const existing = readSessionValue(CHECKOUT_KEY);

  if (existing) {
    return existing;
  }

  const nextKey = createIdempotencyKey();
  writeSessionValue(CHECKOUT_KEY, nextKey);

  return nextKey;
}

export function resetCheckoutIdempotencyKey() {
  const nextKey = createIdempotencyKey();
  writeSessionValue(CHECKOUT_KEY, nextKey);

  return nextKey;
}

export function clearCheckoutIdempotencyKey() {
  removeSessionValue(CHECKOUT_KEY);
}
