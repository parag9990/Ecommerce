export type AppConfig = {
  apiBaseUrl: string;
  requestTimeoutMs: number;
};

const defaultRequestTimeoutMs = 10_000;

export function loadAppConfig(env: ImportMetaEnv = import.meta.env): AppConfig {
  return {
    apiBaseUrl: trimTrailingSlash(env.VITE_API_BASE_URL || ""),
    requestTimeoutMs: readPositiveInteger(
      env.VITE_REQUEST_TIMEOUT_MS,
      defaultRequestTimeoutMs
    )
  };
}

function readPositiveInteger(value: string | undefined, fallback: number): number {
  if (!value) {
    return fallback;
  }

  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback;
}

function trimTrailingSlash(value: string): string {
  return value.endsWith("/") ? value.slice(0, -1) : value;
}

export const appConfig = loadAppConfig();
