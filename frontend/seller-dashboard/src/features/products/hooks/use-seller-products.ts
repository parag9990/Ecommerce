import { useQuery } from "@tanstack/react-query";

import { listSellerProducts } from "../api/seller-product-api";
import type { ProductListRequest } from "../types";
import { productQueryKeys } from "./query-keys";

export function useSellerProducts(params: ProductListRequest) {
  return useQuery({
    queryKey: productQueryKeys.list(params),
    queryFn: () => listSellerProducts(params),
    enabled: Boolean(params.seller_id),
  });
}
