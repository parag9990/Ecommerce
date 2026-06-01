import { Menu, X } from 'lucide-react';

import { useUiStore } from '../stores/ui-store';
import { NavLinks } from './nav-links';

export function MobileNav() {
  const isOpen = useUiStore((state) => state.isMobileNavOpen);
  const setOpen = useUiStore((state) => state.setMobileNavOpen);

  return (
    <div className="border-b border-slate-200 bg-white md:hidden">
      <button
        aria-controls="mobile-navigation"
        aria-expanded={isOpen}
        className="flex w-full items-center justify-between px-4 py-3 text-sm font-medium text-slate-700 transition hover:bg-slate-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-inset focus-visible:outline-blue-600"
        onClick={() => {
          setOpen(!isOpen);
        }}
        type="button"
      >
        <span>Browse</span>
        {isOpen ? (
          <X aria-hidden="true" className="h-5 w-5" />
        ) : (
          <Menu aria-hidden="true" className="h-5 w-5" />
        )}
      </button>

      {isOpen ? (
        <div className="px-4 pb-4" id="mobile-navigation">
          <NavLinks
            onNavigate={() => {
              setOpen(false);
            }}
          />
        </div>
      ) : null}
    </div>
  );
}
