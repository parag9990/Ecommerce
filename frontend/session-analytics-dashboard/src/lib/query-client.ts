import { QueryClient } from "@tanstack/react-query";

import { ApiError } from "./http";

export function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        gcTime: 5 * 60 * 1000,
        refetchOnWindowFocus: false,
        retry: (failureCount, error) => {
          if (error instanceof ApiError && error.status) {
            return error.status >= 500 && failureCount < 2;
          }

          return failureCount < 2;
        },
        staleTime: 30 * 1000
      }
    }
  });
}

export const queryClient = createQueryClient();
