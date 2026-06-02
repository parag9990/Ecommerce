import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";

import { AppError } from "../../lib/api-error";
import { AsyncStateBoundary } from "./async-state-boundary";

type Data = {
  rows: string[];
};

function renderBoundary(props: Partial<Parameters<typeof AsyncStateBoundary<Data>>[0]>) {
  const defaultProps: Parameters<typeof AsyncStateBoundary<Data>>[0] = {
    isLoading: false,
    isError: false,
    error: null,
    data: { rows: ["row-1"] },
    isEmpty: (data) => data.rows.length === 0,
    loadingFallback: <div>Loading rows</div>,
    emptyFallback: <div>No rows</div>,
    children: (data) => <div>{data.rows.join(", ")}</div>,
  };

  return render(
    <MemoryRouter>
      <AsyncStateBoundary {...defaultProps} {...props} />
    </MemoryRouter>,
  );
}

describe("AsyncStateBoundary", () => {
  it("renders loading, empty, and success states", () => {
    const { rerender } = renderBoundary({
      isLoading: true,
      data: undefined,
    });
    expect(screen.getByText("Loading rows")).toBeTruthy();

    rerender(
      <MemoryRouter>
        <AsyncStateBoundary<Data>
          isLoading={false}
          isError={false}
          error={null}
          data={{ rows: [] }}
          isEmpty={(data) => data.rows.length === 0}
          loadingFallback={<div>Loading rows</div>}
          emptyFallback={<div>No rows</div>}
          children={(data) => <div>{data.rows.join(", ")}</div>}
        />
      </MemoryRouter>,
    );
    expect(screen.getByText("No rows")).toBeTruthy();

    rerender(
      <MemoryRouter>
        <AsyncStateBoundary<Data>
          isLoading={false}
          isError={false}
          error={null}
          data={{ rows: ["row-2"] }}
          isEmpty={(data) => data.rows.length === 0}
          loadingFallback={<div>Loading rows</div>}
          emptyFallback={<div>No rows</div>}
          children={(data) => <div>{data.rows.join(", ")}</div>}
        />
      </MemoryRouter>,
    );
    expect(screen.getByText("row-2")).toBeTruthy();
  });

  it("renders failed and permission states", () => {
    const retry = vi.fn();
    const { rerender } = renderBoundary({
      isError: true,
      error: new AppError({
        status: 500,
        code: "SERVER_ERROR",
        message: "panic",
        requestId: "req_failed",
      }),
      data: undefined,
      onRetry: retry,
    });

    expect(screen.getByText("Server side issue aa gaya. Retry karein.")).toBeTruthy();
    expect(screen.getByText("Request ID: req_failed")).toBeTruthy();

    rerender(
      <MemoryRouter>
        <AsyncStateBoundary<Data>
          isLoading={false}
          isError
          error={new AppError({
            status: 403,
            code: "FORBIDDEN",
            message: "forbidden",
          })}
          data={undefined}
          isEmpty={(data) => data.rows.length === 0}
          loadingFallback={<div>Loading rows</div>}
          emptyFallback={<div>No rows</div>}
          children={(data) => <div>{data.rows.join(", ")}</div>}
        />
      </MemoryRouter>,
    );

    expect(screen.getByText("Access unavailable")).toBeTruthy();
  });

  it("keeps stale data visible when a background refresh fails", async () => {
    const user = userEvent.setup();
    const retry = vi.fn();
    renderBoundary({
      isError: true,
      error: new AppError({ status: 0, code: "NETWORK_ERROR", message: "failed" }),
      data: { rows: ["stale-row"] },
      onRetry: retry,
    });

    expect(screen.getByText("stale-row")).toBeTruthy();
    await user.click(screen.getByRole("button", { name: "Retry" }));
    expect(retry).toHaveBeenCalled();
  });
});
