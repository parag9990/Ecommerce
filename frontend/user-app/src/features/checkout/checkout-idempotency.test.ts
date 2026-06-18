import { afterEach, describe, expect, test } from 'vitest';

import {
  clearCheckoutIdempotencyKey,
  getCheckoutIdempotencyKey,
  resetCheckoutIdempotencyKey,
} from './checkout-idempotency';

afterEach(() => {
  window.sessionStorage.clear();
});

describe('checkout idempotency', () => {
  test('reuses a key for the same checkout attempt', () => {
    const firstKey = getCheckoutIdempotencyKey();
    const secondKey = getCheckoutIdempotencyKey();

    expect(firstKey).toBe(secondKey);
  });

  test('can reset and clear the stored key', () => {
    const firstKey = getCheckoutIdempotencyKey();
    const resetKey = resetCheckoutIdempotencyKey();

    expect(resetKey).not.toBe(firstKey);
    expect(getCheckoutIdempotencyKey()).toBe(resetKey);

    clearCheckoutIdempotencyKey();

    expect(getCheckoutIdempotencyKey()).not.toBe(resetKey);
  });
});
