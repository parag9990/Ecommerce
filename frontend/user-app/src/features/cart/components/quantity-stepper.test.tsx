import { cleanup, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, expect, test, vi } from 'vitest';

import { QuantityStepper } from './quantity-stepper';

afterEach(() => {
  cleanup();
});

test('keeps decrease disabled at quantity one', () => {
  render(<QuantityStepper onChange={() => undefined} value={1} />);

  expect(
    screen.getByRole('button', { name: 'Decrease quantity' }).hasAttribute('disabled'),
  ).toBe(true);
});

test('emits next quantity when increased', async () => {
  const user = userEvent.setup();
  const handleChange = vi.fn();

  render(<QuantityStepper onChange={handleChange} value={2} />);

  await user.click(screen.getByRole('button', { name: 'Increase quantity' }));

  expect(handleChange).toHaveBeenCalledWith(3);
});
