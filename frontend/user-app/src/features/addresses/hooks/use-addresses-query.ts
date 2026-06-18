import { useQuery } from '@tanstack/react-query';

import { queryKeys } from '../../../lib/query-keys';
import { listAddresses } from '../api/addresses.api';

export function useAddressesQuery() {
  return useQuery({
    queryFn: ({ signal }) => listAddresses(signal),
    queryKey: queryKeys.profile.addresses(),
    staleTime: 60 * 1000,
  });
}
