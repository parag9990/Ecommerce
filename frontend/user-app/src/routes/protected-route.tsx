import type { ReactNode } from 'react';

import { RequireAuth } from '../features/auth/components/require-auth';

type ProtectedRouteProps = {
  children: ReactNode;
  roles?: string[] | undefined;
};

export function ProtectedRoute({ children, roles }: ProtectedRouteProps) {
  return <RequireAuth roles={roles}>{children}</RequireAuth>;
}
