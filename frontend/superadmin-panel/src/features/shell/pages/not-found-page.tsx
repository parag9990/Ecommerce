import { Link } from "react-router-dom";

export function NotFoundPage() {
  return (
    <main className="flex min-h-screen items-center justify-center bg-slate-50 px-4">
      <section className="w-full max-w-md rounded-lg border border-slate-200 bg-white p-5 text-center shadow-sm">
        <p className="text-sm font-semibold text-slate-500">404</p>
        <h1 className="mt-2 text-lg font-semibold text-slate-950">Page not found</h1>
        <Link
          to="/admin"
          className="mt-4 inline-flex h-10 items-center rounded-lg bg-slate-950 px-4 text-sm font-medium text-white hover:bg-slate-800"
        >
          Back to admin
        </Link>
      </section>
    </main>
  );
}
