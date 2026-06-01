import type { ReactNode } from 'react';
import { useEffect } from 'react';
import { Navigate, useLocation } from 'react-router-dom';

import { clearAuthSession, persistAuthSession } from '../../../lib/auth-session';
import { routePaths } from '../../../routes/route-paths';
import { useAuthStore } from '../../../stores/auth-store';
import { useCurrentUserQuery } from '../hooks/use-current-user-query';

type RequireAuthProps = {
  children: ReactNode;
  roles?: string[] | undefined;
};

export function RequireAuth({ children, roles }: RequireAuthProps) {
  const location = useLocation();
  const setUser = useAuthStore((state) => state.setUser);
  const clearUser = useAuthStore((state) => state.clearUser);
  const currentUserQuery = useCurrentUserQuery();

  useEffect(() => {
    const user = currentUserQuery.data;

    if (!user) {
      return;
    }

    persistAuthSession({ user });

    if (user.user_id) {
      setUser({
        email: user.email,
        name: user.full_name,
        roles: user.roles ?? ['buyer'],
        userId: user.user_id,
      });
    }
  }, [currentUserQuery.data, setUser]);

  useEffect(() => {
    if (currentUserQuery.isError) {
      clearAuthSession();
      clearUser();
    }
  }, [clearUser, currentUserQuery.isError]);

  if (currentUserQuery.isLoading) {
    return (
      <div className="rounded-md border border-slate-200 bg-white p-4 text-sm text-slate-600">
        Checking session...
      </div>
    );
  }

  if (currentUserQuery.isError || !currentUserQuery.data) {
    const redirectTo = encodeURIComponent(location.pathname + location.search);

    return <Navigate replace to={`${routePaths.login}?redirect=${redirectTo}`} />;
  }

  if (
    roles?.length &&
    !roles.some((role) => currentUserQuery.data.roles?.includes(role))
  ) {
    return (
      <div className="rounded-md border border-slate-200 bg-white p-4 text-sm text-slate-600">
        You do not have permission to open this page.
      </div>
    );
  }

  return <>{children}</>;
}
