import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, test, vi } from 'vitest';

import { WishlistItemCard } from './wishlist-item-card';

describe('WishlistItemCard', () => {
  test('disables move to cart for unavailable products', () => {
    render(
      <MemoryRouter>
        <WishlistItemCard
          item={{
            availability: 'out_of_stock',
            product_id: 'product-1',
            title: 'Saved product',
          }}
          onMoveToCart={vi.fn()}
          onRemove={vi.fn()}
        />
      </MemoryRouter>,
    );

    const moveButton = screen.getByRole('button', { name: /move to cart/i });

    expect((moveButton as HTMLButtonElement).disabled).toBe(true);
  });
});
