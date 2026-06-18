import { useMutation, useQueryClient } from '@tanstack/react-query';

import { persistAuthSession } from '../../../lib/auth-session';
import { queryKeys } from '../../../lib/query-keys';
import { useAuthStore } from '../../../stores/auth-store';
import { login } from '../api/auth.api';
import type { AuthSessionResponse, LoginRequest } from '../types';

function storeAuthSession(
  session: AuthSessionResponse,
  setUser: ReturnType<typeof useAuthStore.getState>['setUser'],
) {
  persistAuthSession(session);

  if (!session.user?.user_id) {
    return;
  }

  setUser(
    {
      email: session.user.email,
      name: session.user.full_name,
      roles: session.user.roles ?? ['buyer'],
      userId: session.user.user_id,
    },
    session.session_id,
  );
}

export function useLoginMutation() {
  const queryClient = useQueryClient();
  const setUser = useAuthStore((state) => state.setUser);

  return useMutation({
    mutationFn: (body: LoginRequest) => login(body),
    onSuccess: (session) => {
      storeAuthSession(session, setUser);

      if (session.user) {
        queryClient.setQueryData(queryKeys.auth.me(), session.user);
        queryClient.setQueryData(queryKeys.profile.detail(), session.user);
      }

      void queryClient.invalidateQueries({ queryKey: queryKeys.cart.all });
      void queryClient.invalidateQueries({ queryKey: queryKeys.wishlist.all });
    },
  });
}
