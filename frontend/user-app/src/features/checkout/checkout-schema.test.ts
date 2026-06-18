import { describe, expect, test } from 'vitest';

import { checkoutSchema } from './checkout-schema';

describe('checkout schema', () => {
  test('requires address and payment provider', () => {
    const result = checkoutSchema.safeParse({
      address_id: '',
      coupon_code: '',
      payment_provider: '',
    });

    expect(result.success).toBe(false);
  });

  test('accepts a valid checkout form', () => {
    const result = checkoutSchema.safeParse({
      address_id: 'addr_123',
      coupon_code: 'SAVE10',
      payment_provider: 'stripe',
    });

    expect(result.success).toBe(true);
  });
});
