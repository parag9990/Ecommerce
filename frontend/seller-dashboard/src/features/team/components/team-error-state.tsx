import { FailedState } from "../../../components/state/failed-state";
import { getRequestId, getSafeErrorMessage } from "../../../lib/api-error";

type TeamErrorStateProps = {
  error: unknown;
  onRetry: () => void;
};

export function TeamErrorState({ error, onRetry }: TeamErrorStateProps) {
  return (
    <FailedState
      title="Team members could not be loaded"
      description={getSafeErrorMessage(error)}
      requestId={getRequestId(error)}
      onRetry={onRetry}
    />
  );
}
