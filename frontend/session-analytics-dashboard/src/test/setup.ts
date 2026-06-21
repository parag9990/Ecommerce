import * as matchers from "@testing-library/jest-dom/matchers";
import { afterEach, expect, vi } from "vitest";

expect.extend(matchers);

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

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});
