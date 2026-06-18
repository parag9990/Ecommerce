import { QueryClientProvider } from "@tanstack/react-query";
import { RouterProvider, createBrowserRouter } from "react-router-dom";

import { queryClient } from "./lib/query-client";
import { sellerRoutes } from "./routes/seller-routes";

const router = createBrowserRouter(sellerRoutes);

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  );
}
