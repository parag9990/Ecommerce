import type { ReactNode } from "react";

import {
  AppError,
  getSafeErrorMessage,
  isPermissionError,
} from "../../lib/api-error";
import { FailedState } from "./failed-state";
import { PermissionDeniedState } from "./permission-denied-state";
import { RefreshingNotice } from "./refreshing-notice";

type AsyncStateBoundaryProps<TData> = {
  isLoading: boolean;
  isError: boolean;
  error: unknown;
  data: TData | undefined;
  isEmpty: (data: TData) => boolean;
  loadingFallback: ReactNode;
  emptyFallback: ReactNode;
  children: (data: TData) => ReactNode;
  onRetry?: () => void;
};

export function AsyncStateBoundary<TData>({
  isLoading,
  isError,
  error,
  data,
  isEmpty,
  loadingFallback,
  emptyFallback,
  children,
  onRetry,
}: AsyncStateBoundaryProps<TData>) {
  if (isLoading && !data) {
    return <>{loadingFallback}</>;
  }

  if (isError) {
    if (data && !isEmpty(data)) {
      return (
        <div className="space-y-3">
          <RefreshingNotice
            show
            failed
            message={getSafeErrorMessage(error)}
            onRetry={onRetry}
          />
          {children(data)}
        </div>
      );
    }

    if (isPermissionError(error)) {
      return <PermissionDeniedState description={getSafeErrorMessage(error)} />;
    }

    return (
      <FailedState
        description={getSafeErrorMessage(error)}
        requestId={error instanceof AppError ? error.requestId : undefined}
        onRetry={onRetry}
      />
    );
  }

  if (!data || isEmpty(data)) {
    return <>{emptyFallback}</>;
  }

  return <>{children(data)}</>;
}
