import { env } from './env';
import { getAccessToken } from './auth-session';

type ParsedApiError = {
  code: string;
  message: string;
  requestId?: string;
  details?: unknown;
};

type JsonRequestOptions = {
  auth?: boolean | undefined;
  signal?: AbortSignal | undefined;
};

export class ApiError extends Error {
  readonly code: string;
  readonly status: number;
  readonly requestId?: string;
  readonly details?: unknown;

  constructor(
    message: string,
    code: string,
    status: number,
    requestId?: string,
    details?: unknown,
  ) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
    this.status = status;

    if (requestId) {
      this.requestId = requestId;
    }

    if (details !== undefined) {
      this.details = details;
    }
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

function readString(value: unknown): string | undefined {
  return typeof value === 'string' && value.trim().length > 0
    ? value
    : undefined;
}

function parseApiError(payload: unknown): ParsedApiError | undefined {
  if (!isRecord(payload)) {
    return undefined;
  }

  const requestId = readString(payload.request_id);
  const rawError = payload.error;

  if (isRecord(rawError)) {
    const parsedError: ParsedApiError = {
      code: readString(rawError.code) ?? 'REQUEST_FAILED',
      message: readString(rawError.message) ?? 'Request failed. Please try again.',
    };

    if (requestId) {
      parsedError.requestId = requestId;
    }

    if ('details' in rawError) {
      parsedError.details = rawError.details;
    }

    return parsedError;
  }

  const directMessage = readString(payload.message);

  if (!directMessage) {
    return undefined;
  }

  const parsedError: ParsedApiError = {
    code: readString(payload.code) ?? 'REQUEST_FAILED',
    message: directMessage,
  };

  if (requestId) {
    parsedError.requestId = requestId;
  }

  if ('details' in payload) {
    parsedError.details = payload.details;
  }

  return parsedError;
}

function unwrapResponseData<TResponse>(payload: unknown): TResponse {
  if (isRecord(payload) && 'data' in payload) {
    if (payload.data === undefined || payload.data === null) {
      throw new ApiError(
        'The server returned an empty response.',
        'EMPTY_RESPONSE',
        200,
      );
    }

    return payload.data as TResponse;
  }

  if (payload === undefined || payload === null) {
    throw new ApiError(
      'The server returned an empty response.',
      'EMPTY_RESPONSE',
      200,
    );
  }

  return payload as TResponse;
}

async function readJsonPayload(response: Response): Promise<unknown> {
  const bodyText = await response.text();

  if (!bodyText) {
    return undefined;
  }

  try {
    return JSON.parse(bodyText) as unknown;
  } catch {
    if (response.ok) {
      throw new ApiError(
        'The server returned an invalid response.',
        'INVALID_JSON',
        response.status,
      );
    }

    return undefined;
  }
}

function buildApiUrl(path: string): string {
  const baseUrl = env.apiBaseUrl.replace(/\/+$/, '');
  const normalizedPath = path.startsWith('/') ? path : `/${path}`;

  return `${baseUrl}${normalizedPath}`;
}

async function requestJson<TResponse>(
  path: string,
  requestInit: RequestInit,
): Promise<TResponse> {
  let response: Response;

  try {
    response = await fetch(buildApiUrl(path), requestInit);
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') {
      throw error;
    }

    throw new ApiError(
      'Network error. Please check your connection and try again.',
      'NETWORK_ERROR',
      0,
    );
  }

  const payload = await readJsonPayload(response);
  const parsedError = parseApiError(payload);

  if (!response.ok || parsedError) {
    throw new ApiError(
      parsedError?.message ?? 'Request failed. Please try again.',
      parsedError?.code ?? 'REQUEST_FAILED',
      response.status,
      parsedError?.requestId,
      parsedError?.details,
    );
  }

  return unwrapResponseData<TResponse>(payload);
}

function buildHeaders(
  options: JsonRequestOptions,
  headers: Record<string, string> = {},
) {
  const token = options.auth ? getAccessToken() : undefined;

  return {
    Accept: 'application/json',
    'X-Request-Source': 'user-app',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...headers,
  };
}

export function apiGet<TResponse>(
  path: string,
  options: JsonRequestOptions = {},
): Promise<TResponse> {
  const requestInit: RequestInit = {
    credentials: 'include',
    headers: buildHeaders(options),
    method: 'GET',
  };

  if (options.signal) {
    requestInit.signal = options.signal;
  }

  return requestJson<TResponse>(path, requestInit);
}

export async function postJson<TResponse, TBody>(
  path: string,
  body: TBody,
  options: JsonRequestOptions = {},
): Promise<TResponse> {
  const requestInit: RequestInit = {
    body: JSON.stringify(body),
    credentials: 'include',
    headers: buildHeaders(options, {
      'Content-Type': 'application/json',
    }),
    method: 'POST',
  };

  if (options.signal) {
    requestInit.signal = options.signal;
  }

  return requestJson<TResponse>(path, requestInit);
}

export const apiPost = postJson;

export async function apiPatch<TResponse, TBody>(
  path: string,
  body: TBody,
  options: JsonRequestOptions = {},
): Promise<TResponse> {
  const requestInit: RequestInit = {
    body: JSON.stringify(body),
    credentials: 'include',
    headers: buildHeaders(options, {
      'Content-Type': 'application/json',
    }),
    method: 'PATCH',
  };

  if (options.signal) {
    requestInit.signal = options.signal;
  }

  return requestJson<TResponse>(path, requestInit);
}

export function apiDelete<TResponse>(
  path: string,
  options: JsonRequestOptions = {},
): Promise<TResponse> {
  const requestInit: RequestInit = {
    credentials: 'include',
    headers: buildHeaders(options),
    method: 'DELETE',
  };

  if (options.signal) {
    requestInit.signal = options.signal;
  }

  return requestJson<TResponse>(path, requestInit);
}
