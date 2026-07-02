import { type FormEvent, useState } from "react";
import { Navigate, useLocation, useNavigate } from "react-router-dom";

import { sendJSON } from "../../../lib/http";
import {
  readAdminSession,
  writeAdminSession
} from "../../../lib/auth-session";

type LoginResponse = {
  user?: {
    user_id?: string;
    roles?: string[];
  };
  tokens?: {
    access_token?: string;
    expires_in?: number;
  };
};

export function AdminLoginPage() {
  const [identifier, setIdentifier] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();

  if (readAdminSession()) {
    return <Navigate replace to="/" />;
  }

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);
    if (!identifier.trim() || password.length < 8) {
      setError("Enter your admin identifier and password.");
      return;
    }

    setSubmitting(true);
    try {
      const response = await sendJSON<LoginResponse>("/api/v1/auth/login", {
        body: {
          identifier: identifier.trim(),
          password,
          device: {
            channel: "session-analytics-dashboard",
            user_agent: window.navigator.userAgent
          }
        },
        method: "POST"
      });
      const accessToken = response.tokens?.access_token;
      const userId = response.user?.user_id;
      const roles = response.user?.roles ?? [];
      if (!accessToken || !userId || roles.length === 0) {
        throw new Error("Login did not return a valid admin session.");
      }
      writeAdminSession({
        accessToken,
        expiresAt: response.tokens?.expires_in
          ? new Date(Date.now() + response.tokens.expires_in * 1000).toISOString()
          : undefined,
        roles,
        userId
      });
      const destination =
        (location.state as { from?: string } | null)?.from ?? "/";
      navigate(destination, { replace: true });
    } catch (loginError) {
      setError(
        loginError instanceof Error ? loginError.message : "Admin login failed."
      );
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <main className="flex min-h-screen items-center justify-center bg-zinc-50 px-4">
      <section className="w-full max-w-sm rounded-xl border border-zinc-200 bg-white p-6 shadow-sm">
        <p className="text-xs font-semibold uppercase tracking-wide text-zinc-500">
          Session Analytics
        </p>
        <h1 className="mt-1 text-2xl font-semibold text-zinc-950">Admin sign in</h1>
        <p className="mt-2 text-sm text-zinc-600">
          Use an account with session analytics access.
        </p>
        <form className="mt-6 space-y-4" onSubmit={submit}>
          <label className="block text-sm font-medium text-zinc-700">
            Email or phone
            <input
              autoComplete="username"
              className="mt-1 h-10 w-full rounded-md border border-zinc-300 px-3 outline-none focus:border-zinc-900"
              onChange={(event) => setIdentifier(event.target.value)}
              value={identifier}
            />
          </label>
          <label className="block text-sm font-medium text-zinc-700">
            Password
            <input
              autoComplete="current-password"
              className="mt-1 h-10 w-full rounded-md border border-zinc-300 px-3 outline-none focus:border-zinc-900"
              onChange={(event) => setPassword(event.target.value)}
              type="password"
              value={password}
            />
          </label>
          {error ? (
            <p className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
              {error}
            </p>
          ) : null}
          <button
            className="h-10 w-full rounded-md bg-zinc-950 text-sm font-medium text-white disabled:bg-zinc-400"
            disabled={submitting}
            type="submit"
          >
            {submitting ? "Signing in..." : "Continue"}
          </button>
        </form>
      </section>
    </main>
  );
}
