import { apiFetch } from "../../../lib/http";
import type {
  PlatformSetting,
  PlatformSettingInput,
  PlatformSettingsResponse,
  SettingKey
} from "../types";
import { PLATFORM_SETTING_KEYS } from "../types";

const PLATFORM_SETTINGS_PATH = "/api/v1/admin/settings";
const KNOWN_SETTING_KEYS = new Set<string>(PLATFORM_SETTING_KEYS);

type RawPlatformSetting = {
  key?: string | null;
  value?: unknown;
  updated_at?: string | null;
  version?: number | null;
};

type RawPlatformSettingsResponse = {
  settings?: RawPlatformSetting[] | null;
};

function isSettingKey(key: string): key is SettingKey {
  return KNOWN_SETTING_KEYS.has(key);
}

function normalizePlatformSetting(setting: RawPlatformSetting): PlatformSetting | null {
  const key = setting.key?.trim();

  if (!key || !isSettingKey(key)) {
    return null;
  }

  return {
    key,
    value: setting.value ?? {},
    updated_at: setting.updated_at ?? null,
    version: setting.version ?? null
  };
}

export async function listPlatformSettings(): Promise<PlatformSettingsResponse> {
  const response = await apiFetch<RawPlatformSettingsResponse>(PLATFORM_SETTINGS_PATH);

  return {
    settings: (response.settings ?? [])
      .map(normalizePlatformSetting)
      .filter((setting): setting is PlatformSetting => setting !== null)
  };
}

export async function updatePlatformSetting<TValue>(
  key: SettingKey,
  input: PlatformSettingInput<TValue>
): Promise<PlatformSetting<TValue>> {
  const response = await apiFetch<RawPlatformSetting>(
    `${PLATFORM_SETTINGS_PATH}/${encodeURIComponent(key)}`,
    {
      method: "PATCH",
      body: JSON.stringify(input)
    }
  );
  const normalized = normalizePlatformSetting({ ...response, key, value: response.value ?? input.value });

  return {
    ...(normalized ?? {
      key,
      value: input.value,
      updated_at: null,
      version: null
    }),
    value: (normalized?.value ?? input.value) as TValue
  };
}
