import { cleanup, render, screen } from '@testing-library/react';
import { afterEach, expect, test } from 'vitest';
import { MemoryRouter } from 'react-router-dom';

import { createTestQueryWrapper } from '../../../test/create-test-query-wrapper';
import { ProductCard } from './product-card';

afterEach(() => {
  cleanup();
});

test('renders product title, price, stock, and detail link', () => {
  const TestQueryWrapper = createTestQueryWrapper();

  render(
    <TestQueryWrapper>
      <MemoryRouter>
        <ProductCard
          product={{
            brand: 'Acme',
            product_id: 'prod_1',
            title: 'Running Shoes',
            variants: [
              {
                price: { amount: 1999, currency: 'INR' },
                sku: 'RUN-BLK-8',
                stock_quantity: 8,
              },
            ],
          }}
        />
      </MemoryRouter>
    </TestQueryWrapper>,
  );

  const link = screen.getByRole('link', { name: /running shoes/i });

  expect(link.getAttribute('href')).toBe('/products/prod_1');
  expect(screen.getByText('Running Shoes')).toBeTruthy();
  expect(screen.getByText('In stock')).toBeTruthy();
  expect(screen.getByText(/1,999/)).toBeTruthy();
});
