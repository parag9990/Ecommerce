export type ApiErrorPayload = {
  code?: string;
  message?: string;
  details?: unknown;
};

export class AppError extends Error {
  readonly status: number;
  readonly code: string;
  readonly requestId?: string;
  readonly details?: unknown;

  constructor(params: {
    message: string;
    status: number;
    code?: string;
    requestId?: string;
    details?: unknown;
  }) {
    super(params.message);
    this.name = "AppError";
    this.status = params.status;
    this.code = params.code ?? "UNKNOWN_ERROR";
    this.requestId = params.requestId;
    this.details = params.details;
  }
}

export function getSafeErrorMessage(error: unknown) {
  if (!(error instanceof AppError)) {
    return "Unexpected error aa gaya. Retry karein.";
  }

  if (error.status === 0) {
    return error.code === "REQUEST_TIMEOUT"
      ? "Request time out ho gaya. Retry karein."
      : "Network connection issue lag raha hai.";
  }

  if (error.status === 401) {
    return "Session expire ho gayi. Login dobara karein.";
  }

  if (error.status === 403) {
    return "Is section ka access aapke role me nahi hai.";
  }

  if (error.status === 404) {
    return "Requested record nahi mila.";
  }

  if (error.status === 429) {
    return "Too many requests. Thoda wait karke retry karein.";
  }

  if (error.status >= 500) {
    return "Server side issue aa gaya. Retry karein.";
  }

  return error.message || "Request complete nahi ho payi.";
}

export function isPermissionError(error: unknown) {
  return error instanceof AppError && (error.status === 401 || error.status === 403);
}

export function getRequestId(error: unknown) {
  return error instanceof AppError ? error.requestId : undefined;
}

export { AppError as ApiError };
