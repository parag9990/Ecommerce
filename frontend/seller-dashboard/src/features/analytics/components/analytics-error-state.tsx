import { FailedState } from "../../../components/state/failed-state";
import { getRequestId, getSafeErrorMessage } from "../../../lib/api-error";

type AnalyticsErrorStateProps = {
  error: unknown;
  onRetry: () => void;
};

export function AnalyticsErrorState({ error, onRetry }: AnalyticsErrorStateProps) {
  return (
    <FailedState
      title="Analytics could not be loaded"
      description={getSafeErrorMessage(error)}
      requestId={getRequestId(error)}
      onRetry={onRetry}
    />
  );
}
