import { Lock } from "lucide-react";

import { StateShell } from "./state-shell";
import type { StateAction } from "./state-types";

type PermissionDeniedStateProps = {
  title?: string;
  description?: string;
  action?: StateAction | null;
  className?: string;
};

export function PermissionDeniedState({
  title = "Access unavailable",
  description = "Aapke current seller role ke paas is section ka permission nahi hai.",
  action = { label: "Back to dashboard", href: "/seller" },
  className,
}: PermissionDeniedStateProps) {
  return (
    <StateShell
      kind="permissionDenied"
      title={title}
      description={description}
      icon={<Lock className="h-5 w-5" aria-hidden="true" />}
      action={action ?? undefined}
      className={className}
    />
  );
}
