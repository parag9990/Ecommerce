import { describe, expect, test } from 'vitest';

import { formatMoney } from './format-money';

describe('formatMoney', () => {
  test('formats minor-unit amounts', () => {
    expect(formatMoney({ amount: 12345, currency: 'INR' })).toBe('₹123.45');
  });

  test('falls back for missing money', () => {
    expect(formatMoney()).toBe('N/A');
  });
});
