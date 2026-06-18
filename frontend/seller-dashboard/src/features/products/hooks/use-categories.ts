import { useQuery } from "@tanstack/react-query";

import { listCategories } from "../api/seller-product-api";
import { productQueryKeys } from "./query-keys";

export function useCategories(enabled = true) {
  return useQuery({
    queryKey: productQueryKeys.categories(),
    queryFn: listCategories,
    enabled,
    staleTime: 10 * 60 * 1000,
  });
}
