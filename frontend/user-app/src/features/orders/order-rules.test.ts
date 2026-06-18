import { describe, expect, test } from 'vitest';

import { canCancelOrder, formatOrderStatus } from './order-rules';

describe('order rules', () => {
  test('allows cancellation only before fulfilment is too far along', () => {
    expect(canCancelOrder('created')).toBe(true);
    expect(canCancelOrder('packed')).toBe(true);
    expect(canCancelOrder('shipped')).toBe(false);
    expect(canCancelOrder('delivered')).toBe(false);
  });

  test('formats statuses for display', () => {
    expect(formatOrderStatus('pending_payment')).toBe('pending payment');
    expect(formatOrderStatus()).toBe('unknown');
  });
});
