import { describe, expect, test } from 'vitest';

import { formatMoney } from './price';

describe('formatMoney', () => {
  test('formats minor-unit money values', () => {
    expect(formatMoney({ amount: 199900, currency: 'INR' })).toMatch(/1,999/);
  });

  test('falls back to zero INR when money is missing', () => {
    expect(formatMoney()).toMatch(/0\.00/);
  });
});
