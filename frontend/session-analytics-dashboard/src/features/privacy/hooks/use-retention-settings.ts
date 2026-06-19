import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import {
  getRetentionSettings,
  updateRetentionSettings,
  type RetentionSettings
} from "../../../api/session-api";

export function useRetentionSettings() {
  return useQuery({
    queryFn: getRetentionSettings,
    queryKey: ["analytics", "privacy", "retention"],
    staleTime: 60_000
  });
}

export function useUpdateRetentionSettings() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      reason,
      settings
    }: {
      reason: string;
      settings: RetentionSettings;
    }) => updateRetentionSettings(settings, reason),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: ["analytics", "privacy", "retention"]
      });
    }
  });
}
