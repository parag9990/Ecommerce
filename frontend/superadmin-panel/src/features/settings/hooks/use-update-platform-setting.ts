import { useMutation, useQueryClient } from "@tanstack/react-query";

import { updatePlatformSetting } from "../api/settings-api";
import { PLATFORM_SETTINGS_QUERY_KEY } from "./use-platform-settings";
import type { PlatformSettingInput, SettingKey } from "../types";

type UpdatePlatformSettingMutationInput<TValue> = {
  key: SettingKey;
  input: PlatformSettingInput<TValue>;
};

export function useUpdatePlatformSetting<TValue>() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ key, input }: UpdatePlatformSettingMutationInput<TValue>) =>
      updatePlatformSetting(key, input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: PLATFORM_SETTINGS_QUERY_KEY });
    }
  });
}
