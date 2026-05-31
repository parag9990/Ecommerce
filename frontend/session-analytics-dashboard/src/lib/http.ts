import { appConfig, type AppConfig } from "./config";

type ApiEnvelope<T> = {
  data?: T;
  request_id?: string;
  requestId?: string;
  error?: {
    code?: string;
    message?: string;
    details?: unknown;
  };
};

type HttpOptions = {
  config?: AppConfig;
  headers?: HeadersInit;
  signal?: AbortSignal;
};

type JsonRequestOptions = HttpOptions & {
  body?: unknown;
  expectEmpty?: boolean;
  method?: "POST" | "PUT" | "PATCH" | "DELETE";
};

export type BlobResponse = {
  blob: Blob;
  filename?: string;
};

type ApiErrorParams = {
  code: string;
  message: string;
  status?: number;
  requestId?: string;
  details?: unknown;
  cause?: unknown;
};

export class ApiError extends Error {
  readonly code: string;
  readonly status?: number;
  readonly requestId?: string;
  readonly details?: unknown;

  constructor(params: ApiErrorParams) {
    super(params.message, { cause: params.cause });
    this.name = "ApiError";
    this.code = params.code;
    this.status = params.status;
    this.requestId = params.requestId;
    this.details = params.details;
  }
}

export async function getJSON<T>(path: string, options: HttpOptions = {}): Promise<T> {
  const config = options.config ?? appConfig;
  const requestId = createRequestId();
  const controller = new AbortController();
  const timeoutId = globalThis.setTimeout(
    () => controller.abort("request-timeout"),
    config.requestTimeoutMs
  );
  const removeAbortListener = bindAbortSignal(options.signal, controller);

  const headers = new Headers(options.headers);
  headers.set("Accept", "application/json");
  headers.set("X-Request-ID", requestId);

  try {
    const response = await fetch(buildRequestURL(path, config.apiBaseUrl), {
      credentials: "include",
      headers,
      method: "GET",
      signal: controller.signal
    });

    const payload = await readPayload(response);
    const responseRequestId = extractRequestId(payload, response, requestId);

    if (!response.ok) {
      throw buildHTTPError(response, payload, responseRequestId);
    }

    return unwrapPayload<T>(payload, responseRequestId);
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }

    if (controller.signal.aborted) {
      throw new ApiError({
        code: "REQUEST_TIMEOUT",
        message: "Analytics request timed out.",
        requestId,
        cause: error
      });
    }

    throw new ApiError({
      code: "NETWORK_ERROR",
      message: "Analytics service is unreachable.",
      requestId,
      cause: error
    });
  } finally {
    globalThis.clearTimeout(timeoutId);
    removeAbortListener();
  }
}

export async function sendJSON<T>(
  path: string,
  options: JsonRequestOptions = {}
): Promise<T> {
  const config = options.config ?? appConfig;
  const requestId = createRequestId();
  const controller = new AbortController();
  const timeoutId = globalThis.setTimeout(
    () => controller.abort("request-timeout"),
    config.requestTimeoutMs
  );
  const removeAbortListener = bindAbortSignal(options.signal, controller);

  const headers = new Headers(options.headers);
  headers.set("Accept", "application/json");
  headers.set("X-Request-ID", requestId);

  let body: BodyInit | undefined;
  if (options.body !== undefined) {
    headers.set("Content-Type", "application/json");
    body = JSON.stringify(options.body);
  }

  try {
    const response = await fetch(buildRequestURL(path, config.apiBaseUrl), {
      body,
      credentials: "include",
      headers,
      method: options.method ?? "POST",
      signal: controller.signal
    });

    const payload = await readPayload(response);
    const responseRequestId = extractRequestId(payload, response, requestId);

    if (!response.ok) {
      throw buildHTTPError(response, payload, responseRequestId);
    }

    if (payload === undefined && options.expectEmpty) {
      return undefined as T;
    }

    return unwrapPayload<T>(payload, responseRequestId);
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }

    if (controller.signal.aborted) {
      throw new ApiError({
        code: "REQUEST_TIMEOUT",
        message: "Analytics request timed out.",
        requestId,
        cause: error
      });
    }

    throw new ApiError({
      code: "NETWORK_ERROR",
      message: "Analytics service is unreachable.",
      requestId,
      cause: error
    });
  } finally {
    globalThis.clearTimeout(timeoutId);
    removeAbortListener();
  }
}

