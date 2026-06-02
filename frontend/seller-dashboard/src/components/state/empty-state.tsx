import { Inbox } from "lucide-react";

import { StateShell } from "./state-shell";
import type { StateAction } from "./state-types";

type EmptyStateProps = {
  title: string;
  description: string;
  action?: StateAction;
  secondaryAction?: StateAction;
  className?: string;
};

export function EmptyState({
  title,
  description,
  action,
  secondaryAction,
  className,
}: EmptyStateProps) {
  return (
    <StateShell
      kind="empty"
      title={title}
      description={description}
      icon={<Inbox className="h-5 w-5" aria-hidden="true" />}
      action={action}
      secondaryAction={secondaryAction}
      className={className}
    />
  );
}
