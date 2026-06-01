import { useQuery } from '@tanstack/react-query';

import { queryKeys } from '../../../lib/query-keys';
import {
  AUTOCOMPLETE_MIN_QUERY_LENGTH,
  DEFAULT_AUTOCOMPLETE_LIMIT,
  getAutocompleteSuggestions,
} from '../api/autocomplete.grpc';

export function useAutocompleteQuery(query: string, limit = DEFAULT_AUTOCOMPLETE_LIMIT) {
  const normalizedQuery = query.trim();

  return useQuery({
    enabled: normalizedQuery.length >= AUTOCOMPLETE_MIN_QUERY_LENGTH,
    queryFn: ({ signal }) =>
      getAutocompleteSuggestions(normalizedQuery, { limit, signal }),
    queryKey: queryKeys.grpc.autocomplete(normalizedQuery, limit),
    staleTime: 60 * 1000,
  });
}
