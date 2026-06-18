import { Bell, Heart, MapPin, Package, UserRound } from 'lucide-react';
import { Link } from 'react-router-dom';

import { routePaths } from '../../../routes/route-paths';

const overviewLinks = [
  {
    description: 'Review your name, phone, and account identity.',
    icon: UserRound,
    label: 'Profile',
    to: routePaths.profile,
  },
  {
    description: 'Add, edit, delete, and choose default delivery addresses.',
    icon: MapPin,
    label: 'Addresses',
    to: routePaths.addresses,
  },
  {
    description: 'Check order history, details, and eligible cancellations.',
    icon: Package,
    label: 'Orders',
    to: routePaths.orders,
  },
  {
    description: 'Move saved products to your cart or remove them.',
    icon: Heart,
    label: 'Wishlist',
    to: routePaths.wishlist,
  },
  {
    description: 'Control notification and marketing preferences.',
    icon: Bell,
    label: 'Notifications',
    to: routePaths.notificationPreferences,
  },
] as const;

export function AccountOverviewPage() {
  return (
    <section className="space-y-5">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-slate-950">
          Account
        </h1>
        <p className="mt-1 text-sm text-slate-600">
          Manage the details tied to your buyer account.
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        {overviewLinks.map((link) => {
          const Icon = link.icon;

          return (
            <Link
              className="rounded-md border border-slate-200 bg-white p-4 transition hover:border-slate-300 hover:shadow-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
              key={link.to}
              to={link.to}
            >
              <Icon aria-hidden="true" className="h-5 w-5 text-blue-700" />
              <h2 className="mt-3 font-semibold text-slate-950">
                {link.label}
              </h2>
              <p className="mt-1 text-sm leading-6 text-slate-600">
                {link.description}
              </p>
            </Link>
          );
        })}
      </div>
    </section>
  );
}
