import { Suspense } from 'react';
import { RouterProvider } from 'react-router-dom';

import { router } from './routes';

export default function App() {
  return (
    <Suspense
      fallback={
        <main
          aria-busy="true"
          aria-live="polite"
          className="mx-auto w-full max-w-7xl px-4 py-6 text-sm text-slate-600 sm:px-6 lg:px-8"
        >
          Loading page...
        </main>
      }
    >
      <RouterProvider router={router} />
    </Suspense>
  );
}
