import type { ProductListRequest } from "../types";

export const productQueryKeys = {
  all: ["seller-products"] as const,
  lists: () => [...productQueryKeys.all, "list"] as const,
  list: (params: ProductListRequest) => [...productQueryKeys.lists(), params] as const,
  details: () => [...productQueryKeys.all, "detail"] as const,
  detail: (productId: string) => [...productQueryKeys.details(), productId] as const,
  categories: () => ["categories"] as const,
};
