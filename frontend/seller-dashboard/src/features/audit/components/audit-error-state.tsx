import { FailedState } from "../../../components/state/failed-state";
import { getRequestId, getSafeErrorMessage } from "../../../lib/api-error";

type AuditErrorStateProps = {
  error: unknown;
  onRetry: () => void;
};

export function AuditErrorState({ error, onRetry }: AuditErrorStateProps) {
  return (
    <FailedState
      title="Audit activity could not be loaded"
      description={getSafeErrorMessage(error)}
      requestId={getRequestId(error)}
      onRetry={onRetry}
    />
  );
}