export async function getBlob(
  path: string,
  options: HttpOptions = {}
): Promise<BlobResponse> {
  const config = options.config ?? appConfig;
  const requestId = createRequestId();
  const controller = new AbortController();
  const timeoutId = globalThis.setTimeout(
    () => controller.abort("request-timeout"),
    config.requestTimeoutMs
  );
  const removeAbortListener = bindAbortSignal(options.signal, controller);

  const headers = new Headers(options.headers);
  headers.set("Accept", headers.get("Accept") ?? "*/*");
  headers.set("X-Request-ID", requestId);

  try {
    const response = await fetch(buildRequestURL(path, config.apiBaseUrl), {
      credentials: "include",
      headers,
      method: "GET",
      signal: controller.signal
    });

    if (!response.ok) {
      const payload = await readPayload(response);
      const responseRequestId = extractRequestId(payload, response, requestId);
      throw buildHTTPError(response, payload, responseRequestId);
    }

    return {
      blob: await response.blob(),
      filename: extractDownloadFilename(response.headers)
    };
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }

    if (controller.signal.aborted) {
      throw new ApiError({
        code: "REQUEST_TIMEOUT",
        message: "Analytics request timed out.",
        requestId,
        cause: error
      });
    }

    throw new ApiError({
      code: "NETWORK_ERROR",
      message: "Analytics service is unreachable.",
      requestId,
      cause: error
    });
  } finally {
    globalThis.clearTimeout(timeoutId);
    removeAbortListener();
  }
}

function buildRequestURL(path: string, apiBaseUrl: string): string {
  if (!apiBaseUrl) {
    return path;
  }

  return new URL(path, `${apiBaseUrl}/`).toString();
}

function bindAbortSignal(
  source: AbortSignal | undefined,
  controller: AbortController
): () => void {
  if (!source) {
    return () => undefined;
  }

  if (source.aborted) {
    controller.abort(source.reason);
    return () => undefined;
  }

  const abort = () => controller.abort(source.reason);
  source.addEventListener("abort", abort, { once: true });

  return () => source.removeEventListener("abort", abort);
}

async function readPayload(response: Response): Promise<unknown> {
  const text = await response.text();
  if (!text) {
    return undefined;
  }

  try {
    return JSON.parse(text) as unknown;
  } catch (error) {
    throw new ApiError({
      code: "INVALID_JSON",
      message: "Analytics service returned invalid JSON.",
      status: response.status,
      requestId: response.headers.get("x-request-id") ?? undefined,
      cause: error
    });
  }
}

function buildHTTPError(
  response: Response,
  payload: unknown,
  requestId: string
): ApiError {
  const envelope = isRecord(payload) ? (payload as ApiEnvelope<unknown>) : undefined;

  return new ApiError({
    code: envelope?.error?.code || `HTTP_${response.status}`,
    details: envelope?.error?.details,
    message:
      envelope?.error?.message ||
      "Analytics metrics could not be loaded. Please try again.",
    requestId,
    status: response.status
  });
}

function unwrapPayload<T>(payload: unknown, requestId: string): T {
  if (payload === undefined) {
    throw new ApiError({
      code: "EMPTY_RESPONSE",
      message: "Analytics service returned an empty response.",
      requestId
    });
  }

  if (isRecord(payload) && "error" in payload && payload.error) {
    const envelope = payload as ApiEnvelope<T>;
    throw new ApiError({
      code: envelope.error?.code || "API_ERROR",
      details: envelope.error?.details,
      message:
        envelope.error?.message ||
        "Analytics metrics could not be loaded. Please try again.",
      requestId
    });
  }

  if (isRecord(payload) && "data" in payload) {
    return (payload as ApiEnvelope<T>).data as T;
  }

  return payload as T;
}

function extractRequestId(
  payload: unknown,
  response: Response,
  fallback: string
): string {
  const headerRequestId = response.headers.get("x-request-id");
  if (headerRequestId) {
    return headerRequestId;
  }

  if (isRecord(payload)) {
    const envelope = payload as ApiEnvelope<unknown>;
    return envelope.request_id || envelope.requestId || fallback;
  }

  return fallback;
}

function createRequestId(): string {
  return globalThis.crypto?.randomUUID?.() || `web-${Date.now()}-${Math.random()}`;
}

function extractDownloadFilename(headers: Headers): string | undefined {
  const contentDisposition = headers.get("content-disposition");
  if (!contentDisposition) {
    return undefined;
  }

  const encodedMatch = contentDisposition.match(/filename\*=UTF-8''([^;]+)/i);
  if (encodedMatch?.[1]) {
    try {
      return sanitizeDownloadFilename(decodeURIComponent(encodedMatch[1]));
    } catch {
      return sanitizeDownloadFilename(encodedMatch[1]);
    }
  }

  const quotedMatch = contentDisposition.match(/filename="([^"]+)"/i);
  if (quotedMatch?.[1]) {
    return sanitizeDownloadFilename(quotedMatch[1]);
  }

  const plainMatch = contentDisposition.match(/filename=([^;]+)/i);
  if (plainMatch?.[1]) {
    return sanitizeDownloadFilename(plainMatch[1]);
  }

  return undefined;
}

function sanitizeDownloadFilename(value: string): string | undefined {
  const filename = value
    .trim()
    .replace(/[\\/:*?"<>|\u0000-\u001f]/g, "_")
    .slice(0, 160);

  return filename || undefined;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
