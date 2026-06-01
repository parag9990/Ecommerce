import { describe, expect, test } from 'vitest';

import { readProductFilters, writeProductFilters } from './product-url-state';

describe('product URL state', () => {
  test('reads supported filters from URL params', () => {
    const params = new URLSearchParams(
      'q=shoes&category_id=cat_1&brand=Nike&seller_id=seller_1&min_price=100&max_price=5000&min_rating=4&in_stock=true&page=2&page_size=48&sort=price%3Aasc',
    );

    expect(readProductFilters(params)).toEqual({
      brand: 'Nike',
      categoryId: 'cat_1',
      inStock: true,
      maxPrice: '5000',
      minPrice: '100',
      minRating: '4',
      page: 2,
      pageSize: 48,
      q: 'shoes',
      sellerId: 'seller_1',
      sort: 'price:asc',
    });
  });

  test('falls back from invalid page and sort params', () => {
    const params = new URLSearchParams('page=-2&page_size=120&sort=unknown');

    expect(readProductFilters(params)).toMatchObject({
      page: 1,
      pageSize: 60,
      sort: undefined,
    });
  });

  test('writes clean URL params without default pagination', () => {
    const params = writeProductFilters({
      page: 1,
      pageSize: 24,
      q: 'laptop',
      sort: 'price:asc',
    });

    expect(params.toString()).toBe('q=laptop&sort=price%3Aasc');
  });
});
