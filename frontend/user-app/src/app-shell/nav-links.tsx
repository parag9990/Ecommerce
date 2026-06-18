import { NavLink } from 'react-router-dom';

import { routePaths } from '../routes/route-paths';

const navItems = [
  { label: 'Home', to: routePaths.home },
  { label: 'Categories', to: routePaths.categories },
  { label: 'Deals', to: routePaths.deals },
  { label: 'Wishlist', to: routePaths.wishlist },
] as const;

type NavLinksProps = {
  onNavigate?: () => void;
};

export function NavLinks({ onNavigate }: NavLinksProps) {
  return (
    <nav
      aria-label="Primary navigation"
      className="flex flex-col gap-1 md:flex-row md:items-center md:gap-2"
    >
      {navItems.map((item) => (
        <NavLink
          className={({ isActive }) =>
            [
              'rounded-md px-3 py-2 text-sm font-medium transition',
              'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600',
              isActive
                ? 'bg-blue-50 text-blue-700'
                : 'text-slate-700 hover:bg-slate-100 hover:text-slate-950',
            ].join(' ')
          }
          key={item.to}
          onClick={onNavigate}
          to={item.to}
        >
          {item.label}
        </NavLink>
      ))}
    </nav>
  );
}
