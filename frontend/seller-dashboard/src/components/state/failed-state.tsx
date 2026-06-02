import { AlertTriangle } from "lucide-react";

import { StateShell } from "./state-shell";
import type { StateAction } from "./state-types";

type FailedStateProps = {
  title?: string;
  description?: string;
  requestId?: string;
  onRetry?: () => void;
  action?: StateAction;
  secondaryAction?: StateAction;
  className?: string;
};

export function FailedState({
  title = "Something went wrong",
  description = "Data load nahi ho paya. Connection check karke retry karein.",
  requestId,
  onRetry,
  action,
  secondaryAction,
  className,
}: FailedStateProps) {
  return (
    <StateShell
      kind="failed"
      title={title}
      description={description}
      requestId={requestId}
      icon={<AlertTriangle className="h-5 w-5" aria-hidden="true" />}
      action={action ?? (onRetry ? { label: "Retry", onClick: onRetry } : undefined)}
      secondaryAction={secondaryAction}
      className={className}
    />
  );
}
