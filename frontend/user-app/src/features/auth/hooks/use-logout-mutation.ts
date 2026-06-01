import { useMutation, useQueryClient } from '@tanstack/react-query';

import { clearAuthSession } from '../../../lib/auth-session';
import { queryKeys } from '../../../lib/query-keys';
import { useAuthStore } from '../../../stores/auth-store';
import { logout } from '../api/auth.api';

export function useLogoutMutation() {
  const queryClient = useQueryClient();
  const clearUser = useAuthStore((state) => state.clearUser);

  return useMutation({
    mutationFn: logout,
    onSettled: () => {
      clearAuthSession();
      clearUser();
      queryClient.removeQueries({ queryKey: queryKeys.auth.all });
      queryClient.removeQueries({ queryKey: queryKeys.cart.all });
      queryClient.removeQueries({ queryKey: queryKeys.wishlist.all });
      queryClient.removeQueries({ queryKey: queryKeys.profile.all });
      queryClient.removeQueries({ queryKey: queryKeys.orders.all });
    },
  });
}
