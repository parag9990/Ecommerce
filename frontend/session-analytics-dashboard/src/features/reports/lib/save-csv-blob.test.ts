import { describe, expect, it, vi } from "vitest";

import { saveCsvBlob } from "./save-csv-blob";

describe("saveCsvBlob", () => {
  it("creates an object URL, clicks a download link, and revokes the URL", () => {
    const click = vi.fn();
    const createObjectURL = vi.fn(() => "blob:report");
    const revokeObjectURL = vi.fn();
    const originalCreateElement = document.createElement.bind(document);
    const URLConstructor = globalThis.URL;
    const URLMock = function URL(value: string | URL, base?: string | URL) {
      return new URLConstructor(value, base);
    } as typeof globalThis.URL;
    Object.assign(URLMock, URLConstructor, {
      createObjectURL,
      revokeObjectURL
    });

    vi.stubGlobal("URL", URLMock);
    vi.spyOn(document, "createElement").mockImplementation((tagName) => {
      const element = originalCreateElement(tagName);
      if (tagName === "a") {
        Object.defineProperty(element, "click", {
          configurable: true,
          value: click
        });
      }

      return element;
    });

    saveCsvBlob(new Blob(["a,b\n1,2\n"], { type: "text/csv" }), "report.csv");

    expect(createObjectURL).toHaveBeenCalledOnce();
    expect(click).toHaveBeenCalledOnce();
    expect(revokeObjectURL).toHaveBeenCalledWith("blob:report");
  });
});
