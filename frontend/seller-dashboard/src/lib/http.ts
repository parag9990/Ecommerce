import { AppError } from "./api-error";

const DEFAULT_API_BASE_URL = "http://localhost:8080";
const DEFAULT_TIMEOUT_MS = 15000;

export { ApiError, AppError } from "./api-error";

type ApiEnvelope<T> = {
  data?: T;
  request_id?: string;
  error?: {
    code?: string;
    message?: string;
    details?: unknown;
  };
};

export type HttpOptions = RequestInit & {
  timeoutMs?: number;
};

function getApiBaseUrl() {
  return import.meta.env.VITE_API_BASE_URL ?? DEFAULT_API_BASE_URL;
}

function getTimeoutMs(timeoutMs?: number) {
  const configuredTimeout = Number(import.meta.env.VITE_API_TIMEOUT_MS);

  if (timeoutMs && timeoutMs > 0) {
    return timeoutMs;
  }

  if (Number.isFinite(configuredTimeout) && configuredTimeout > 0) {
    return configuredTimeout;
  }

  return DEFAULT_TIMEOUT_MS;
}

function buildUrl(path: string) {
  if (/^https?:\/\//i.test(path)) {
    return path;
  }

  const baseUrl = getApiBaseUrl().replace(/\/+$/, "");
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;

  return `${baseUrl}${normalizedPath}`;
}

function createRequestId() {
  if (globalThis.crypto?.randomUUID) {
    return globalThis.crypto.randomUUID();
  }

  return `seller-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

async function readResponseBody<T>(response: Response): Promise<ApiEnvelope<T>> {
  const contentType = response.headers.get("content-type") ?? "";

  if (!contentType.includes("application/json")) {
    return {};
  }

  return (await response.json()) as ApiEnvelope<T>;
}

export async function http<T>(path: string, options: HttpOptions = {}): Promise<T> {
  const controller = new AbortController();
  const timeout = window.setTimeout(() => controller.abort(), getTimeoutMs(options.timeoutMs));
  const headers = new Headers(options.headers);

  if (!headers.has("content-type") && options.body) {
    headers.set("content-type", "application/json");
  }

  if (!headers.has("x-client-app")) {
    headers.set("x-client-app", "seller-dashboard");
  }

  if (!headers.has("x-request-id")) {
    headers.set("x-request-id", createRequestId());
  }

  const requestId = headers.get("x-request-id") ?? undefined;

  try {
    const response = await fetch(buildUrl(path), {
      ...options,
      credentials: options.credentials ?? "include",
      headers,
      signal: options.signal ?? controller.signal,
    });

    const body = await readResponseBody<T>(response);

    if (!response.ok || body.error) {
      throw new AppError({
        message: body.error?.message ?? `Request failed with status ${response.status}`,
        status: response.status,
        code: body.error?.code,
        requestId: body.request_id ?? requestId,
        details: body.error?.details,
      });
    }

    if ("data" in body) {
      return body.data as T;
    }

    return body as T;
  } catch (error) {
    if (error instanceof AppError) {
      throw error;
    }

    if (error instanceof DOMException && error.name === "AbortError") {
      throw new AppError({
        message: "Request timed out",
        status: 0,
        code: "REQUEST_TIMEOUT",
        requestId,
      });
    }

    throw new AppError({
      message: error instanceof Error ? error.message : "Network request failed",
      status: 0,
      code: "NETWORK_ERROR",
      requestId,
      details: error,
    });
  } finally {
    window.clearTimeout(timeout);
  }
}
