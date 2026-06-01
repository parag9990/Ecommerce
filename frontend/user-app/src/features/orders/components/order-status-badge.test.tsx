import { render, screen } from '@testing-library/react';
import { describe, expect, test } from 'vitest';

import { OrderStatusBadge } from './order-status-badge';

describe('OrderStatusBadge', () => {
  test('renders readable status text', () => {
    render(<OrderStatusBadge status="pending_payment" />);

    expect(screen.getByText('pending payment')).toBeTruthy();
  });
});
