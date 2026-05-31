import {
  Activity,
  BarChart3,
  Download,
  Filter,
  Gauge,
  MousePointerClick,
  Repeat,
  Route,
  ShieldCheck
} from "lucide-react";
import { NavLink, Outlet } from "react-router-dom";

const navItems = [
  {
    href: "/",
    icon: Gauge,
    label: "Overview"
  },
  {
    href: "/live",
    icon: Activity,
    label: "Live Sessions"
  },
  {
    href: "/journey",
    icon: Route,
    label: "Journey Explorer"
  },
  {
    href: "/funnels",
    icon: Filter,
    label: "Funnels"
  },
  {
    href: "/heatmaps",
    icon: MousePointerClick,
    label: "Heatmaps"
  },
  {
    href: "/cohorts",
    icon: Repeat,
    label: "Cohorts"
  },
  {
    href: "/reports",
    icon: Download,
    label: "Reports"
  },
  {
    href: "/privacy",
    icon: ShieldCheck,
    label: "Privacy"
  }
];

export function AnalyticsLayout() {
  return (
    <div className="min-h-screen bg-zinc-50 text-zinc-950">
      <aside className="fixed inset-y-0 left-0 hidden w-64 border-r border-zinc-200 bg-white lg:block">
        <div className="border-b border-zinc-200 px-5 py-4">
          <div className="flex items-center gap-3">
            <div className="flex h-9 w-9 items-center justify-center rounded-md bg-zinc-950 text-white">
              <BarChart3 className="h-5 w-5" aria-hidden="true" />
            </div>
            <div>
              <p className="text-xs font-semibold uppercase text-zinc-500">
                Analytics
              </p>
              <h1 className="text-base font-semibold">Session Dashboard</h1>
            </div>
          </div>
        </div>

        <nav aria-label="Session analytics" className="space-y-1 p-3">
          {navItems.map((item) => {
            const Icon = item.icon;

            return (
              <NavLink
                end={item.href === "/"}
                key={item.href}
                to={item.href}
                className={({ isActive }) =>
                  [
                    "flex h-10 items-center gap-3 rounded-md px-3 text-sm font-medium transition-colors",
                    isActive
                      ? "bg-zinc-950 text-white"
                      : "text-zinc-700 hover:bg-zinc-100 hover:text-zinc-950"
                  ].join(" ")
                }
              >
                <Icon className="h-4 w-4" aria-hidden="true" />
                {item.label}
              </NavLink>
            );
          })}
        </nav>
      </aside>

      <div className="border-b border-zinc-200 bg-white px-4 py-3 lg:hidden">
        <div className="flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-md bg-zinc-950 text-white">
            <BarChart3 className="h-5 w-5" aria-hidden="true" />
          </div>
          <div>
            <p className="text-xs font-semibold uppercase text-zinc-500">
              Analytics
            </p>
            <h1 className="text-base font-semibold">Session Dashboard</h1>
          </div>
        </div>

        <nav
          aria-label="Session analytics mobile"
          className="mt-3 flex gap-2 overflow-x-auto"
        >
          {navItems.map((item) => {
            const Icon = item.icon;

            return (
              <NavLink
                end={item.href === "/"}
                key={item.href}
                to={item.href}
                className={({ isActive }) =>
                  [
                    "inline-flex h-9 shrink-0 items-center gap-2 rounded-md px-3 text-sm font-medium transition-colors",
                    isActive
                      ? "bg-zinc-950 text-white"
                      : "bg-zinc-100 text-zinc-700 hover:bg-zinc-200 hover:text-zinc-950"
                  ].join(" ")
                }
              >
                <Icon className="h-4 w-4" aria-hidden="true" />
                {item.label}
              </NavLink>
            );
          })}
        </nav>
      </div>

      <main className="min-h-screen lg:pl-64">
        <Outlet />
      </main>
    </div>
  );
}
