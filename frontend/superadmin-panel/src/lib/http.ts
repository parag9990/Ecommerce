import { useAuthStore } from "../stores/auth-store";

const DEFAULT_TIMEOUT_MS = 15_000;
const DEFAULT_API_BASE_URL = "http://localhost:8080";

export type ApiErrorBody = {
  code?: string;
  message?: string;
  request_id?: string;
  details?: unknown;
  error?: {
    code?: string;
    message?: string;
    details?: unknown;
  };
};

export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly requestId: string | null;
  readonly details: unknown;

  constructor(message: string, options: { status: number; code?: string; requestId?: string; details?: unknown }) {
    super(message);
    this.name = "ApiError";
    this.status = options.status;
    this.code = options.code ?? "api_error";
    this.requestId = options.requestId ?? null;
    this.details = options.details;
  }
}

function buildApiUrl(path: string): string {
  const baseUrl = import.meta.env.VITE_API_BASE_URL || DEFAULT_API_BASE_URL;

  return `${baseUrl.replace(/\/$/, "")}/${path.replace(/^\/+/, "")}`;
}

async function parseErrorBody(response: Response): Promise<ApiErrorBody | null> {
  const contentType = response.headers.get("content-type");

  if (!contentType?.includes("application/json")) {
    return null;
  }

  try {
    const payload = (await response.json()) as ApiErrorBody;

    if (payload.error) {
      return {
        code: payload.error.code,
        message: payload.error.message,
        request_id: payload.request_id,
        details: payload.error.details
      };
    }

    return payload;
  } catch {
    return null;
  }
}

export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const controller = new AbortController();
  const timeout = window.setTimeout(() => controller.abort(), DEFAULT_TIMEOUT_MS);
  const token = useAuthStore.getState().accessToken;
  const requestId = crypto.randomUUID();

  try {
    const response = await fetch(buildApiUrl(path), {
      ...init,
      signal: init.signal ?? controller.signal,
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        "X-Request-ID": requestId,
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...init.headers
      }
    });

    if (!response.ok) {
      const errorBody = await parseErrorBody(response);

      throw new ApiError(errorBody?.message ?? `API request failed with status ${response.status}`, {
        status: response.status,
        code: errorBody?.code,
        requestId: errorBody?.request_id ?? requestId,
        details: errorBody?.details
      });
    }

    if (response.status === 204) {
      return undefined as T;
    }

    return (await response.json()) as T;
  } finally {
    window.clearTimeout(timeout);
  }
}

export async function apiFetchBlob(path: string, init: RequestInit = {}): Promise<Blob> {
  const controller = new AbortController();
  const timeout = window.setTimeout(() => controller.abort(), DEFAULT_TIMEOUT_MS);
  const token = useAuthStore.getState().accessToken;
  const requestId = crypto.randomUUID();

  try {
    const response = await fetch(buildApiUrl(path), {
      ...init,
      signal: init.signal ?? controller.signal,
      headers: {
        Accept: "text/csv,application/octet-stream,*/*",
        "Content-Type": "application/json",
        "X-Request-ID": requestId,
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...init.headers
      }
    });

    if (!response.ok) {
      const errorBody = await parseErrorBody(response);

      throw new ApiError(errorBody?.message ?? `API request failed with status ${response.status}`, {
        status: response.status,
        code: errorBody?.code,
        requestId: errorBody?.request_id ?? requestId,
        details: errorBody?.details
      });
    }

    return response.blob();
  } finally {
    window.clearTimeout(timeout);
  }
}
