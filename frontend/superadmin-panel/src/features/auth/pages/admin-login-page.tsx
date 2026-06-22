import { FormEvent, useState } from "react";
import { Navigate, useLocation, useNavigate } from "react-router-dom";

import { apiFetch } from "../../../lib/http";
import { isAdminUser } from "../../../lib/admin-rbac";
import { RouteLoader } from "../../../components/ui/route-loader";
import { useAuthStore, type AdminSession, type AdminUser } from "../../../stores/auth-store";

type LoginResponse = {
  user?: {
    user_id?: string;
    id?: string;
    email?: string;
    full_name?: string;
    name?: string;
    roles?: string[];
  };
  tokens?: {
    access_token?: string;
    expires_in?: number;
  };
};

function toAdminSession(response: LoginResponse): AdminSession {
  const accessToken = response.tokens?.access_token;
  const user = response.user;

  if (!accessToken || !user?.roles?.length) {
    throw new Error("Login response did not include a valid admin session.");
  }

  const adminUser: AdminUser = {
    id: user.user_id ?? user.id ?? user.email ?? "admin",
    email: user.email ?? "",
    name: user.full_name ?? user.name ?? user.email ?? "Admin",
    roles: user.roles
  };

  if (!isAdminUser(adminUser.roles)) {
    throw new Error("This account does not have access to the Superadmin Panel.");
  }

  return {
    accessToken,
    user: adminUser,
    expiresAt: response.tokens?.expires_in
      ? new Date(Date.now() + response.tokens.expires_in * 1000).toISOString()
      : undefined
  };
}

export function AdminLoginPage() {
  const [identifier, setIdentifier] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const setSession = useAuthStore((state) => state.setSession);
  const accessToken = useAuthStore((state) => state.accessToken);
  const user = useAuthStore((state) => state.user);
  const isHydrated = useAuthStore((state) => state.isHydrated);
  const navigate = useNavigate();
  const location = useLocation();

  if (!isHydrated) {
    return <RouteLoader label="Checking admin session" />;
  }

  if (isHydrated && accessToken && user && isAdminUser(user.roles)) {
    return <Navigate to="/admin" replace />;
  }

  const from = (location.state as { from?: { pathname?: string } } | null)?.from?.pathname ?? "/admin";

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);

    if (!identifier.trim() || password.length < 8) {
      setError("Enter your admin identifier and a valid password.");
      return;
    }

    setIsSubmitting(true);

    try {
      const response = await apiFetch<LoginResponse>("/api/v1/auth/login", {
        method: "POST",
        body: JSON.stringify({
          identifier: identifier.trim(),
          password,
          device: {
            user_agent: window.navigator.userAgent,
            source: "superadmin-panel"
          }
        })
      });

      setSession(toAdminSession(response));
      navigate(from, { replace: true });
    } catch (loginError) {
      setError(loginError instanceof Error ? loginError.message : "Admin login failed.");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <main className="flex min-h-screen items-center justify-center bg-slate-50 px-4 py-8">
      <section className="w-full max-w-sm rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
        <div className="mb-5">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            Superadmin Panel
          </p>
          <h1 className="mt-1 text-xl font-semibold text-slate-950">Admin Login</h1>
          <p className="mt-1 text-sm text-slate-500">Use an account with an assigned admin role.</p>
        </div>

        <form className="space-y-4" onSubmit={handleSubmit}>
          <label className="block">
            <span className="text-sm font-medium text-slate-700">Email or phone</span>
            <input
              value={identifier}
              onChange={(event) => setIdentifier(event.target.value)}
              autoComplete="username"
              className="mt-1 h-10 w-full rounded-lg border border-slate-300 px-3 text-sm outline-none focus:border-slate-900 focus:ring-2 focus:ring-slate-900/10"
              placeholder="admin@example.com"
            />
          </label>

          <label className="block">
            <span className="text-sm font-medium text-slate-700">Password</span>
            <input
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              type="password"
              autoComplete="current-password"
              className="mt-1 h-10 w-full rounded-lg border border-slate-300 px-3 text-sm outline-none focus:border-slate-900 focus:ring-2 focus:ring-slate-900/10"
              placeholder="Minimum 8 characters"
            />
          </label>

          {error ? (
            <p className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
              {error}
            </p>
          ) : null}

          <button
            type="submit"
            disabled={isSubmitting}
            className="h-10 w-full rounded-lg bg-slate-950 text-sm font-medium text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-400"
          >
            {isSubmitting ? "Signing in..." : "Continue"}
          </button>
        </form>
      </section>
    </main>
  );
}
