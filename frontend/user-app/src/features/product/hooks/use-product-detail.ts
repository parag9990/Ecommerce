import { useProductDetailQuery } from './use-product-detail-query';

export function useProductDetail(productId: string | undefined) {
  const query = useProductDetailQuery(productId);

  if (!productId) {
    return {
      error: 'Product id is missing.',
      isLoading: false,
      product: undefined,
    };
  }

  return {
    error:
      query.error instanceof Error
        ? query.error.message
        : query.isError
          ? 'Product details could not be loaded.'
          : undefined,
    isLoading: query.isLoading,
    product: query.data,
  };
}
