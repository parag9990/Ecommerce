import * as matchers from "@testing-library/jest-dom/matchers";
import { afterEach, beforeEach, expect, vi } from "vitest";

expect.extend(matchers);

const originalConsoleWarn = console.warn.bind(console);

if (!("text" in Blob.prototype)) {
  Object.defineProperty(Blob.prototype, "text", {
    configurable: true,
    value(this: Blob) {
      return new Promise<string>((resolve, reject) => {
        const reader = new FileReader();
        reader.onerror = () => reject(reader.error);
        reader.onload = () => resolve(String(reader.result ?? ""));
        reader.readAsText(this);
      });
    }
  });
}

beforeEach(() => {
  vi.spyOn(console, "warn").mockImplementation((...args: unknown[]) => {
    if (isExpectedAnalyticsRequestWarning(args[0])) {
      return;
    }

    originalConsoleWarn(...args);
  });
});

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

function isExpectedAnalyticsRequestWarning(value: unknown) {
  return (
    isRecord(value) &&
    value.level === "warn" &&
    typeof value.event === "string" &&
    value.event.startsWith("analytics.") &&
    value.event.endsWith(".request_failed")
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
