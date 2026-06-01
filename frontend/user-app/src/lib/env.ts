const APP_ENVS = ['local', 'development', 'staging', 'production'] as const;

export type AppEnv = (typeof APP_ENVS)[number];

export type PublicEnv = Readonly<{
  apiBaseUrl: string;
  appEnv: AppEnv;
  grpcWebBaseUrl: string;
  grpcWebTimeoutMs: number;
  paymentProviders: readonly string[];
}>;

const DEFAULT_API_BASE_URL = 'http://localhost:8080';
const DEFAULT_APP_ENV: AppEnv = 'local';
const DEFAULT_GRPC_WEB_BASE_URL = 'http://localhost:8082';
const DEFAULT_GRPC_WEB_TIMEOUT_MS = 5000;
const DEFAULT_PAYMENT_PROVIDERS = ['stripe', 'razorpay'] as const;

function readOptionalEnv(value: string | undefined): string | undefined {
  const trimmedValue = value?.trim();

  return trimmedValue && trimmedValue.length > 0 ? trimmedValue : undefined;
}

function isAppEnv(value: string): value is AppEnv {
  return APP_ENVS.includes(value as AppEnv);
}

function readPaymentProviders(value: string | undefined): readonly string[] {
  const providers = readOptionalEnv(value)
    ?.split(',')
    .map((provider) => provider.trim().toLowerCase())
    .filter((provider) => provider.length > 0);

  return providers && providers.length > 0
    ? Array.from(new Set(providers))
    : DEFAULT_PAYMENT_PROVIDERS;
}

function readPositiveIntegerEnv(
  value: string | undefined,
  fallback: number,
): number {
  const numericValue = Number(readOptionalEnv(value));

  return Number.isInteger(numericValue) && numericValue > 0
    ? numericValue
    : fallback;
}

function readEnv(): PublicEnv {
  const apiBaseUrl =
    readOptionalEnv(import.meta.env.VITE_API_BASE_URL) ?? DEFAULT_API_BASE_URL;
  const appEnv = readOptionalEnv(import.meta.env.VITE_APP_ENV) ?? DEFAULT_APP_ENV;
  const grpcWebBaseUrl =
    readOptionalEnv(import.meta.env.VITE_GRPC_WEB_BASE_URL) ??
    DEFAULT_GRPC_WEB_BASE_URL;
  const grpcWebTimeoutMs = readPositiveIntegerEnv(
    import.meta.env.VITE_GRPC_WEB_TIMEOUT_MS,
    DEFAULT_GRPC_WEB_TIMEOUT_MS,
  );
  const paymentProviders = readPaymentProviders(
    import.meta.env.VITE_PAYMENT_PROVIDERS,
  );

  if (!isAppEnv(appEnv)) {
    throw new Error(
      `Invalid VITE_APP_ENV "${appEnv}". Expected one of: ${APP_ENVS.join(', ')}.`,
    );
  }

  return {
    apiBaseUrl,
    appEnv,
    grpcWebBaseUrl,
    grpcWebTimeoutMs,
    paymentProviders,
  };
}

export const env = Object.freeze(readEnv());
