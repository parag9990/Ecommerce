/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string;
  readonly VITE_APP_ENV?: string;
  readonly VITE_GRPC_WEB_BASE_URL?: string;
  readonly VITE_GRPC_WEB_TIMEOUT_MS?: string;
  readonly VITE_PAYMENT_PROVIDERS?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
