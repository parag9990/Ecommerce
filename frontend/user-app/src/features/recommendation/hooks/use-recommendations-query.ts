import { useQuery } from '@tanstack/react-query';

import { queryKeys } from '../../../lib/query-keys';
import {
  DEFAULT_RECOMMENDATION_LIMIT,
  getRecommendations,
  type RecommendationContextInput,
} from '../api/recommendation.grpc';

type UseRecommendationsQueryOptions = Readonly<{
  enabled?: boolean | undefined;
  limit?: number | undefined;
}>;

export function useRecommendationsQuery(
  context: RecommendationContextInput,
  options: UseRecommendationsQueryOptions = {},
) {
  const limit = options.limit ?? DEFAULT_RECOMMENDATION_LIMIT;

  return useQuery({
    enabled: options.enabled ?? true,
    queryFn: ({ signal }) => getRecommendations(context, { limit, signal }),
    queryKey: queryKeys.grpc.recommendations(context, limit),
    staleTime: 5 * 60 * 1000,
  });
}
