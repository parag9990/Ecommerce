import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import {
  getPrivacySettings,
  updatePrivacySettings,
  type PrivacyMaskingSettings
} from "../../../api/session-api";

export function usePrivacySettings() {
  return useQuery({
    queryFn: getPrivacySettings,
    queryKey: ["analytics", "privacy", "settings"],
    staleTime: 60_000
  });
}

export function useUpdatePrivacySettings() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (masking: PrivacyMaskingSettings) =>
      updatePrivacySettings(masking),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: ["analytics", "privacy", "settings"]
      });
    }
  });
}
