import { type FormEvent, useState } from "react";
import { LogIn } from "lucide-react";
import { Navigate, useLocation, useNavigate } from "react-router-dom";

import { http } from "../lib/http";
import {
  readSellerAuthSession,
  writeSellerAuthSession,
  type SellerAuthSession,
} from "../lib/auth-session";

type LoginResponse = {
  user?: {
    user_id?: string;
    roles?: string[];
    seller_id?: string;
  };
  tokens?: {
    access_token?: string;
    expires_in?: number;
  };
};

function toSellerSession(response: LoginResponse): SellerAuthSession {
  const accessToken = response.tokens?.access_token;
  const userId = response.user?.user_id;
  const roles = response.user?.roles ?? [];

  if (!accessToken || !userId || roles.length === 0) {
    throw new Error("Login did not return a valid seller session.");
  }

  if (!roles.some((role) => role === "seller" || role.startsWith("seller_") || role === "superadmin")) {
    throw new Error("This account does not have Seller Dashboard access.");
  }

  return {
    accessToken,
    expiresAt: response.tokens?.expires_in
      ? new Date(Date.now() + response.tokens.expires_in * 1000).toISOString()
      : undefined,
    roles,
    sellerId: response.user?.seller_id,
    userId,
  };
}

export function LoginRedirectPage() {
  const location = useLocation();
  const navigate = useNavigate();
  const [identifier, setIdentifier] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const fromPath = typeof location.state?.from?.pathname === "string"
    ? location.state.from.pathname
    : "/seller";

  if (readSellerAuthSession()) {
    return <Navigate to={fromPath} replace />;
  }

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);

    if (!identifier.trim() || password.length < 8) {
      setError("Enter your seller email and password.");
      return;
    }

    setIsSubmitting(true);
    try {
      const response = await http<LoginResponse>("/api/v1/auth/login", {
        method: "POST",
        body: JSON.stringify({
          identifier: identifier.trim(),
          password,
          device: {
            channel: "seller-dashboard",
            user_agent: window.navigator.userAgent,
          },
        }),
      });

      writeSellerAuthSession(toSellerSession(response));
      navigate(fromPath, { replace: true });
    } catch (loginError) {
      setError(loginError instanceof Error ? loginError.message : "Seller login failed.");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <main className="flex min-h-screen items-center justify-center bg-slate-100 px-4 py-8">
      <section className="w-full max-w-sm rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
        <div className="mb-5">
          <p className="text-xs font-semibold uppercase tracking-wide text-blue-700">
            Seller Dashboard
          </p>
          <h1 className="mt-1 text-xl font-semibold text-slate-950">Seller login</h1>
          <p className="mt-1 text-sm text-slate-500">
            Use an active seller account to manage catalog, orders, offers, and analytics.
          </p>
        </div>

        <form className="space-y-4" onSubmit={submit}>
          <label className="block">
            <span className="text-sm font-medium text-slate-700">Email or phone</span>
            <input
              autoComplete="username"
              className="mt-1 h-10 w-full rounded-lg border border-slate-300 px-3 text-sm outline-none focus:border-blue-700 focus:ring-2 focus:ring-blue-700/10"
              onChange={(event) => setIdentifier(event.target.value)}
              placeholder="seller@example.com"
              value={identifier}
            />
          </label>

          <label className="block">
            <span className="text-sm font-medium text-slate-700">Password</span>
            <input
              autoComplete="current-password"
              className="mt-1 h-10 w-full rounded-lg border border-slate-300 px-3 text-sm outline-none focus:border-blue-700 focus:ring-2 focus:ring-blue-700/10"
              onChange={(event) => setPassword(event.target.value)}
              placeholder="Minimum 8 characters"
              type="password"
              value={password}
            />
          </label>

          {error ? (
            <p className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
              {error}
            </p>
          ) : null}

          <button
            className="inline-flex h-10 w-full items-center justify-center gap-2 rounded-lg bg-blue-700 px-4 text-sm font-medium text-white transition hover:bg-blue-800 disabled:cursor-not-allowed disabled:bg-slate-400"
            disabled={isSubmitting}
            type="submit"
          >
            <LogIn className="h-4 w-4" aria-hidden="true" />
            <span>{isSubmitting ? "Signing in..." : "Continue"}</span>
          </button>
        </form>
      </section>
    </main>
  );
}
