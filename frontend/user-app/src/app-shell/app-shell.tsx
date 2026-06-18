import { Outlet } from 'react-router-dom';

import { Header } from './header';
import { MobileNav } from './mobile-nav';

export function AppShell() {
  return (
    <div className="min-h-screen bg-slate-50 text-slate-950">
      <Header />
      <MobileNav />

      <main className="mx-auto w-full max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
        <Outlet />
      </main>
    </div>
  );
}
