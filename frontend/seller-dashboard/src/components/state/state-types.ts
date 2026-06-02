import type { ReactNode } from "react";

export type DashboardStateKind =
  | "loading"
  | "empty"
  | "failed"
  | "permissionDenied"
  | "unavailable";

export type StateAction = {
  label: string;
  onClick?: () => void;
  href?: string;
  variant?: "primary" | "secondary";
  disabled?: boolean;
  title?: string;
};

export type StateCopy = {
  title: string;
  description: string;
  icon?: ReactNode;
  action?: StateAction;
  secondaryAction?: StateAction;
  requestId?: string;
};
