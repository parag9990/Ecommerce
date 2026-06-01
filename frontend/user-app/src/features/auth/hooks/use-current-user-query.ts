import { useQuery } from '@tanstack/react-query';

import { queryKeys } from '../../../lib/query-keys';
import { getMyProfile } from '../../profile/api/profile.api';

export function useCurrentUserQuery(enabled = true) {
  return useQuery({
    enabled,
    queryFn: ({ signal }) => getMyProfile(signal),
    queryKey: queryKeys.auth.me(),
    retry: false,
    staleTime: 60 * 1000,
  });
}
