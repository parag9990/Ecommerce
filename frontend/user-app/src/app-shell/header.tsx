import { Link } from 'react-router-dom';

import { routePaths } from '../routes/route-paths';
import { AccountMenu } from './account-menu';
import { CartBadge } from './cart-badge';
import { NavLinks } from './nav-links';
import { SearchBox } from './search-box';

export function Header() {
  return (
    <header className="sticky top-0 z-20 border-b border-slate-200 bg-white/95 backdrop-blur">
      <div className="mx-auto flex min-h-16 w-full max-w-7xl items-center gap-4 px-4 sm:px-6 lg:px-8">
        <Link
          className="shrink-0 rounded-md text-xl font-bold tracking-tight text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
          to={routePaths.home}
        >
          Ecom
        </Link>

        <div className="hidden md:block">
          <NavLinks />
        </div>

        <div className="ml-auto hidden flex-1 justify-center md:flex">
          <SearchBox />
        </div>

        <div className="ml-auto flex items-center gap-1 md:ml-0">
          <AccountMenu />
          <CartBadge />
        </div>
      </div>

      <div className="border-t border-slate-100 px-4 py-3 md:hidden">
        <SearchBox />
      </div>
    </header>
  );
}
