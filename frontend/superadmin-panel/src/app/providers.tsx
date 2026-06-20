import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useEffect } from "react";
import { RouterProvider } from "react-router-dom";

import { useAuthStore } from "../stores/auth-store";
import { router } from "./router";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1
    }
  }
});

function AuthHydrator() {
  const hydrateSession = useAuthStore((state) => state.hydrateSession);

  useEffect(() => {
    hydrateSession();
  }, [hydrateSession]);

  return <RouterProvider router={router} />;
}

export function AppProviders() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthHydrator />
    </QueryClientProvider>
  );
}
