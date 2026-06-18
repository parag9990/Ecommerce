import { useCategoriesQuery } from './use-categories-query';

export function useCategories() {
  const query = useCategoriesQuery();

  return {
    categories: query.data?.categories ?? [],
    error:
      query.error instanceof Error
        ? query.error.message
        : query.isError
          ? 'Categories could not be loaded.'
          : undefined,
    isLoading: query.isLoading,
  };
}
