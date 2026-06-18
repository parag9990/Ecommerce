import { BarChart3 } from "lucide-react";

import { StateShell } from "./state-shell";
import type { StateAction } from "./state-types";

type UnavailableStateProps = {
  title?: string;
  description?: string;
  action?: StateAction;
  className?: string;
};

export function UnavailableState({
  title = "Data unavailable",
  description = "Is data source ke liye backend response abhi available nahi hai. Available data visible rahega.",
  action,
  className,
}: UnavailableStateProps) {
  return (
    <StateShell
      kind="unavailable"
      title={title}
      description={description}
      icon={<BarChart3 className="h-5 w-5" aria-hidden="true" />}
      action={action}
      className={className}
    />
  );
}
