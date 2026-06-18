import { Outlet } from 'react-router-dom';

import { AccountNav } from './account-nav';

export function AccountLayout() {
  return (
    <div className="grid gap-6 lg:grid-cols-[14rem_minmax(0,1fr)]">
      <aside className="lg:sticky lg:top-24 lg:self-start">
        <div className="rounded-md border border-slate-200 bg-white p-3">
          <AccountNav />
        </div>
      </aside>

      <div className="min-w-0">
        <Outlet />
      </div>
    </div>
  );
}
