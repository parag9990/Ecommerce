import type { ProductFilters } from '../features/product/types';

export const queryKeys = {
  auth: {
    all: ['auth'] as const,
    me: () => [...queryKeys.auth.all, 'me'] as const,
  },
  cart: {
    all: ['cart'] as const,
    detail: () => [...queryKeys.cart.all, 'detail'] as const,
  },
  grpc: {
    all: ['grpc'] as const,
    autocomplete: (query: string, limit: number) =>
      [...queryKeys.grpc.all, 'search', 'autocomplete', query.trim(), limit] as const,
    recommendations: (context: unknown, limit: number) =>
      [...queryKeys.grpc.all, 'recommendations', context, limit] as const,
  },
  orders: {
    all: ['orders'] as const,
    detail: (orderId: string) =>
      [...queryKeys.orders.all, 'detail', orderId] as const,
    list: (page: number, pageSize: number) =>
      [...queryKeys.orders.all, 'list', { page, pageSize }] as const,
  },
  products: {
    all: ['products'] as const,
    autocomplete: (q: string, limit: number) =>
      [...queryKeys.products.all, 'autocomplete', q.trim(), limit] as const,
    categories: () => [...queryKeys.products.all, 'categories'] as const,
    detail: (productId: string) =>
      [...queryKeys.products.all, 'detail', productId] as const,
    list: (filters: ProductFilters, mode: 'home' | 'search') =>
      [...queryKeys.products.lists(), mode, filters] as const,
    lists: () => [...queryKeys.products.all, 'list'] as const,
  },
  profile: {
    addresses: () => [...queryKeys.profile.all, 'addresses'] as const,
    all: ['profile'] as const,
    detail: () => [...queryKeys.profile.all, 'detail'] as const,
  },
  wishlist: {
    all: ['wishlist'] as const,
    detail: () => [...queryKeys.wishlist.all, 'detail'] as const,
  },
} as const;
