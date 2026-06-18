import { LogOut, User } from 'lucide-react';
import { Link } from 'react-router-dom';

import { useLogoutMutation } from '../features/auth/hooks/use-logout-mutation';
import { routePaths } from '../routes/route-paths';
import { useAuthStore } from '../stores/auth-store';
import { useUiStore } from '../stores/ui-store';

const accountLinkClasses =
  'block rounded px-3 py-2 text-sm text-slate-700 transition hover:bg-slate-100 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600';

export function AccountMenu() {
  const user = useAuthStore((state) => state.user);
  const isOpen = useUiStore((state) => state.isAccountMenuOpen);
  const setOpen = useUiStore((state) => state.setAccountMenuOpen);
  const logoutMutation = useLogoutMutation();

  function closeMenu() {
    setOpen(false);
  }

  return (
    <div className="relative">
      <button
        aria-expanded={isOpen}
        className="flex h-10 items-center gap-2 rounded-md px-3 text-sm font-medium text-slate-700 transition hover:bg-slate-100 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
        onClick={() => {
          setOpen(!isOpen);
        }}
        type="button"
      >
        <User aria-hidden="true" className="h-5 w-5" />
        <span className="hidden max-w-32 truncate lg:inline">
          {user?.name ?? 'Account'}
        </span>
      </button>

      {isOpen ? (
        <div className="absolute right-0 z-30 mt-2 w-56 rounded-md border border-slate-200 bg-white p-2 shadow-lg">
          {user ? (
            <>
              <Link
                className={accountLinkClasses}
                onClick={closeMenu}
                to={routePaths.profile}
              >
                Profile
              </Link>
              <Link
                className={accountLinkClasses}
                onClick={closeMenu}
                to={routePaths.addresses}
              >
                Addresses
              </Link>
              <Link
                className={accountLinkClasses}
                onClick={closeMenu}
                to={routePaths.orders}
              >
                Orders
              </Link>
              <Link
                className={accountLinkClasses}
                onClick={closeMenu}
                to={routePaths.notificationPreferences}
              >
                Notifications
              </Link>
              <button
                className="flex w-full items-center rounded px-3 py-2 text-left text-sm text-red-700 transition hover:bg-red-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-red-600 disabled:cursor-not-allowed disabled:opacity-60"
                disabled={logoutMutation.isPending}
                onClick={() => {
                  logoutMutation.mutate();
                  closeMenu();
                }}
                type="button"
              >
                <LogOut aria-hidden="true" className="mr-2 h-4 w-4" />
                Logout
              </button>
            </>
          ) : (
            <>
              <Link
                className={accountLinkClasses}
                onClick={closeMenu}
                to={routePaths.login}
              >
                Login
              </Link>
              <Link
                className={accountLinkClasses}
                onClick={closeMenu}
                to={routePaths.signup}
              >
                Create account
              </Link>
            </>
          )}
        </div>
      ) : null}
    </div>
  );
}
