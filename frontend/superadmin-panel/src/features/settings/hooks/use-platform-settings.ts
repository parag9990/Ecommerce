import { useQuery } from "@tanstack/react-query";

import { listPlatformSettings } from "../api/settings-api";

export const PLATFORM_SETTINGS_QUERY_KEY = ["admin", "settings"] as const;

export function usePlatformSettings() {
  return useQuery({
    queryKey: PLATFORM_SETTINGS_QUERY_KEY,
    queryFn: listPlatformSettings,
    staleTime: 60_000
  });
}
