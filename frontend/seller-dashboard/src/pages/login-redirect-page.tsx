import { useEffect } from "react";
import { useLocation } from "react-router-dom";

export function LoginRedirectPage() {
  const location = useLocation();
  const configuredLoginUrl = import.meta.env.VITE_LOGIN_URL;
  const fromPath = typeof location.state?.from?.pathname === "string"
    ? location.state.from.pathname
    : "/seller";

  useEffect(() => {
    if (configuredLoginUrl) {
      const nextUrl = new URL(configuredLoginUrl, window.location.origin);
      nextUrl.searchParams.set("next", fromPath);
      window.location.assign(nextUrl.toString());
    }
  }, [configuredLoginUrl, fromPath]);

  return (
    <main className="flex min-h-screen items-center justify-center bg-slate-100 p-6">
      <section className="w-full max-w-md rounded-md border border-slate-200 bg-white p-6 text-center shadow-sm">
        <p className="text-xs font-semibold uppercase tracking-wide text-blue-700">
          Authentication required
        </p>
        <h1 className="mt-2 text-lg font-semibold text-slate-950">
          Sign in required
        </h1>
        <p className="mt-2 text-sm leading-6 text-slate-500">
          Sign in with an active seller account to continue to the seller
          dashboard.
        </p>
      </section>
    </main>
  );
}
