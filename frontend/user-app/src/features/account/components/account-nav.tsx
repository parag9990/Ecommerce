import { Bell, Heart, MapPin, Package, UserRound } from 'lucide-react';
import { NavLink } from 'react-router-dom';

import { routePaths } from '../../../routes/route-paths';

const accountLinks = [
  { icon: UserRound, label: 'Profile', to: routePaths.profile },
  { icon: MapPin, label: 'Addresses', to: routePaths.addresses },
  { icon: Package, label: 'Orders', to: routePaths.orders },
  { icon: Heart, label: 'Wishlist', to: routePaths.wishlist },
  { icon: Bell, label: 'Notifications', to: routePaths.notificationPreferences },
] as const;

export function AccountNav() {
  return (
    <nav
      aria-label="Account navigation"
      className="flex gap-2 overflow-x-auto pb-1 lg:grid lg:overflow-visible lg:pb-0"
    >
      {accountLinks.map((link) => {
        const Icon = link.icon;

        return (
          <NavLink
            className={({ isActive }) =>
              [
                'inline-flex min-h-10 shrink-0 items-center gap-2 rounded-md px-3 text-sm font-semibold transition',
                'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600',
                isActive
                  ? 'bg-blue-600 text-white'
                  : 'text-slate-700 hover:bg-slate-100 hover:text-slate-950',
              ].join(' ')
            }
            key={link.to}
            to={link.to}
          >
            <Icon aria-hidden="true" className="h-4 w-4" />
            {link.label}
          </NavLink>
        );
      })}
    </nav>
  );
}
