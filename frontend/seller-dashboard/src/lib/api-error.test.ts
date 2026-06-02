import { describe, expect, it } from "vitest";

import {
  AppError,
  getRequestId,
  getSafeErrorMessage,
  isPermissionError,
} from "./api-error";

describe("api error helpers", () => {
  it("maps permission and server errors to safe copy", () => {
    expect(
      getSafeErrorMessage(
        new AppError({ status: 403, code: "FORBIDDEN", message: "raw forbidden" }),
      ),
    ).toBe("Is section ka access aapke role me nahi hai.");

    expect(
      getSafeErrorMessage(
        new AppError({ status: 500, code: "PANIC", message: "stack trace" }),
      ),
    ).toBe("Server side issue aa gaya. Retry karein.");
  });

  it("identifies permission errors and exposes request ids", () => {
    const error = new AppError({
      status: 401,
      code: "UNAUTHORIZED",
      message: "expired",
      requestId: "req_123",
    });

    expect(isPermissionError(error)).toBe(true);
    expect(getRequestId(error)).toBe("req_123");
  });

  it("maps network and unknown errors", () => {
    expect(
      getSafeErrorMessage(
        new AppError({ status: 0, code: "NETWORK_ERROR", message: "fetch failed" }),
      ),
    ).toBe("Network connection issue lag raha hai.");
    expect(getSafeErrorMessage(new Error("boom"))).toBe(
      "Unexpected error aa gaya. Retry karein.",
    );
  });
});
