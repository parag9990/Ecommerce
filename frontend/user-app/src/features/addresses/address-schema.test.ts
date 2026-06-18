import { describe, expect, test } from 'vitest';

import { addressSchema } from './address-schema';

describe('addressSchema', () => {
  test('accepts a complete address and rejects incomplete address line', () => {
    expect(
      addressSchema.safeParse({
        city: 'Mumbai',
        country: 'India',
        is_default: true,
        line1: '221 Market Street',
        line2: '',
        name: 'Buyer User',
        phone: '+919999999999',
        postal_code: '400001',
        state: 'Maharashtra',
      }).success,
    ).toBe(true);

    expect(
      addressSchema.safeParse({
        city: 'A',
        country: 'India',
        is_default: false,
        line1: '12',
        line2: '',
        name: 'B',
        phone: '',
        postal_code: '1',
        state: 'M',
      }).success,
    ).toBe(false);
  });
});
