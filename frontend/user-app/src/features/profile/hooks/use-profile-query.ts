import { useQuery } from '@tanstack/react-query';

import { queryKeys } from '../../../lib/query-keys';
import { getMyProfile } from '../api/profile.api';

export function useProfileQuery() {
  return useQuery({
    queryFn: ({ signal }) => getMyProfile(signal),
    queryKey: queryKeys.profile.detail(),
    staleTime: 60 * 1000,
  });
}
