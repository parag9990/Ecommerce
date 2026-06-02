import { QueryClient } from "@tanstack/react-query";

import { AppError } from "./api-error";

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: (failureCount, error) => {
        if (error instanceof AppError) {
          if ([400, 401, 403, 404, 422].includes(error.status)) {
            return false;
          }

          if (error.status === 429) {
            return failureCount < 1;
          }
        }

        return failureCount < 2;
      },
      staleTime: 30_000,
      gcTime: 5 * 60_000,
    },
    mutations: {
      retry: false,
    },
  },
});
