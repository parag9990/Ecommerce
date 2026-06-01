import { useMutation, useQueryClient } from '@tanstack/react-query';

import { queryKeys } from '../../../lib/query-keys';
import { useAuthStore } from '../../../stores/auth-store';
import { updateMyProfile } from '../api/profile.api';
import type { UpdateUserProfileInput } from '../types';

export function useProfileMutation() {
  const queryClient = useQueryClient();
  const setUser = useAuthStore((state) => state.setUser);

  return useMutation({
    mutationFn: (body: UpdateUserProfileInput) => updateMyProfile(body),
    onSuccess: (profile) => {
      queryClient.setQueryData(queryKeys.profile.detail(), profile);
      queryClient.setQueryData(queryKeys.auth.me(), profile);

      if (profile.user_id) {
        setUser({
          email: profile.email,
          name: profile.full_name,
          roles: profile.roles ?? ['buyer'],
          userId: profile.user_id,
        });
      }
    },
  });
}
